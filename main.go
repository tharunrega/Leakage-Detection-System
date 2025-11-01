package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"stackguard-detector/internal/scanner"
)

var (
	scanMu      sync.Mutex
	scanRunning bool
)

func startScanIfNotRunning() bool {
	scanMu.Lock()
	defer scanMu.Unlock()
	if scanRunning {
		return false
	}
	scanRunning = true
	go func() {
		defer func() {
			scanMu.Lock()
			scanRunning = false
			scanMu.Unlock()
		}()
		// recover to avoid crashing the server if scanner panics
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ERROR] scanner panicked: %v", r)
			}
		}()
		scanner.StartScan()
	}()
	return true
}

func isScanRunning() bool {
	scanMu.Lock()
	defer scanMu.Unlock()
	return scanRunning
}

func main() {
	// Load .env if present (optional for local runs)
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] no .env file loaded (this is optional)")
	}

	// Basic sanity checks and warnings
	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Println("[WARN] GITHUB_TOKEN not set. GitHub API calls may be rate limited or fail.")
	}
	if os.Getenv("SLACK_WEBHOOK_URL") == "" {
		log.Println("[WARN] SLACK_WEBHOOK_URL not set. Slack alerts will fail unless configured.")
	}
	// Check SMTP env vars used by SendEmailAlert
	if os.Getenv("SMTP_EMAIL") == "" || os.Getenv("SMTP_PASS") == "" || os.Getenv("SMTP_HOST") == "" || os.Getenv("SMTP_PORT") == "" || os.Getenv("ALERT_EMAIL") == "" {
		log.Println("[WARN] SMTP env vars not fully set. Email alerts may fail.")
	}

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Start a scan (guarded against concurrent runs)
	r.POST("/scan", func(c *gin.Context) {
		if !startScanIfNotRunning() {
			c.JSON(http.StatusConflict, gin.H{"error": "scan already running"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "Scan started"})
	})

	// Scan status
	r.GET("/scan/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"running": isScanRunning()})
	})

	// Get latest results
	r.GET("/results", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"results": scanner.GetLatestResults()})
	})

	// Send a test Slack alert
	r.POST("/alerts/test", func(c *gin.Context) {
		if err := scanner.SendSlackAlert("✅ Test Alert from StackGuard"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Test Slack alert sent"})
	})

	// Send a test Email alert
	r.POST("/alerts/email/test", func(c *gin.Context) {
		if err := scanner.SendEmailAlert("✅ Test Email from StackGuard", "This is a Mailtrap test email."); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Test email sent via Mailtrap"})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server
	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[INFO] received signal %s, shutting down server", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[FATAL] server forced to shutdown: %v", err)
	}
	log.Println("[INFO] server exited cleanly")
}

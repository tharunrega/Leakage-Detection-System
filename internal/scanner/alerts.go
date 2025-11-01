package scanner

import (
	"bytes"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
)

func SendSlackAlert(msg string) error {
	webhook := os.Getenv("SLACK_WEBHOOK_URL")
	if webhook == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL not set")
	}
	payload := []byte(fmt.Sprintf(`{"text": %q}`, msg))
	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook error: %s", resp.Status)
	}
	return nil
}

func SendEmailAlert(subject, body string) error {
	from := os.Getenv("SMTP_EMAIL")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	to := os.Getenv("ALERT_EMAIL")

	if from == "" || pass == "" || host == "" || port == "" || to == "" {
		return fmt.Errorf("SMTP env vars not fully set (SMTP_EMAIL, SMTP_PASS, SMTP_HOST, SMTP_PORT, ALERT_EMAIL)")
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", from, pass, host)

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

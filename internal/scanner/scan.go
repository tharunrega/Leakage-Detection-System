package scanner

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"stackguard-detector/internal/models"
)

var (
	mu            sync.RWMutex
	latestResults []models.LeakResult
)

// allow tests to override snippet fetcher
var fetchSnippet = FetchSnippet

// EnrichResultsWithSnippets populates LeakResult.Snippet using fetchSnippet.
// It is best-effort and will not fail the scan on error.
func EnrichResultsWithSnippets(results []models.LeakResult) []models.LeakResult {
	for i, r := range results {
		// skip if already set
		if r.Snippet != "" {
			continue
		}

		// Build ownerRepo like "owner/repo"
		ownerRepo := r.Repo
		if ownerRepo == "" && r.Owner != "" && r.Path != "" {
			// if Repo is empty but we have Owner and a Path, we cannot form owner/repo reliably -> skip
			continue
		}
		// if Repo does not contain "/" but Owner is present, try combining them
		if !strings.Contains(ownerRepo, "/") && r.Owner != "" && ownerRepo != "" {
			ownerRepo = r.Owner + "/" + ownerRepo
		}
		// require ownerRepo and path
		if ownerRepo == "" || r.Path == "" {
			continue
		}

		snip, err := fetchSnippet(ownerRepo, r.Path, "")
		if err != nil {
			log.Printf("[WARN] failed to fetch snippet for %s:%s: %v", ownerRepo, r.Path, err)
			continue
		}
		results[i].Snippet = snip
	}
	return results
}

func GetLatestResults() []models.LeakResult {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]models.LeakResult, len(latestResults))
	copy(out, latestResults)
	return out
}

func StartScan() {
	tokens, err := LoadInventory("inventory.json")
	if err != nil {
		log.Printf("[ERROR] loading inventory: %v", err)
		return
	}

	var wg sync.WaitGroup
	var results []models.LeakResult
	var rmu sync.Mutex

	for _, t := range tokens {
		tok := t
		wg.Add(1)
		go func() {
			defer wg.Done()
			leaks, err := SearchGitHubForToken(tok.Value)
			if err != nil {
				log.Printf("[WARN] search error for %s: %v", tok.Type, err)
				return
			}

			// Enrich results with file snippets (best-effort)
			leaks = EnrichResultsWithSnippets(leaks)

			for i := range leaks {
				leaks[i].TokenType = tok.Type
				loc, _ := GetUserLocation(leaks[i].User)
				leaks[i].Location = loc
				leaks[i].Confidence = 1.0

				msg := fmt.Sprintf("🚨 Leak Detected!\nType: %s\nRepo: %s\nUser: %s\nLocation: %s\nURL: %s",
					tok.Type, leaks[i].Repo, leaks[i].User, loc, leaks[i].URL)

				// Best-effort alerts (don't fail the whole scan)
				if err := SendSlackAlert(msg); err != nil {
					log.Printf("[WARN] slack alert failed: %v", err)
				}
				if err := SendEmailAlert("Leak Alert - "+tok.Type, msg); err != nil {
					log.Printf("[WARN] email alert failed: %v", err)
				}
			}

			if len(leaks) > 0 {
				rmu.Lock()
				results = append(results, leaks...)
				rmu.Unlock()
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	latestResults = results
	mu.Unlock()
	log.Printf("[INFO] scan finished. %d result(s).", len(results))
}

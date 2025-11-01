package scanner

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"stackguard-detector/internal/models"
)

type gitHubSearchResult struct {
	Items []struct {
		Name       string `json:"name"`
		Path       string `json:"path"`
		HTMLURL    string `json:"html_url"`
		Repository struct {
			FullName string `json:"full_name"`
			Owner    struct {
				Login string `json:"login"`
			} `json:"owner"`
		} `json:"repository"`
	} `json:"items"`
}

func SearchGitHubForToken(token string) ([]models.LeakResult, error) {
	ghToken := os.Getenv("GITHUB_TOKEN")
	query := url.QueryEscape(fmt.Sprintf("\"%s\"", token)) // quote the token for exact match
	apiURL := fmt.Sprintf("https://api.github.com/search/code?q=%s&per_page=10", query)

	req, _ := http.NewRequest("GET", apiURL, nil)
	if ghToken != "" {
		req.Header.Add("Authorization", "token "+ghToken)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", res.Status, string(body))
	}

	body, _ := io.ReadAll(res.Body)
	var ghResult gitHubSearchResult
	if err := json.Unmarshal(body, &ghResult); err != nil {
		return nil, err
	}

	var leaks []models.LeakResult
	for _, item := range ghResult.Items {
		ownerRepo := item.Repository.FullName // "owner/repo"
		// default snippet is the path (fallback)
		snippet := strings.TrimSpace(item.Path)
		// best-effort: fetch file contents and extract a snippet around the token
		if snip, err := FetchSnippet(ownerRepo, item.Path, token); err == nil && snip != "" {
			snippet = snip
		}
		leaks = append(leaks, models.LeakResult{
			Repo:       ownerRepo,
			URL:        item.HTMLURL,
			User:       item.Repository.Owner.Login,
			Owner:      item.Repository.Owner.Login,
			Path:       item.Path,
			Snippet:    snippet,
			Confidence: 1.0,
		})
	}
	return leaks, nil
}

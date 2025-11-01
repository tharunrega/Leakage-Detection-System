package scanner

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ghContent struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

func FetchSnippet(ownerRepo, path string, needle string) (string, error) {
	// ownerRepo is like "user/repo"
	api := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", ownerRepo, url.PathEscape(path))
	req, _ := http.NewRequest("GET", api, nil)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("contents api: %s - %s", res.Status, string(body))
	}

	var c ghContent
	if err := json.NewDecoder(res.Body).Decode(&c); err != nil {
		return "", err
	}
	if c.Encoding != "base64" {
		return "", fmt.Errorf("unexpected encoding: %s", c.Encoding)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(c.Content, "\n", ""))
	if err != nil {
		return "", err
	}

	// Make a small snippet around the needle (±60 chars)
	text := string(raw)
	idx := strings.Index(text, needle)
	if idx < 0 {
		return "", nil
	}
	start := idx - 60
	if start < 0 {
		start = 0
	}
	end := idx + len(needle) + 60
	if end > len(text) {
		end = len(text)
	}
	snip := text[start:end]
	snip = strings.ReplaceAll(snip, "\n", " ")
	return snip, nil
}

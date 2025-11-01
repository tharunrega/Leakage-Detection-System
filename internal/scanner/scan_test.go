package scanner

import (
	"testing"

	"stackguard-detector/internal/models"
)

func TestEnrichResultsWithSnippets(t *testing.T) {
	orig := fetchSnippet
	defer func() { fetchSnippet = orig }()

	// stub fetchSnippet for deterministic behavior
	fetchSnippet = func(owner, repo, path string) (string, error) {
		return owner + "/" + repo + ":" + path, nil
	}

	input := []models.LeakResult{
		{Owner: "alice", Repo: "app", Path: "secrets.txt"},
		{Owner: "bob", Repo: "lib", Path: "config.yaml", Snippet: "existing-snippet"},
		{Repo: "carol/tool", Path: "main.go"},     // repo combined "owner/repo"
		{User: "dave", Repo: "project", Path: ""}, // missing path -> skipped
	}

	got := EnrichResultsWithSnippets(input)

	if got[0].Snippet != "alice/app:secrets.txt" {
		t.Fatalf("expected snippet for first item, got %q", got[0].Snippet)
	}
	if got[1].Snippet != "existing-snippet" {
		t.Fatalf("existing snippet was overwritten")
	}
	if got[2].Snippet != "carol/tool:main.go" {
		t.Fatalf("expected snippet for third item, got %q", got[2].Snippet)
	}
	if got[3].Snippet != "" {
		t.Fatalf("expected empty snippet for item with missing path, got %q", got[3].Snippet)
	}
}

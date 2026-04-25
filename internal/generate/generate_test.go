package generate_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moneyforward/agito/internal/generate"
	"github.com/moneyforward/agito/internal/ollama"
)

func validContent() generate.Content {
	return generate.Content{
		BranchName:    "feat/add-tests",
		CommitType:    "feat",
		CommitSubject: "add test coverage",
		CommitBody:    "Added comprehensive tests.",
		PRTitle:       "feat: add test coverage",
		PRBody:        "## Context\n\n...\n\n## Changes\n\n...\n\n## How to verify\n\n...",
	}
}

func chatResp(content string) map[string]any {
	return map[string]any{
		"message": map[string]string{"role": "assistant", "content": content},
	}
}

func TestRun_SuccessOnFirstTry(t *testing.T) {
	payload := validContent()
	b, _ := json.Marshal(payload)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResp(string(b)))
	}))
	defer srv.Close()

	cfg := ollama.Config{Host: srv.URL, Model: "test", Temperature: 0.3, NumCtx: 4096}
	result, err := generate.Run(cfg, "test diff", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BranchName != "feat/add-tests" {
		t.Errorf("expected feat/add-tests, got %q", result.BranchName)
	}
}

func TestRun_RetriesOnInvalidJSON(t *testing.T) {
	callCount := 0
	payload := validContent()
	b, _ := json.Marshal(payload)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(chatResp("not valid json"))
			return
		}
		json.NewEncoder(w).Encode(chatResp(string(b)))
	}))
	defer srv.Close()

	cfg := ollama.Config{Host: srv.URL, Model: "test", Temperature: 0.3, NumCtx: 4096}
	result, err := generate.Run(cfg, "diff", "")
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls (retry), got %d", callCount)
	}
}

func TestRun_FailsAfterMaxRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// commit_type "invalid" cannot be sanitized, will always fail validation
		json.NewEncoder(w).Encode(chatResp(`{"branch_name":"feat/test","commit_type":"invalid","commit_subject":"x","pr_title":"x","pr_body":"x","commit_body":"x"}`))
	}))
	defer srv.Close()

	cfg := ollama.Config{Host: srv.URL, Model: "test", Temperature: 0.3, NumCtx: 4096}
	_, err := generate.Run(cfg, "diff", "")
	if err == nil {
		t.Fatal("expected error after max retries")
	}
}

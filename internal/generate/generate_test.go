package generate_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moneyforward/figaro/internal/generate"
	"github.com/moneyforward/figaro/internal/ollama"
)

func validContent() generate.Content {
	return generate.Content{
		BranchName:    "feat/add-tests",
		CommitType:    "feat",
		CommitSubject: "add test coverage",
		CommitBody:    "Added comprehensive tests.",
		PRTitle:       "feat: add test coverage",
		PRBody:        "## 背景\n\n...\n\n## 変更内容\n\n...\n\n## 確認方法\n\n...",
	}
}

func TestRun_SuccessOnFirstTry(t *testing.T) {
	payload := validContent()
	b, _ := json.Marshal(payload)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"response": string(b)})
	}))
	defer srv.Close()

	cfg := ollama.Config{Host: srv.URL, Model: "test", Temperature: 0.3, NumCtx: 4096}
	result, err := generate.Run(cfg, "test diff", "ja", "")
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
			// First attempt: invalid response
			json.NewEncoder(w).Encode(map[string]string{"response": "not valid json"})
			return
		}
		// Second attempt: valid response
		json.NewEncoder(w).Encode(map[string]string{"response": string(b)})
	}))
	defer srv.Close()

	cfg := ollama.Config{Host: srv.URL, Model: "test", Temperature: 0.3, NumCtx: 4096}
	result, err := generate.Run(cfg, "diff", "ja", "")
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
		json.NewEncoder(w).Encode(map[string]string{"response": `{"branch_name":"INVALID"}`})
	}))
	defer srv.Close()

	cfg := ollama.Config{Host: srv.URL, Model: "test", Temperature: 0.3, NumCtx: 4096}
	_, err := generate.Run(cfg, "diff", "ja", "")
	if err == nil {
		t.Fatal("expected error after max retries")
	}
}

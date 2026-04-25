package ollama_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tatsuo48/agito/internal/ollama"
)

func chatResp(content string) map[string]any {
	return map[string]any{
		"message": map[string]string{"role": "assistant", "content": content},
	}
}

func TestGenerate_Success(t *testing.T) {
	want := `{"branch_name":"feat/test","commit_type":"feat","commit_scope":"api","commit_subject":"add test","commit_body":"body","pr_title":"feat: add test","pr_body":"## Context\n\ndetails"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResp(want))
	}))
	defer srv.Close()

	cfg := ollama.Config{Host: srv.URL, Model: "test-model", Temperature: 0.3, NumCtx: 4096}
	out, err := ollama.Generate(cfg, "test diff", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(out, "feat/test") {
		t.Errorf("expected branch name in output, got %q", out)
	}
}

func TestGenerate_OllamaError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"error": "model not found"})
	}))
	defer srv.Close()

	cfg := ollama.Config{Host: srv.URL, Model: "bad-model", Temperature: 0.3, NumCtx: 4096}
	_, err := ollama.Generate(cfg, "diff", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildPrompt_ContainsEnglishSections(t *testing.T) {
	prompt := ollama.BuildPrompt("some diff", "")
	if !strings.Contains(prompt, "## Context") {
		t.Error("expected English PR sections")
	}
	if !strings.Contains(prompt, "## Changes") {
		t.Error("expected Changes section")
	}
}

func TestBuildPrompt_ExtraInstruction(t *testing.T) {
	prompt := ollama.BuildPrompt("diff", "keep it brief")
	if !strings.Contains(prompt, "keep it brief") {
		t.Error("expected extra instruction in prompt")
	}
}

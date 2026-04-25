package gh_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/moneyforward/agito/internal/gh"
	"github.com/moneyforward/agito/internal/git"
)

func TestCreatePR_BuildsArgs(t *testing.T) {
	fake := &git.FakeRunner{Output: "https://github.com/org/repo/pull/1"}
	cfg := gh.PRConfig{
		Title: "feat: add test",
		Body:  "## 背景\n\ntest",
		Base:  "main",
		Draft: false,
	}
	url, err := gh.CreatePR(fake, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://github.com/org/repo/pull/1" {
		t.Errorf("expected PR URL, got %q", url)
	}
	// Verify gh pr create was called with correct args
	if len(fake.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fake.Calls))
	}
	call := fake.Calls[0]
	if call[0] != "gh" || call[1] != "pr" || call[2] != "create" {
		t.Errorf("expected 'gh pr create', got %v", call)
	}
	// Check --title is present
	found := false
	for i, arg := range call {
		if arg == "--title" && i+1 < len(call) && call[i+1] == "feat: add test" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected --title 'feat: add test' in args: %v", call)
	}
}

func TestCreatePR_WithDraft(t *testing.T) {
	fake := &git.FakeRunner{Output: "https://github.com/org/repo/pull/2"}
	cfg := gh.PRConfig{
		Title: "draft PR",
		Body:  "body",
		Base:  "main",
		Draft: true,
	}
	url, err := gh.CreatePR(fake, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Fatal("expected PR URL")
	}
	// Verify --draft flag was passed
	call := fake.Calls[0]
	hasDraft := false
	for _, arg := range call {
		if arg == "--draft" {
			hasDraft = true
		}
	}
	if !hasDraft {
		t.Errorf("expected --draft in args: %v", call)
	}
}

func TestCreatePR_Failure(t *testing.T) {
	fake := &git.FuncRunner{
		RunFn: func(name string, args ...string) (string, error) {
			return "", fmt.Errorf("gh: not authenticated")
		},
	}
	cfg := gh.PRConfig{Title: "test", Body: "body", Base: "main"}
	_, err := gh.CreatePR(fake, cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "gh operation failed") {
		t.Errorf("expected gh error, got %q", err.Error())
	}
}

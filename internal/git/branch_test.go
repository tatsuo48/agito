package git_test

import (
	"fmt"
	"testing"

	"github.com/moneyforward/figaro/internal/git"
)

func TestDetectDefaultBranch_SymbolicRef(t *testing.T) {
	fake := &git.FakeRunner{Output: "origin/main"}
	branch, err := git.DetectDefaultBranch(fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Fatalf("expected 'main', got %q", branch)
	}
}

func TestDetectDefaultBranch_Fallback(t *testing.T) {
	callCount := 0
	fake := &git.FuncRunner{
		RunFn: func(name string, args ...string) (string, error) {
			callCount++
			// First call (symbolic-ref) fails
			if callCount == 1 {
				return "", fmt.Errorf("symbolic-ref failed")
			}
			// Second call (remote show) returns HEAD branch
			if callCount == 2 {
				return "  HEAD branch: develop\n", nil
			}
			return "", nil
		},
	}
	branch, err := git.DetectDefaultBranch(fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "develop" {
		t.Fatalf("expected 'develop', got %q", branch)
	}
}

func TestCurrentBranch(t *testing.T) {
	fake := &git.FakeRunner{Output: "feature/foo"}
	branch, err := git.CurrentBranch(fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "feature/foo" {
		t.Fatalf("expected 'feature/foo', got %q", branch)
	}
}

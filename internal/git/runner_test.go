package git_test

import (
	"testing"

	"github.com/tatsuo48/agito/internal/git"
)

func TestFakeRunner(t *testing.T) {
	fake := &git.FakeRunner{
		Output: "hello",
		Err:    nil,
	}
	out, err := fake.Run("git", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello" {
		t.Fatalf("expected 'hello', got %q", out)
	}
	if len(fake.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fake.Calls))
	}
}

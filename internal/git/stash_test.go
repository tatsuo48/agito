package git_test

import (
	"strings"
	"testing"
	"time"

	"github.com/moneyforward/figaro/internal/git"
)

func TestStashMessage(t *testing.T) {
	ts := time.Date(2026, 4, 24, 14, 30, 22, 0, time.UTC)
	msg := git.StashMessage(ts)
	if !strings.HasPrefix(msg, "figaro-auto-stash-") {
		t.Errorf("unexpected stash message: %q", msg)
	}
	if msg != "figaro-auto-stash-20260424-143022" {
		t.Errorf("unexpected stash message format: %q", msg)
	}
}

func TestHasFigaroStash_Found(t *testing.T) {
	fake := &git.FakeRunner{Output: "stash@{0}: figaro-auto-stash-20260424-143022"}
	found, err := git.HasFigaroStash(fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected figaro stash to be found")
	}
}

func TestHasFigaroStash_NotFound(t *testing.T) {
	fake := &git.FakeRunner{Output: "stash@{0}: WIP on main: abc123"}
	found, err := git.HasFigaroStash(fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected no figaro stash")
	}
}

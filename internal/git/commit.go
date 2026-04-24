package git

import (
	"fmt"
	"strings"
)

// Checkout switches to the specified branch.
func Checkout(r Runner, branch string) error {
	_, err := r.Run("git", "checkout", branch)
	if err != nil {
		return fmt.Errorf("checkout %s: %w", branch, err)
	}
	return nil
}

// Pull fetches and fast-forward merges from origin.
// Uses fetch + merge --ff-only instead of pull to avoid FETCH_HEAD ambiguity
// when multiple branches have been fetched previously.
func Pull(r Runner, branch string) error {
	if _, err := r.Run("git", "fetch", "origin", branch); err != nil {
		return err
	}
	_, err := r.Run("git", "merge", "--ff-only", "origin/"+branch)
	return err
}

// CreateBranch creates a new branch with the given name, appending -2, -3 etc. on collision.
func CreateBranch(r Runner, name string) (string, error) {
	candidate := name
	for i := 2; i <= 10; i++ {
		_, err := r.Run("git", "checkout", "-b", candidate)
		if err == nil {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", name, i)
	}
	return "", fmt.Errorf("could not create branch %s (all variants taken)", name)
}

// AddAll stages all changes.
func AddAll(r Runner) error {
	_, err := r.Run("git", "add", "-A")
	return err
}

// CommitWithMessage creates a commit with the given message.
func CommitWithMessage(r Runner, message string) error {
	_, err := r.Run("git", "commit", "-m", message)
	return err
}

// Push pushes the current branch to origin.
func Push(r Runner) error {
	_, err := r.Run("git", "push", "-u", "origin", "HEAD")
	return err
}

// HasChanges reports whether there are any uncommitted changes.
func HasChanges(r Runner) (bool, error) {
	out, err := r.Run("git", "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

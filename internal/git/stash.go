package git

import (
	"fmt"
	"strings"
	"time"
)

// StashMessage generates the figaro stash message with a timestamp.
func StashMessage(t time.Time) string {
	return fmt.Sprintf("figaro-auto-stash-%s", t.Format("20060102-150405"))
}

// StashPush stashes all changes including untracked files.
// Returns the stash ref (e.g. "stash@{0}").
func StashPush(r Runner, msg string) (string, error) {
	_, err := r.Run("git", "stash", "push", "--include-untracked", "-m", msg)
	if err != nil {
		return "", fmt.Errorf("stash push: %w", err)
	}
	ref, err := r.Run("git", "stash", "list", "--format=%gd", "-1")
	if err != nil {
		return "stash@{0}", nil
	}
	return strings.TrimSpace(ref), nil
}

// StashPop applies the most recent stash.
func StashPop(r Runner) error {
	_, err := r.Run("git", "stash", "pop")
	return err
}

// StashDrop removes the specified stash.
func StashDrop(r Runner, ref string) error {
	_, err := r.Run("git", "stash", "drop", ref)
	return err
}

// HasFigaroStash checks whether any figaro-auto-stash entries remain.
func HasFigaroStash(r Runner) (bool, error) {
	out, err := r.Run("git", "stash", "list")
	if err != nil {
		return false, err
	}
	return strings.Contains(out, "figaro-auto-stash-"), nil
}

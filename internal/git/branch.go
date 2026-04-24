package git

import (
	"fmt"
	"strings"
)

// DetectDefaultBranch returns the default branch name.
// Falls back in order: symbolic-ref → remote show → main/master/develop.
func DetectDefaultBranch(r Runner) (string, error) {
	out, err := r.Run("git", "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	if err == nil {
		// "origin/main" → "main"
		parts := strings.SplitN(out, "/", 2)
		if len(parts) == 2 {
			return parts[1], nil
		}
	}

	// Fallback: git remote show origin
	out, err = r.Run("git", "remote", "show", "origin")
	if err == nil {
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "HEAD branch:") {
				branch := strings.TrimSpace(strings.TrimPrefix(line, "HEAD branch:"))
				if branch != "" && branch != "(unknown)" {
					return branch, nil
				}
			}
		}
	}

	// Final fallback: check known branch names
	for _, candidate := range []string{"main", "master", "develop"} {
		if _, err := r.Run("git", "rev-parse", "--verify", candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not detect default branch")
}

// CurrentBranch returns the name of the current branch.
func CurrentBranch(r Runner) (string, error) {
	return r.Run("git", "rev-parse", "--abbrev-ref", "HEAD")
}

package gh

import (
	"fmt"
	"os"
	"strings"

	"github.com/tatsuo48/agito/internal/apperr"
	"github.com/tatsuo48/agito/internal/git"
)

// PRConfig holds PR creation parameters.
type PRConfig struct {
	Title string
	Body  string
	Base  string
	Draft bool
}

// CreatePR creates a GitHub PR using gh CLI and returns the PR URL.
func CreatePR(r git.Runner, cfg PRConfig) (string, error) {
	tmp, err := os.CreateTemp("", "agito-pr-body-*.md")
	if err != nil {
		return "", fmt.Errorf("%w: creating temp file: %v", apperr.ErrGH, err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(cfg.Body); err != nil {
		tmp.Close()
		return "", fmt.Errorf("%w: writing PR body: %v", apperr.ErrGH, err)
	}
	tmp.Close()

	args := []string{"pr", "create",
		"--title", cfg.Title,
		"--body-file", tmp.Name(),
		"--base", cfg.Base,
	}
	if cfg.Draft {
		args = append(args, "--draft")
	}

	url, err := r.Run("gh", args...)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apperr.ErrGH, err)
	}
	return strings.TrimSpace(url), nil
}

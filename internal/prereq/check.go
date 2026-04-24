package prereq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/moneyforward/figaro/internal/apperr"
	"github.com/moneyforward/figaro/internal/git"
)

// Config holds the configuration needed for prerequisite checks.
type Config struct {
	OllamaHost string
	Model      string
}

// Check verifies all prerequisites are met.
func Check(r git.Runner, cfg Config) error {
	checks := []struct {
		name string
		fn   func() error
	}{
		{"git repository", func() error { return checkGitRepo(r) }},
		{"gh auth", func() error { return checkGHAuth(r) }},
		{"git remote origin", func() error { return checkRemote(r) }},
		{"changes exist", func() error { return checkChanges(r) }},
		{"ollama running", func() error { return checkOllama(cfg.OllamaHost) }},
		{"ollama model", func() error { return checkModel(cfg.OllamaHost, cfg.Model) }},
	}

	for _, c := range checks {
		if err := c.fn(); err != nil {
			return fmt.Errorf("%w: %s: %v", apperr.ErrPrereq, c.name, err)
		}
	}
	return nil
}

func checkGitRepo(r git.Runner) error {
	_, err := r.Run("git", "rev-parse", "--is-inside-work-tree")
	return err
}

func checkGHAuth(r git.Runner) error {
	_, err := r.Run("gh", "auth", "status")
	return err
}

func checkRemote(r git.Runner) error {
	_, err := r.Run("git", "remote", "get-url", "origin")
	return err
}

func checkChanges(r git.Runner) error {
	out, err := r.Run("git", "status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) == "" {
		return fmt.Errorf("no changes detected")
	}
	return nil
}

func checkOllama(host string) error {
	resp, err := http.Get(host + "/api/tags") //nolint:noctx
	if err != nil {
		return fmt.Errorf("ollama not running at %s: %v", host, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}
	return nil
}

func checkModel(host, model string) error {
	resp, err := http.Get(host + "/api/tags") //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	for _, m := range result.Models {
		if m.Name == model || strings.HasPrefix(m.Name, model+":") || strings.HasPrefix(model, m.Name) {
			return nil
		}
	}
	return fmt.Errorf("model %q not found (run: ollama pull %s)", model, model)
}

package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetDiff returns combined diff of tracked changes and untracked files.
func GetDiff(r Runner) (string, error) {
	tracked, err := r.Run("git", "diff", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git diff HEAD: %w", err)
	}

	// Also include staged-only changes not yet in HEAD
	staged, err := r.Run("git", "diff", "--cached")
	if err != nil {
		staged = ""
	}

	untracked, err := r.Run("git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return "", fmt.Errorf("ls-files: %w", err)
	}

	var parts []string
	if tracked != "" {
		parts = append(parts, tracked)
	}
	// Add staged diff only if not already covered by HEAD diff
	if staged != "" && tracked == "" {
		parts = append(parts, staged)
	}

	for _, f := range strings.Split(untracked, "\n") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		content, err := os.ReadFile(filepath.Clean(f))
		if err != nil {
			continue
		}
		pseudo := buildPseudoDiff(f, string(content))
		parts = append(parts, pseudo)
	}

	combined := strings.Join(parts, "\n")
	return TruncateDiff(combined), nil
}

func buildPseudoDiff(filename, content string) string {
	lines := strings.Split(content, "\n")
	var sb strings.Builder
	fmt.Fprintf(&sb, "diff --git a/%s b/%s\nnew file mode 100644\n+++ b/%s\n@@ -0,0 +1,%d @@\n",
		filename, filename, filename, len(lines))
	for _, line := range lines {
		sb.WriteString("+" + line + "\n")
	}
	return sb.String()
}

// TruncateDiff reduces diff size if it exceeds ~30,000 characters.
func TruncateDiff(diff string) string {
	const maxChars = 30000
	if len(diff) <= maxChars {
		return diff
	}

	filtered := filterNoisyFiles(diff)
	if len(filtered) <= maxChars {
		return filtered
	}

	return filtered[:maxChars] + "\n... (diff truncated)"
}

var noisyPatterns = []string{
	"package-lock.json", "yarn.lock", "go.sum", "Gemfile.lock",
	"node_modules/", "vendor/",
	".pb.go", ".generated.ts", ".pb.swift", ".min.js", ".min.css",
}

func filterNoisyFiles(diff string) string {
	var kept []string
	for _, block := range splitDiffByFile(diff) {
		noisy := false
		for _, pat := range noisyPatterns {
			if strings.Contains(block, pat) {
				noisy = true
				break
			}
		}
		if !noisy {
			kept = append(kept, block)
		}
	}
	return strings.Join(kept, "\n")
}

func splitDiffByFile(diff string) []string {
	var blocks []string
	var current strings.Builder
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git") && current.Len() > 0 {
			blocks = append(blocks, current.String())
			current.Reset()
		}
		current.WriteString(line + "\n")
	}
	if current.Len() > 0 {
		blocks = append(blocks, current.String())
	}
	return blocks
}

// MultiRunner is a test stub that returns different responses per command.
type MultiRunner struct {
	Responses map[string]string
	Err       error
}

func (m *MultiRunner) Run(name string, args ...string) (string, error) {
	key := name + " " + strings.Join(args, " ")
	if out, ok := m.Responses[key]; ok {
		return out, m.Err
	}
	return "", m.Err
}

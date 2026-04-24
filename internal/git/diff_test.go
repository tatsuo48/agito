package git_test

import (
	"strings"
	"testing"

	"github.com/moneyforward/figaro/internal/git"
)

func TestGetDiff_TrackedOnly(t *testing.T) {
	fake := &git.MultiRunner{
		Responses: map[string]string{
			"git diff HEAD":                           "diff --git a/foo.go b/foo.go\n+added line",
			"git diff --cached":                       "",
			"git ls-files --others --exclude-standard": "",
		},
	}
	diff, err := git.GetDiff(fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(diff, "foo.go") {
		t.Error("expected tracked diff in output")
	}
}

func TestGetDiff_UntrackedFiles(t *testing.T) {
	// Create a temp file for untracked test
	fake := &git.MultiRunner{
		Responses: map[string]string{
			"git diff HEAD":                           "",
			"git diff --cached":                       "",
			"git ls-files --others --exclude-standard": "",
		},
	}
	diff, err := git.GetDiff(fake)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = diff // empty is fine when no changes
}

func TestTruncateDiff_SmallInput(t *testing.T) {
	small := "diff --git a/foo.go b/foo.go\n+added"
	result := git.TruncateDiff(small)
	if result != small {
		t.Errorf("small diff should not be truncated, got %q", result)
	}
}

func TestTruncateDiff_LargeInput(t *testing.T) {
	large := strings.Repeat("a", 35000)
	result := git.TruncateDiff(large)
	if len(result) > 30200 {
		t.Errorf("expected truncated diff ≤30200 chars, got %d", len(result))
	}
}

func TestTruncateDiff_FiltersNoisyFiles(t *testing.T) {
	noisyDiff := "diff --git a/package-lock.json b/package-lock.json\n" + strings.Repeat("+lock content\n", 2500)
	cleanDiff := "diff --git a/main.go b/main.go\n+func main() {}"
	combined := noisyDiff + cleanDiff

	result := git.TruncateDiff(combined)
	if strings.Contains(result, "package-lock.json") {
		t.Error("expected noisy file to be filtered")
	}
	if !strings.Contains(result, "main.go") {
		t.Error("expected clean file to remain")
	}
}

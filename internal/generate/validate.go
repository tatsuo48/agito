package generate

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var branchNameRe = regexp.MustCompile(`^(feat|fix|chore|docs|refactor|test|perf)/[a-z0-9-]+$`)

var validPrefixes = []string{"feat", "fix", "chore", "docs", "refactor", "test", "perf"}

var validCommitTypes = map[string]bool{
	"feat": true, "fix": true, "chore": true, "docs": true,
	"refactor": true, "test": true, "perf": true,
}

// ParseAndValidate parses raw JSON bytes, sanitizes, and validates the content.
func ParseAndValidate(raw []byte) (*Content, error) {
	var c Content
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	sanitize(&c)
	return &c, validate(&c)
}

// sanitize normalizes fields that models commonly get slightly wrong.
func sanitize(c *Content) {
	c.BranchName = sanitizeBranchName(c.BranchName)

	// Normalize commit_type to lowercase.
	c.CommitType = strings.ToLower(strings.TrimSpace(c.CommitType))

	// Truncate commit_subject if over 72 runes.
	if utf8.RuneCountInString(c.CommitSubject) > 72 {
		runes := []rune(c.CommitSubject)
		c.CommitSubject = string(runes[:72])
	}
}

var (
	nonAlphaNumHyphen = regexp.MustCompile(`[^a-z0-9-]+`)
	multipleDashes    = regexp.MustCompile(`-{2,}`)
)

func sanitizeBranchName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	// Split on first "/" to separate prefix from slug.
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		// No slash — treat the whole thing as slug under "chore/".
		slug := toSlug(name)
		if slug == "" {
			return "chore/changes"
		}
		return "chore/" + slug
	}

	prefix := strings.ToLower(strings.TrimSpace(parts[0]))
	slug := toSlug(parts[1])

	// If prefix is not a valid conventional prefix, fall back to chore.
	validPrefix := false
	for _, p := range validPrefixes {
		if prefix == p {
			validPrefix = true
			break
		}
	}
	if !validPrefix {
		prefix = "chore"
	}
	if slug == "" {
		slug = "changes"
	}

	result := prefix + "/" + slug
	// Enforce 50-char limit by trimming the slug.
	if utf8.RuneCountInString(result) > 50 {
		max := 50 - len(prefix) - 1 // len("/")
		slugRunes := []rune(slug)
		if max > 0 && len(slugRunes) > max {
			slug = strings.TrimRight(string(slugRunes[:max]), "-")
		}
		result = prefix + "/" + slug
	}
	return result
}

// toSlug converts arbitrary text to kebab-case [a-z0-9-]+.
func toSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// Replace underscores and spaces with hyphens.
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	// Remove anything that's not a-z, 0-9, or hyphen.
	s = nonAlphaNumHyphen.ReplaceAllString(s, "-")
	// Collapse multiple hyphens.
	s = multipleDashes.ReplaceAllString(s, "-")
	// Strip leading/trailing hyphens.
	s = strings.Trim(s, "-")
	return s
}

func validate(c *Content) error {
	if c.BranchName == "" {
		return fmt.Errorf("branch_name is required")
	}
	if !branchNameRe.MatchString(c.BranchName) {
		return fmt.Errorf("branch_name %q could not be sanitized to a valid pattern", c.BranchName)
	}
	if utf8.RuneCountInString(c.BranchName) > 50 {
		return fmt.Errorf("branch_name exceeds 50 characters")
	}
	if !validCommitTypes[c.CommitType] {
		return fmt.Errorf("commit_type %q is not valid", c.CommitType)
	}
	if c.CommitSubject == "" {
		return fmt.Errorf("commit_subject is required")
	}
	if utf8.RuneCountInString(c.CommitSubject) > 72 {
		return fmt.Errorf("commit_subject exceeds 72 characters")
	}
	if c.PRTitle == "" {
		return fmt.Errorf("pr_title is required")
	}
	if c.PRBody == "" {
		return fmt.Errorf("pr_body is required")
	}
	return nil
}

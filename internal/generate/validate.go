package generate

import (
	"encoding/json"
	"fmt"
	"regexp"
	"unicode/utf8"
)

var branchNameRe = regexp.MustCompile(`^(feat|fix|chore|docs|refactor|test|perf)/[a-z0-9-]+$`)

var validCommitTypes = map[string]bool{
	"feat": true, "fix": true, "chore": true, "docs": true,
	"refactor": true, "test": true, "perf": true,
}

// ParseAndValidate parses raw JSON bytes and validates the content.
func ParseAndValidate(raw []byte) (*Content, error) {
	var c Content
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return &c, validate(&c)
}

func validate(c *Content) error {
	if c.BranchName == "" {
		return fmt.Errorf("branch_name is required")
	}
	if !branchNameRe.MatchString(c.BranchName) {
		return fmt.Errorf("branch_name %q does not match pattern ^(feat|fix|chore|docs|refactor|test|perf)/[a-z0-9-]+$", c.BranchName)
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

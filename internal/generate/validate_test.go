package generate_test

import (
	"strings"
	"testing"

	"github.com/tatsuo48/agito/internal/generate"
)

func TestValidate_Valid(t *testing.T) {
	raw := `{
		"branch_name": "feat/add-login",
		"commit_type": "feat",
		"commit_scope": "auth",
		"commit_subject": "add login endpoint",
		"commit_body": "Added POST /login endpoint.\n\nReturns JWT token.",
		"pr_title": "feat(auth): add login endpoint",
		"pr_body": "## 背景\n\nlogin needed\n\n## 変更内容\n\nadded\n\n## 確認方法\n\ncurl"
	}`
	result, err := generate.ParseAndValidate([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BranchName != "feat/add-login" {
		t.Errorf("expected branch_name 'feat/add-login', got %q", result.BranchName)
	}
}

func TestSanitize_BranchNameUppercaseUnderscore(t *testing.T) {
	// docs/README_copy → docs/readme-copy
	raw := `{
		"branch_name": "docs/README_copy",
		"commit_type": "docs",
		"commit_scope": "",
		"commit_subject": "update readme",
		"commit_body": "Updated README.",
		"pr_title": "docs: update readme",
		"pr_body": "## Context\n\n...\n\n## Changes\n\n...\n\n## How to verify\n\n..."
	}`
	result, err := generate.ParseAndValidate([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error after sanitize: %v", err)
	}
	if result.BranchName != "docs/readme-copy" {
		t.Errorf("expected 'docs/readme-copy', got %q", result.BranchName)
	}
}

func TestSanitize_BranchNameNoPrefix(t *testing.T) {
	// readme-update → chore/readme-update
	raw := `{
		"branch_name": "readme-update",
		"commit_type": "chore",
		"commit_scope": "",
		"commit_subject": "update readme",
		"commit_body": "Updated README.",
		"pr_title": "chore: update readme",
		"pr_body": "## Context\n\n...\n\n## Changes\n\n...\n\n## How to verify\n\n..."
	}`
	result, err := generate.ParseAndValidate([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BranchName != "chore/readme-update" {
		t.Errorf("expected 'chore/readme-update', got %q", result.BranchName)
	}
}

func TestSanitize_BranchNameNoSlashConvertsToChore(t *testing.T) {
	// "INVALID_BRANCH" has no prefix slash, gets sanitized to chore/invalid-branch
	raw := `{
		"branch_name": "INVALID_BRANCH",
		"commit_type": "feat",
		"commit_scope": "",
		"commit_subject": "test",
		"commit_body": "body",
		"pr_title": "title",
		"pr_body": "body"
	}`
	result, err := generate.ParseAndValidate([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BranchName != "chore/invalid-branch" {
		t.Errorf("expected 'chore/invalid-branch', got %q", result.BranchName)
	}
}

func TestSanitize_SubjectTooLongIsTruncated(t *testing.T) {
	// 73-rune subject gets truncated to 72 runes by sanitize
	raw := `{
		"branch_name": "feat/something",
		"commit_type": "feat",
		"commit_scope": "",
		"commit_subject": "` + strings.Repeat("あ", 73) + `",
		"commit_body": "body",
		"pr_title": "title",
		"pr_body": "body"
	}`
	result, err := generate.ParseAndValidate([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error after sanitize: %v", err)
	}
	if len([]rune(result.CommitSubject)) != 72 {
		t.Errorf("expected 72 runes, got %d", len([]rune(result.CommitSubject)))
	}
}

func TestValidate_InvalidCommitType(t *testing.T) {
	raw := `{
		"branch_name": "feat/something",
		"commit_type": "invalid",
		"commit_scope": "",
		"commit_subject": "test subject",
		"commit_body": "body",
		"pr_title": "title",
		"pr_body": "body"
	}`
	_, err := generate.ParseAndValidate([]byte(raw))
	if err == nil {
		t.Fatal("expected validation error for invalid commit_type")
	}
}

func TestBuildCommitMessage_WithScope(t *testing.T) {
	c := &generate.Content{
		CommitType:    "fix",
		CommitScope:   "cache",
		CommitSubject: "fix TTL policy",
		CommitBody:    "Changed allkeys-lru.",
	}
	msg := c.CommitMessage()
	expected := "fix(cache): fix TTL policy\n\nChanged allkeys-lru."
	if msg != expected {
		t.Errorf("expected %q\ngot %q", expected, msg)
	}
}

func TestBuildCommitMessage_NoScope(t *testing.T) {
	c := &generate.Content{
		CommitType:    "chore",
		CommitSubject: "update deps",
		CommitBody:    "Updated go.sum.",
	}
	msg := c.CommitMessage()
	expected := "chore: update deps\n\nUpdated go.sum."
	if msg != expected {
		t.Errorf("expected %q\ngot %q", expected, msg)
	}
}

func TestBuildCommitMessage_NoBody(t *testing.T) {
	c := &generate.Content{
		CommitType:    "docs",
		CommitSubject: "update readme",
	}
	msg := c.CommitMessage()
	if msg != "docs: update readme" {
		t.Errorf("expected 'docs: update readme', got %q", msg)
	}
}

package generate_test

import (
	"strings"
	"testing"

	"github.com/moneyforward/figaro/internal/generate"
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

func TestValidate_InvalidBranchName(t *testing.T) {
	raw := `{
		"branch_name": "INVALID_BRANCH",
		"commit_type": "feat",
		"commit_scope": "",
		"commit_subject": "test",
		"commit_body": "body",
		"pr_title": "title",
		"pr_body": "body"
	}`
	_, err := generate.ParseAndValidate([]byte(raw))
	if err == nil {
		t.Fatal("expected validation error for invalid branch_name")
	}
}

func TestValidate_SubjectTooLong(t *testing.T) {
	raw := `{
		"branch_name": "feat/something",
		"commit_type": "feat",
		"commit_scope": "",
		"commit_subject": "` + strings.Repeat("あ", 73) + `",
		"commit_body": "body",
		"pr_title": "title",
		"pr_body": "body"
	}`
	_, err := generate.ParseAndValidate([]byte(raw))
	if err == nil {
		t.Fatal("expected validation error for long subject")
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

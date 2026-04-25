package ui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/moneyforward/agito/internal/generate"
	"gopkg.in/yaml.v3"
)

// EditContent opens $EDITOR to edit the specified part of the content.
func EditContent(c *generate.Content, target EditTarget) (*generate.Content, error) {
	switch target {
	case EditBranch:
		tmp, err := writeTempFile("branch_name: "+c.BranchName+"\n", "*.txt")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmp)
		if err := openEditor(tmp); err != nil {
			return nil, err
		}
		b, err := os.ReadFile(tmp)
		if err != nil {
			return nil, err
		}
		// Simple parse: take the value after "branch_name: "
		val := string(b)
		if len(val) > len("branch_name: ") {
			c.BranchName = trimNewlines(val[len("branch_name: "):])
		}
		return c, nil

	case EditCommit:
		tmp, err := writeTempFile(c.CommitMessage(), "*.txt")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmp)
		if err := openEditor(tmp); err != nil {
			return nil, err
		}
		b, err := os.ReadFile(tmp)
		if err != nil {
			return nil, err
		}
		c.CommitSubject = trimNewlines(string(b))
		return c, nil

	case EditPR:
		tmp, err := writeTempFile(c.PRBody, "*.md")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmp)
		if err := openEditor(tmp); err != nil {
			return nil, err
		}
		b, err := os.ReadFile(tmp)
		if err != nil {
			return nil, err
		}
		c.PRBody = string(b)
		return c, nil

	default:
		return editAll(c)
	}
}

type contentYAML struct {
	BranchName    string `yaml:"branch_name"`
	CommitType    string `yaml:"commit_type"`
	CommitScope   string `yaml:"commit_scope"`
	CommitSubject string `yaml:"commit_subject"`
	CommitBody    string `yaml:"commit_body"`
	PRTitle       string `yaml:"pr_title"`
	PRBody        string `yaml:"pr_body"`
}

func editAll(c *generate.Content) (*generate.Content, error) {
	cy := contentYAML{
		BranchName:    c.BranchName,
		CommitType:    c.CommitType,
		CommitScope:   c.CommitScope,
		CommitSubject: c.CommitSubject,
		CommitBody:    c.CommitBody,
		PRTitle:       c.PRTitle,
		PRBody:        c.PRBody,
	}
	b, err := yaml.Marshal(cy)
	if err != nil {
		return nil, err
	}

	tmp, err := writeTempFile(string(b), "*.yaml")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp)

	if err := openEditor(tmp); err != nil {
		return nil, err
	}

	edited, err := os.ReadFile(tmp)
	if err != nil {
		return nil, err
	}

	var result contentYAML
	if err := yaml.Unmarshal(edited, &result); err != nil {
		return nil, fmt.Errorf("YAML parse error: %w", err)
	}

	c.BranchName = result.BranchName
	c.CommitType = result.CommitType
	c.CommitScope = result.CommitScope
	c.CommitSubject = result.CommitSubject
	c.CommitBody = result.CommitBody
	c.PRTitle = result.PRTitle
	c.PRBody = result.PRBody
	return c, nil
}

func writeTempFile(content, pattern string) (string, error) {
	tmp, err := os.CreateTemp("", "agito-edit-"+pattern)
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := tmp.WriteString(content); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func trimNewlines(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

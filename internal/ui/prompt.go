package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/moneyforward/figaro/internal/generate"
)

var (
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// Action represents the user's chosen action.
type Action int

const (
	ActionYes        Action = iota
	ActionNo
	ActionEdit
	ActionRegenerate
	ActionPreview
)

// ShowSummary prints the generated content summary to w.
func ShowSummary(w io.Writer, c *generate.Content) {
	divider := dividerStyle.Render(repeatStr("─", 50))
	fmt.Fprintln(w, divider)
	fmt.Fprintf(w, "  %s  %s\n", labelStyle.Render("Branch: "), c.BranchName)
	fmt.Fprintf(w, "  %s  %s\n", labelStyle.Render("Commit: "), c.CommitMessage())
	fmt.Fprintf(w, "  %s  %s\n", labelStyle.Render("PR Title:"), c.PRTitle)
	fmt.Fprintf(w, "  %s  (%d lines) — preview with 'p'\n", labelStyle.Render("PR Body: "), lineCount(c.PRBody))
	fmt.Fprintln(w, divider)
}

// AskAction prompts the user to choose an action.
// Returns the action, optional extra instruction (for regenerate), and any error.
func AskAction() (Action, string, error) {
	var choice string
	err := huh.NewSelect[string]().
		Title("Proceed?").
		Options(
			huh.NewOption("[y] Yes — proceed", "y"),
			huh.NewOption("[e] Edit generated content", "e"),
			huh.NewOption("[r] Regenerate", "r"),
			huh.NewOption("[p] Preview PR body", "p"),
			huh.NewOption("[n] No — cancel", "n"),
		).
		Value(&choice).
		Run()
	if err != nil {
		return ActionNo, "", err
	}

	switch choice {
	case "y":
		return ActionYes, "", nil
	case "n":
		return ActionNo, "", nil
	case "e":
		return ActionEdit, "", nil
	case "p":
		return ActionPreview, "", nil
	case "r":
		var extra string
		inputErr := huh.NewInput().
			Title("Regenerate with extra instruction (empty to just retry):").
			Value(&extra).
			Run()
		return ActionRegenerate, extra, inputErr
	}
	return ActionNo, "", nil
}

// EditTarget indicates which part of the content to edit.
type EditTarget int

const (
	EditBranch EditTarget = iota
	EditCommit
	EditPR
	EditAll
)

// AskEditTarget prompts the user to choose what to edit.
func AskEditTarget() (EditTarget, error) {
	var choice string
	err := huh.NewSelect[string]().
		Title("What would you like to edit?").
		Options(
			huh.NewOption("Branch name", "branch"),
			huh.NewOption("Commit message", "commit"),
			huh.NewOption("PR title / body", "pr"),
			huh.NewOption("Everything (YAML)", "all"),
		).
		Value(&choice).
		Run()
	if err != nil {
		return EditAll, err
	}
	switch choice {
	case "branch":
		return EditBranch, nil
	case "commit":
		return EditCommit, nil
	case "pr":
		return EditPR, nil
	default:
		return EditAll, nil
	}
}

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func lineCount(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

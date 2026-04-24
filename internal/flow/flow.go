package flow

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/moneyforward/figaro/internal/apperr"
	"github.com/moneyforward/figaro/internal/generate"
	"github.com/moneyforward/figaro/internal/gh"
	"github.com/moneyforward/figaro/internal/git"
	"github.com/moneyforward/figaro/internal/ollama"
	"github.com/moneyforward/figaro/internal/prereq"
	"github.com/moneyforward/figaro/internal/ui"
)

// Config holds the full workflow configuration.
type Config struct {
	Model      string
	OllamaHost string
	DryRun     bool
	Yes        bool
	Draft      bool
	Base       string
	NoPull     bool
}

// Run executes the main Figaro workflow.
func Run(cfg Config) error {
	r := &git.RealRunner{}
	w := os.Stdout

	// Step 1: Prerequisites
	fmt.Fprintln(w, "Checking prerequisites...")
	prereqCfg := prereq.Config{OllamaHost: cfg.OllamaHost, Model: cfg.Model}
	if err := prereq.Check(r, prereqCfg); err != nil {
		return err
	}
	fmt.Fprintln(w, "✓ git, gh, ollama all available")

	// Step 2: Detect default branch
	defaultBranch := cfg.Base
	if defaultBranch == "" {
		var err error
		defaultBranch, err = git.DetectDefaultBranch(r)
		if err != nil {
			return fmt.Errorf("%w: %v", apperr.ErrGit, err)
		}
	}
	fmt.Fprintf(w, "✓ default branch: %s\n", defaultBranch)

	// Step 3: Get current branch
	currentBranch, err := git.CurrentBranch(r)
	if err != nil {
		return fmt.Errorf("%w: %v", apperr.ErrGit, err)
	}
	fmt.Fprintf(w, "✓ current branch: %s\n", currentBranch)

	// Step 4: Move to default branch via stash if needed
	var stashRef string
	stashed := false

	if currentBranch != defaultBranch {
		fmt.Fprintln(w, "\nMoving to default branch...")
		stashMsg := git.StashMessage(time.Now())
		ref, err := git.StashPush(r, stashMsg)
		if err != nil {
			return fmt.Errorf("%w: stash push failed: %v", apperr.ErrGit, err)
		}
		stashRef = ref
		stashed = true
		fmt.Fprintf(w, "  ✓ stashed changes (%s)\n", stashMsg)

		if err := git.Checkout(r, defaultBranch); err != nil {
			_ = git.StashPop(r)
			return fmt.Errorf("%w: checkout %s failed: %v", apperr.ErrGit, defaultBranch, err)
		}
		fmt.Fprintf(w, "  ✓ checked out %s\n", defaultBranch)
	}

	if !cfg.NoPull {
		if err := git.Pull(r, defaultBranch); err != nil {
			if stashed {
				_ = git.StashPop(r)
			}
			return fmt.Errorf("%w: pull --ff-only failed. Use --no-pull to skip, or fix divergence manually: %v",
				apperr.ErrGit, err)
		}
		fmt.Fprintln(w, "  ✓ pulled latest (up to date)")
	}

	if stashed {
		if err := git.StashPop(r); err != nil {
			return fmt.Errorf("%w: stash pop conflict detected.\n  Stash ref: %s\n  Resolve conflicts, then run:\n    git stash drop %s\n  Then re-run figaro",
				apperr.ErrGit, stashRef, stashRef)
		}
		fmt.Fprintln(w, "  ✓ popped stash")
	}

	// Step 5: Get diff
	diff, err := git.GetDiff(r)
	if err != nil {
		return fmt.Errorf("%w: %v", apperr.ErrGit, err)
	}

	if cfg.DryRun {
		preview := diff
		if len(preview) > 500 {
			preview = preview[:500] + "\n... (truncated)"
		}
		fmt.Fprintln(w, "\n[dry-run] diff obtained:")
		fmt.Fprintln(w, preview)
		return nil
	}

	// Step 6: Generate with Ollama
	fmt.Fprintf(w, "\nGenerating with %s...\n", cfg.Model)
	ollamaCfg := ollama.Config{
		Host:        cfg.OllamaHost,
		Model:       cfg.Model,
		Temperature: 0.3,
		NumCtx:      16384,
	}

	content, err := generate.Run(ollamaCfg, diff, "")
	if err != nil {
		return err
	}

	// Step 7: User confirmation loop
	for {
		ui.ShowSummary(w, content)

		if cfg.Yes {
			break
		}

		action, extra, err := ui.AskAction()
		if err != nil {
			return fmt.Errorf("%w: prompt error: %v", apperr.ErrCancelled, err)
		}

		switch action {
		case ui.ActionYes:
			goto execute
		case ui.ActionNo:
			fmt.Fprintln(w, "Cancelled.")
			return apperr.ErrCancelled
		case ui.ActionPreview:
			fmt.Fprintln(w, "\n"+content.PRBody+"\n")
			continue
		case ui.ActionRegenerate:
			fmt.Fprintf(w, "Regenerating with %s...\n", cfg.Model)
			content, err = generate.Run(ollamaCfg, diff, extra)
			if err != nil {
				return err
			}
		case ui.ActionEdit:
			target, err := ui.AskEditTarget()
			if err != nil {
				return fmt.Errorf("%w: edit target selection failed: %v", apperr.ErrCancelled, err)
			}
			content, err = ui.EditContent(content, target)
			if err != nil {
				return fmt.Errorf("%w: edit failed: %v", apperr.ErrCancelled, err)
			}
		}
	}

execute:
	return executeFlow(r, w, content, defaultBranch, cfg)
}

func executeFlow(r git.Runner, w io.Writer, content *generate.Content, defaultBranch string, cfg Config) error {
	// Create branch
	branchName, err := git.CreateBranch(r, content.BranchName)
	if err != nil {
		return fmt.Errorf("%w: %v", apperr.ErrGit, err)
	}
	fmt.Fprintf(w, "✓ created branch: %s\n", branchName)

	// Stage and commit
	if err := git.AddAll(r); err != nil {
		return fmt.Errorf("%w: git add: %v", apperr.ErrGit, err)
	}
	commitMsg := content.CommitMessage()
	if err := git.CommitWithMessage(r, commitMsg); err != nil {
		return fmt.Errorf("%w: git commit: %v", apperr.ErrGit, err)
	}
	fmt.Fprintln(w, "✓ committed")

	// Push
	if err := git.Push(r); err != nil {
		return fmt.Errorf("%w: git push failed. Run manually:\n  git push -u origin HEAD\n%v",
			apperr.ErrGit, err)
	}
	fmt.Fprintln(w, "✓ pushed")

	// Create PR
	prCfg := gh.PRConfig{
		Title: content.PRTitle,
		Body:  content.PRBody,
		Base:  defaultBranch,
		Draft: cfg.Draft,
	}
	prURL, err := gh.CreatePR(r, prCfg)
	if err != nil {
		return fmt.Errorf("%w: run manually:\n  gh pr create --title %q --base %s\n%v",
			apperr.ErrGH, content.PRTitle, defaultBranch, err)
	}
	fmt.Fprintf(w, "\n✓ PR created: %s\n", prURL)

	// Warn if figaro stash was accidentally left behind
	if hasStash, _ := git.HasFigaroStash(r); hasStash {
		fmt.Fprintln(w, "⚠  Warning: figaro-auto-stash still in stash list. Check with: git stash list")
	}

	return nil
}

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/moneyforward/figaro/internal/apperr"
	"github.com/moneyforward/figaro/internal/flow"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	switch {
	case errors.Is(err, apperr.ErrCancelled):
		return 1
	case errors.Is(err, apperr.ErrGit):
		return 2
	case errors.Is(err, apperr.ErrOllama):
		return 3
	case errors.Is(err, apperr.ErrGH):
		return 4
	case errors.Is(err, apperr.ErrPrereq):
		return 10
	default:
		return 1
	}
}

func run() error {
	cfg := flow.Config{}

	flag.StringVar(&cfg.Model, "model", "gemma4:latest", "Ollama model to use")
	flag.StringVar(&cfg.OllamaHost, "ollama-host", "http://localhost:11434", "Ollama endpoint")
	flag.StringVar(&cfg.Language, "language", "ja", "Generation language (ja or en)")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Show generated content only, no git/gh operations")

	// --yes and -y both set the same field
	flag.BoolVar(&cfg.Yes, "yes", false, "Skip confirmation prompts (CI use)")
	flag.BoolVar(&cfg.Yes, "y", false, "Skip confirmation prompts (CI use)")

	flag.BoolVar(&cfg.Draft, "draft", false, "Create PR as draft")
	flag.StringVar(&cfg.Base, "base", "", "PR base branch (auto-detected if empty)")
	flag.BoolVar(&cfg.NoPull, "no-pull", false, "Skip git pull on default branch")

	flag.Parse()

	return flow.Run(cfg)
}

---
name: agito
description: |
  Automates Git workflow using local Ollama LLM. Generates branch names, commit messages,
  and pull request content from git diff — without sending code to external services.
  Use when user says "create a PR with agito", "agito PR", "make a PR locally", or
  when they explicitly want to use local LLM for git automation.
  Requires Ollama running locally with gemma4:latest (or another model) pulled.
allowed-tools: Bash Read
---

# agito — Local LLM Git Automation

Generates branch names, commit messages, and PRs from your working tree diff using a local
Ollama model. Code never leaves your machine.

## Prerequisites Check (always run first)

```bash
# Check Ollama is running
curl -s http://localhost:11434/api/tags | jq -r '.models[].name' 2>/dev/null \
  || echo "OLLAMA_NOT_RUNNING"

# Check agito binary exists
which agito || echo "AGITO_NOT_INSTALLED"
```

**If `OLLAMA_NOT_RUNNING`:** Tell user to run `ollama serve` and stop.
**If `AGITO_NOT_INSTALLED`:** Tell user to run `go install github.com/tatsuo48/agito/cmd/agito@latest` and stop.
**If `gemma4:latest` not in the list:** Tell user to run `ollama pull gemma4:latest`.

## Usage

```bash
# Standard run — creates branch, commits, pushes, and opens a PR
# --yes is REQUIRED in Claude Code (non-TTY environment)
agito --yes

# Draft PR
agito --yes --draft

# Preview only — shows generated content without executing git/gh operations
# Note: Ollama is still required even for dry-run
agito --dry-run --yes

# Specify base branch
agito --yes --base develop

# Skip git pull on default branch
agito --yes --no-pull

# Use a different Ollama model
agito --yes --model llama3.2:latest

# Use a remote Ollama instance
agito --yes --ollama-host http://remote-host:11434
```

## Workflow Before Running

1. Run the prerequisites check (curl + which agito)
2. Check `git status` to confirm there are changes to commit
3. Ask the user which options they want (--draft? --base? --model?)
4. Show the exact command you will run and ask for confirmation
5. Run `agito --yes [options]`
6. Check exit code and handle errors

## Exit Codes

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Report PR URL to user |
| 1 | User cancelled | Treat as normal — no action needed |
| 2 | git failure | Show `git status` output |
| 3 | Ollama failure | Check `ollama serve` and `ollama list` |
| 4 | gh failure | Check `gh auth status` |
| 10 | Prerequisites not met | Show the error message from agito |

## agito vs pr-create skill

- **agito (this skill)**: Use when privacy matters — code stays local via Ollama
- **pr-create skill**: Use when Ollama isn't available or Claude's quality is preferred

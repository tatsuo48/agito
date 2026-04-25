# Agito

A CLI tool that automates the entire git workflow — branch creation, commit, and PR — from your working tree changes, using a local LLM (Ollama).

## Features

- **Offline** — All LLM inference runs locally via Ollama. Your code never leaves your machine.
- **Fast** — Single Go binary with minimal startup overhead.
- **Safe** — Review, edit, or regenerate the output before any git operations are performed.

## Installation

### Via Go

```bash
go install github.com/tatsuo48/agito/cmd/agito@latest
```

### From source

```bash
git clone https://github.com/tatsuo48/agito
cd agito
make install
```

## Prerequisites

- `git` installed and the current directory is a git repository
- `gh` (GitHub CLI) installed and authenticated (`gh auth login`)
- `ollama serve` running
- Default model `gemma4:latest` pulled (`ollama pull gemma4:latest`)

## Usage

```bash
agito [options]
```

Agito detects changes in your working tree and automatically:

1. Generates a branch name and creates the branch
2. Generates a commit message and commits
3. Generates a PR title and description and opens the PR

### Options

| Option          | Default                  | Description                                            |
| --------------- | ------------------------ | ------------------------------------------------------ |
| `--model`       | `gemma4:latest`          | Ollama model to use                                    |
| `--ollama-host` | `http://localhost:11434` | Ollama endpoint                                        |
| `--dry-run`     | false                    | Generate and preview output only; no git/gh operations |
| `--yes` / `-y`  | false                    | Skip confirmation prompts (for CI use)                 |
| `--draft`       | false                    | Create the PR as a draft                               |
| `--base`        | auto-detected            | Base branch for the PR                                 |
| `--no-pull`     | false                    | Skip `git pull` on the default branch                  |

### Exit codes

| Code | Meaning               |
| ---- | --------------------- |
| 0    | Success               |
| 1    | Cancelled by user     |
| 2    | Git operation failed  |
| 3    | Ollama failed         |
| 4    | gh failed             |
| 10   | Prerequisites not met |

## Troubleshooting

### Ollama is not running

```bash
ollama serve
```

### Model not found

```bash
ollama pull gemma4:latest
```

### Conflict on `stash pop`

Agito does not attempt to resolve conflicts automatically. It exits leaving the stash intact.
Check the stash ID shown in the error and resolve manually:

```bash
# Check what's conflicting
git status

# After resolving conflicts
git add .
git stash drop stash@{0}

# Re-run Agito
agito
```

### Push failed — manual recovery

```bash
git push -u origin HEAD
```

### PR creation failed — manual recovery

```bash
gh pr create --title "..." --body-file /tmp/agito-pr-body-*.md --base main
```

## Claude Code Integration

Use agito directly from [Claude Code](https://claude.ai/code) as a plugin.

### Plugin Installation

```bash
claude plugin marketplace add tatsuo48/agito
claude plugin install agito@agito
```

Or manage plugins interactively with `/plugin` inside Claude Code.

### Triggering the Skill

After installing, trigger the skill in any Claude Code conversation:

- *"Create a PR with agito"*
- *"Use agito to make a PR"*
- *"agito PR, draft mode"*

Claude Code will check that Ollama is running and `agito` is installed, then run
`agito --yes` (with your chosen options) automatically.

## Development

```bash
# Run tests
make test

# Build
make build

# Lint
make lint
```

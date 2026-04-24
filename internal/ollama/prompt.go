package ollama

import "fmt"

// BuildPrompt builds the full prompt from diff content and options.
func BuildPrompt(diffContent, extraInstruction string) string {
	extra := ""
	if extraInstruction != "" {
		extra = fmt.Sprintf("\nAdditional user instruction: %s\n", extraInstruction)
	}

	return fmt.Sprintf(`You are Figaro, a git assistant that generates branch names, commit messages, and pull requests from git diffs.

Analyze the following git diff and produce a JSON object with exactly these fields:

- "branch_name": kebab-case branch name prefixed with one of: feat/, fix/, chore/, docs/, refactor/, test/, perf/. Max 50 chars. English only.
- "commit_type": one of: feat, fix, chore, docs, refactor, test, perf
- "commit_scope": short scope in English (e.g., "api", "cache", "auth") or null
- "commit_subject": imperative mood, max 72 chars. English.
- "commit_body": markdown explaining what changed and why. English. 2-5 sentences.
- "pr_title": similar to commit_subject but can include conventional commit prefix. English.
- "pr_body": markdown with sections "## Context", "## Changes", "## How to verify".

Rules:
- Output ONLY a JSON object. No markdown code fences. No preamble.
- Do not include file paths verbatim in the subject line.
- If the diff is trivial (typo, whitespace), keep the messages short.
- If you cannot determine intent from the diff, use "chore" as the type.
%s
<diff>
%s
</diff>`, extra, diffContent)
}

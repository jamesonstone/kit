package templates

import "strings"

func memoryTooling() string {
	content := strings.Replace(agentsTooling, "# Tooling\n", `# Tooling

## Kit Evidence Sequence

- Use `+"`kit capabilities <command> --json`"+` when side effects are not already established.
- Resolve `+"`kit context resolve --workflow <slug> --json`"+` before coding-agent work and load the selected local evidence.
- Rerun resolution after material scope changes; never treat resolved JSON as a new source of truth.
`, 1)
	content = strings.ReplaceAll(content,
		"Use `kit dispatch` when broad work must be turned into a safe Agent Team Plan",
		"Use `kit dispatch` after native planning when an accepted plan needs a safe multi-lane execution topology",
	)
	content = strings.ReplaceAll(content,
		"except for preparing the writable worktree and its exact `.env` link when needed",
		"except for preparing the writable worktree and its exact `.env` and `.envrc` links when needed",
	)
	return content
}

func memoryRLM() string {
	content := strings.Replace(agentsRLM, "## Runtime Loop\n", `## Coding Agent Contract

1. Run `+"`kit context resolve --workflow <slug> --json`"+` with relevant feature and path hints.
2. Load every required selected artifact before acting.
3. Treat blocked resolution as a hard evidence gap.
4. Rerun resolution after material scope changes.

## Runtime Loop
`, 1)
	content = strings.ReplaceAll(content,
		"For v2 feature-scoped work",
		"For living-spec feature work",
	)
	return strings.ReplaceAll(content,
		"Use `kit dispatch` only when the work moves from broad discovery into multi-lane execution planning",
		"Use `kit dispatch` only after native planning has established a narrow implementation topology",
	)
}

func memoryReferencesREADME() string {
	content := referencesREADME
	content = strings.Replace(content, "## Starter Files\n", "- Use `rules/coding-agent-context-usage.md` for the capability, resolution, loading, and re-resolution sequence\n- Store declarative coding-agent workflow contracts under `workflows/<slug>.md`\n\n## Starter Files\n", 1)
	return strings.ReplaceAll(content,
		"Use `worktrees.md` when present for the canonical native Git worktree hierarchy, naming, shared-state model, safety contract, and optional manual convenience commands",
		"Use `worktrees.md` for the canonical native Git worktree hierarchy, naming, shared-state model, environment ownership, and safety contract",
	)
}

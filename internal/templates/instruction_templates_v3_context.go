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
		"Use `kit dispatch --loop --pr <target>` when current unresolved CodeRabbit PR review feedback should become a human-reviewed dispatch prompt instead of an agent repair loop.",
		"Use `kit dispatch --loop --pr <target> --watch` only for bounded expected CodeRabbit intake; waiting is deterministic and model-free.",
	)
	content = strings.ReplaceAll(content,
		"except for preparing the writable worktree and its exact `.env` link when needed",
		"except for preparing the writable worktree and its exact `.env` and `.envrc` links when needed",
	)
	content = strings.ReplaceAll(content,
		"Link the invoking checkout's `.env` into writable lanes by default when it exists, using only an exact verified symlink; omit the link when isolation is required",
		"Link the primary checkout's `.env` and `.envrc` into writable lanes by default when each exists, using only exact verified symlinks; omit both links when isolation is required",
	)
	content = strings.ReplaceAll(content,
		"Never copy `.env` contents or automatically share `.envrc`; worktree tooling does not manage runtime services, databases, ports, Temporal state, processes, or sibling repositories",
		"Never copy environment contents or overwrite destination environment material; preserve a repository- or user-supplied `.envrc`, and remember that direnv approval remains path-specific; worktree tooling does not manage runtime services, databases, ports, Temporal state, processes, or sibling repositories",
	)
	return strings.ReplaceAll(content,
		"Load `docs/references/worktrees.md` when present and worktree creation",
		"Load `docs/references/worktrees.md` when worktree creation",
	)
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

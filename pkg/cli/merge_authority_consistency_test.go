package cli

import (
	"strings"
	"testing"
)

func TestActivePolicyUsesOneMergeAuthorityModel(t *testing.T) {
	files := map[string]string{
		"AGENTS.md":  readRepositoryFile(t, "AGENTS.md"),
		"CLAUDE.md":  readRepositoryFile(t, "CLAUDE.md"),
		"Copilot":    readRepositoryFile(t, ".github/copilot-instructions.md"),
		"Guardrails": readRepositoryFile(t, "docs/agents/GUARDRAILS.md"),
	}
	for name, content := range files {
		for _, check := range []string{
			"Merge is a distinct mutation boundary",
			"never invent merge readiness",
			"Do not stop for a separate merge-consent prompt",
			"Only exact current `MERGE_READY` nodes may merge",
			"A changed head invalidates readiness, not accepted-task authority",
			"Never bypass protection",
		} {
			if !strings.Contains(content, check) {
				t.Errorf("%s missing active merge policy %q", name, check)
			}
		}
	}
	for _, path := range []string{"AGENTS.md", "CLAUDE.md"} {
		if !strings.Contains(files[path], "Issue, branch, staging, commit, push, PR, and merge actions are distinct mutation boundaries") {
			t.Errorf("%s mutation list does not include merge distinctly", path)
		}
	}
}

func TestActivePolicyRejectsContradictoryMergeAuthority(t *testing.T) {
	activePaths := []string{
		"docs/references/rules/safety-guardrails.md",
		"docs/references/rules/work-lane-gating.md",
		"docs/references/rules/github-pr-delivery.md",
		"docs/references/rules/github-pr-merge.md",
		"docs/references/rules/agent-team-orchestration.md",
		"docs/references/rules/cross-repository-program-coordination.md",
		"docs/references/rules/infrastructure-change-approval.md",
		"docs/references/rules/testing-and-environment-validation.md",
	}
	combined := ""
	for _, path := range activePaths {
		combined += "\n" + readRepositoryFile(t, path)
	}

	required := []string{
		"The ledger records and reconciles that authority; it never creates it",
		"Read-only verification agents never merge",
		"Protection bypass, admin override, review bypass, required-check bypass",
		"Do not stop for a separate merge-consent prompt",
		"Proceed autonomously when the graph contains only additive or rollback-preserving effects",
	}
	for _, check := range required {
		if !strings.Contains(combined, check) {
			t.Errorf("active policy missing consistency boundary %q", check)
		}
	}

	for _, forbidden := range []string{
		"GitHub access is never permission to:\n\n- Merge.",
		"program ledger creates merge authority",
		"verification agents may merge",
		"passing checks authorize merge",
		"Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set",
		"never imply merge consent",
		"Obtain one explicit user confirmation for the complete bounded batch",
	} {
		if strings.Contains(combined, forbidden) {
			t.Errorf("active policy contains contradictory authority %q", forbidden)
		}
	}
}

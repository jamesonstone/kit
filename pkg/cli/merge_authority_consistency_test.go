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
			"Standing merge authority exists only when a human explicitly authorizes a bounded task, goal, or program",
			"may bind later-created in-scope PRs and refreshed heads",
			"Only exact current `MERGE_READY` nodes may merge",
			"A changed in-scope head invalidates readiness, not standing authority",
			"A commit SHA or head OID identifies readiness evidence only",
			"Never request exact-head reauthorization",
			"Pause, hold, or revocation stops affected actions and dependents",
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
	combined = normalizeMergeAuthorityPolicy(combined)

	required := []string{
		"it never creates authority",
		"Read-only verification agents never merge",
		"Protection bypass, admin override, review bypass, required-check bypass",
		"explicitly authorizes a bounded task, goal, or program",
		"### Standard Deployments Under Standing Authority",
		"IAM, network, KMS, secrets, database schema/data-loss",
	}
	for _, check := range required {
		if !strings.Contains(combined, check) {
			t.Errorf("active policy missing consistency boundary %q", check)
		}
	}

	for _, forbidden := range append([]string{
		"GitHub access is never permission to:\n\n- Merge.",
		"program ledger creates merge authority",
		"verification agents may merge",
		"passing checks authorize merge",
		"Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set",
		"never imply merge consent",
		"accepted task or active `/goal` authorizes",
		"Proceed autonomously when the graph contains only additive or rollback-preserving effects",
	}, exactHeadReauthorizationPhrases()...) {
		if strings.Contains(combined, normalizeMergeAuthorityPolicy(forbidden)) {
			t.Errorf("active policy contains contradictory authority %q", forbidden)
		}
	}
}

func TestNormalizeMergeAuthorityPolicyCollapsesWhitespace(t *testing.T) {
	policy := "GitHub access is never permission to: - Merge."
	forbidden := "GitHub access is never permission to:\n\n- Merge."
	if !strings.Contains(normalizeMergeAuthorityPolicy(policy), normalizeMergeAuthorityPolicy(forbidden)) {
		t.Fatal("normalized multiline phrase did not match normalized policy")
	}
}

func normalizeMergeAuthorityPolicy(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func TestProjectSummaryDoesNotRequireExactHeadReauthorization(t *testing.T) {
	summary := normalizeStandingAuthorityPolicy(readRepositoryFile(t, "docs/PROJECT_PROGRESS_SUMMARY.md"))
	for _, forbidden := range exactHeadReauthorizationPhrases() {
		if strings.Contains(summary, normalizeStandingAuthorityPolicy(forbidden)) {
			t.Errorf("project summary contains superseded authority requirement %q", forbidden)
		}
	}
}

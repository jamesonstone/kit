package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGitHubPRMergeRulesetIsValid(t *testing.T) {
	ruleset := loadGitHubPRMergeRuleset(t)
	normalized := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"Standing merge authority exists only when a human explicitly authorizes a bounded task, goal, or program",
		"Bind later pull requests only when each is directly required",
		"Record the standing-authority source and selector",
		"`MERGE_READY`",
		"`BLOCKED`",
		"`UNKNOWN`",
		"authenticated GitHub actor",
		"current ready frontier",
		"Read-only verification agents never merge",
		"Merge success never implies workflow success",
		"Treat routine remediation as an update to the existing pull request",
		"Treat the resolved existing pull-request set as explicit continuation",
		"Do not create a new coordination issue, branch, worktree",
		"A changed in-scope head invalidates readiness, not standing authority",
		"Only an explicit human resume or new grant restores it",
		"IAM, network, KMS, secrets, database schema or data-loss changes",
		"continue authorized merge and deployment work without interleaving UI or browser walkthrough verification after each result",
		"After every result in the authorized set is delivered, run one final UI verification",
	} {
		if !strings.Contains(normalized, check) {
			t.Errorf("github-pr-merge ruleset missing %q", check)
		}
	}
}

func TestGitHubPRMergeRulesetCoversRequiredScenarios(t *testing.T) {
	rules := strings.Join(strings.Fields(loadGitHubPRMergeRuleset(t).Body), " ")
	scenarios := map[string][]string{
		"later blocker PR uses standing authority": {
			"Later in-scope blocker PR:", "Blocker PR #84 was created later in a governed lane",
		},
		"generic task is not standing authority": {
			"Accepting a generic implementation task", "do not create standing authority or `MERGE_READY`",
		},
		"semantic selector binds exact current target": {
			"grant may use a semantic scope before pull-request numbers", "Materialize the exact PR and current head",
		},
		"independent ready concurrency": {
			"Independent `MERGE_READY` nodes may merge concurrently",
		},
		"dependency ordering": {
			"Dependency chains and", "remain serialized",
		},
		"head drift": {
			"Head or base drift invalidates readiness",
		},
		"in-place repair keeps the PR": {
			"An in-scope in-place repair keeps the same pull request", "invalidates its readiness until",
		},
		"routine remediation preserves the PR": {
			"Do not create recursive corrective pull requests", "push to the same branch without rebasing, force-pushing",
		},
		"merge coordination preserves existing lanes": {
			"resolved existing pull-request set as explicit continuation", "Do not create a new coordination issue, branch, worktree",
		},
		"replacement PR is exceptional": {
			"Use a replacement pull request only when remediation materially changes",
		},
		"pending or missing checks": {
			"pending or missing expected checks", "never passing",
		},
		"unauthorized extra PR": {
			"#91 changes the production network and is not covered", "BLOCKED pending explicit expanded authority",
		},
		"participant authority": {
			"A participant may merge only specifically assigned PR nodes", "Subagent assignment alone does not create merge authority",
		},
		"infrastructure risk remains separate": {
			"infrastructure creation, replacement, or deletion", "outside standing merge/deploy authority",
		},
		"unresolved risk classification": {
			"Unresolved classification makes the node `UNKNOWN` or `BLOCKED`",
		},
		"standard deployment boundary": {
			"repository-approved existing standard workflow", "already-provisioned targets",
		},
		"partial wave failure": {
			"A failure on one node stops that node and its dependents", "Preserve exact",
		},
		"deadline-mode UI deferral": {
			"continue authorized merge and deployment work without interleaving UI or browser walkthrough verification after each result",
			"After every result in the authorized set is delivered, run one final UI verification",
		},
		"actor mismatch isolation": {
			"Identity failure blocks only that repository node and its dependents",
		},
		"merge queue and docs-only squash": {
			"Use the required merge queue when policy requires it", "For documentation-only squash merges",
		},
	}
	scenarios["pause and revoke"] = []string{
		"Pause and revocation:", "Passing checks and a new head do not resume it",
	}
	for scenario, checks := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			for _, check := range checks {
				if !strings.Contains(rules, check) {
					t.Errorf("scenario contract missing %q", check)
				}
			}
		})
	}
	if strings.Contains(rules, "Source remediation becomes a normal corrective pull request") {
		t.Fatal("routine remediation regressed to a mandatory replacement pull request")
	}
}

func loadGitHubPRMergeRuleset(t *testing.T) rulesetDocument {
	t.Helper()
	path := filepath.Join("..", "..", "docs", "references", "rules", "github-pr-merge.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "github-pr-merge"); len(issues) > 0 {
		t.Fatalf("github-pr-merge ruleset issues = %#v", issues)
	}
	return ruleset
}

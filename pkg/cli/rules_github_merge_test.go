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
		"accepted task or active `/goal`",
		"Do not stop for a separate merge-consent prompt",
		"Record the accepted-scope source and exact in-scope pull-request set",
		"`MERGE_READY`",
		"`BLOCKED`",
		"`UNKNOWN`",
		"authenticated GitHub actor",
		"current ready frontier",
		"Read-only verification agents never merge",
		"Merge success never implies workflow success",
		"Treat routine remediation as an update to the existing pull request",
		"Treat that exact existing pull-request set as explicit continuation",
		"Do not create a new coordination issue, branch, worktree",
		"A changed head invalidates readiness, not accepted-task authority",
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
		"accepted goal merges without another prompt": {
			"Accepted task/goal includes owner/service#84.", "Merge #84 without another consent prompt",
		},
		"ready PR is not MERGE_READY": {
			"Opening, updating, or approving a ready pull request does not invent",
		},
		"automatic preflight is not MERGE_READY": {
			"Automatic clean-preflight delivery allocation", "are also not `MERGE_READY`",
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
			"An in-scope in-place repair keeps the same pull", "invalidates its readiness until",
		},
		"routine remediation preserves the PR": {
			"Do not create recursive corrective pull requests", "push to the same branch without rebasing, force-pushing",
		},
		"merge coordination preserves existing lanes": {
			"exact existing pull-request set as explicit continuation", "Do not create a new coordination issue, branch, worktree",
		},
		"replacement PR is exceptional": {
			"Use a replacement pull request only when remediation materially changes",
		},
		"pending or missing checks": {
			"pending or missing expected checks", "never passing",
		},
		"unauthorized extra PR": {
			"#91 is green and related but is not", "BLOCKED pending product-scope clarification",
		},
		"participant authority": {
			"A participant may merge only specifically assigned PR nodes", "Subagent assignment alone does not create merge authority",
		},
		"destructive merge confirmation": {
			"A merge known to trigger a destructive effect is not `MERGE_READY`",
		},
		"routine application operation merge": {
			"routine application operation", "do not require infrastructure-change-approval",
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

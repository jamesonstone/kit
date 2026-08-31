package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGitHubPRMergeRulesetIsValid(t *testing.T) {
	ruleset := loadGitHubPRMergeRuleset(t)
	for _, check := range []string{
		"direct user",
		"request or accepted bounded merge plan authorizes it",
		"Record the authorization source and exact approved pull-request set",
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
		"exact-head merge authorization before it can become",
		"continue authorized merge and deployment work without interleaving UI or",
		"the authorized set is delivered, run one final UI verification",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Errorf("github-pr-merge ruleset missing %q", check)
		}
	}
}

func TestGitHubPRMergeRulesetCoversRequiredScenarios(t *testing.T) {
	rules := loadGitHubPRMergeRuleset(t).Body
	scenarios := map[string][]string{
		"authorized single PR": {
			"Direct request: merge owner/service#84.", "State: MERGE_READY. Merge only #84.",
		},
		"ready PR is not authorization": {
			"Opening, updating, or approving a ready pull request does not trigger this rule",
		},
		"automatic preflight is not authorization": {
			"Automatic clean-preflight delivery allocation", "are also not authorization",
		},
		"approved multi-repository batch": {
			"accepts a bounded merge plan naming the complete", "approved PR set",
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
		"in-place remediation invalidates old-head authority": {
			"An authorized in-place repair keeps the same pull", "invalidates its readiness and prior exact-head merge",
		},
		"routine remediation preserves the PR": {
			"Do not create recursive corrective pull requests", "push to the same branch without rebasing, force-pushing",
		},
		"merge coordination preserves existing lanes": {
			"exact existing pull-request set as explicit continuation", "Do not create a new coordination issue, branch, worktree",
		},
		"replacement PR is exceptional": {
			"Use a replacement pull request only when remediation materially changes", "is a new node and is not automatically added",
		},
		"pending or missing checks": {
			"pending or missing expected checks", "never passing",
		},
		"unauthorized extra PR": {
			"#91 is green and related but is not", "BLOCKED pending follow-up authorization",
		},
		"participant authority": {
			"A participant may merge only specifically assigned PR nodes", "Subagent assignment alone does not create merge authority",
		},
		"infrastructure-triggering merge": {
			"known to trigger a covered infrastructure mutation", "accepted plan must identify",
		},
		"routine application operation merge": {
			"routine application operation", "does not require infrastructure-change-approval",
		},
		"partial wave failure": {
			"A failure on one node stops that node and its dependents", "Preserve exact",
		},
		"deadline-mode UI deferral": {
			"continue authorized merge and deployment work without interleaving UI or",
			"the authorized set is delivered, run one final UI verification",
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

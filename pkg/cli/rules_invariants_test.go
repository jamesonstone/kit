package cli

import (
	"context"
	"strings"
	"testing"
)

// ruleInvariants pins one marker per safety-relevant invariant in each shipped
// rule. Markers pin meaning, not prose; reword freely while keeping each stated.
var ruleInvariants = map[string][]string{
	"delivery": {
		"Default to a new lane without asking",
		"reuses that pull request's head branch",
		"Never ask the user to choose between lanes",
		"The primary checkout is read-only for agent work",
		"except the fast-forward-only default-branch update in `work-cleanup`",
		"Never add co-author trailers",
		"Assign every created or reused issue and pull request to the human",
		"Never commit to the default or a protected branch, force-push, or rewrite pushed history",
		"Stage explicit paths only",
	},
	"github-pr-merge": {
		"Merge only under explicit human authorization",
		"never request exact-head reauthorization",
		"Only current `MERGE_READY` pull requests merge",
		"A human pause, hold, or revocation stops",
		"Never bypass branch protection",
		"A merge is not proof of deployment",
		"It never applies to a check that ran and failed",
		"Never report a not-run hosted check as passed",
		"The human's first bypass approval is a standing grant",
		"It does not create merge authority",
		"passes on the exact head being merged",
		"Never use admin overrides",
		"post a pull-request comment that names the outage, the local validation commands, their results, and the head SHA",
		"Re-run CI on every pull request merged under the bypass",
	},
	"infrastructure-change-approval": {
		"present one outline",
		"Get one explicit approval for that batch",
		"Deleting, destroying, or removing infrastructure always needs explicit confirmation",
		"Generic task acceptance never authorizes deployment",
	},
	"deletion-safety": {
		"An unqualified \"delete\" or \"remove\" means a recoverable soft delete",
		"get a specific human confirmation for those targets",
		"if they changed, outline and confirm again",
	},
	"slack-read-only": {
		"Never post, reply, react, edit, delete, forward, share, or change channel state without approval",
		"Approval is single-use",
	},
	"aws-agent-toolkit-guidance": {
		"run `kit aws verify`",
		"never fall back to default, ambient, or other profiles",
		"Never read secret values into the session",
	},
	"testing-and-environment-validation": {
		"they stay primary",
		"Agents may drive any browser session the host exposes",
		"never touch customer data",
	},
	"agent-team-orchestration": {
		"Delegated agents never create issues, branches, commits, pushes, pull requests, comments, or merges",
		"Never let two agents write overlapping files concurrently",
		"Report only real topology",
	},
	"deadline-mode": {
		"Only when the user explicitly declares a real deadline",
		"Never weaken merge authority",
	},
}

func TestShippedRulesKeepTheirInvariants(t *testing.T) {
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatalf("embeddedRulesetRegistry() error = %v", err)
	}
	bySlug := map[string]string{}
	for _, item := range registry {
		bySlug[item.Slug] = strings.Join(strings.Fields(item.Content), " ")
	}
	for slug, markers := range ruleInvariants {
		content, ok := bySlug[slug]
		if !ok {
			t.Errorf("rule %s is not shipped", slug)
			continue
		}
		for _, marker := range markers {
			if !strings.Contains(content, marker) {
				t.Errorf("rule %s lost invariant %q", slug, marker)
			}
		}
	}
	for _, retired := range []string{"CAPABILITY_NEGOTIATING", "Capability Manifest", "single-lane, because", "Pull-Request Landing Plan", "Delivery Contract:"} {
		for slug, content := range bySlug {
			if strings.Contains(content, retired) {
				t.Errorf("rule %s reintroduced retired ceremony %q", slug, retired)
			}
		}
	}
	for _, retired := range []string{"Never drive the user's active browser profile", "Prefer the host's built-in or an isolated browser"} {
		if strings.Contains(bySlug["testing-and-environment-validation"], retired) {
			t.Errorf("rule testing-and-environment-validation reintroduced retired browser restriction %q", retired)
		}
	}
}

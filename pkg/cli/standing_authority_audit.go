package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func auditStandingAuthorityPolicy(projectRoot string) []reconcileFinding {
	var findings []reconcileFinding
	for _, check := range standingAuthorityChecks() {
		absolutePath := filepath.Join(projectRoot, filepath.FromSlash(check.path))
		content, err := os.ReadFile(absolutePath)
		if err != nil {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("failed to read standing-authority policy document: %v", err),
				templateSource(projectRoot),
				"restore policy document readability before reconciling standing-authority guidance",
				[]string{fmt.Sprintf("ls -l %s", absolutePath)},
			))
			continue
		}
		body := string(content)
		normalizedBody := normalizeStandingAuthorityPolicy(body)
		for _, required := range check.required {
			if strings.Contains(normalizedBody, normalizeStandingAuthorityPolicy(required)) {
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("policy document is missing standing-authority guidance %q", required),
				templateSource(projectRoot),
				"restore Kit's explicit bounded standing-authority, current-head readiness, standard-deployment, and pause/revocation boundaries",
				[]string{fmt.Sprintf("rg -n %q %s", required, absolutePath)},
			))
			break
		}
		for _, forbidden := range check.forbidden {
			if !strings.Contains(normalizedBody, normalizeStandingAuthorityPolicy(forbidden)) {
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("policy document contains superseded standing-authority guidance %q", forbidden),
				templateSource(projectRoot),
				"remove generic accepted-task merge/deploy authority and additive-infrastructure autonomy; keep dynamic in-scope binding plus exact current readiness and separate risk approvals",
				[]string{fmt.Sprintf("rg -n %q %s", forbidden, absolutePath)},
			))
			break
		}
	}
	return findings
}

type standingAuthorityCheck struct {
	path      string
	required  []string
	forbidden []string
}

func standingAuthorityChecks() []standingAuthorityCheck {
	return []standingAuthorityCheck{
		{
			path: "docs/references/rules/github-pr-merge.md",
			required: []string{
				"Standing merge authority exists only when a human explicitly authorizes a",
				"Bind later pull requests only when each is directly required",
				"A changed in-scope head invalidates readiness, not standing",
				"Only exact current `MERGE_READY` nodes may merge",
				"A commit SHA or head OID identifies readiness evidence only",
				"without exact-head reauthorization",
				"Only an explicit human resume or new grant restores it",
			},
			forbidden: append(exactHeadReauthorizationPhrases(),
				"accepted task or active `/goal`",
				"accepted-task authority",
				"Additive and routine application operations proceed autonomously",
				"Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set",
			),
		},
		{
			path: "docs/references/rules/work-lane-gating.md",
			required: []string{
				"A commit SHA or head OID is readiness evidence, never an authorization",
				"without exact-head reauthorization",
			},
			forbidden: exactHeadReauthorizationPhrases(),
		},
		{
			path: "docs/references/rules/github-pr-delivery.md",
			required: []string{
				"A SHA or head OID identifies the evidence to revalidate",
				"already-authorized standard deployment and browser retry",
			},
			forbidden: exactHeadReauthorizationPhrases(),
		},
		{
			path: "docs/references/rules/cross-repository-program-coordination.md",
			required: []string{
				"A commit SHA or head OID is an evidence pointer, never an authorization",
				"without exact-head reauthorization",
			},
			forbidden: exactHeadReauthorizationPhrases(),
		},
		{
			path: "docs/references/workflows/pull-request-merge.md",
			required: []string{
				"Treat the SHA/head OID as the evidence key only",
				"Do not request exact-head reauthorization after checks pass",
			},
			forbidden: exactHeadReauthorizationPhrases(),
		},
		{
			path: "docs/references/workflows/release-orchestration.md",
			required: []string{
				"Each SHA/head OID identified readiness evidence only",
				"No changed head required exact-head reauthorization",
			},
			forbidden: exactHeadReauthorizationPhrases(),
		},
		{
			path: "docs/references/rules/infrastructure-change-approval.md",
			required: []string{
				"Standing merge/deploy authority never authorizes a covered mutation",
				"### Standard Deployments Under Standing Authority",
				"It excludes novel provider commands, new targets, workflow mutation, IAM",
				"Deleting, destroying, or removing infrastructure always requires explicit",
			},
			forbidden: []string{
				"Proceed autonomously when the graph contains only additive or rollback-preserving effects",
				"Additive IAM, network topology",
			},
		},
	}
}

func exactHeadReauthorizationPhrases() []string {
	return []string{
		"prior merge authority is invalid",
		"changed head invalidates prior merge authority",
		"changed head loses prior readiness and merge authority",
		"requires fresh exact-head authorization",
		"require fresh exact-head authorization",
		"fresh exact-head authorization",
		"exact-head merge authorization",
		"reauthorize the current head",
		"refreshed head needs exact-head authorization",
	}
}

func normalizeStandingAuthorityPolicy(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

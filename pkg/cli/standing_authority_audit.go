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
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("failed to read standing-authority policy document: %v", err),
				"restore policy document readability before reconciling standing-authority guidance",
			))
			continue
		}
		body := string(content)
		normalizedBody := normalizeStandingAuthorityPolicy(body)
		for _, forbidden := range check.forbidden {
			if !strings.Contains(normalizedBody, normalizeStandingAuthorityPolicy(forbidden)) {
				continue
			}
			findings = append(findings, newFinding(
				reconcileSeverityWarning,
				absolutePath,
				fmt.Sprintf("policy document contains superseded standing-authority guidance %q", forbidden),
				"remove generic accepted-task merge/deploy authority and additive-infrastructure autonomy; keep dynamic in-scope binding plus exact current readiness and separate risk approvals",
			))
			break
		}
	}
	return findings
}

type standingAuthorityCheck struct {
	path      string
	forbidden []string
}

// standingAuthorityChecks guards locally customized rule copies against
// reintroducing superseded merge and infrastructure authority (see spec 0075).
func standingAuthorityChecks() []standingAuthorityCheck {
	return []standingAuthorityCheck{
		{
			path: "docs/references/rules/github-pr-merge.md",
			forbidden: append(exactHeadReauthorizationPhrases(),
				"accepted task or active `/goal`",
				"accepted-task authority",
				"Additive and routine application operations proceed autonomously",
				"Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set",
			),
		},
		{path: "docs/references/rules/delivery.md", forbidden: exactHeadReauthorizationPhrases()},
		{path: "docs/references/rules/cross-repository-program-coordination.md", forbidden: exactHeadReauthorizationPhrases()},
		{
			path: "docs/references/rules/infrastructure-change-approval.md",
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
		"requires exact-head reauthorization",
		"require exact-head reauthorization",
		"fresh exact-head authorization",
		"exact-head merge authorization",
		"reauthorize the current head",
		"refreshed head needs exact-head authorization",
	}
}

func normalizeStandingAuthorityPolicy(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

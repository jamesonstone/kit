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
			continue
		}
		body := string(content)
		for _, required := range check.required {
			if strings.Contains(body, required) {
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
			if !strings.Contains(body, forbidden) {
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
				"Only an explicit human resume or new grant restores it",
			},
			forbidden: []string{
				"accepted task or active `/goal`",
				"accepted-task authority",
				"Additive and routine application operations proceed autonomously",
				"Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set",
			},
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

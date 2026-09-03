package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func auditNondestructiveAuthorityPolicy(projectRoot string) []reconcileFinding {
	var findings []reconcileFinding
	for _, check := range nondestructiveAuthorityChecks() {
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
				fmt.Sprintf("policy document is missing autonomous non-destructive guidance %q", required),
				templateSource(projectRoot),
				"restore Kit's accepted-task merge readiness and destructive-only infrastructure confirmation; local-custom copies must not reintroduce non-destructive consent pauses",
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
				fmt.Sprintf("policy document still requires superseded non-destructive consent %q", forbidden),
				templateSource(projectRoot),
				"remove the extra merge or additive-infrastructure confirmation pause; keep MERGE_READY evidence and exact confirmation only for deletes, removals, or otherwise destructive effects",
				[]string{fmt.Sprintf("rg -n %q %s", forbidden, absolutePath)},
			))
			break
		}
	}
	return findings
}

type nondestructiveAuthorityCheck struct {
	path      string
	required  []string
	forbidden []string
}

func nondestructiveAuthorityChecks() []nondestructiveAuthorityCheck {
	return []nondestructiveAuthorityCheck{
		{
			path: "docs/references/rules/github-pr-merge.md",
			required: []string{
				"accepted task or active `/goal`",
				"Do not stop for a separate merge-consent prompt",
				"A changed head invalidates readiness, not accepted-task authority",
			},
			forbidden: []string{
				"Merge only after a direct user request or accepted bounded merge plan names the exact authorized PR set",
				"The rule applies only when one of these creates merge authority",
			},
		},
		{
			path: "docs/references/rules/infrastructure-change-approval.md",
			required: []string{
				"Proceed autonomously when the graph contains only additive or rollback-preserving effects",
				"always requires explicit user confirmation after the consolidated outline",
			},
			forbidden: []string{
				"Obtain one explicit user confirmation of the complete outline before editing covered infrastructure source or performing a live mutation.",
				"Obtain one explicit user confirmation for the complete bounded batch",
			},
		},
	}
}

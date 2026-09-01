package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNondestructiveAuthorityPolicyLocksRequiredScenarios(t *testing.T) {
	merge := readRepositoryFile(t, "docs/references/rules/github-pr-merge.md")
	infra := readRepositoryFile(t, "docs/references/rules/infrastructure-change-approval.md")
	safety := readRepositoryFile(t, "docs/references/rules/safety-guardrails.md")
	combined := merge + "\n" + infra + "\n" + safety

	required := []string{
		"Accepted task/goal includes owner/service#84.",
		"Merge #84 without another consent prompt",
		"A changed head invalidates readiness, not accepted-task authority",
		"Additive create proceeds autonomously",
		"Provider replacement containing a destroy",
		"Proceed with this complete deletion batch?",
		"An explicit user hold such as \"do not merge\" prevails",
		"Unresolved destructive-effect classification fails closed",
		"Additive IAM, network topology",
		"keep production default-off",
		"are not infrastructure-approval batches",
		"never passing",
	}
	for _, check := range required {
		if !strings.Contains(combined, check) {
			t.Errorf("policy missing required scenario %q", check)
		}
	}
}

func TestAuditNondestructiveAuthorityPolicyFindsSupersededConsent(t *testing.T) {
	projectRoot := t.TempDir()
	path := filepath.Join(projectRoot, "docs", "references", "rules", "github-pr-merge.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	stale := "# stale\nThe rule applies only when one of these creates merge authority\n"
	if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}
	findings := auditNondestructiveAuthorityPolicy(projectRoot)
	if len(findings) == 0 {
		t.Fatal("expected a superseded-consent finding")
	}
}

func TestAuditNondestructiveAuthorityPolicyAcceptsCurrentRules(t *testing.T) {
	projectRoot := filepath.Join("..", "..")
	if findings := auditNondestructiveAuthorityPolicy(projectRoot); len(findings) != 0 {
		t.Fatalf("current rules produced policy findings: %#v", findings)
	}
}

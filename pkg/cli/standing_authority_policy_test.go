package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStandingAuthorityPolicyLocksRequiredScenarios(t *testing.T) {
	merge := readRepositoryFile(t, "docs/references/rules/github-pr-merge.md")
	infra := readRepositoryFile(t, "docs/references/rules/infrastructure-change-approval.md")
	safety := readRepositoryFile(t, "docs/references/rules/safety-guardrails.md")
	combined := merge + "\n" + infra + "\n" + safety
	combined = strings.Join(strings.Fields(combined), " ")

	required := []string{
		"Later in-scope blocker PR:",
		"Do not ask again",
		"A changed in-scope head invalidates readiness, not standing",
		"### Standard Deployments Under Standing Authority",
		"provider plan now replaces the database",
		"Proceed with this complete deletion batch?",
		"Pause and revocation:",
		"Only an explicit human resume or new grant restores it",
		"IAM, network, KMS, secrets, database-schema or data-loss changes",
		"infrastructure creation, replacement, or deletion",
		"pending or missing expected checks",
		"skipped checks without verified policy eligibility",
		"Protection bypass, admin override, review bypass, required-check bypass",
		"Identity failure blocks only that repository node and its dependents",
		"Adding or changing repository, target base, environment, actor, identity, merge method, deployment workflow, product scope, or material effect",
		"are not infrastructure-approval batches",
		"never passing",
	}
	for _, check := range required {
		if !strings.Contains(combined, check) {
			t.Errorf("policy missing required scenario %q", check)
		}
	}
}

func TestAuditStandingAuthorityPolicyFindsSupersededGuidance(t *testing.T) {
	projectRoot := t.TempDir()
	path := filepath.Join(projectRoot, "docs", "references", "rules", "github-pr-merge.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	stale := "# stale\naccepted task or active `/goal` authorizes every merge\n"
	if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}
	findings := auditStandingAuthorityPolicy(projectRoot)
	if len(findings) == 0 {
		t.Fatal("expected a superseded-consent finding")
	}
}

func TestAuditStandingAuthorityPolicyAcceptsCurrentRules(t *testing.T) {
	projectRoot := filepath.Join("..", "..")
	if findings := auditStandingAuthorityPolicy(projectRoot); len(findings) != 0 {
		t.Fatalf("current rules produced policy findings: %#v", findings)
	}
}

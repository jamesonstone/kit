package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditStandingAuthorityPolicyRejectsExactHeadReauthorization(t *testing.T) {
	projectRoot := copyStandingAuthorityPolicies(t)
	path := filepath.Join(projectRoot, "docs", "references", "rules", "delivery.md")
	stale := "After that check passes, the refreshed head needs exact-head authorization, deployment, and one browser retry.\n"
	if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}
	assertStandingAuthorityFinding(t, auditStandingAuthorityPolicy(projectRoot), path, "superseded standing-authority guidance")
}

func TestExactHeadReauthorizationPhrasesAreNormalized(t *testing.T) {
	body := normalizeStandingAuthorityPolicy("A changed head loses prior readiness AND merge authority.")
	phrase := normalizeStandingAuthorityPolicy("changed head loses prior readiness and merge authority")
	if !strings.Contains(body, phrase) {
		t.Fatal("normalized exact-head reauthorization phrase did not match")
	}
}

func TestAuditStandingAuthorityPolicyFindsSupersededGuidance(t *testing.T) {
	projectRoot := copyStandingAuthorityPolicies(t)
	path := filepath.Join(projectRoot, "docs", "references", "rules", "github-pr-merge.md")
	stale := "# stale\naccepted task or active `/goal` authorizes every merge\n"
	if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}
	assertStandingAuthorityFinding(t, auditStandingAuthorityPolicy(projectRoot), path, "superseded standing-authority guidance")
}

func TestAuditStandingAuthorityPolicyAcceptsCurrentRules(t *testing.T) {
	projectRoot := filepath.Join("..", "..")
	if findings := auditStandingAuthorityPolicy(projectRoot); len(findings) != 0 {
		t.Fatalf("current rules produced policy findings: %#v", findings)
	}
}

func TestAuditStandingAuthorityPolicyWarnsWhenPolicyUnreadable(t *testing.T) {
	projectRoot := copyStandingAuthorityPolicies(t)
	path := filepath.Join(projectRoot, "docs", "references", "rules", "github-pr-merge.md")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	assertStandingAuthorityFinding(t, auditStandingAuthorityPolicy(projectRoot), path, "failed to read standing-authority policy document")
}

func copyStandingAuthorityPolicies(t *testing.T) string {
	t.Helper()
	projectRoot := t.TempDir()
	writeStandingAuthorityPolicies(t, projectRoot)
	return projectRoot
}

func writeStandingAuthorityPolicies(t *testing.T, projectRoot string) {
	t.Helper()
	for _, check := range standingAuthorityChecks() {
		path := filepath.Join(projectRoot, filepath.FromSlash(check.path))
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(readRepositoryFile(t, check.path)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func assertStandingAuthorityFinding(t *testing.T, findings []reconcileFinding, path, issue string) {
	t.Helper()
	for _, finding := range findings {
		if finding.FilePath == path && strings.Contains(finding.Issue, issue) && finding.Severity == reconcileSeverityWarning {
			return
		}
	}
	t.Fatalf("missing warning for %s containing %q: %#v", path, issue, findings)
}

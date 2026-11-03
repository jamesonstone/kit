package cli

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestDeliveryEvidenceCopiesMatchSources(t *testing.T) {
	if got := readRepositoryFile(t, pullRequestTemplatePath); got != templates.PullRequestTemplate {
		t.Fatal("checked-in pull request template drifted from the scaffold source")
	}
	if got := readRepositoryFile(t, issueTemplatePath); got != templates.IssueTemplate {
		t.Fatal("checked-in issue template drifted from the scaffold source")
	}
	for _, want := range []string{"## Verified when", "before an agent starts", "- Done:", "- Out of scope:", "- Evidence: passing test, screenshot, log, or trace", "- Hypothesis (optional):"} {
		if !strings.Contains(templates.IssueTemplate, want) {
			t.Errorf("issue template missing %q", want)
		}
	}
	for _, want := range []string{"## Evidence", "- Commands: command and real output or excerpt", "- Decisions and assumptions: defaults, scope calls, conflicting requirements reinterpreted", "- Checks: for each check, the property it verified", "- Not checked:", "- Fresh-context review:"} {
		if !strings.Contains(templates.PullRequestTemplate, want) {
			t.Errorf("pull request template missing %q", want)
		}
	}
	linear := templates.BuildPullRequestTemplate("linear")
	if !strings.Contains(linear, "## Evidence") || strings.Contains(linear, "Closes #") || !strings.Contains(linear, "[Linear issue URL]") {
		t.Fatalf("linear pull request template = %q", linear)
	}
}

func TestReconcileIssueTemplateCreateAndPreserve(t *testing.T) {
	root := t.TempDir()
	created, err := planRefreshInitScaffoldFile(root, initRefreshOptions{}, nil, issueTemplatePath, templates.IssueTemplate, false, false)
	if err != nil || created.result != instructionFileCreated || created.after != templates.IssueTemplate {
		t.Fatalf("create %#v %v", created, err)
	}
	custom := templates.IssueTemplate + "\nCustom\n"
	writeFile(t, root+"/"+issueTemplatePath, custom)
	kept, err := planRefreshInitScaffoldFile(root, initRefreshOptions{}, nil, issueTemplatePath, templates.IssueTemplate, false, false)
	if err != nil || kept.result != instructionFileSkipped || kept.after != custom {
		t.Fatalf("preserve %#v %v", kept, err)
	}
}

func TestDeliveryRegistryHashMatchesRule(t *testing.T) {
	content := readRepositoryFile(t, "docs/references/rules/delivery.md")
	hash, err := normalizedRulesetContentHash(content, "active")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	artifact, ok := cfg.RegistryArtifact(rulesetKind, "delivery")
	if !ok || artifact.InstalledHash != hash {
		t.Fatalf("delivery installed hash = %q, content hash = %q", artifact.InstalledHash, hash)
	}
}

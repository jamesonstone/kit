package cli

import (
	"bytes"
	"os"
	"path/filepath"
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
	if !strings.HasPrefix(issueTemplatePath, ".github/ISSUE_TEMPLATE/") || strings.HasSuffix(issueTemplatePath, "ISSUE_TEMPLATE.md") {
		t.Fatalf("issue template path %s uses the retired single-file location", issueTemplatePath)
	}
	if _, err := os.Stat(filepath.Join("..", "..", ".github", "ISSUE_TEMPLATE.md")); !os.IsNotExist(err) {
		t.Fatal("retired .github/ISSUE_TEMPLATE.md must not exist")
	}
	for _, want := range []string{"name: Issue", "about:", "## Verified when", "before an agent starts", "- Done:", "- Out of scope:", "- Evidence: passing test, screenshot, log, or trace", "- Hypothesis or expected answer (optional):"} {
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
	skipped, err := planRefreshInitScaffoldFile(root, initRefreshOptions{}, nil, issueTemplatePath, templates.IssueTemplate, false, false)
	if err != nil || skipped.result != instructionFileSkipped || skipped.after != "" {
		t.Fatalf("opt-in skip %#v %v", skipped, err)
	}
	created, err := planRefreshInitScaffoldFile(root, initRefreshOptions{issueTemplate: true}, nil, issueTemplatePath, templates.IssueTemplate, false, false)
	if err != nil || created.result != instructionFileCreated || created.after != templates.IssueTemplate {
		t.Fatalf("create %#v %v", created, err)
	}
	writeFile(t, root+"/"+issueTemplatePath, templates.IssueTemplate)
	shipped, err := planRefreshInitScaffoldFile(root, initRefreshOptions{updateShippedIssueTemplate: true}, nil, issueTemplatePath, templates.IssueTemplate+"\n", false, false)
	if err != nil || shipped.result != instructionFileUpdated || shipped.after != templates.IssueTemplate+"\n" {
		t.Fatalf("shipped update %#v %v", shipped, err)
	}
	custom := templates.IssueTemplate + "\nCustom\n"
	writeFile(t, root+"/"+issueTemplatePath, custom)
	kept, err := planRefreshInitScaffoldFile(root, initRefreshOptions{issueTemplate: true, updateShippedIssueTemplate: true}, nil, issueTemplatePath, templates.IssueTemplate, false, false)
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

func TestSkippedIssueTemplateDoesNotWarnAboutOrganization(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	writeFile(t, filepath.Join(root, ".github", "ISSUE_TEMPLATE", "config.yml"), "blank_issues_enabled: false\n")
	previous := readmeCommandRunner
	t.Cleanup(func() { readmeCommandRunner = previous })
	readmeCommandRunner = prewriteCommandRunner{}
	var output string
	withInitFlags(t, func() {
		initPlanDryRun = true
		initIssueTemplate = true
		output = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	if strings.Contains(output, "organization") || strings.Contains(output, "cannot detect") {
		t.Fatalf("skipped template write warned about organization templates:\n%s", output)
	}
	if !strings.Contains(output, "skip .github/ISSUE_TEMPLATE/issue.md") {
		t.Fatal(output)
	}
}

func TestSuppressedIssueTemplatePlanIsReported(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	cfg := defaultInitConfig()
	enabled := true
	cfg.Issues.IssueTemplate = &enabled
	if err := config.Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, ".github", "ISSUE_TEMPLATE", "config.yml"), "blank_issues_enabled: false\n")
	var out bytes.Buffer
	reported, err := reportSuppressedIssueTemplate(&out, root, initRefreshOptions{issueTemplate: true, updateShippedIssueTemplate: true})
	if err != nil || !reported {
		t.Fatalf("reported=%v err=%v\n%s", reported, err, out.String())
	}
	for _, want := range []string{
		"skip .github/ISSUE_TEMPLATE/issue.md",
		"existing issue template config: .github/ISSUE_TEMPLATE/config.yml",
		"No files written.",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q\n%s", want, out.String())
		}
	}
}

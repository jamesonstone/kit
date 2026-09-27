package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/document"
)

func TestRepositoryMemoryWorkflowFixturesCoverMaterialAndCodeSufficientOutcomes(t *testing.T) {
	materialPath := filepath.Join("testdata", "repository-memory", "material-why", "SPEC.md")
	material, err := os.ReadFile(materialPath)
	if err != nil {
		t.Fatal(err)
	}
	doc := document.Parse(string(material), materialPath, document.TypeSpec)
	if errors := doc.Validate(); len(errors) != 0 {
		t.Fatalf("material-memory fixture validation errors = %#v", errors)
	}

	codeSufficient, err := os.ReadFile(filepath.Join("testdata", "repository-memory", "code-sufficient", "EXPECTED_FINAL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Repository Memory", "Decision: not required", "Artifacts: none"} {
		if !strings.Contains(string(codeSufficient), want) {
			t.Fatalf("code-sufficient fixture missing %q", want)
		}
	}
}

func TestBuildAutoAssignWorkflowRendersSafeGitHubActionsWorkflow(t *testing.T) {
	content := BuildAutoAssignWorkflow([]string{"jamesonstone", "octocat"})

	for _, check := range []string{
		"# Kit-managed auto-assignment workflow.",
		"pull_request_target:",
		"issues: write",
		"pull-requests: read",
		"continue-on-error: true",
		`"jamesonstone"`,
		`"octocat"`,
		"const configured = [",
		"context.payload?.issue?.user?.login",
		"context.payload?.pull_request?.user?.login",
		`key.endsWith("[bot]")`,
		"login.toLowerCase()",
		"includes initiator",
		"github.rest.issues.addAssignees",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected workflow to contain %q, got:\n%s", check, content)
		}
	}
	if strings.Contains(content, "actions/checkout") {
		t.Fatalf("auto-assign workflow must not check out pull request code:\n%s", content)
	}
}

func TestBuildAutoAssignWorkflowNoOpsWithoutAssignees(t *testing.T) {
	content := BuildAutoAssignWorkflow(nil)

	for _, check := range []string{
		"const configured = [];",
		"context.payload?.issue?.user?.login",
		"No assignees resolved (no configured maintainers and no human initiator); skipping.",
		"continue-on-error: true",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected empty-assignee workflow to contain %q, got:\n%s", check, content)
		}
	}
}

func TestMakefileTemplateProvidesSafeStarter(t *testing.T) {
	for _, check := range []string{
		".DEFAULT_GOAL := help",
		".PHONY: help",
		"help:",
		"Run the Kit initialization prompt to add project-specific targets.",
	} {
		if !strings.Contains(Makefile, check) {
			t.Fatalf("expected Makefile template to contain %q, got:\n%s", check, Makefile)
		}
	}

	for _, unverified := range []string{"TODO", "dev:", "npm ", "go run ", "docker compose"} {
		if strings.Contains(Makefile, unverified) {
			t.Fatalf("Makefile template must not contain unverified command %q:\n%s", unverified, Makefile)
		}
	}
}

func TestBrainstormTemplateUsesReferenceProseSection(t *testing.T) {
	checks := []string{
		"## DEPENDENCIES",
		"References are tracked in front matter.",
	}

	for _, check := range checks {
		if !strings.Contains(BrainstormArtifact, check) {
			t.Fatalf("expected BrainstormArtifact to contain %q", check)
		}
	}
}

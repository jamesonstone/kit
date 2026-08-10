package cli

import (
	"os"
	"path/filepath"
	stdreflect "reflect"
	"testing"

	contextcontract "github.com/jamesonstone/kit/internal/context"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestRunInitMaterializesEveryContextWorkflow(t *testing.T) {
	root := initializeContextWorkflowProject(t)
	artifacts, err := templates.ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts) != 5 {
		t.Fatalf("workflow count = %d, want 5", len(artifacts))
	}
	for _, artifact := range artifacts {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(artifact.Path)))
		if err != nil {
			t.Fatalf("read %s: %v", artifact.Path, err)
		}
		if string(content) != artifact.Content {
			t.Fatalf("materialized workflow %s differs from embedded artifact", artifact.Path)
		}
	}
}

func TestRunInitPreservesCustomizedContextWorkflow(t *testing.T) {
	root := initializeContextWorkflowProject(t)
	path := filepath.Join(root, "docs/references/workflows/implementation-delivery.md")
	custom := "# Project-owned workflow customization\n"
	if err := os.WriteFile(path, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("repeated runInit() error = %v", err)
			}
		})
	})
	if got := readFile(t, path); got != custom {
		t.Fatalf("customized workflow was overwritten:\n%s", got)
	}
}

func TestRunInitRefreshDryRunDoesNotRestoreMissingWorkflow(t *testing.T) {
	root := initializeContextWorkflowProject(t)
	path := filepath.Join(root, "docs/references/workflows/repository-maintenance.md")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	before := contextCLISnapshot(t, root)

	withInitFlags(t, func() {
		initRefresh = true
		initDryRun = true
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("dry-run refresh error = %v", err)
			}
		})
	})
	after := contextCLISnapshot(t, root)
	if !stdreflect.DeepEqual(before, after) {
		t.Fatalf("dry-run changed files:\nbefore=%#v\nafter=%#v", before, after)
	}
}

func TestFreshInitResolvesRepositoryBootstrapContract(t *testing.T) {
	contextRule := readRepositoryFile(t, "docs/references/rules/coding-agent-context-usage.md")
	constitutionRule := readRepositoryFile(t, "docs/references/rules/constitution-curation.md")
	capabilitiesRule := readRepositoryFile(t, "docs/references/rules/kit-capabilities-usage.md")
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	stubRulesetRegistry(t,
		registryRulesetWithContentForTest("coding-agent-context-usage", contextRule, "test-context"),
		registryRulesetWithContentForTest("constitution-curation", constitutionRule, "test-constitution"),
		registryRulesetWithContentForTest("kit-capabilities-usage", capabilitiesRule, "test-capabilities"),
	)
	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})
	contract := contextcontract.Resolve(root, contextcontract.Request{Workflow: "repository-bootstrap"})
	if contract.Blocked {
		t.Fatalf("repository-bootstrap contract blocked: %#v", contract.Diagnostics)
	}
	for _, path := range []string{
		"docs/references/workflows/repository-bootstrap.md",
		"docs/references/rules/coding-agent-context-usage.md",
		"docs/references/rules/constitution-curation.md",
		"docs/references/rules/kit-capabilities-usage.md",
	} {
		if !contractEvidencePathPresent(contract, path) {
			t.Errorf("bootstrap contract missing %s", path)
		}
	}
}

func initializeContextWorkflowProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})
	return root
}

func contractEvidencePathPresent(contract contextcontract.Contract, path string) bool {
	for _, evidence := range contract.Evidence {
		if evidence.Path == path {
			return true
		}
	}
	return false
}

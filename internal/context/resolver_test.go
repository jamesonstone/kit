package context

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestResolveIsDeterministicAndOrdersWorkflowDependencies(t *testing.T) {
	root := contextProject(t)
	writeContextFile(t, root, "docs/references/workflows/base.md", workflowDocument("base", nil, []string{"base-rule"}, nil))
	writeContextFile(t, root, "docs/references/workflows/main.md", workflowDocument("main", []string{"base"}, []string{"main-rule"}, []string{"docs/agents/README.md"}))
	writeContextFile(t, root, "docs/references/rules/base-rule.md", rulesetDocument("base-rule"))
	writeContextFile(t, root, "docs/references/rules/main-rule.md", rulesetDocument("main-rule"))
	writeContextFile(t, root, "docs/agents/README.md", "# Agents\n")
	writeContextFile(t, root, "z.txt", "z\n")
	writeContextFile(t, root, "a.txt", "a\n")

	request := Request{Workflow: "main", Paths: []string{"z.txt", "a.txt"}}
	first := Resolve(root, request)
	second := Resolve(root, request)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("resolution is not deterministic:\n%#v\n%#v", first, second)
	}
	if first.Blocked {
		t.Fatalf("resolution unexpectedly blocked: %#v", first.Diagnostics)
	}
	if len(first.Workflows) != 2 || first.Workflows[0].Slug != "base" || first.Workflows[1].Slug != "main" {
		t.Fatalf("workflow order = %#v", first.Workflows)
	}
	if first.Request.Paths[0] != "a.txt" || first.Request.Paths[1] != "z.txt" {
		t.Fatalf("path hints not normalized: %#v", first.Request.Paths)
	}
	for _, item := range first.Evidence {
		if item.State == "present" && !strings.HasPrefix(item.Digest, "sha256:") {
			t.Fatalf("missing digest for %#v", item)
		}
		if filepath.IsAbs(item.Path) {
			t.Fatalf("absolute evidence path leaked: %s", item.Path)
		}
	}
}

func TestResolveBlocksMissingRequiredButNotOptionalEvidence(t *testing.T) {
	root := contextProject(t)
	document := `---
kind: workflow
slug: check
description: test
rules: []
evidence:
  - kind: required
    path: required.md
    required: true
  - kind: optional
    path: optional.md
    required: false
---
# Check
`
	writeContextFile(t, root, "docs/references/workflows/check.md", document)
	contract := Resolve(root, Request{Workflow: "check"})
	if !contract.Blocked {
		t.Fatalf("missing required evidence did not block: %#v", contract)
	}
	levels := map[string]string{}
	for _, diagnostic := range contract.Diagnostics {
		levels[diagnostic.Path] = diagnostic.Level
	}
	if levels["required.md"] != "error" || levels["optional.md"] != "warning" {
		t.Fatalf("diagnostic levels = %#v", levels)
	}
}

func TestResolveRejectsEvidenceSymlinkEscapingProject(t *testing.T) {
	root := contextProject(t)
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	writeContextFile(t, root, "docs/references/workflows/check.md", workflowDocument("check", nil, nil, []string{"escape.md"}))
	contract := Resolve(root, Request{Workflow: "check"})
	if !contract.Blocked {
		t.Fatal("escaping symlink did not block")
	}
	if !diagnosticCodePresent(contract, "invalid-evidence-path") {
		t.Fatalf("missing path diagnostic: %#v", contract.Diagnostics)
	}
}

func TestResolveRejectsWorkflowSymlinkEscapingProject(t *testing.T) {
	root := contextProject(t)
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte(workflowDocument("check", nil, nil, nil)), 0o644); err != nil {
		t.Fatal(err)
	}
	workflowDir := filepath.Join(root, "docs", "references", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(workflowDir, "check.md")); err != nil {
		t.Fatal(err)
	}
	contract := Resolve(root, Request{Workflow: "check"})
	if !contract.Blocked || len(contract.Workflows) != 0 {
		t.Fatalf("escaping workflow was selected: %#v", contract)
	}
	if !diagnosticCodePresent(contract, "invalid-workflow") || !diagnosticCodePresent(contract, "invalid-evidence-path") {
		t.Fatalf("missing workflow confinement diagnostics: %#v", contract.Diagnostics)
	}
}

func TestResolveFeatureIncludesRelationshipsAndLocalReferences(t *testing.T) {
	root := contextProject(t)
	writeContextFile(t, root, "docs/references/workflows/check.md", workflowDocument("check", nil, nil, nil))
	writeContextFile(t, root, "docs/specs/0002-history/SPEC.md", v3Spec("0002", "history", "0002-history", "", ""))
	relationships := "relationships:\n  - type: builds_on\n    target: 0002-history"
	references := `references:
  - id: local
    name: Local rule
    type: rule
    target: docs/references/rules/local.md
    relation: guides
    read_policy: must
    used_for: test
    status: active`
	writeContextFile(t, root, "docs/specs/0001-alpha/SPEC.md", v3Spec("0001", "alpha", "0001-alpha", relationships, references))
	writeContextFile(t, root, "docs/references/rules/local.md", rulesetDocument("local"))
	contract := Resolve(root, Request{Workflow: "check", Feature: "alpha"})
	if contract.Blocked {
		t.Fatalf("feature resolution blocked: %#v", contract.Diagnostics)
	}
	for _, path := range []string{"docs/specs/0001-alpha/SPEC.md", "docs/specs/0002-history/SPEC.md", "docs/references/rules/local.md"} {
		if !evidencePathPresent(contract, path) {
			t.Fatalf("missing feature evidence %s: %#v", path, contract.Evidence)
		}
	}
}

func contextProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := config.Save(root, config.Default()); err != nil {
		t.Fatal(err)
	}
	return root
}

func writeContextFile(t *testing.T, root, relativePath, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func diagnosticCodePresent(contract Contract, code string) bool {
	for _, diagnostic := range contract.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func evidencePathPresent(contract Contract, path string) bool {
	for _, evidence := range contract.Evidence {
		if evidence.Path == path {
			return true
		}
	}
	return false
}

package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/registry"
)

func TestResolveGoldenIncludesHintsDependenciesAndCustomization(t *testing.T) {
	root := writeContractProject(t)
	resolved, err := Resolve(root, Hints{
		Paths: []string{"internal/deep/agent.go"}, Workflows: []string{"delivery"},
	})
	if err != nil {
		t.Fatal(err)
	}
	resolved.ProjectRoot = "<project>"
	normalizeGoldenDigests(&resolved)
	content, err := json.MarshalIndent(resolved, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, '\n')
	want, err := os.ReadFile("testdata/resolved.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != string(want) {
		t.Fatalf("resolved contract differs from golden\n--- got\n%s\n--- want\n%s", content, want)
	}

	again, err := Resolve(root, Hints{Workflows: []string{"delivery"}, Paths: []string{"internal/deep/agent.go"}})
	if err != nil {
		t.Fatal(err)
	}
	again.ProjectRoot = "<project>"
	normalizeGoldenDigests(&again)
	againJSON, _ := json.MarshalIndent(again, "", "  ")
	if string(content[:len(content)-1]) != string(againJSON) {
		t.Fatal("equivalent hints did not resolve deterministically")
	}
}

func TestResolveBlocksMissingRequiredArtifactsAndUnknownWorkflow(t *testing.T) {
	root := writeContractProject(t)
	if err := os.Remove(filepath.Join(root, "docs/references/rules/safety.md")); err != nil {
		t.Fatal(err)
	}
	resolved, err := Resolve(root, Hints{Workflows: []string{"absent"}})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.State != "blocked" {
		t.Fatalf("state = %s", resolved.State)
	}
	if len(resolved.Diagnostics) != 2 || len(resolved.NextActions) != 1 {
		t.Fatalf("diagnostics = %v, next actions = %v", resolved.Diagnostics, resolved.NextActions)
	}
}

func TestResolveBlocksSelectedConditionalArtifact(t *testing.T) {
	root := writeContractProject(t)
	if err := os.Remove(filepath.Join(root, "docs/references/rules/go-paths.md")); err != nil {
		t.Fatal(err)
	}
	resolved, err := Resolve(root, Hints{Paths: []string{"internal/deep/agent.go"}})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.State != "blocked" || len(resolved.Rules["conditional"]) != 1 {
		t.Fatalf("resolved state = %s, conditional rules = %#v", resolved.State, resolved.Rules["conditional"])
	}
}

func TestResolveBlocksSchemaV1WithoutNetworkOrWrites(t *testing.T) {
	root := t.TempDir()
	writeContractFile(t, filepath.Join(root, registry.ProjectFile), "schema_version: 1\n")
	before, _ := os.ReadFile(filepath.Join(root, registry.ProjectFile))
	resolved, err := Resolve(root, Hints{})
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(root, registry.ProjectFile))
	if resolved.State != "blocked" || string(before) != string(after) {
		t.Fatalf("resolution state = %s or project was mutated", resolved.State)
	}
}

func TestResolvePRFeedbackWorkflowIncludesTeamAndLaneDependencies(t *testing.T) {
	root := writeContractProject(t)
	resolved, err := Resolve(root, Hints{Workflows: []string{"pr-feedback-repair"}})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.State != "ready" || !hasArtifact(resolved.Workflows, "pr-feedback-repair") {
		t.Fatalf("resolved state = %s, workflows = %#v", resolved.State, resolved.Workflows)
	}
	for _, slug := range []string{"agent-team", "github-delivery", "work-lane"} {
		if !hasArtifact(resolved.Rules["conditional"], slug) {
			t.Fatalf("dependency %q was not selected: %#v", slug, resolved.Rules["conditional"])
		}
	}
}

func writeContractProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	artifacts := []registry.ArtifactRecord{
		contractRecord("ruleset", "agent-team", "docs/references/rules/agent-team.md", "conditional", []string{"subagent"}, nil, nil),
		contractRecord("ruleset", "base", "docs/references/rules/base.md", "conditional", []string{"core"}, nil, nil),
		contractRecord("ruleset", "feature-specification", "docs/references/rules/feature-specification.md", "conditional", []string{"feature"}, nil, nil),
		contractRecord("ruleset", "github-delivery", "docs/references/rules/github-delivery.md", "conditional", []string{"github"}, nil, nil),
		contractRecord("ruleset", "go-paths", "docs/references/rules/go-paths.md", "conditional", []string{"go"}, []string{"**/*.go"}, nil),
		contractRecord("ruleset", "safety", "docs/references/rules/safety.md", "must", []string{"safety"}, nil, nil),
		contractRecord("ruleset", "work-lane", "docs/references/rules/work-lane.md", "conditional", []string{"lane"}, nil, nil),
		contractRecord("workflow", "delivery", "docs/references/workflows/delivery.md", "conditional", []string{"delivery"}, nil,
			[]string{"ruleset/base", "ruleset/feature-specification"}),
		contractRecord("workflow", "implementation-delivery", "docs/references/workflows/implementation-delivery.md", "conditional", []string{"implementation"}, nil,
			[]string{"ruleset/base", "ruleset/feature-specification"}),
		contractRecord("workflow", "pr-feedback-repair", "docs/references/workflows/pr-feedback-repair.md", "conditional", []string{"review"}, nil,
			[]string{"ruleset/agent-team", "ruleset/github-delivery", "ruleset/work-lane"}),
	}
	for index := range artifacts {
		content := contractDocument(artifacts[index])
		if artifacts[index].Slug == "base" {
			artifacts[index].ContentHash = registry.HashContent(content)
			content += "\n## Local Addition\n\nProject customization.\n"
		} else {
			artifacts[index].ContentHash = registry.HashContent(content)
		}
		writeContractFile(t, filepath.Join(root, filepath.FromSlash(artifacts[index].Path)), content)
	}
	cfg := registry.NewProjectConfig(registry.SourceConfig{Path: ".", CatalogPath: "registry/catalog.yaml"})
	cfg.Registry.Artifacts = artifacts
	content, err := registry.MarshalProject(cfg)
	if err != nil {
		t.Fatal(err)
	}
	writeContractFile(t, filepath.Join(root, registry.ProjectFile), string(content))
	writeContractFile(t, filepath.Join(root, "AGENTS.md"), "# Agents\n")
	return root
}

func contractRecord(kind, slug, path, policy string, applies, paths, dependencies []string) registry.ArtifactRecord {
	return registry.ArtifactRecord{
		Kind: kind, Slug: slug, Description: slug + " contract", Path: path, Version: 1,
		ReadPolicy: policy, AppliesTo: applies, Paths: paths, Dependencies: dependencies,
		State: registry.StateManaged,
	}
}

func normalizeGoldenDigests(resolved *Resolved) {
	for index := range resolved.Workflows {
		resolved.Workflows[index].Digest = "DIGEST_" + strings.ToUpper(strings.ReplaceAll(resolved.Workflows[index].Slug, "-", "_"))
	}
	for policy := range resolved.Rules {
		for index := range resolved.Rules[policy] {
			slug := resolved.Rules[policy][index].Slug
			if slug == "base" {
				slug = "base-custom"
			}
			resolved.Rules[policy][index].Digest = "DIGEST_" + strings.ToUpper(strings.ReplaceAll(slug, "-", "_"))
		}
	}
}

func contractDocument(record registry.ArtifactRecord) string {
	content := "---\nkind: " + record.Kind + "\nslug: " + record.Slug + "\ndescription: " + record.Description + "\nstatus: active\nregistry_scope: downstream\nread_policy_default: " + record.ReadPolicy + "\n"
	if len(record.Dependencies) > 0 {
		content += "dependencies:\n"
		for _, dependency := range record.Dependencies {
			content += "  - " + dependency + "\n"
		}
	}
	return content + "---\n\n# Contract\n\n## Purpose\n\nFollow this contract.\n"
}

func writeContractFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

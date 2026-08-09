package agentcli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

func TestRootExposesOnlyMajorReleaseSurface(t *testing.T) {
	root := NewRoot()
	assertCommandNames(t, root, []string{"contract", "init", "reconcile", "registry", "rules"})
	assertCommandNames(t, commandNamed(t, root, "contract"), []string{"resolve"})
	assertCommandNames(t, commandNamed(t, root, "rules"), []string{"add", "list", "view"})
	assertCommandNames(t, commandNamed(t, root, "registry"), []string{"add", "list", "status", "view"})

	removed := []string{
		"aws", "capabilities", "check", "config", "dispatch", "health", "improve", "instructions",
		"legacy", "loop", "map", "plan", "pr", "prompt", "replay", "review", "scaffold", "skill",
		"spec", "state", "status", "verify",
	}
	for _, name := range removed {
		if command := findDirectCommand(root, name); command != nil {
			t.Errorf("removed command %q is still exposed", name)
		}
	}
	if findDirectCommand(root, "rule") != nil {
		t.Error("legacy singular rules alias is still exposed")
	}
}

func TestReadmeDocumentsMajorReleaseAndPrimaryFlow(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readme := string(content)
	required := []string{
		"Major update", "kit init", "kit contract resolve", "agent implementation", "kit reconcile",
		"docs/migrations/coding-agent-first-major.md",
	}
	for _, phrase := range required {
		if !strings.Contains(readme, phrase) {
			t.Errorf("README is missing %q", phrase)
		}
	}
	for _, removed := range []string{"kit spec", "kit loop", "kit health", "kit capabilities"} {
		if strings.Contains(readme, removed) {
			t.Errorf("README still advertises removed surface %q", removed)
		}
	}
	if lines := strings.Count(readme, "\n") + 1; lines > 150 {
		t.Errorf("README has %d lines; major-release overview should stay concise", lines)
	}
}

func TestInitIsIdempotentButDefersDriftToReconcile(t *testing.T) {
	registryRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(registryRoot, "registry"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(registryRoot, "registry/catalog.yaml"), []byte("schema_version: 1\nartifacts: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	sourceCfg := registry.SourceConfig{Path: registryRoot, CatalogPath: "registry/catalog.yaml"}
	plan, err := registry.BuildInitPlan(context.Background(), project, registry.LocalSource{Root: registryRoot}, sourceCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.ApplyPlan(project, plan); err != nil {
		t.Fatal(err)
	}

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(project); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	command := NewRoot()
	output := &bytes.Buffer{}
	command.SetOut(output)
	command.SetArgs([]string{"init", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("repeat init: %v", err)
	}
	if !strings.Contains(output.String(), `"state": "current"`) {
		t.Fatalf("repeat init output = %s", output)
	}

	agentsPath := filepath.Join(project, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("project-owned drift\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	command = NewRoot()
	command.SetArgs([]string{"init"})
	err = command.Execute()
	if err == nil || !strings.Contains(err.Error(), "use `kit reconcile`") {
		t.Fatalf("drift init error = %v", err)
	}
	content, _ := os.ReadFile(agentsPath)
	if string(content) != "project-owned drift\n" {
		t.Fatal("repeat init changed a drifted project")
	}
}

func assertCommandNames(t *testing.T, command interface{ Commands() []*cobra.Command }, want []string) {
	t.Helper()
	var got []string
	for _, child := range command.Commands() {
		if !child.Hidden {
			got = append(got, child.Name())
		}
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %v, want %v", got, want)
	}
}

func commandNamed(t *testing.T, root *cobra.Command, name string) *cobra.Command {
	t.Helper()
	command := findDirectCommand(root, name)
	if command == nil {
		t.Fatalf("command %q not found", name)
	}
	return command
}

func findDirectCommand(root *cobra.Command, name string) *cobra.Command {
	for _, command := range root.Commands() {
		if command.Name() == name {
			return command
		}
	}
	return nil
}

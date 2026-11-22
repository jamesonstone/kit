package cli

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

// Ledger writes are native agent actions. This pins the shipped instructions.
func TestEmbeddedWorkLedgerHook(t *testing.T) {
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var delivery, cleanup string
	for _, rule := range registry {
		switch rule.Slug {
		case "delivery":
			delivery = rule.Content
		case "work-cleanup":
			cleanup = rule.Content
		}
	}
	for _, marker := range []string{
		"### Work ledger hook",
		"`$XDG_CONFIG_HOME/kit/.kit.yaml`",
		"overrides the user file key by key",
		"A missing user file changes nothing",
		"never committed",
		"`plane.states.in_review`",
		"`plane` is implemented",
		"`linear`, `jira`, and `notion` are reserved",
		"does not call the ledger API",
		"Plane MCP",
		"work_ledger.plane.workspace",
		"work_ledger.plane.project",
		"work_ledger.owner_label",
		"`<workspace>`",
		"`<project>`",
		"The repository owner supplies",
		"only when the repository `.kit.yaml` sets `work_ledger.provider`",
		"configured active state (default `Active`)",
		"configured in-review state (default `In review`)",
		"Do not move that item back to the active state",
		"configured done state (default `Done`)",
		"A failed search is not an empty project",
		"do not by themselves stop the ticket",
	} {
		if !strings.Contains(delivery, marker) {
			t.Errorf("embedded delivery missing %q", marker)
		}
	}
	if !strings.Contains(cleanup, "A configured `work_ledger` follows `delivery` after a satisfying merge and sets Done.") {
		t.Fatal("work-cleanup does not route a configured ledger to Done")
	}
	contract := templates.UniversalContractBlock()
	if strings.Contains(contract, "BEGIN KIT-MANAGED WORK LEDGER") {
		t.Fatal("unconfigured contract emitted a work ledger note")
	}
	generated := []string{delivery, cleanup, contract}
	for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
		data, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		generated = append(generated, string(data))
	}
	needle := "app.plane." + "so"
	for _, bad := range []string{"Jameson", "reasoning-machines", needle} {
		for _, text := range generated {
			if strings.Contains(text, bad) {
				t.Errorf("generated ledger guidance contains %q", bad)
			}
		}
	}
	root := filepath.Join("..", "..")
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "bin", "tmp":
				return filepath.SkipDir
			}
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(data), needle) {
			t.Errorf("%s contains a Plane URL", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPublishedWorkLedgerExamplesUseTheConfigSchema(t *testing.T) {
	for _, name := range []string{"README.md", "docs/references/rules/delivery.md"} {
		content := readRepositoryFile(t, name)
		_, sample, found := strings.Cut(content, "```yaml\nwork_ledger:\n")
		if !found {
			t.Fatalf("%s has no work ledger YAML example", name)
		}
		sample, _, found = strings.Cut(sample, "```")
		if !found {
			t.Fatalf("%s has an unclosed YAML example", name)
		}
		example := strings.NewReplacer("<workspace>", "example-workspace", "<project>", "WL", "<owner-label>", "agent:example").Replace(sample)
		cfg, inspection, err := config.ParseWithInspection([]byte("schema_version: 2\nwork_ledger:\n" + example))
		if err != nil || inspection.HasErrors() || !cfg.WorkLedger.Enabled() || cfg.WorkLedger.Plane.Project != "WL" {
			t.Fatalf("%s example: cfg=%#v inspection=%+v err=%v", name, cfg, inspection, err)
		}
	}
}

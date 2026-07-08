package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func setupMapProject(t *testing.T) string {
	t.Helper()

	projectRoot := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "CONSTITUTION.md"), "# CONSTITUTION\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), "# PROJECT PROGRESS SUMMARY\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "agents", "README.md"), "# Agents Docs\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "references", "README.md"), "# References\n")
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"), `---
kit_metadata_version: 1
artifact: spec
feature:
  id: "0001"
  slug: alpha
  dir: 0001-alpha
references:
  - name: docs/agents/RLM.md
    type: doc
    target: docs/agents/RLM.md
    relation: informs
    read_policy: conditional
    used_for: context routing
    status: active
  - name: old-context.md
    type: doc
    target: docs/old-context.md
    relation: informs
    read_policy: skip
    used_for: legacy context
    status: stale
---
# SPEC

## RELATIONSHIPS

- builds on: `+"`0002-beta`"+`
- follows: 0003-gamma

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| docs/agents/RLM.md | doc | docs/agents/RLM.md | context routing | active |
| old-context.md | doc | docs/old-context.md | legacy context | stale |
`)
	writeMapProjectFile(t, filepath.Join(projectRoot, "docs", "specs", "0002-beta", "SPEC.md"), `# SPEC

## RELATIONSHIPS

none
`)

	return projectRoot
}

func writeMapProjectFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

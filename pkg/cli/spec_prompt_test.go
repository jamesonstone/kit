package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func TestOutputCompiledPrompt_IncludesSkillsDiscoveryInputs(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := filepath.Join(projectRoot, "home")
	codexDir := filepath.Join(homeDir, ".codex")

	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexDir)

	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")
	writeFile(t, filepath.Join(homeDir, ".claude", "CLAUDE.md"), "# global claude\n")
	writeFile(t, filepath.Join(codexDir, "AGENTS.md"), "# global codex agents\n")
	writeFile(t, filepath.Join(codexDir, "instructions.md"), "# global codex instructions\n")

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0009-spec-skills-discovery")
	specPath := filepath.Join(featurePath, "SPEC.md")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")
	writeFile(t, brainstormPath, "# BRAINSTORM\n")
	writeFile(t, specPath, documentTemplateWithSummary())

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()

	output := captureStdout(t, func() {
		err := outputCompiledPrompt(
			specPath,
			brainstormPath,
			"spec-skills-discovery",
			projectRoot,
			cfg,
			&specAnswers{Problem: "skills are undocumented"},
			true,
		)
		if err != nil {
			t.Fatalf("outputCompiledPrompt() error = %v", err)
		}
	})

	checks := []string{
		filepath.Join(projectRoot, "AGENTS.md"),
		filepath.Join(projectRoot, "CLAUDE.md"),
		filepath.Join(projectRoot, ".github", "copilot-instructions.md"),
		filepath.Join(projectRoot, ".agents", "skills", "*", "SKILL.md"),
		filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		filepath.Join(codexDir, "AGENTS.md"),
		filepath.Join(codexDir, "instructions.md"),
		filepath.Join(codexDir, "skills", "*", "SKILL.md"),
		"Perform a skills discovery phase before treating SPEC.md as complete",
		"write the selected skills into the `## SKILLS` table",
		"Populate or refresh the `## DEPENDENCIES` table",
		"keep `## SKILLS` focused on execution-time agent skills and track broader supporting inputs in `## DEPENDENCIES`",
		"for Figma or MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`",
		"do not use `.claude/skills` as canonical discovery input",
		"no section in `SPEC.md` may remain empty or contain only an HTML TODO comment",
		"`not applicable`, `not required`, or `no additional information required`",
		"## Skills",
		"read that feature's SPEC.md and the `## SKILLS` table first",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}

func TestRunSpecTemplate_IncludesSkillsSectionGuidance(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := filepath.Join(projectRoot, "home")
	codexDir := filepath.Join(homeDir, ".codex")

	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexDir)

	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	specPath := filepath.Join(projectRoot, "docs", "specs", "0010-sample", "SPEC.md")

	output := captureStdout(t, func() {
		err := runSpecTemplate(specPath, "", "sample", projectRoot, cfg, true)
		if err != nil {
			t.Fatalf("runSpecTemplate() error = %v", err)
		}
	})

	checks := []string{
		"Perform a skills discovery phase before asking sign-off questions",
		"populate the `## SKILLS` table",
		"Populate or refresh the `## DEPENDENCIES` table",
		"keep the required `none | n/a | n/a | no additional skills required | no` row",
		"keep `## SKILLS` focused on execution-time agent skills and track broader supporting inputs in `## DEPENDENCIES`",
		"the ## SKILLS section is mandatory and must be populated before sign-off",
		"the ## DEPENDENCIES section must be current before sign-off",
		"no section in `SPEC.md` may remain empty or contain only an HTML TODO comment",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = original
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close() error = %v", err)
	}

	return buf.String()
}

func chdirForTest(t *testing.T, dir string) func() {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}

	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("os.Chdir() restore error = %v", err)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := document.Write(path, content); err != nil {
		t.Fatalf("document.Write(%q) error = %v", path, err)
	}
}

func defaultKitConfig() string {
	return "goal_percentage: 95\nspecs_dir: docs/specs\nskills_dir: .agents/skills\nconstitution_path: docs/CONSTITUTION.md\nallow_out_of_order: false\nagents:\n  - AGENTS.md\n  - CLAUDE.md\nfeature_naming:\n  numeric_width: 4\n  separator: '-'\n"
}

func documentTemplateWithSummary() string {
	return "# SPEC\n\n## SUMMARY\n\nsummary\n"
}

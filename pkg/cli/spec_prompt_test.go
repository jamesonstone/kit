package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	stdreflect "reflect"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
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
		"Use an RLM-style prior-work discovery pass over",
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		"conditional reads only",
		"shared interface or contract",
		"inspect at most 5 prior feature directories",
		"do not paraphrase entire prior docs into chat",
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
		"Perform a skills discovery phase before treating SPEC.md as complete",
		"write the selected skills into the `## SKILLS` table",
		"Populate or refresh the `## DEPENDENCIES` table",
		"Use an RLM-style prior-work discovery pass over",
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		"conditional reads only",
		"shared interface or contract",
		"inspect at most 5 prior feature directories",
		"do not paraphrase entire prior docs into chat",
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

func TestRunSpecTemplate_IncludesRLMGuidanceWhenBrainstormHintsLargeRepo(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := filepath.Join(projectRoot, "home")
	codexDir := filepath.Join(homeDir, ".codex")

	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexDir)

	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0011-repository-audit")
	specPath := filepath.Join(featurePath, "SPEC.md")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")
	writeFile(t, brainstormPath, "scan repository for auth and FHIR integration points\n")

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	output := captureStdout(t, func() {
		err := runSpecTemplate(specPath, brainstormPath, "repository-audit", projectRoot, cfg, true)
		if err != nil {
			t.Fatalf("runSpecTemplate() error = %v", err)
		}
	})

	checks := []string{
		"# Use RLM Pattern",
		"parallelization_mode: \"rlm\"",
		"Index → Filter → Map → Reduce",
		"add `rlm` to the `## SKILLS` table",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}

func TestRunSpecInteractive_UsesEditorByDefault(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0010-sample")
	specPath := filepath.Join(featurePath, "SPEC.md")
	writeFile(t, specPath, "# SPEC\n")

	restore := chdirForTest(t, projectRoot)
	defer restore()

	previousWait := awaitEditorLaunchConfirmation
	previousRunner := editorInputRunner
	defer func() {
		awaitEditorLaunchConfirmation = previousWait
		editorInputRunner = previousRunner
	}()

	var sequence []string
	awaitEditorLaunchConfirmation = func(_ *os.File, _ io.Writer) error {
		sequence = append(sequence, "wait")
		return nil
	}
	editorInputRunner = func(_ freeTextInputConfig, fieldName, _ string) (string, bool, error) {
		sequence = append(sequence, fieldName)
		return fieldName + " answer", true, nil
	}

	cfg := config.Default()
	feat := &feature.Feature{Slug: "sample", DirName: "0010-sample", Path: featurePath}

	output := captureStdout(t, func() {
		err := runSpecInteractive(
			specPath,
			"",
			feat,
			projectRoot,
			cfg,
			newFreeTextInputConfig(false, "", false, true),
			true,
		)
		if err != nil {
			t.Fatalf("runSpecInteractive() error = %v", err)
		}
	})

	wantSequence := []string{
		"wait",
		"problem",
		"wait",
		"goals",
		"wait",
		"non-goals",
		"wait",
		"users",
		"wait",
		"requirements",
		"wait",
		"acceptance",
		"wait",
		"edge-cases",
	}
	if !stdreflect.DeepEqual(sequence, wantSequence) {
		t.Fatalf("unexpected editor prompt sequence: got %v want %v", sequence, wantSequence)
	}

	checks := []string{
		"A vim-compatible editor will open for each free-text response.",
		"**PROBLEM**: problem answer",
		"**EDGE-CASES**: edge-cases answer",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}

func TestOutputCompiledPrompt_IncludesRLMGuidanceWhenContextRequiresIt(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := filepath.Join(projectRoot, "home")
	codexDir := filepath.Join(homeDir, ".codex")

	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexDir)

	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "AGENTS.md"), "# AGENTS\n")
	writeFile(t, filepath.Join(projectRoot, "CLAUDE.md"), "# CLAUDE\n")
	writeFile(t, filepath.Join(projectRoot, ".github", "copilot-instructions.md"), "# COPILOT\n")

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0012-codebase-audit")
	specPath := filepath.Join(featurePath, "SPEC.md")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")
	writeFile(t, brainstormPath, "# BRAINSTORM\n")
	writeFile(t, specPath, documentTemplateWithSummary())

	restore := chdirForTest(t, projectRoot)
	defer restore()

	cfg := config.Default()
	answers := &specAnswers{Problem: "Need codebase-wide analysis of all FHIR and auth flows."}

	output := captureStdout(t, func() {
		err := outputCompiledPrompt(specPath, brainstormPath, "codebase-audit", projectRoot, cfg, answers, true)
		if err != nil {
			t.Fatalf("outputCompiledPrompt() error = %v", err)
		}
	})

	checks := []string{
		"# Use RLM Pattern",
		"parallelization_mode: \"rlm\"",
		"Index → Filter → Map → Reduce",
		"add `rlm` to the `## SKILLS` table",
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
	return "goal_percentage: 95\nspecs_dir: docs/specs\nskills_dir: .agents/skills\nconstitution_path: docs/CONSTITUTION.md\nallow_out_of_order: false\nagents:\n  - AGENTS.md\n  - CLAUDE.md\n  - .github/copilot-instructions.md\nfeature_naming:\n  numeric_width: 4\n  separator: '-'\n"
}

func documentTemplateWithSummary() string {
	return "# SPEC\n\n## SUMMARY\n\nsummary\n"
}

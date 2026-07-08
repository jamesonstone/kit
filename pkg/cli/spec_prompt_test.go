package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
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
		"`SPEC.md` is the single durable feature artifact",
		"Keep front matter `references`",
		"Keep front matter `skills` focused on execution-time skills",
		"Use an RLM-style just-in-time prior-work pass over",
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		"conditional reads only",
		"shared interface or contract",
		"inspect at most 5 prior feature directories",
		"do not paraphrase entire prior docs into chat",
		"Do not use `.claude/skills` as canonical discovery input",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
	assertV2SpecPromptContract(t, output)
	assertV2SpecPromptExcludesV1StageAssumptions(t, output)
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
		err := runSpecTemplate(specPath, "", "sample", projectRoot, cfg, true, false)
		if err != nil {
			t.Fatalf("runSpecTemplate() error = %v", err)
		}
	})

	checks := []string{
		"`SPEC.md` is the single durable feature artifact",
		"Keep front matter `references`",
		"Keep front matter `skills` focused on execution-time skills",
		"Use an RLM-style just-in-time prior-work pass over",
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		"conditional reads only",
		"shared interface or contract",
		"inspect at most 5 prior feature directories",
		"do not paraphrase entire prior docs into chat",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
	assertV2SpecPromptContract(t, output)
	assertV2SpecPromptExcludesV1StageAssumptions(t, output)
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
		err := runSpecTemplate(specPath, brainstormPath, "repository-audit", projectRoot, cfg, true, false)
		if err != nil {
			t.Fatalf("runSpecTemplate() error = %v", err)
		}
	})

	checks := []string{
		"# Use RLM Pattern",
		"parallelization_mode: \"rlm\"",
		"immediate decision → smallest artifact → required facts → act or recurse",
		"add `rlm` to canonical front matter `skills`",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q", check)
		}
	}
}

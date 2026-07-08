package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildBrainstormPrompt(t *testing.T) {
	prompt := buildBrainstormPrompt(
		"/tmp/docs/specs/0001-sample/BRAINSTORM.md",
		"sample-feature",
		"/tmp/project",
		"Need better import validation for malformed CSV uploads.",
		95,
	)

	checks := []string{
		"Research and document feature: **sample-feature**",
		"Do NOT implement code, write production changes, or move into execution",
		"Ask clarifying questions until you reach ≥95% confidence that you understand the problem and desired solution",
		"Use numbered lists",
		"Ask questions in batches of up to 10",
		"For every question, include your current best recommended default, proposed solution, or assumption",
		"State uncertainties",
		"\"yes\" or \"y\" approves all recommended defaults in the batch",
		"\"yes 3, 4, 5\" or \"y 3, 4, 5\" approves only those numbered defaults in the batch",
		"If the user approves only specific question numbers, treat all other questions in that batch as unresolved",
		"After each batch of up to 10 questions, output your current percentage understanding so the user can see progress",
		"Only update BRAINSTORM.md and supporting documentation; do not modify product code, tests, runtime config, generated artifacts, or implementation files.",
		"research and documentation only; no implementation",
		"kit spec sample-feature",
		"/tmp/docs/specs/0001-sample/BRAINSTORM.md",
		"/tmp/project/docs/notes/0001-sample",
		"Inspect the feature notes directory",
		"ignore `.gitkeep`",
		"read only the notes relevant to the user thesis",
		"record specific note files that shaped the brainstorm",
		"leave the notes directory reference as `optional`",
		"/tmp/project/docs/CONSTITUTION.md",
		"canonical front matter references",
		"`name`, `type`, `target`, `relation`, `read_policy`, `used_for`, and `status`",
		"Use an RLM-style just-in-time prior-work pass over `/tmp/docs/specs` before broad repository reads",
		"/tmp/project/docs/PROJECT_PROGRESS_SUMMARY.md",
		"conditional reads only",
		"shared interface or contract",
		"inspect at most 5 prior feature directories",
		"do not paraphrase entire prior docs into chat",
		"for Figma or other MCP-driven design references, store the exact design URL or file/node reference in `target` and use stable selectors when needed",
		"`status: stale`",
		"no section in `BRAINSTORM.md` may remain empty or contain only an HTML TODO comment",
		"`not applicable`, `not required`, or `no additional information required`",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}
	assertFinalResponseContractHeadings(t, prompt,
		"Summary",
		"Artifacts Updated",
		"Key Decisions",
		"Open Questions",
		"Next Step",
	)

	if !strings.HasPrefix(prompt, "Research and document feature: **sample-feature**\n\n") {
		t.Fatalf("expected prompt to start with research header, got %q", prompt[:64])
	}
	if strings.Contains(prompt, "/plan") || strings.Contains(prompt, "planning mode") {
		t.Fatalf("expected prompt to avoid native plan-mode triggers, got %q", prompt)
	}
}

func TestRunBrainstorm_CreatesFeatureNotesDirAndSeedsReference(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreEditor := stubBrainstormEditor(t, "Need better import validation for malformed CSV uploads.")
	defer restoreEditor()
	restoreFlags := setBrainstormFlagState(false, "", false, false, false, false)
	defer restoreFlags()

	cmd := newBrainstormTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set() error = %v", err)
	}
	_ = captureStdout(t, func() {
		if err := runBrainstorm(cmd, []string{"sample-feature"}); err != nil {
			t.Fatalf("runBrainstorm() error = %v", err)
		}
	})

	notesPath := filepath.Join(projectRoot, "docs", "notes", "0001-sample-feature")
	if _, err := os.Stat(filepath.Join(notesPath, ".gitkeep")); err != nil {
		t.Fatalf("expected feature notes .gitkeep, got %v", err)
	}

	brainstormPath := filepath.Join(projectRoot, "docs", "specs", "0001-sample-feature", "BRAINSTORM.md")
	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	checks := []string{
		"kit_metadata_version: 1",
		"artifact: brainstorm",
		"dir: 0001-sample-feature",
		"Need better import validation for malformed CSV uploads.",
		"name: Feature notes",
		"target: docs/notes/0001-sample-feature",
		"relation: informs",
		"read_policy: conditional",
		"used_for: optional pre-brainstorm research input",
		"status: optional",
	}
	for _, check := range checks {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected BRAINSTORM.md to contain %q, got %q", check, string(content))
		}
	}
}

func TestRunBrainstormFrontendProfileCreatesDesignMaterialsAndSeedsReferences(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)

	restoreWD, err := ensureHandoffTestWorkingDirectory(projectRoot)
	if err != nil {
		t.Fatalf("ensureHandoffTestWorkingDirectory() error = %v", err)
	}
	defer restoreWD()

	restoreEditor := stubBrainstormEditor(t, "Need a responsive dashboard redesign.")
	defer restoreEditor()
	restoreFlags := setBrainstormFlagState(false, "", false, false, false, false)
	defer restoreFlags()
	restorePromptProfileState(t, promptProfileFrontend, true)

	cmd := newBrainstormTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set() error = %v", err)
	}
	output := captureStdout(t, func() {
		if err := runBrainstorm(cmd, []string{"dashboard-redesign"}); err != nil {
			t.Fatalf("runBrainstorm() error = %v", err)
		}
	})

	featureDir := "0001-dashboard-redesign"
	designPath := filepath.Join(projectRoot, "docs", "notes", featureDir, "design")
	for _, path := range []string{
		filepath.Join(designPath, ".gitkeep"),
		filepath.Join(designPath, "screenshots", ".gitkeep"),
		filepath.Join(designPath, "references", ".gitkeep"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected frontend design placeholder %s, got %v", path, err)
		}
	}

	brainstormPath := filepath.Join(projectRoot, "docs", "specs", featureDir, "BRAINSTORM.md")
	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	text := string(content)
	checks := []string{
		"name: Feature notes",
		"target: docs/notes/0001-dashboard-redesign",
		"name: Frontend profile",
		"target: --profile=frontend",
		"name: Design materials",
		"target: docs/notes/0001-dashboard-redesign/design",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected BRAINSTORM.md to contain %q, got:\n%s", check, text)
		}
	}

	promptChecks := []string{
		"DESIGN MATERIALS",
		designPath,
		"ignore `.gitkeep`",
		"## Frontend Profile",
	}
	for _, check := range promptChecks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected frontend brainstorm prompt to contain %q, got:\n%s", check, output)
		}
	}
}

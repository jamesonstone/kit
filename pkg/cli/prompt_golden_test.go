package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func assertPromptGolden(t *testing.T, name, got string) {
	t.Helper()

	wantBytes, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", name, err)
	}

	want := string(wantBytes)
	got = normalizeGoldenText(got)
	want = normalizeGoldenText(want)
	if got != want {
		t.Fatalf("prompt mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

func normalizeGoldenText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimSuffix(s, "\n")
}

func TestCodeReviewInstructions_Golden(t *testing.T) {
	assertPromptGolden(t, "code_review_prompt.golden", codeReviewInstructions())
}

func TestGenericSummarizeInstructions_Golden(t *testing.T) {
	assertPromptGolden(t, "summarize_generic_prompt.golden", genericSummarizeInstructions())
}

func TestFeatureScopedSummarizeInstructions_Golden(t *testing.T) {
	assertPromptGolden(
		t,
		"summarize_feature_prompt.golden",
		featureScopedSummarizeInstructions("/repo", "alpha", "/repo/docs/specs/0001-alpha"),
	)
}

func TestBuildSpecV2SupervisorPrompt_Golden(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	t.Setenv("CODEX_HOME", "/home/tester/.codex")

	cfg := config.Default()
	prompt := buildSpecV2SupervisorPrompt(specV2PromptInput{
		SpecPath:       "/repo/docs/specs/0001-alpha/SPEC.md",
		BrainstormPath: "/repo/docs/specs/0001-alpha/BRAINSTORM.md",
		FeatureSlug:    "alpha",
		ProjectRoot:    "/repo",
		Config:         cfg,
		Answers: &specAnswers{
			Problem:        "Build the alpha workflow.",
			Requirements:   "Keep the workflow deterministic.",
			Acceptance:     "Acceptance criteria map to validation evidence.",
			DeliveryIntent: "Use existing in-flight changes; defer PR until validation passes.",
		},
	})

	assertPromptGolden(t, "spec_v2_supervisor_prompt.golden", prompt)
	assertV2SpecPromptContract(t, prompt)
	assertV2SpecPromptExcludesV1StageAssumptions(t, prompt)
}

func TestBuildSpecV2SupervisorPrompt_SingleAgentMode(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	t.Setenv("CODEX_HOME", "/home/tester/.codex")

	prompt := buildSpecV2SupervisorPrompt(specV2PromptInput{
		SpecPath:    "/repo/docs/specs/0001-alpha/SPEC.md",
		FeatureSlug: "alpha",
		ProjectRoot: "/repo",
		Config:      config.Default(),
		SingleAgent: true,
	})

	checks := []string{
		"`--single-agent` is active. Keep execution in one supervisor lane and do not require implementation or verification subagents.",
		"Even in single-agent mode, record logical work lanes in the Agent Team Plan when they clarify sequencing, validation, or risk.",
		"In final responses, state `single supervisor lane; no specialist or verification agents spawned` and cite `--single-agent` as the exception.",
		"single supervisor lane; no specialist or verification agents spawned",
		"state the exception that justified single-lane execution",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected single-agent prompt to contain %q", check)
		}
	}

	unwanted := []string{
		"Default to a subagent team for implementation and verification.",
		"do not keep work single-lane merely because subagents were not explicitly re-requested.",
	}
	for _, check := range unwanted {
		if strings.Contains(prompt, check) {
			t.Fatalf("single-agent prompt should not contain default-mode instruction %q", check)
		}
	}
	assertV2SpecPromptExcludesV1StageAssumptions(t, prompt)
}

func TestBuildReflectPrompt_Golden(t *testing.T) {
	prompt := buildReflectPrompt(
		"/repo",
		"/repo/docs/CONSTITUTION.md",
		"/repo/docs/PROJECT_PROGRESS_SUMMARY.md",
		"/repo/docs/specs/0001-alpha/BRAINSTORM.md",
		"/repo/docs/specs/0001-alpha/SPEC.md",
		"/repo/docs/specs/0001-alpha/PLAN.md",
		"/repo/docs/specs/0001-alpha/TASKS.md",
		"alpha",
	)

	assertPromptGolden(
		t,
		"reflect_feature_prompt.golden",
		prompt,
	)
	assertFinalResponseContractHeadings(t, prompt,
		"Changeset",
		"Verification",
		"Review Findings",
		"Doc Trace",
		"Final Status",
		"Follow-ups",
	)
}

func TestBuildReflectPromptIncludesProjectRefreshDueState(t *testing.T) {
	projectRoot, cfg := setupProjectRefreshTestProject(t)
	cfg.ProjectRefresh.Constitution.FeatureInterval = 1
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	createCompletedV2Feature(t, projectRoot, "0001-alpha")

	prompt := buildReflectPrompt(
		projectRoot,
		filepath.Join(projectRoot, "docs", "CONSTITUTION.md"),
		filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"),
		"",
		filepath.Join(projectRoot, "docs", "specs", "0001-alpha", "SPEC.md"),
		"",
		"",
		"alpha",
	)

	if !strings.Contains(prompt, "current due state: due") {
		t.Fatalf("expected reflect prompt to include due project refresh state, got:\n%s", prompt)
	}
	if !strings.Contains(prompt, "kit project refresh") {
		t.Fatalf("expected reflect prompt to point to kit project refresh, got:\n%s", prompt)
	}
}

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
	path := filepath.Join("testdata", name)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, []byte(strings.TrimSuffix(got, "\n")+"\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(%q): %v", name, err)
		}
		return
	}

	wantBytes, err := os.ReadFile(path)
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
		"`--single-agent` is active: keep execution and verification in one supervisor lane",
		"single supervisor lane; no specialist or verification agents spawned",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected single-agent prompt to contain %q", check)
		}
	}

	unwanted := []string{
		"For nontrivial separable work",
		"use at least one read-only verifier",
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

func TestBuildReflectPromptTrimsFeatureSlugBeforeUse(t *testing.T) {
	prompt := buildReflectPrompt(
		"/repo",
		"/repo/docs/CONSTITUTION.md",
		"/repo/docs/PROJECT_PROGRESS_SUMMARY.md",
		"/repo/docs/specs/0001-alpha/BRAINSTORM.md",
		"/repo/docs/specs/0001-alpha/SPEC.md",
		"/repo/docs/specs/0001-alpha/PLAN.md",
		"/repo/docs/specs/0001-alpha/TASKS.md",
		"  alpha  ",
	)

	for _, check := range []string{
		"## Reflection — Feature: alpha",
		"run `kit legacy verify alpha`",
	} {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected trimmed feature slug in %q, got:\n%s", check, prompt)
		}
	}
	if strings.Contains(prompt, "  alpha  ") {
		t.Fatalf("expected raw whitespace-padded feature slug to be absent, got:\n%s", prompt)
	}
}

func TestBuildReflectPromptWhitespaceOnlyFeatureSlugUsesGenericScope(t *testing.T) {
	prompt := buildReflectPrompt(
		"/repo",
		"/repo/docs/CONSTITUTION.md",
		"/repo/docs/PROJECT_PROGRESS_SUMMARY.md",
		"/repo/docs/specs/0001-alpha/BRAINSTORM.md",
		"/repo/docs/specs/0001-alpha/SPEC.md",
		"/repo/docs/specs/0001-alpha/PLAN.md",
		"/repo/docs/specs/0001-alpha/TASKS.md",
		"   ",
	)

	for _, check := range []string{
		"## Reflection\n",
		"no feature-scoped verification evidence is required for generic reflection",
	} {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected whitespace-only slug to use generic reflection scope with %q, got:\n%s", check, prompt)
		}
	}
	for _, unwanted := range []string{
		"## Reflection — Feature:",
		"no local verification run found",
		"`kit legacy verify  ",
	} {
		if strings.Contains(prompt, unwanted) {
			t.Fatalf("expected whitespace-only slug to omit feature verification requirement %q, got:\n%s", unwanted, prompt)
		}
	}
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

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestBuildReflectPrompt_Golden(t *testing.T) {
	assertPromptGolden(
		t,
		"reflect_feature_prompt.golden",
		buildReflectPrompt(
			"/repo",
			"/repo/docs/CONSTITUTION.md",
			"/repo/docs/PROJECT_PROGRESS_SUMMARY.md",
			"/repo/docs/specs/0001-alpha/BRAINSTORM.md",
			"/repo/docs/specs/0001-alpha/SPEC.md",
			"/repo/docs/specs/0001-alpha/PLAN.md",
			"/repo/docs/specs/0001-alpha/TASKS.md",
			"alpha",
		),
	)
}

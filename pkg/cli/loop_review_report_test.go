package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadLoopReviewConfirmation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "yes shorthand", input: "y\n", want: true},
		{name: "yes word", input: "yes\n", want: true},
		{name: "no shorthand", input: "n\n", want: false},
		{name: "blank default", input: "\n", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readLoopReviewConfirmation(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("readLoopReviewConfirmation() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("confirmation = %t, want %t", got, tt.want)
			}
		})
	}

	if _, err := readLoopReviewConfirmation(strings.NewReader("maybe\n")); err == nil {
		t.Fatal("readLoopReviewConfirmation(invalid) error = nil")
	}
}

func TestLatestLoopReviewReportIgnoresWorkflowRuns(t *testing.T) {
	projectRoot := t.TempDir()
	writeLoopReviewRunJSON(t, projectRoot, "workflow", `{
  "schema_version": 1,
  "run_id": "workflow",
  "feature": "0001-demo",
  "status": "stopped",
  "started_at": "2026-06-17T13:00:00Z"
}`)
	writeLoopReviewRunJSON(t, projectRoot, "old-review", `{
  "schema_version": 1,
  "run_id": "old-review",
  "status": "complete",
  "base_ref": "origin/main",
  "started_at": "2026-06-17T13:05:00Z"
}`)
	writeLoopReviewRunJSON(t, projectRoot, "new-review", `{
  "schema_version": 1,
  "run_id": "new-review",
  "status": "stopped",
  "stop_reason": "max iterations reached: 10",
  "base_ref": "origin/main",
  "started_at": "2026-06-17T13:10:00Z"
}`)

	report, found, err := latestLoopReviewReport(projectRoot)
	if err != nil {
		t.Fatalf("latestLoopReviewReport() error = %v", err)
	}
	if !found {
		t.Fatal("latestLoopReviewReport() found = false, want true")
	}
	if report.RunID != "new-review" {
		t.Fatalf("RunID = %q, want new-review", report.RunID)
	}
}

func writeLoopReviewRunJSON(t *testing.T, projectRoot, runID, content string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".kit", "loops", runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(run.json) error = %v", err)
	}
}

package verify

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestExecuteRunReportsNoDeclaredChecks(t *testing.T) {
	run := ExecuteRun(context.Background(), RunOptions{
		ProjectRoot: "/tmp",
		Feature:     FeatureRef{DirName: "fixture"},
		TaskIDs:     []string{"T001"},
	})

	if run.Status != RunStatusNoDeclaredChecks {
		t.Fatalf("Status = %q, want %q", run.Status, RunStatusNoDeclaredChecks)
	}
	if len(run.Results) != 0 {
		t.Fatalf("len(Results) = %d, want 0", len(run.Results))
	}
}

func TestExecuteRunPreservesExpectedFiles(t *testing.T) {
	run := ExecuteRun(context.Background(), RunOptions{
		ProjectRoot:   "/tmp",
		Feature:       FeatureRef{DirName: "fixture"},
		TaskIDs:       []string{"T001"},
		ExpectedFiles: []string{"internal/verify/"},
		DryRun:        true,
	})

	if len(run.ExpectedFiles) != 1 || run.ExpectedFiles[0] != "internal/verify/" {
		t.Fatalf("ExpectedFiles = %#v", run.ExpectedFiles)
	}
}

func TestNewRunIDIncludesTimestampAndEntropy(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 123, time.UTC)
	id := NewRunID(now)

	if !strings.HasPrefix(id, "20260520T120000.000000123Z-") {
		t.Fatalf("RunID = %q, want timestamp prefix", id)
	}
	if len(id) != len("20260520T120000.000000123Z-000000") {
		t.Fatalf("RunID length = %d", len(id))
	}
}

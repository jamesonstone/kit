package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/usage"
)

func TestRunUsageReportEmitsAggregatedVersionedJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Test binaries never record usage, so seed the store directly.
	dir := filepath.Join(home, ".config", "kit", "usage")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	event, err := json.Marshal(usage.Event{
		SchemaVersion: usage.SchemaVersion, Timestamp: now, Command: "status",
		Version: "v2.0.0", Success: true, ElapsedMS: 1, ProjectID: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, now.Format("2006-01")+"-0001.jsonl"), append(event, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	if err := runUsageReport(cmd, &usageReportOptions{since: "90d", jsonOutput: true}); err != nil {
		t.Fatal(err)
	}
	var report usage.Report
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatalf("invalid usage JSON: %v\n%s", err, output.String())
	}
	if report.SchemaVersion != usage.SchemaVersion || report.TotalCalls != 1 {
		t.Fatalf("unexpected usage report: %#v", report)
	}
	if len(report.Commands) != 1 || report.Commands[0].Command != "status" {
		t.Fatalf("unexpected command summary: %#v", report.Commands)
	}
	if len(report.ZeroUseCommands) == 0 {
		t.Fatal("expected report to expose preserved commands with no observed use")
	}
}

func TestParseUsageDurationAcceptsBoundedDayWindow(t *testing.T) {
	duration, err := parseUsageDuration("90d")
	if err != nil || duration != 90*24*time.Hour {
		t.Fatalf("parseUsageDuration(90d) = %v, %v", duration, err)
	}
	for _, value := range []string{"0d", "-1d", "forever"} {
		if _, err := parseUsageDuration(value); err == nil {
			t.Errorf("parseUsageDuration(%q) succeeded", value)
		}
	}
}

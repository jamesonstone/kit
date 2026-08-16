package cli

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/usage"
)

func TestRunUsageReportEmitsAggregatedVersionedJSON(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := usage.Record(usage.RecordInput{
		Command: "status", Version: "v2.0.0", Elapsed: time.Millisecond,
	}); err != nil {
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

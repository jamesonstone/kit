package worktree

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSyncReportHumanJSONParityAndOutputFailure(t *testing.T) {
	fixture := newGitFixture(t)
	report := fixture.app.runSync(
		context.Background(),
		fixture.primary,
		SyncOptions{DryRun: true},
	)

	var human bytes.Buffer
	if err := writeSyncHuman(&human, report); err != nil {
		t.Fatalf("writeSyncHuman() error = %v", err)
	}
	var encoded bytes.Buffer
	if err := writeSyncJSON(&encoded, report); err != nil {
		t.Fatalf("writeSyncJSON() error = %v", err)
	}
	var decoded SyncReport
	if err := json.Unmarshal(encoded.Bytes(), &decoded); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	if !reflect.DeepEqual(report, decoded) {
		t.Fatalf("JSON report changed data:\nreport=%#v\ndecoded=%#v", report, decoded)
	}
	for _, worktree := range report.Worktrees {
		for _, value := range []string{
			worktree.State,
			worktree.Head,
			worktree.LastUpdated,
			worktree.Path,
		} {
			if !strings.Contains(human.String(), value) {
				t.Fatalf("human report omitted %q:\n%s", value, human.String())
			}
		}
	}

	fixture.app.out = failingWriter{}
	err := fixture.app.Run(
		context.Background(),
		fixture.primary,
		[]string{"sync", "--dry-run", "--json"},
	)
	if err == nil || !strings.Contains(err.Error(), "write output") {
		t.Fatalf("output failure error = %v", err)
	}
}

func TestSyncHumanFetchDetailIsSingleLine(t *testing.T) {
	report := SyncReport{
		Repository: "example/project",
		Result:     "failed",
		Fetch: SyncOperation{
			Status: "failed",
			Detail: "fetch failed\nwith a second line",
		},
	}

	var output bytes.Buffer
	if err := writeSyncHuman(&output, report); err != nil {
		t.Fatalf("writeSyncHuman() error = %v", err)
	}
	if got, want := output.String(), "SYNC example/project (failed)\nFETCH\tfailed\tfetch failed with a second line\nDEFAULT\t\t\t\nACTION\tBRANCH\tREASON\tPATH\nPRUNE\t\t\nSTATE\tHEAD\tLAST UPDATED\tPATH\n"; got != want {
		t.Fatalf("human output = %q, want %q", got, want)
	}
}

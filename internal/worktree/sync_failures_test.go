package worktree

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSyncFailuresFailClosedAndIndependentCandidatesContinue(t *testing.T) {
	t.Run("fetch", func(t *testing.T) {
		fixture := newGitFixture(t)
		run := fixture.app.run
		fixture.app.run = func(
			ctx context.Context,
			cwd string,
			name string,
			args ...string,
		) ([]byte, error) {
			if name == "git" && len(args) >= 1 && args[0] == "fetch" {
				return []byte("simulated fetch failure"), fmt.Errorf("simulated failure")
			}
			return run(ctx, cwd, name, args...)
		}
		before := captureSyncState(t, fixture)
		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err == nil {
			t.Fatal("sync expected fetch failure")
		}
		if report.Result != "failed" || report.Fetch.Status != "failed" {
			t.Fatalf("fetch failure report = %#v", report)
		}
		after := captureSyncState(t, fixture)
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("fetch failure mutated state:\nbefore=%#v\nafter=%#v", before, after)
		}
	})

	t.Run("github", func(t *testing.T) {
		fixture := newGitFixture(t)
		runGit(t, fixture.primary, "branch", "topic/github", "origin/main")
		runWT(t, fixture.app, fixture.primary, "add", "topic/github", "--no-link-env")
		path := canonicalTestLanePath(fixture, "topic/github")
		fixture.app.resolveSyncPRs = func(
			context.Context,
			string,
			string,
			[]string,
		) (map[string][]SyncPullRequest, error) {
			return nil, fmt.Errorf("simulated GitHub failure")
		}

		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err == nil {
			t.Fatal("sync expected GitHub failure")
		}
		if decision := findSyncLane(t, report, "topic/github"); decision.Reason != "github-unavailable" {
			t.Fatalf("decision = %#v", decision)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("GitHub failure removed lane: %v", err)
		}
	})

	t.Run("removal continues", func(t *testing.T) {
		fixture := newGitFixture(t)
		failedPath, failedOID := createMergedLane(t, fixture, "topic/fail-remove")
		safePath, safeOID := createMergedLane(t, fixture, "topic/continue")
		fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
			"topic/fail-remove": {
				mergedSyncPR(50, "topic/fail-remove", "main", failedOID),
			},
			"topic/continue": {
				mergedSyncPR(51, "topic/continue", "main", safeOID),
			},
		})
		run := fixture.app.run
		fixture.app.run = func(
			ctx context.Context,
			cwd string,
			name string,
			args ...string,
		) ([]byte, error) {
			if name == "git" &&
				len(args) >= 3 &&
				args[0] == "worktree" &&
				args[1] == "remove" &&
				samePath(args[2], failedPath) {
				return []byte("simulated removal failure"), fmt.Errorf("simulated failure")
			}
			return run(ctx, cwd, name, args...)
		}

		report, err := runSyncJSON(t, fixture, fixture.primary)
		if err == nil {
			t.Fatal("sync expected aggregate failure")
		}
		if decision := findSyncLane(t, report, "topic/fail-remove"); decision.Reason != "worktree-removal-failed" {
			t.Fatalf("failed decision = %#v", decision)
		}
		if decision := findSyncLane(t, report, "topic/continue"); decision.Action != "removed" {
			t.Fatalf("continued decision = %#v", decision)
		}
		if _, err := os.Stat(failedPath); err != nil {
			t.Fatalf("failed candidate was not preserved: %v", err)
		}
		if _, err := os.Stat(safePath); !os.IsNotExist(err) {
			t.Fatalf("safe candidate still exists or stat failed: %v", err)
		}
	})
}

func TestSyncBranchDeletionFailureIsReportedWithoutForce(t *testing.T) {
	fixture := newGitFixture(t)
	path, headOID := createMergedLane(t, fixture, "topic/branch-delete")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/branch-delete": {
			mergedSyncPR(60, "topic/branch-delete", "main", headOID),
		},
	})
	run := fixture.app.run
	fixture.app.run = func(
		ctx context.Context,
		cwd string,
		name string,
		args ...string,
	) ([]byte, error) {
		if name == "git" &&
			strings.Contains(strings.Join(args, "\x00"), "branch\x00-d") {
			return []byte("simulated branch deletion failure"), fmt.Errorf("simulated failure")
		}
		return run(ctx, cwd, name, args...)
	}

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err == nil {
		t.Fatal("sync expected branch deletion failure")
	}
	decision := findSyncLane(t, report, "topic/branch-delete")
	if decision.Action != "worktree-removed" ||
		decision.Reason != "branch-deletion-failed" {
		t.Fatalf("decision = %#v", decision)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists or stat failed: %v", err)
	}
	if got := gitText(t, fixture.primary, "show-ref", "--verify", "refs/heads/topic/branch-delete"); got == "" {
		t.Fatal("branch should remain after ordinary deletion failure")
	}
}

func TestSyncInspectionAndPruneFailuresAreAggregated(t *testing.T) {
	fixture := newGitFixture(t)
	runGit(t, fixture.primary, "branch", "topic/inspect-failure", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "topic/inspect-failure", "--no-link-env")
	path := canonicalTestLanePath(fixture, "topic/inspect-failure")
	headOID := gitText(t, path, "rev-parse", "HEAD")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/inspect-failure": {
			mergedSyncPR(61, "topic/inspect-failure", "main", headOID),
		},
	})
	run := fixture.app.run
	fixture.app.run = func(
		ctx context.Context,
		cwd string,
		name string,
		args ...string,
	) ([]byte, error) {
		switch {
		case name == "git" &&
			len(args) >= 1 &&
			args[0] == "status" &&
			samePath(cwd, path):
			return []byte("simulated status failure"), fmt.Errorf("simulated failure")
		case name == "git" &&
			len(args) >= 2 &&
			args[0] == "worktree" &&
			args[1] == "prune":
			return []byte("simulated prune failure"), fmt.Errorf("simulated failure")
		default:
			return run(ctx, cwd, name, args...)
		}
	}

	report, err := runSyncJSON(t, fixture, fixture.primary)
	if err == nil {
		t.Fatal("sync expected aggregate failure")
	}
	decision := findSyncLane(t, report, "topic/inspect-failure")
	if decision.Reason != "inspection-failed" || decision.Action != "preserved" {
		t.Fatalf("decision = %#v", decision)
	}
	if report.Prune.Status != "failed" || len(report.Failures) != 2 {
		t.Fatalf("failure report = %#v", report.Failures)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("inspection failure removed lane: %v", err)
	}
}

func TestResolveSyncPullRequestsUsesBatchAndExactFallback(t *testing.T) {
	app := NewApp(io.Discard, io.Discard)
	var calls []string
	app.run = func(
		_ context.Context,
		_ string,
		name string,
		args ...string,
	) ([]byte, error) {
		if name != "gh" {
			return nil, fmt.Errorf("unexpected command %s", name)
		}
		head := ""
		for i := range args {
			if args[i] == "--head" && i+1 < len(args) {
				head = args[i+1]
			}
		}
		calls = append(calls, head)
		var prs []SyncPullRequest
		switch head {
		case "":
			prs = []SyncPullRequest{
				mergedSyncPR(80, "topic/batch", "main", strings.Repeat("1", 40)),
			}
		case "topic/batch":
			prs = []SyncPullRequest{
				mergedSyncPR(80, "topic/batch", "main", strings.Repeat("1", 40)),
				mergedSyncPR(82, "other-head", "main", strings.Repeat("3", 40)),
			}
		case "topic/fallback":
			prs = []SyncPullRequest{
				mergedSyncPR(81, "topic/fallback", "main", strings.Repeat("2", 40)),
				mergedSyncPR(83, "other-head", "main", strings.Repeat("4", 40)),
			}
		}
		return json.Marshal(prs)
	}

	result, err := app.resolveSyncPullRequests(
		context.Background(),
		t.TempDir(),
		"example/project",
		[]string{"topic/batch", "topic/fallback"},
	)
	if err != nil {
		t.Fatalf("resolveSyncPullRequests() error = %v", err)
	}
	if len(result["topic/batch"]) != 1 ||
		len(result["topic/fallback"]) != 1 ||
		!reflect.DeepEqual(calls, []string{"", "topic/batch", "topic/fallback"}) {
		t.Fatalf("result=%#v calls=%#v", result, calls)
	}
}

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

func TestSyncIsIdempotent(t *testing.T) {
	fixture := newGitFixture(t)
	_, headOID := createMergedLane(t, fixture, "topic/idempotent")
	fixture.app.resolveSyncPRs = staticSyncPRs(map[string][]SyncPullRequest{
		"topic/idempotent": {
			mergedSyncPR(70, "topic/idempotent", "main", headOID),
		},
	})
	first, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("first sync error = %v", err)
	}
	if decision := findSyncLane(t, first, "topic/idempotent"); decision.Action != "removed" {
		t.Fatalf("first decision = %#v", decision)
	}
	second, err := runSyncJSON(t, fixture, fixture.primary)
	if err != nil {
		t.Fatalf("second sync error = %v", err)
	}
	for _, decision := range second.Lanes {
		if decision.Branch == "topic/idempotent" {
			t.Fatalf("second sync repeated removed lane: %#v", decision)
		}
	}
}

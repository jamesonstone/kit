package worktree

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func runSyncJSON(
	t *testing.T,
	fixture gitFixture,
	cwd string,
	args ...string,
) (SyncReport, error) {
	t.Helper()
	fixture.out.Reset()
	commandArgs := []string{"sync", "--json"}
	commandArgs = append(commandArgs, args...)
	err := fixture.app.Run(context.Background(), cwd, commandArgs)
	var report SyncReport
	if decodeErr := json.Unmarshal(fixture.out.Bytes(), &report); decodeErr != nil {
		t.Fatalf(
			"decode sync report: %v\ncommand error: %v\noutput:\n%s",
			decodeErr,
			err,
			fixture.out.String(),
		)
	}
	return report, err
}

func staticSyncPRs(
	prs map[string][]SyncPullRequest,
) syncPRResolverFunc {
	return func(
		_ context.Context,
		_ string,
		_ string,
		branches []string,
	) (map[string][]SyncPullRequest, error) {
		result := make(map[string][]SyncPullRequest, len(branches))
		for _, branch := range branches {
			result[branch] = append([]SyncPullRequest(nil), prs[branch]...)
		}
		return result, nil
	}
}

func mergedSyncPR(
	number int,
	branch string,
	base string,
	headOID string,
) SyncPullRequest {
	mergedAt := time.Unix(1_700_000_000+int64(number), 0).UTC()
	return SyncPullRequest{
		Number:      number,
		State:       "MERGED",
		MergedAt:    &mergedAt,
		BaseRefName: base,
		HeadRefName: branch,
		HeadRefOID:  headOID,
		URL:         fmt.Sprintf("https://github.com/example/project/pull/%d", number),
	}
}

func createMergedLane(
	t *testing.T,
	fixture gitFixture,
	branch string,
) (string, string) {
	t.Helper()
	runGit(t, fixture.primary, "branch", "--track", branch, "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", branch)
	path := canonicalTestLanePath(fixture, branch)
	filename := strings.ReplaceAll(branch, "/", "-") + ".txt"
	if err := os.WriteFile(filepath.Join(path, filename), []byte("merged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, path, "add", filename)
	runGit(t, path, "commit", "-m", "merged lane")
	runGit(t, path, "push", "-u", "origin", branch)
	headOID := gitText(t, path, "rev-parse", "HEAD")
	runGit(t, fixture.primary, "merge", "--no-ff", "--no-edit", branch)
	runGit(t, fixture.primary, "push", "origin", "main")
	return path, headOID
}

func advanceRemoteMain(
	t *testing.T,
	fixture gitFixture,
	filename string,
) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "remote-main")
	runGit(t, filepath.Dir(path), "clone", fixture.remote, path)
	runGit(t, path, "config", "user.name", "Test User")
	runGit(t, path, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(path, filename), []byte("remote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, path, "add", filename)
	runGit(t, path, "commit", "-m", "advance remote main")
	runGit(t, path, "push", "origin", "main")
	return gitText(t, path, "rev-parse", "HEAD")
}

func canonicalTestLanePath(fixture gitFixture, branch string) string {
	return filepath.Join(
		fixture.worktreeRoot,
		"example",
		"project",
		filepath.FromSlash(branch),
	)
}

func findSyncLane(
	t *testing.T,
	report SyncReport,
	branch string,
) SyncLaneDecision {
	t.Helper()
	for _, decision := range report.Lanes {
		if decision.Branch == branch {
			return decision
		}
	}
	t.Fatalf("report has no lane for branch %s: %#v", branch, report.Lanes)
	return SyncLaneDecision{}
}

func findSyncLaneByPath(
	t *testing.T,
	report SyncReport,
	path string,
) SyncLaneDecision {
	t.Helper()
	for _, decision := range report.Lanes {
		if samePath(decision.Path, path) {
			return decision
		}
	}
	t.Fatalf("report has no lane for path %s: %#v", path, report.Lanes)
	return SyncLaneDecision{}
}

type syncState struct {
	Head        string
	Main        string
	OriginMain  string
	Branches    []string
	Worktrees   string
	Status      string
	ProjectDirs []string
}

func captureSyncState(t *testing.T, fixture gitFixture) syncState {
	t.Helper()
	state := syncState{
		Head:       gitText(t, fixture.primary, "rev-parse", "HEAD"),
		Main:       gitText(t, fixture.primary, "rev-parse", "refs/heads/main"),
		OriginMain: gitText(t, fixture.primary, "rev-parse", "refs/remotes/origin/main"),
		Worktrees:  gitText(t, fixture.primary, "worktree", "list", "--porcelain"),
		Status:     gitText(t, fixture.primary, "status", "--porcelain=v1", "--untracked-files=all"),
	}
	branchOutput := gitText(
		t,
		fixture.primary,
		"for-each-ref",
		"--format=%(refname):%(objectname)",
		"refs/heads",
	)
	state.Branches = strings.Split(branchOutput, "\n")
	sort.Strings(state.Branches)
	err := filepath.WalkDir(
		filepath.Join(fixture.worktreeRoot, "example", "project"),
		func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				if errorsIsNotExist(err) {
					return nil
				}
				return err
			}
			relative, relErr := filepath.Rel(fixture.worktreeRoot, path)
			if relErr != nil {
				return relErr
			}
			state.ProjectDirs = append(state.ProjectDirs, relative)
			return nil
		},
	)
	if err != nil && !errorsIsNotExist(err) {
		t.Fatal(err)
	}
	sort.Strings(state.ProjectDirs)
	return state
}

func errorsIsNotExist(err error) bool {
	return os.IsNotExist(err)
}

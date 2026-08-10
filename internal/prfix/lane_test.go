package prfix

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/worktree"
)

type fakePreparer struct {
	repair worktree.PullRequestRepair
	err    error
}

func (preparer fakePreparer) PreparePullRequestRepair(context.Context, string, int, bool) (worktree.PullRequestRepair, error) {
	return preparer.repair, preparer.err
}

func TestLaneResolverPinsExactWritableHeadAndDirtyState(t *testing.T) {
	commands := []string{}
	runner := scriptedRunner{run: func(directory, name string, args []string) ([]byte, error) {
		commands = append(commands, name+" "+strings.Join(args, " ")+" @ "+directory)
		switch args[0] {
		case "rev-parse":
			return []byte("abc\n"), nil
		case "symbolic-ref":
			return []byte("GH-9\n"), nil
		case "status":
			return []byte(" M internal/app.go\n?? notes.txt\n"), nil
		default:
			return nil, fmt.Errorf("unexpected git arguments")
		}
	}}
	resolver := LaneResolver{runner: runner, preparer: fakePreparer{repair: worktree.PullRequestRepair{
		PreparedWorktree: worktree.PreparedWorktree{Path: "/tmp/GH-9", Branch: "GH-9"},
		Repository:       "acme/app", Number: 9, URL: "https://github.com/acme/app/pull/9", HeadRefOID: "abc",
	}}}
	pullRequest := PullRequest{URL: "https://github.com/acme/app/pull/9", State: "OPEN", HeadRefName: "GH-9", HeadRefOID: "abc"}
	lane, err := resolver.Resolve(context.Background(), "/repo", Target{Repository{"acme", "app"}, 9}, pullRequest)
	if err != nil {
		t.Fatal(err)
	}
	if lane.WorktreePath != "/tmp/GH-9" || lane.PushTarget != "origin/GH-9" || lane.LocalHead != "abc" {
		t.Fatalf("lane = %#v", lane)
	}
	if strings.Join(lane.DirtyPaths, ",") != "internal/app.go,notes.txt" {
		t.Fatalf("dirty paths = %#v", lane.DirtyPaths)
	}
	if len(commands) != 3 {
		t.Fatalf("commands = %#v", commands)
	}
}

func TestLaneResolverRejectsHeadMismatchAndDetachedBranch(t *testing.T) {
	runner := scriptedRunner{run: func(_ string, _ string, args []string) ([]byte, error) {
		if args[0] == "rev-parse" {
			return []byte("local\n"), nil
		}
		return []byte("GH-9\n"), nil
	}}
	base := worktree.PullRequestRepair{
		PreparedWorktree: worktree.PreparedWorktree{Path: "/tmp/GH-9", Branch: "GH-9"},
		Repository:       "acme/app", HeadRefOID: "remote",
	}
	pullRequest := PullRequest{HeadRefName: "GH-9", HeadRefOID: "remote"}
	resolver := LaneResolver{runner: runner, preparer: fakePreparer{repair: base}}
	if _, err := resolver.Resolve(context.Background(), "/repo", Target{Repository{"acme", "app"}, 9}, pullRequest); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("head mismatch error = %v", err)
	}
	base.Branch = "PR-9"
	base.HeadRefOID = "local"
	pullRequest.HeadRefOID = "local"
	resolver.preparer = fakePreparer{repair: base}
	if _, err := resolver.Resolve(context.Background(), "/repo", Target{Repository{"acme", "app"}, 9}, pullRequest); err == nil || !strings.Contains(err.Error(), "durable") {
		t.Fatalf("detached branch error = %v", err)
	}
}

func TestDirtyOwnershipFailsClosedOnExcludedOverlap(t *testing.T) {
	lane := Lane{DirtyPaths: []string{"a.go", "local.txt"}}
	feedback := []Feedback{{Path: "a.go"}}
	if _, err := ApplyDirtyOwnership(lane, "exclude", feedback); err == nil || !strings.Contains(err.Error(), "overlaps") {
		t.Fatalf("overlap error = %v", err)
	}
	owned, err := ApplyDirtyOwnership(lane, "include", feedback)
	if err != nil || owned.DirtyOwnership != "include" {
		t.Fatalf("owned=%#v err=%v", owned, err)
	}
}

func TestParseDirtyPathsHandlesNULAndRenameRecords(t *testing.T) {
	status := " M path with space.go\x00R  new name.go\x00old name.go\x00?? notes.txt\x00"
	got := strings.Join(parseDirtyPaths(status), ",")
	if got != "new name.go,notes.txt,old name.go,path with space.go" {
		t.Fatalf("dirty paths = %s", got)
	}
}

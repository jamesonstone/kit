package prfix

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/worktree"
)

type WorktreePreparer interface {
	PreparePullRequestRepair(context.Context, string, int, bool) (worktree.PullRequestRepair, error)
}

type LaneResolver struct {
	runner   Runner
	preparer WorktreePreparer
}

func NewLaneResolver(runner Runner, output Output) LaneResolver {
	out, errOut := output.Stdout, output.Stderr
	if out == nil {
		out = io.Discard
	}
	if errOut == nil {
		errOut = io.Discard
	}
	return LaneResolver{runner: runner, preparer: worktree.NewApp(out, errOut)}
}

func (resolver LaneResolver) Resolve(
	ctx context.Context,
	cwd string,
	target Target,
	pullRequest PullRequest,
) (Lane, error) {
	prepared, err := resolver.preparer.PreparePullRequestRepair(ctx, cwd, target.Number, true)
	if err != nil {
		return Lane{}, fmt.Errorf("prepare writable PR-head lane: %w", err)
	}
	if !strings.EqualFold(prepared.Repository, target.Slug()) {
		return Lane{}, fmt.Errorf("prepared repository %s does not match target %s", prepared.Repository, target.Slug())
	}
	if prepared.HeadRefOID != pullRequest.HeadRefOID {
		return Lane{}, fmt.Errorf("pull request head changed from %s to %s during lane preparation",
			pullRequest.HeadRefOID, prepared.HeadRefOID)
	}
	if prepared.Branch != pullRequest.HeadRefName || strings.HasPrefix(strings.ToUpper(prepared.Branch), "PR-") {
		return Lane{}, fmt.Errorf("prepared branch %q is not the exact durable PR head %q", prepared.Branch, pullRequest.HeadRefName)
	}
	localHead, err := resolver.gitText(ctx, prepared.Path, "rev-parse", "HEAD")
	if err != nil {
		return Lane{}, fmt.Errorf("read repair lane HEAD: %w", err)
	}
	if localHead != pullRequest.HeadRefOID {
		return Lane{}, fmt.Errorf("repair lane HEAD %s does not match expected remote head %s",
			localHead, pullRequest.HeadRefOID)
	}
	branch, err := resolver.gitText(ctx, prepared.Path, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil || branch != pullRequest.HeadRefName {
		return Lane{}, fmt.Errorf("repair lane is not on exact head branch %q", pullRequest.HeadRefName)
	}
	statusOutput, err := resolver.runner.Run(ctx, prepared.Path, "git", "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return Lane{}, fmt.Errorf("inspect repair lane changes: %w", err)
	}
	status := strings.TrimSpace(strings.ReplaceAll(string(statusOutput), "\x00", "\n"))
	return Lane{
		Repository: target.Slug(), PRNumber: target.Number, PRURL: pullRequest.URL,
		HeadBranch: pullRequest.HeadRefName, ExpectedHead: pullRequest.HeadRefOID,
		LocalHead: localHead, WorktreePath: prepared.Path,
		PushTarget: "origin/" + pullRequest.HeadRefName, Created: prepared.Created,
		DirtyStatus: status, DirtyPaths: parseDirtyPaths(string(statusOutput)), DirtyOwnership: "none",
	}, nil
}

func (resolver LaneResolver) gitText(ctx context.Context, directory string, args ...string) (string, error) {
	output, err := resolver.runner.Run(ctx, directory, "git", args...)
	return strings.TrimSpace(string(output)), err
}

func parseDirtyPaths(status string) []string {
	seen := map[string]bool{}
	if strings.ContainsRune(status, '\x00') {
		parseNULDirtyPaths(status, seen)
		return sortedDirtyPaths(seen)
	}
	for _, line := range strings.Split(status, "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if _, destination, found := strings.Cut(path, " -> "); found {
			path = destination
		}
		path = filepath.ToSlash(strings.Trim(path, `"`))
		if path != "" {
			seen[path] = true
		}
	}
	return sortedDirtyPaths(seen)
}

func parseNULDirtyPaths(status string, seen map[string]bool) {
	records := strings.Split(strings.TrimSuffix(status, "\x00"), "\x00")
	for index := 0; index < len(records); index++ {
		record := records[index]
		if len(record) < 4 {
			continue
		}
		seen[filepath.ToSlash(record[3:])] = true
		if (strings.ContainsAny(record[:2], "RC")) && index+1 < len(records) {
			index++
			if records[index] != "" {
				seen[filepath.ToSlash(records[index])] = true
			}
		}
	}
}

func sortedDirtyPaths(seen map[string]bool) []string {
	result := make([]string, 0, len(seen))
	for path := range seen {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func ApplyDirtyOwnership(lane Lane, ownership string, feedback []Feedback) (Lane, error) {
	if len(lane.DirtyPaths) == 0 {
		lane.DirtyOwnership = "none"
		return lane, nil
	}
	if ownership != "include" && ownership != "exclude" {
		return Lane{}, fmt.Errorf("dirty repair lane requires explicit include or exclude ownership")
	}
	if ownership == "exclude" {
		feedbackPaths := map[string]bool{}
		for _, item := range feedback {
			if item.Path != "" {
				feedbackPaths[filepath.ToSlash(item.Path)] = true
			}
		}
		for _, path := range lane.DirtyPaths {
			if feedbackPaths[path] {
				return Lane{}, fmt.Errorf("excluded dirty path %s overlaps active review feedback", path)
			}
		}
	}
	lane.DirtyOwnership = ownership
	return lane, nil
}

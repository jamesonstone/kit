package cli

import (
	"fmt"
	"sort"
	"strings"
)

func resolveLoopReviewTarget(opts loopReviewOptions) (loopReviewTarget, error) {
	baseRef, err := resolveLoopReviewBase(opts.ProjectRoot, opts.BaseRef)
	if err != nil {
		return loopReviewTarget{}, err
	}

	files := map[string]bool{}
	for _, args := range [][]string{
		{"diff", "--name-only", baseRef + "...HEAD"},
		{"diff", "--name-only"},
		{"diff", "--cached", "--name-only"},
	} {
		output, err := reviewLoopRunner.Output(opts.ProjectRoot, "git", args...)
		if err != nil {
			return loopReviewTarget{}, fmt.Errorf("failed to inspect review diff with git %s: %w", strings.Join(args, " "), err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				files[line] = true
			}
		}
	}

	changedFiles := make([]string, 0, len(files))
	for path := range files {
		changedFiles = append(changedFiles, path)
	}
	sort.Strings(changedFiles)

	diffStat, _ := loopReviewOptionalGitOutput(opts.ProjectRoot, "diff", "--stat", baseRef+"...HEAD")
	workingStat, _ := loopReviewOptionalGitOutput(opts.ProjectRoot, "diff", "--stat")
	stagedStat, _ := loopReviewOptionalGitOutput(opts.ProjectRoot, "diff", "--cached", "--stat")

	return loopReviewTarget{
		BaseRef:        baseRef,
		ChangedFiles:   changedFiles,
		DiffStat:       diffStat,
		WorkingStat:    workingStat,
		StagedStat:     stagedStat,
		NoLocalChanges: len(changedFiles) == 0,
	}, nil
}

func resolveLoopReviewBase(projectRoot, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override), nil
	}
	for _, candidate := range []string{"origin/main", "main"} {
		if _, err := reviewLoopRunner.Output(projectRoot, "git", "rev-parse", "--verify", candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not resolve review base: tried origin/main then main; pass --base <ref> to choose a base")
}

func loopReviewOptionalGitOutput(projectRoot string, args ...string) (string, error) {
	output, err := reviewLoopRunner.Output(projectRoot, "git", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

package worktree

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type migration struct {
	source      string
	destination string
	branch      string
}

func (a *App) migrate(ctx context.Context, cwd string, args []string) error {
	apply := false
	switch {
	case len(args) == 0:
	case len(args) == 1 && args[0] == "--apply":
		apply = true
	default:
		return fmt.Errorf("usage: git wt migrate [--apply]")
	}
	if _, err := a.repository(ctx, cwd); err != nil {
		return err
	}
	root, err := a.baseRoot()
	if err != nil {
		return err
	}
	plans, err := a.migrationPlan(ctx, root)
	if err != nil {
		return err
	}
	if len(plans) == 0 {
		return a.writef("No legacy flat linked worktrees found beneath %s\n", root)
	}
	for _, plan := range plans {
		if !apply {
			if err := a.writef("WOULD MOVE\t%s\t%s\n", plan.source, plan.destination); err != nil {
				return err
			}
			continue
		}
		if err := a.mkdirAll(filepath.Dir(plan.destination), 0o755); err != nil {
			return fmt.Errorf("create destination parent for %s: %w", plan.destination, err)
		}
		if _, err := a.git(ctx, plan.source, "worktree", "move", plan.source, plan.destination); err != nil {
			return fmt.Errorf("move %s to %s: %w", plan.source, plan.destination, err)
		}
		if err := a.writef("MOVED\t%s\t%s\n", plan.source, plan.destination); err != nil {
			return err
		}
	}
	if !apply {
		return a.writef("Dry run only. Re-run with --apply after reviewing every destination.\n")
	}
	return nil
}

func (a *App) migrationPlan(ctx context.Context, root string) ([]migration, error) {
	children, err := a.readDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read worktree root %s: %w", root, err)
	}

	plans := make([]migration, 0)
	destinations := map[string]string{}
	for _, child := range children {
		if !child.IsDir() {
			continue
		}
		source := filepath.Join(root, child.Name())
		top, err := a.gitText(ctx, source, "rev-parse", "--show-toplevel")
		if err != nil || !samePath(top, source) {
			continue
		}
		gitDir, err := a.gitText(ctx, source, "rev-parse", "--path-format=absolute", "--git-dir")
		if err != nil {
			return nil, err
		}
		commonDir, err := a.gitText(ctx, source, "rev-parse", "--path-format=absolute", "--git-common-dir")
		if err != nil {
			return nil, err
		}
		if samePath(gitDir, commonDir) {
			continue
		}
		owner, name, err := a.projectIdentity(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("identify legacy worktree %s: %w", source, err)
		}
		branch, branchErr := a.gitText(ctx, source, "symbolic-ref", "--quiet", "--short", "HEAD")
		if branchErr != nil {
			match := regexp.MustCompile(`(?i)(PR-[1-9][0-9]*)$`).FindStringSubmatch(child.Name())
			if match == nil {
				return nil, fmt.Errorf("legacy worktree %s is detached and has no PR-<number> identity", source)
			}
			branch = strings.ToUpper(match[1])
		}
		repo := repository{
			owner:       strings.ToLower(owner),
			name:        strings.ToLower(name),
			projectRoot: filepath.Join(root, strings.ToLower(owner), strings.ToLower(name)),
		}
		destination, err := canonicalLanePath(repo, branch)
		if err != nil {
			return nil, fmt.Errorf("plan migration for %s: %w", source, err)
		}
		exists, err := a.pathExists(destination)
		if err != nil {
			return nil, fmt.Errorf("inspect migration destination %s: %w", destination, err)
		}
		if exists {
			return nil, fmt.Errorf("migration destination already exists: %s", destination)
		}
		if other, duplicate := destinations[destination]; duplicate {
			return nil, fmt.Errorf("legacy worktrees %s and %s map to the same destination %s", other, source, destination)
		}
		destinations[destination] = source
		plans = append(plans, migration{source: source, destination: destination, branch: branch})
	}
	return plans, nil
}

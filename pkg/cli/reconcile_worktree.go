package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/worktreeprep"
)

// reconcileBranch is the branch reconcile uses when run from a primary checkout.
const reconcileBranch = "kit-reconcile"

// reconcileTarget is where reconcile writes.
type reconcileTarget struct {
	projectRoot string
	worktree    bool // writes go to a linked worktree, not the invoking checkout
	path        string
	base        string
	created     bool
	// current marks an already-current primary checkout: nothing is written,
	// not even local-only files such as .env.
	current bool
}

var inspectReconcileWorktree = func(projectRoot string) (worktreeprep.Location, error) {
	return worktreeprep.New().Inspect(context.Background(), projectRoot)
}

var reconcileWorktreeRoot = func() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "worktrees"), nil
}

// resolveReconcileTarget keeps the primary checkout read-only: a writing
// reconcile run there migrates a linked worktree on branch kit-reconcile
// instead. Linked worktrees, non-Git projects, and dry runs use the project in
// place.
func resolveReconcileTarget(out io.Writer, projectRoot string, dryRun, upToDate bool) (reconcileTarget, error) {
	location, err := inspectReconcileWorktree(projectRoot)
	if err != nil {
		return reconcileTarget{}, fmt.Errorf("inspect reconcile worktree: %w", err)
	}
	if !location.InsideGit || !location.IsPrimary {
		return reconcileTarget{projectRoot: projectRoot}, nil
	}
	if upToDate {
		// Nothing to write: a preview plans in place and a writing run is a no-op.
		return reconcileTarget{projectRoot: projectRoot, current: !dryRun}, nil
	}
	if dryRun {
		_, err := fmt.Fprintf(out, "Preview of this checkout. A writing run applies the migration in the %s linked worktree, based on %s; changes that are not in %s are not included.\n", reconcileBranch, reconcileBaseRef(location.Path), reconcileBaseRef(location.Path))
		return reconcileTarget{projectRoot: projectRoot}, err
	}
	target, err := prepareReconcileWorktree(projectRoot, location)
	if err != nil {
		return reconcileTarget{}, err
	}
	verb := "Reusing"
	if target.created {
		verb = "Created"
	}
	_, err = fmt.Fprintf(out, "Primary checkout is read-only; %s worktree %s on branch %s.\n", strings.ToLower(verb), target.path, reconcileBranch)
	return target, err
}

func prepareReconcileWorktree(projectRoot string, location worktreeprep.Location) (reconcileTarget, error) {
	// Git reports the project's place in the repository itself, so a symlinked
	// working directory can never make the target escape the worktree.
	prefix, err := runGit(projectRoot, "rev-parse", "--show-prefix")
	if err != nil {
		return reconcileTarget{}, fmt.Errorf("resolve project path in repository: %w: %s", err, strings.TrimSpace(prefix))
	}
	relative := filepath.FromSlash(strings.TrimSuffix(strings.TrimSpace(prefix), "/"))
	existing, err := worktreeprep.New().WorktreeForBranch(context.Background(), location.Path, reconcileBranch)
	if err != nil {
		return reconcileTarget{}, fmt.Errorf("list worktrees: %w", err)
	}
	target := reconcileTarget{worktree: true, path: existing}
	switch {
	case existing != "" && sameResolvedPath(existing, location.PrimaryPath):
		return reconcileTarget{}, fmt.Errorf("the primary checkout is on branch %s; switch it to the default branch so reconcile can use a linked worktree", reconcileBranch)
	case existing != "":
		if top, err := runGit(existing, "rev-parse", "--show-toplevel"); err != nil || !sameResolvedPath(strings.TrimSpace(top), existing) {
			return reconcileTarget{}, fmt.Errorf("worktree %s for branch %s is missing or broken; run `git worktree prune` and rerun", existing, reconcileBranch)
		}
	default:
		if target.path, err = defaultReconcileWorktreePath(location.Path); err != nil {
			return reconcileTarget{}, err
		}
		if _, err := os.Lstat(target.path); err == nil {
			return reconcileTarget{}, fmt.Errorf("%s exists but is not the %s worktree; move it aside and rerun", target.path, reconcileBranch)
		}
		args := []string{"worktree", "add", target.path, reconcileBranch}
		if _, err := runGit(location.Path, "rev-parse", "--verify", "--quiet", "refs/heads/"+reconcileBranch); err != nil {
			target.base = reconcileBaseRef(location.Path)
			args = []string{"worktree", "add", "-b", reconcileBranch, target.path, target.base}
		}
		if out, err := runGit(location.Path, args...); err != nil {
			return reconcileTarget{}, fmt.Errorf("create reconcile worktree: %w: %s", err, strings.TrimSpace(out))
		}
		target.created = true
	}
	target.projectRoot = filepath.Join(target.path, relative)
	if rel, err := filepath.Rel(target.path, target.projectRoot); err != nil || strings.HasPrefix(rel, "..") {
		return reconcileTarget{}, fmt.Errorf("reconcile target %s is outside worktree %s", target.projectRoot, target.path)
	}
	if err := linkEnvironmentFiles(filepath.Join(location.Path, relative), target.projectRoot); err != nil {
		return reconcileTarget{}, err
	}
	return target, nil
}

func sameResolvedPath(left, right string) bool {
	resolve := func(path string) string {
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			return filepath.Clean(resolved)
		}
		return filepath.Clean(path)
	}
	return resolve(left) == resolve(right)
}

// defaultReconcileWorktreePath follows the lane layout
// ~/worktrees/<owner>/<repository>/<branch>, falling back to the checkout's
// directory name when origin is not a GitHub remote.
func defaultReconcileWorktreePath(primary string) (string, error) {
	root, err := reconcileWorktreeRoot()
	if err != nil {
		return "", fmt.Errorf("resolve worktree root: %w", err)
	}
	if remote, err := runGit(primary, "remote", "get-url", "origin"); err == nil {
		if owner, repo, err := parseGitHubRemoteURL(remote); err == nil {
			return filepath.Join(root, owner, repo, reconcileBranch), nil
		}
	}
	return filepath.Join(root, filepath.Base(primary), reconcileBranch), nil
}

// reconcileBaseRef prefers the locally known default branch of origin and
// never fetches.
func reconcileBaseRef(primary string) string {
	if ref, err := runGit(primary, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"); err == nil && strings.TrimSpace(ref) != "" {
		return strings.TrimSpace(ref)
	}
	return "HEAD"
}

// linkEnvironmentFiles links the primary checkout's local-only .env and .envrc
// into the worktree; it never copies or overwrites them.
func linkEnvironmentFiles(primary, worktree string) error {
	for _, name := range []string{envPath, envrcPath} {
		source := filepath.Join(primary, name)
		if info, err := os.Stat(source); err != nil || !info.Mode().IsRegular() {
			continue
		}
		destination := filepath.Join(worktree, name)
		if _, err := os.Lstat(destination); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.Symlink(source, destination); err != nil {
			return fmt.Errorf("link %s into the reconcile worktree: %w", name, err)
		}
	}
	return nil
}

func printReconcileWorktreeNextSteps(out io.Writer, target reconcileTarget, dryRun bool) {
	if dryRun {
		return
	}
	lines := []string{
		"",
		"Reconcile wrote to the linked worktree " + target.path + " (branch " + reconcileBranch + ").",
		"Next:",
		"  1. Review: git -C " + shellQuoteArgument(target.path) + " status && git -C " + shellQuoteArgument(target.path) + " diff",
		"  2. Commit the changes there and open a pull request from " + reconcileBranch + ".",
		"  3. After it merges, pull the default branch in the primary checkout and remove the worktree with `git worktree remove`.",
	}
	for _, line := range lines {
		_, _ = fmt.Fprintln(out, line)
	}
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// reconcileHasNoChanges plans in place without writing; a project that is
// already current needs no linked worktree.
func reconcileHasNoChanges(projectRoot string, opts initRefreshOptions) (bool, error) {
	opts.dryRun = true
	opts.outputOnly = true
	plan, err := buildInitRefreshPlan(context.Background(), projectRoot, opts)
	if err != nil {
		return false, err
	}
	return len(actionableRefreshChanges(projectRoot, plan.changes)) == 0, nil
}

// actionableRefreshChanges is the one definition of pending Kit-managed
// repository drift, shared by `kit reconcile` and `kit health`: planned writes
// and removals of repository-delivered paths. Local-only files, such as a
// clean clone's missing .env and .envrc, are not drift.
func actionableRefreshChanges(projectRoot string, changes []initRefreshFileChange) []initRefreshFileChange {
	paths := make([]string, 0, len(changes))
	for _, change := range changes {
		if change.result != instructionFileSkipped {
			paths = append(paths, change.relativePath)
		}
	}
	eligible := managedFileDeliveryEligiblePaths(projectRoot, paths)
	var actionable []initRefreshFileChange
	for _, change := range changes {
		if change.result != instructionFileSkipped && eligible[normalizeManagedFileDeliveryPath(change.relativePath)] {
			actionable = append(actionable, change)
		}
	}
	return actionable
}

package worktreeprep

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var environmentFileNames = []string{".env", ".envrc"}

type environmentLinkPlan struct {
	source      string
	destination string
	create      bool
}

func ensureEnvironmentLinks(sourceRoot, destinationRoot string, enabled bool) error {
	if !enabled {
		return nil
	}
	plans := make([]environmentLinkPlan, 0, len(environmentFileNames))
	for _, name := range environmentFileNames {
		plan, err := planEnvironmentLink(sourceRoot, destinationRoot, name)
		if err != nil {
			return err
		}
		plans = append(plans, plan)
	}
	created := make([]environmentLinkPlan, 0, len(plans))
	for _, plan := range plans {
		if !plan.create {
			continue
		}
		if err := os.Symlink(plan.source, plan.destination); err != nil {
			rollbackErr := rollbackEnvironmentLinks(created)
			if rollbackErr != nil {
				return fmt.Errorf("link environment file %s: %w; rollback: %v", plan.destination, err, rollbackErr)
			}
			return fmt.Errorf("link environment file %s: %w", plan.destination, err)
		}
		created = append(created, plan)
	}
	return nil
}

func planEnvironmentLink(sourceRoot, destinationRoot, name string) (environmentLinkPlan, error) {
	source, err := filepath.Abs(filepath.Join(sourceRoot, name))
	if err != nil {
		return environmentLinkPlan{}, fmt.Errorf("resolve source environment path: %w", err)
	}
	destination, err := filepath.Abs(filepath.Join(destinationRoot, name))
	if err != nil {
		return environmentLinkPlan{}, fmt.Errorf("resolve destination environment path: %w", err)
	}
	plan := environmentLinkPlan{source: source, destination: destination}
	if filepath.Clean(source) == filepath.Clean(destination) {
		return plan, nil
	}
	destinationInfo, err := os.Lstat(destination)
	if err == nil {
		if destinationInfo.Mode()&os.ModeSymlink == 0 {
			if name == ".envrc" {
				return plan, nil
			}
			return environmentLinkPlan{}, fmt.Errorf(
				"destination environment file already exists and is not a symlink: %s",
				destination,
			)
		}
		matches, err := environmentSymlinkMatches(destination, source)
		if err != nil {
			return environmentLinkPlan{}, err
		}
		if !matches {
			return environmentLinkPlan{}, fmt.Errorf(
				"destination environment symlink points somewhere unexpected: %s",
				destination,
			)
		}
		if _, err := os.Stat(destination); err != nil {
			return environmentLinkPlan{}, fmt.Errorf(
				"inspect destination environment symlink %s: %w",
				destination,
				err,
			)
		}
		sourceInfo, err := os.Stat(source)
		if err != nil {
			return environmentLinkPlan{}, fmt.Errorf("inspect source environment file %s: %w", source, err)
		}
		if !sourceInfo.Mode().IsRegular() {
			return environmentLinkPlan{}, fmt.Errorf("source environment file must be a regular file: %s", source)
		}
		return plan, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return environmentLinkPlan{}, fmt.Errorf("inspect destination environment file %s: %w", destination, err)
	}
	sourceInfo, err := os.Stat(source)
	if errors.Is(err, os.ErrNotExist) {
		return plan, nil
	}
	if err != nil {
		return environmentLinkPlan{}, fmt.Errorf("inspect source environment file %s: %w", source, err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return environmentLinkPlan{}, fmt.Errorf("source environment file must be a regular file: %s", source)
	}
	plan.create = true
	return plan, nil
}

func environmentSymlinkMatches(path, expectedSource string) (bool, error) {
	target, err := os.Readlink(path)
	if err != nil {
		return false, fmt.Errorf("read environment symlink %s: %w", path, err)
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	return samePath(target, expectedSource), nil
}

func rollbackEnvironmentLinks(plans []environmentLinkPlan) error {
	var rollbackErr error
	for index := len(plans) - 1; index >= 0; index-- {
		plan := plans[index]
		matches, err := environmentSymlinkMatches(plan.destination, plan.source)
		if err != nil {
			rollbackErr = errors.Join(rollbackErr, err)
			continue
		}
		if !matches {
			rollbackErr = errors.Join(rollbackErr, fmt.Errorf(
				"refusing to remove changed environment symlink %s",
				plan.destination,
			))
			continue
		}
		if err := os.Remove(plan.destination); err != nil {
			rollbackErr = errors.Join(rollbackErr, err)
		}
	}
	return rollbackErr
}

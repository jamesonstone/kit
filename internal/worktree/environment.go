package worktree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	environmentFileName   = ".env"
	environmentRCFileName = ".envrc"
)

var environmentFileNames = []string{
	environmentFileName,
	environmentRCFileName,
}

type environmentLinkPlan struct {
	source      string
	destination string
	message     string
	create      bool
}

func (a *App) ensureEnvironmentLinks(sourceRoot, destinationRoot string, enabled bool) error {
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
		linkCreated, err := a.applyEnvironmentLink(plan)
		if linkCreated {
			created = append(created, plan)
		}
		if err != nil {
			if rollbackErr := rollbackEnvironmentLinks(created); rollbackErr != nil {
				return fmt.Errorf(
					"%w; additionally failed to roll back environment links: %v",
					err,
					rollbackErr,
				)
			}
			return err
		}
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
			if name == environmentRCFileName {
				plan.message = fmt.Sprintf(
					"Preserved existing environment file at %s; no link was created.\n",
					destination,
				)
				return plan, nil
			}
			return environmentLinkPlan{}, fmt.Errorf(
				"destination environment file already exists and is not a symlink: %s",
				destination,
			)
		}
		matches, _, err := environmentSymlinkMatches(destination, source)
		if err != nil {
			return environmentLinkPlan{}, err
		}
		if !matches {
			return environmentLinkPlan{}, fmt.Errorf(
				"destination environment symlink points somewhere unexpected: %s",
				destination,
			)
		}
		if _, err := os.Stat(destination); errors.Is(err, os.ErrNotExist) {
			return environmentLinkPlan{}, fmt.Errorf(
				"destination environment symlink is broken: %s",
				destination,
			)
		} else if err != nil {
			return environmentLinkPlan{}, fmt.Errorf(
				"inspect destination environment symlink %s: %w",
				destination,
				err,
			)
		}
		plan.message = fmt.Sprintf("Environment link already present at %s\n", destination)
		return plan, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return environmentLinkPlan{}, fmt.Errorf(
			"inspect destination environment file %s: %w",
			destination,
			err,
		)
	}

	sourceInfo, err := os.Stat(source)
	if errors.Is(err, os.ErrNotExist) {
		plan.message = fmt.Sprintf(
			"No environment file found at %s; no %s link was created.\n",
			source,
			name,
		)
		return plan, nil
	}
	if err != nil {
		return environmentLinkPlan{}, fmt.Errorf(
			"inspect source environment file %s: %w",
			source,
			err,
		)
	}
	if !sourceInfo.Mode().IsRegular() {
		return environmentLinkPlan{}, fmt.Errorf(
			"source environment file must be a regular file: %s",
			source,
		)
	}
	plan.create = true
	plan.message = fmt.Sprintf("Linked %s -> %s\n", destination, source)
	return plan, nil
}

func (a *App) applyEnvironmentLink(plan environmentLinkPlan) (bool, error) {
	if plan.create {
		if err := os.Symlink(plan.source, plan.destination); err != nil {
			return false, fmt.Errorf(
				"link environment file %s to %s: %w",
				plan.destination,
				plan.source,
				err,
			)
		}
	}
	if plan.message == "" {
		return plan.create, nil
	}
	return plan.create, a.writef("%s", plan.message)
}

func rollbackEnvironmentLinks(plans []environmentLinkPlan) error {
	var rollbackErr error
	for index := len(plans) - 1; index >= 0; index-- {
		plan := plans[index]
		matches, _, err := environmentSymlinkMatches(plan.destination, plan.source)
		if err != nil {
			rollbackErr = errors.Join(rollbackErr, err)
			continue
		}
		if !matches {
			rollbackErr = errors.Join(
				rollbackErr,
				fmt.Errorf(
					"refusing to remove changed environment symlink %s",
					plan.destination,
				),
			)
			continue
		}
		if err := os.Remove(plan.destination); err != nil {
			rollbackErr = errors.Join(
				rollbackErr,
				fmt.Errorf("remove environment symlink %s: %w", plan.destination, err),
			)
		}
	}
	return rollbackErr
}

func environmentSymlinkMatches(path, expectedSource string) (bool, string, error) {
	target, err := os.Readlink(path)
	if err != nil {
		return false, "", fmt.Errorf("read environment symlink %s: %w", path, err)
	}
	resolvedTarget := target
	if !filepath.IsAbs(resolvedTarget) {
		resolvedTarget = filepath.Join(filepath.Dir(path), resolvedTarget)
	}
	expected, err := filepath.Abs(expectedSource)
	if err != nil {
		return false, "", fmt.Errorf("resolve expected environment source %s: %w", expectedSource, err)
	}
	return samePath(resolvedTarget, expected), target, nil
}

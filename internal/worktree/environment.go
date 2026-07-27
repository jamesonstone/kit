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

func (a *App) ensureEnvironmentLinks(sourceRoot, destinationRoot string, enabled bool) error {
	if !enabled {
		return nil
	}

	for _, name := range environmentFileNames {
		if err := a.ensureEnvironmentLink(sourceRoot, destinationRoot, name); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) ensureEnvironmentLink(sourceRoot, destinationRoot, name string) error {
	source, err := filepath.Abs(filepath.Join(sourceRoot, name))
	if err != nil {
		return fmt.Errorf("resolve source environment path: %w", err)
	}
	destination, err := filepath.Abs(filepath.Join(destinationRoot, name))
	if err != nil {
		return fmt.Errorf("resolve destination environment path: %w", err)
	}

	if filepath.Clean(source) == filepath.Clean(destination) {
		return nil
	}

	destinationInfo, err := os.Lstat(destination)
	if err == nil {
		if destinationInfo.Mode()&os.ModeSymlink == 0 {
			if name == environmentRCFileName {
				return a.writef("Preserved existing environment file at %s; no link was created.\n", destination)
			}
			return fmt.Errorf(
				"destination environment file already exists and is not a symlink: %s",
				destination,
			)
		}
		matches, _, err := environmentSymlinkMatches(destination, source)
		if err != nil {
			return err
		}
		if !matches {
			return fmt.Errorf(
				"destination environment symlink points somewhere unexpected: %s",
				destination,
			)
		}
		if _, err := os.Stat(destination); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("destination environment symlink is broken: %s", destination)
		} else if err != nil {
			return fmt.Errorf("inspect destination environment symlink %s: %w", destination, err)
		}
		return a.writef("Environment link already present at %s\n", destination)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect destination environment file %s: %w", destination, err)
	}

	sourceInfo, err := os.Stat(source)
	if errors.Is(err, os.ErrNotExist) {
		return a.writef("No environment file found at %s; no %s link was created.\n", source, name)
	}
	if err != nil {
		return fmt.Errorf("inspect source environment file %s: %w", source, err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return fmt.Errorf("source environment file must be a regular file: %s", source)
	}
	if err := os.Symlink(source, destination); err != nil {
		return fmt.Errorf("link environment file %s to %s: %w", destination, source, err)
	}
	return a.writef("Linked %s -> %s\n", destination, source)
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

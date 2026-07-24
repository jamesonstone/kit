package worktree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const environmentFileName = ".env"

func (a *App) ensureEnvironmentLink(sourceRoot, destinationRoot string, enabled bool) error {
	if !enabled {
		return nil
	}

	source, err := filepath.Abs(filepath.Join(sourceRoot, environmentFileName))
	if err != nil {
		return fmt.Errorf("resolve source environment path: %w", err)
	}
	destination, err := filepath.Abs(filepath.Join(destinationRoot, environmentFileName))
	if err != nil {
		return fmt.Errorf("resolve destination environment path: %w", err)
	}

	sourceInfo, err := os.Stat(source)
	if errors.Is(err, os.ErrNotExist) {
		return a.writef("No environment file found at %s; no .env link was created.\n", source)
	}
	if err != nil {
		return fmt.Errorf("inspect source environment file %s: %w", source, err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return fmt.Errorf("source environment file must be a regular file: %s", source)
	}
	if filepath.Clean(source) == filepath.Clean(destination) {
		return nil
	}

	destinationInfo, err := os.Lstat(destination)
	if err == nil {
		if destinationInfo.Mode()&os.ModeSymlink == 0 {
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
		return a.writef("Environment link already present at %s\n", destination)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect destination environment file %s: %w", destination, err)
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

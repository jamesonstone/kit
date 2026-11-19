package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/plugin"
)

const pluginManifestTempName = ".kit-plugin.yaml.tmp"

// copyMissingPluginFiles copies a plugin directory from the invoking checkout
// when the reconcile worktree has no manifest. git worktree add does not
// populate a submodule, and Kit does not clone or fetch one.
func copyMissingPluginFiles(fromRoot, toRoot string) error {
	fromPlugins := filepath.Join(fromRoot, filepath.FromSlash(plugin.RootDir))
	info, err := os.Lstat(fromPlugins)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("%s is not a plugin directory", plugin.RootDir)
	}
	entries, err := os.ReadDir(fromPlugins)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		source := filepath.Join(fromPlugins, entry.Name())
		sourceInfo, err := os.Lstat(source)
		if err != nil {
			return err
		}
		if sourceInfo.Mode()&os.ModeSymlink != 0 || !sourceInfo.IsDir() {
			return fmt.Errorf("%s is not a plugin directory", filepath.ToSlash(filepath.Join(plugin.RootDir, entry.Name())))
		}
		destination := filepath.Join(toRoot, filepath.FromSlash(plugin.RootDir), entry.Name())
		manifest := filepath.Join(destination, plugin.ManifestName)
		if _, err := os.Lstat(manifest); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := copyPluginTree(toRoot, source, destination); err != nil {
			_ = os.Remove(filepath.Join(destination, pluginManifestTempName))
			return err
		}
		if err := publishPluginManifest(toRoot, source, destination); err != nil {
			return err
		}
	}
	return nil
}

func copyPluginTree(worktree, fromDir, toDir string) error {
	return filepath.WalkDir(fromDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(fromDir, path)
		if err != nil {
			return err
		}
		if rel == plugin.ManifestName {
			return nil
		}
		target := filepath.Join(toDir, rel)
		if err := rejectWorktreeSymlink(worktree, target); err != nil {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s is a symlink", filepath.ToSlash(filepath.Join(plugin.RootDir, rel)))
		}
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", filepath.ToSlash(filepath.Join(plugin.RootDir, rel)))
		}
		return copyRegularFile(path, target, info.Mode().Perm())
	})
}

// publishPluginManifest renames the manifest into place only after the other
// plugin files have been copied. A missing manifest is what the next call uses
// to retry an interrupted transfer.
func publishPluginManifest(worktree, fromDir, toDir string) error {
	temp := filepath.Join(toDir, pluginManifestTempName)
	manifest := filepath.Join(toDir, plugin.ManifestName)
	if err := rejectWorktreeSymlink(worktree, temp); err != nil {
		return err
	}
	if err := copyRegularFile(filepath.Join(fromDir, plugin.ManifestName), temp, 0o644); err != nil {
		_ = os.Remove(temp)
		return err
	}
	if err := os.Rename(temp, manifest); err != nil {
		_ = os.Remove(temp)
		return err
	}
	return nil
}

// rejectWorktreeSymlink stops before a write that would follow a symlink out of
// the reconcile worktree. Components that do not exist yet are created later.
func rejectWorktreeSymlink(worktree, full string) error {
	rel, err := filepath.Rel(worktree, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s escapes the worktree", full)
	}
	current := worktree
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s is a symlink", filepath.ToSlash(rel))
		}
	}
	return nil
}

func copyRegularFile(from, to string, mode os.FileMode) (err error) {
	input, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	if err = os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	output, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := output.Close()
		if err == nil {
			err = closeErr
		}
	}()
	_, err = io.Copy(output, input)
	return err
}

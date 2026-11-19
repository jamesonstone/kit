package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/v3/internal/plugin"
)

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
		if err := copyPluginTree(source, destination); err != nil {
			return err
		}
	}
	return nil
}

func copyPluginTree(fromDir, toDir string) error {
	return filepath.WalkDir(fromDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(fromDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(toDir, rel)
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

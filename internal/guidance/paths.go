package guidance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
)

// Resolve finds the nearest project without changing the process directory.
func Resolve(path string) (string, error) {
	if path == "" {
		path = "."
	}
	start, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(start)
	if err != nil {
		return "", fmt.Errorf("inspect path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("inspect path is not a directory: %s", start)
	}
	// Normalize the caller's directory before establishing the project boundary.
	start, err = filepath.EvalSymlinks(start)
	if err != nil {
		return "", err
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		info, err := os.Lstat(filepath.Join(dir, config.ConfigFileName))
		if err == nil {
			if !info.Mode().IsRegular() {
				return "", fmt.Errorf("%s must be a regular file, not a symlink or directory", filepath.Join(dir, config.ConfigFileName))
			}
			return dir, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect project config: %w", err)
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	return "", fmt.Errorf("no kit in this backpack: %s is not a kit project", start)
}

// safePath rejects escapes and every symlink component before any file read.
func safePath(root *os.Root, path string) (string, error) {
	if !filepath.IsLocal(path) {
		return "", fmt.Errorf("path must stay inside the project")
	}
	path = filepath.Clean(path)
	current := ""
	for _, part := range strings.Split(path, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := root.Lstat(current)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlink guidance is not supported")
		}
	}
	return path, nil
}

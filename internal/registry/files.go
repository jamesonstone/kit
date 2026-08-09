package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

func ReadOptional(root, relativePath string) (string, bool, error) {
	path, err := confinedPath(root, relativePath)
	if err != nil {
		return "", false, err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", false, fmt.Errorf("managed path %s cannot be a symbolic link", relativePath)
	}
	if !info.Mode().IsRegular() {
		return "", false, fmt.Errorf("managed path %s must be a regular file", relativePath)
	}
	content, err := os.ReadFile(path)
	return string(content), true, err
}

func confinedPath(root, relativePath string) (string, error) {
	if !safeRelativePath(relativePath) {
		return "", fmt.Errorf("path %q must stay inside the project", relativePath)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	path := filepath.Join(rootAbs, filepath.FromSlash(relativePath))
	rel, err := filepath.Rel(rootAbs, path)
	if err != nil || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path %q escapes the project", relativePath)
	}
	return path, nil
}

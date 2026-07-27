package improve

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

func snapshotDir(root string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(root, func(filePath string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		files[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	return files, err
}

func changedFiles(before, after map[string]string) []string {
	seen := map[string]struct{}{}
	for filePath, hash := range after {
		if before[filePath] != hash {
			seen[filePath] = struct{}{}
		}
	}
	for filePath := range before {
		if _, ok := after[filePath]; !ok {
			seen[filePath] = struct{}{}
		}
	}
	var paths []string
	for filePath := range seen {
		paths = append(paths, filePath)
	}
	sort.Strings(paths)
	return paths
}

func allowedSurfaceViolations(changed, allowed []string) []string {
	if len(changed) == 0 || len(allowed) == 0 {
		return nil
	}
	var violations []string
	for _, filePath := range changed {
		if !matchesAllowedSurface(filePath, allowed) {
			violations = append(violations, filePath)
		}
	}
	return violations
}

func matchesAllowedSurface(filePath string, allowed []string) bool {
	filePath = filepath.ToSlash(filePath)
	for _, raw := range allowed {
		pattern := strings.TrimSpace(filepath.ToSlash(raw))
		if pattern == "" {
			continue
		}
		if pattern == filePath {
			return true
		}
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "**")
			if strings.HasPrefix(filePath, prefix) {
				return true
			}
		}
		if ok, _ := path.Match(pattern, filePath); ok {
			return true
		}
	}
	return false
}

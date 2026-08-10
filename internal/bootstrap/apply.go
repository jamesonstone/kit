package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/registry"
)

type externalSnapshot struct {
	path    string
	before  string
	exists  bool
	changed bool
	mode    os.FileMode
}

func Apply(plan Plan) error {
	userSnapshot, err := applyUserConfig(plan.UserConfig)
	if err != nil {
		return err
	}
	createdDirectories, err := applyDirectories(plan)
	if err != nil {
		return withUserRollback(err, restoreExternal(userSnapshot))
	}
	createdSensitive, err := applySensitive(plan)
	if err != nil {
		rollbackDirectories(plan.root, createdDirectories)
		return withUserRollback(err, restoreExternal(userSnapshot))
	}
	if err := registry.ApplyPlan(plan.root, plan.Registry); err != nil {
		rollbackSensitive(plan.root, createdSensitive)
		rollbackDirectories(plan.root, createdDirectories)
		return withUserRollback(err, restoreExternal(userSnapshot))
	}
	return nil
}

func applyUserConfig(disposition UserConfigDisposition) (externalSnapshot, error) {
	snapshot := externalSnapshot{path: disposition.Path, before: disposition.before, exists: disposition.exists, mode: 0o644}
	if disposition.Action == "" || disposition.Action == "none" {
		return snapshot, nil
	}
	if disposition.exists {
		info, err := os.Lstat(disposition.Path)
		if err != nil {
			return snapshot, fmt.Errorf("recheck user config: %w", err)
		}
		if !info.Mode().IsRegular() {
			return snapshot, fmt.Errorf("user config must be a regular file")
		}
		snapshot.mode = info.Mode().Perm()
	}
	current, err := os.ReadFile(disposition.Path)
	if os.IsNotExist(err) && !disposition.exists {
		current = nil
	} else if err != nil {
		return snapshot, fmt.Errorf("recheck user config: %w", err)
	}
	if string(current) != disposition.before {
		return snapshot, fmt.Errorf("user config changed after planning; rerun `kit init`")
	}
	if err := atomicWrite(disposition.Path, disposition.after, snapshot.mode); err != nil {
		return snapshot, fmt.Errorf("write user config: %w", err)
	}
	snapshot.changed = true
	return snapshot, nil
}

func applyDirectories(plan Plan) ([]string, error) {
	var created []string
	for _, directory := range plan.Directories {
		if directory.Action != "create" {
			continue
		}
		path := filepath.Join(plan.root, filepath.FromSlash(directory.Path))
		// #nosec G301 -- repository documentation directories must remain traversable.
		if err := os.MkdirAll(path, 0o755); err != nil {
			rollbackDirectories(plan.root, created)
			return nil, fmt.Errorf("create %s: %w", directory.Path, err)
		}
		created = append(created, path)
	}
	return created, nil
}

func applySensitive(plan Plan) ([]string, error) {
	var created []string
	for _, item := range plan.exclusive {
		path := filepath.Join(plan.root, filepath.FromSlash(item.path))
		// #nosec G304 -- item.path comes only from the fixed bootstrap environment allowlist.
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.FileMode(item.mode))
		if err != nil {
			rollbackSensitive(plan.root, created)
			return nil, fmt.Errorf("create %s without reading existing content: %w", item.path, err)
		}
		if _, err := file.WriteString(item.content); err != nil {
			_ = file.Close()
			_ = os.Remove(path)
			rollbackSensitive(plan.root, created)
			return nil, err
		}
		if err := file.Close(); err != nil {
			_ = os.Remove(path)
			rollbackSensitive(plan.root, created)
			return nil, err
		}
		created = append(created, item.path)
	}
	return created, nil
}

func atomicWrite(path, content string, mode os.FileMode) error {
	// #nosec G301 -- project and user config parents require normal traversal.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".kit-write-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if _, err := temp.WriteString(content); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Chmod(mode); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func restoreExternal(snapshot externalSnapshot) error {
	if snapshot.path == "" || !snapshot.changed {
		return nil
	}
	if !snapshot.exists {
		if err := os.Remove(snapshot.path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return atomicWrite(snapshot.path, snapshot.before, snapshot.mode)
}

func withUserRollback(primary, rollback error) error {
	if rollback == nil {
		return primary
	}
	return fmt.Errorf("%w; user config rollback failed: %v", primary, rollback)
}

func rollbackSensitive(root string, relativePaths []string) {
	for _, relative := range relativePaths {
		_ = os.Remove(filepath.Join(root, filepath.FromSlash(relative)))
	}
}

func rollbackDirectories(root string, paths []string) {
	for index := len(paths) - 1; index >= 0; index-- {
		path := paths[index]
		_ = os.Remove(path)
		for parent := filepath.Dir(path); parent != root && parent != filepath.Dir(parent); parent = filepath.Dir(parent) {
			if err := os.Remove(parent); err != nil {
				break
			}
		}
	}
}

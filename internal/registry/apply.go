package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

type appliedChange struct {
	change Change
	exists bool
	mode   os.FileMode
}

func ApplyPlan(root string, plan Plan) error {
	var applied []appliedChange
	for _, change := range plan.Changes {
		path, err := confinedPath(root, change.Path)
		if err != nil {
			rollback(root, applied)
			return err
		}
		current, exists, err := ReadOptional(root, change.Path)
		if err != nil {
			rollback(root, applied)
			return err
		}
		if current != change.Before {
			rollback(root, applied)
			return fmt.Errorf("%s changed after planning; rerun the command", change.Path)
		}
		mode := os.FileMode(0o644)
		if exists {
			if info, statErr := os.Stat(path); statErr == nil {
				mode = info.Mode().Perm()
			}
		}
		applied = append(applied, appliedChange{change: change, exists: exists, mode: mode})
		if err := applyFile(path, change, mode); err != nil {
			rollback(root, applied)
			return fmt.Errorf("apply %s: %w", change.Path, err)
		}
	}
	return nil
}

func applyFile(path string, change Change, mode os.FileMode) error {
	if change.Action == "delete" {
		return os.Remove(path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".kit-write-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if _, err := temp.WriteString(change.After); err != nil {
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

func rollback(root string, applied []appliedChange) {
	for index := len(applied) - 1; index >= 0; index-- {
		item := applied[index]
		path, err := confinedPath(root, item.change.Path)
		if err != nil {
			continue
		}
		if !item.exists {
			_ = os.Remove(path)
			continue
		}
		_ = applyFile(path, Change{Action: "update", After: item.change.Before}, item.mode)
	}
}

func RenderDiff(plan Plan) string {
	result := ""
	for _, change := range plan.Changes {
		result += "--- a/" + change.Path + "\n+++ b/" + change.Path + "\n@@ replacement @@\n"
		for _, line := range splitDiffLines(change.Before) {
			result += "-" + line + "\n"
		}
		for _, line := range splitDiffLines(change.After) {
			result += "+" + line + "\n"
		}
	}
	return result
}

func splitDiffLines(content string) []string {
	if content == "" {
		return nil
	}
	lines := []string{}
	start := 0
	for index, char := range content {
		if char == '\n' {
			lines = append(lines, content[start:index])
			start = index + 1
		}
	}
	if start < len(content) {
		lines = append(lines, content[start:])
	}
	return lines
}

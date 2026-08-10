package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/registry"
)

func planEnvironment(plan *Plan) error {
	files := []struct {
		path, strategy, content string
		mode                    uint32
	}{
		{".env", "empty-create-if-missing-sensitive", "", 0o600},
		{".envrc", "create-if-missing-no-trust", envrcStarter, 0o644},
	}
	for _, file := range files {
		path := filepath.Join(plan.root, file.path)
		_, err := os.Lstat(path)
		disposition := FileDisposition{
			Path: file.path, Strategy: file.strategy,
			State: "preserved", Action: "none",
		}
		if os.IsNotExist(err) {
			disposition.State, disposition.Action = "planned", "create"
			plan.exclusive = append(plan.exclusive, exclusiveCreate{
				path: file.path, content: file.content, mode: file.mode,
			})
		} else if err != nil {
			return fmt.Errorf("inspect %s without reading content: %w", file.path, err)
		}
		plan.Files = append(plan.Files, disposition)
	}
	return nil
}

func planDirectory(plan *Plan, relative string) error {
	path := filepath.Join(plan.root, filepath.FromSlash(relative))
	info, err := os.Stat(path)
	disposition := DirectoryDisposition{Path: relative, State: "preserved", Action: "none"}
	if os.IsNotExist(err) {
		disposition.State, disposition.Action = "planned", "create"
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("bootstrap directory %s is occupied by a non-directory", relative)
	}
	plan.Directories = append(plan.Directories, disposition)
	return nil
}

func planGitignore(plan *Plan) error {
	before, exists, err := registry.ReadOptional(plan.root, ".gitignore")
	if err != nil {
		return err
	}
	after := appendIgnorePatterns(before)
	disposition := FileDisposition{
		Path: ".gitignore", Strategy: "append-only", State: "preserved", Action: "none",
	}
	if after != before {
		disposition.State = "planned"
		disposition.Action = "update"
		if !exists {
			disposition.Action = "create"
		}
		plan.Registry.Changes = append(plan.Registry.Changes, registry.Change{
			Path: ".gitignore", Action: disposition.Action, Before: before, After: after,
		})
	}
	plan.Files = append(plan.Files, disposition)
	return nil
}

func appendIgnorePatterns(content string) string {
	existing := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			existing[line] = true
		}
	}
	var missing []string
	for _, pattern := range gitignorePatterns {
		if !existing[pattern] {
			missing = append(missing, pattern)
		}
	}
	if len(missing) == 0 {
		return content
	}
	result := strings.TrimRight(content, "\n")
	if strings.TrimSpace(result) != "" {
		result += "\n\n"
	}
	result += gitignoreHeader + "\n" + strings.Join(missing, "\n") + "\n"
	return result
}

func planManagedFile(plan *Plan, path, strategy string, render func(string) string) error {
	before, exists, err := registry.ReadOptional(plan.root, path)
	if err != nil {
		return err
	}
	after := render(before)
	disposition := FileDisposition{Path: path, Strategy: strategy, State: "preserved", Action: "none"}
	if before != after {
		disposition.State, disposition.Action = "planned", "update"
		if !exists {
			disposition.Action = "create"
		}
		plan.Registry.Changes = append(plan.Registry.Changes, registry.Change{
			Path: path, Action: disposition.Action, Before: before, After: after,
		})
	} else if managedContentCustomized(path, before) {
		disposition.State = registry.StateLocalCustom
	}
	plan.Files = append(plan.Files, disposition)
	return nil
}

func readmeContent(content string) string {
	if strings.TrimSpace(content) == "" {
		content = readmeStarter
	}
	content = upsertBoundedBlock(content, readmeBadgesStart, readmeBadgesEnd, readmeBadgesBlock)
	return upsertBoundedBlock(content, readmeMaintainersStart, readmeMaintainersEnd, readmeMaintainersBlock)
}

func constitutionContent(content string) string {
	if strings.TrimSpace(content) == "" {
		return constitutionStarter
	}
	return upsertBoundedBlock(content, constitutionStart, constitutionEnd, constitutionBaseline)
}

func upsertBoundedBlock(content, start, end, block string) string {
	hasStart, hasEnd := strings.Contains(content, start), strings.Contains(content, end)
	if hasStart || hasEnd {
		return content
	}
	content = strings.TrimRight(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if content == "" {
		return block + "\n"
	}
	return content + "\n\n" + block + "\n"
}

func managedContentCustomized(path, content string) bool {
	switch path {
	case "README.md":
		return strings.Contains(content, readmeBadgesStart) &&
			(!strings.Contains(content, readmeBadgesBlock) || !strings.Contains(content, readmeMaintainersBlock))
	case "docs/CONSTITUTION.md":
		return strings.Contains(content, constitutionStart) && !strings.Contains(content, constitutionBaseline)
	default:
		return false
	}
}

func planRoutingFiles(plan *Plan) error {
	for path, starter := range routingStarters {
		before, exists, err := registry.ReadOptional(plan.root, path)
		if err != nil {
			return err
		}
		disposition := FileDisposition{Path: path, Strategy: "bounded-routing", State: "preserved", Action: "none"}
		change := changeForPath(plan.Registry.Changes, path)
		if change != nil {
			disposition.State, disposition.Action = "planned", change.Action
			if !exists {
				change.After = registry.RoutingContent(starter, plan.Registry.Config.Registry.Artifacts)
			}
		} else if !exists {
			disposition.State, disposition.Action = "planned", "create"
			plan.Registry.Changes = append(plan.Registry.Changes, registry.Change{
				Path: path, Action: "create",
				After: registry.RoutingContent(starter, plan.Registry.Config.Registry.Artifacts),
			})
		}
		_ = before
		plan.Files = append(plan.Files, disposition)
	}
	return nil
}

func changeForPath(changes []registry.Change, path string) *registry.Change {
	for index := range changes {
		if changes[index].Path == path {
			return &changes[index]
		}
	}
	return nil
}

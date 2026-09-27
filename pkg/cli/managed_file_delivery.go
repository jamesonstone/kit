package cli

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const managedFileAbsentState = "absent"

type managedFileDeliverySnapshot struct {
	Path            string `json:"path"`
	Action          string `json:"action"`
	PreCommandState string `json:"pre_command_state"`
	ResultState     string `json:"result_state"`
}

type managedFileDeliveryBaselineEntry struct {
	content string
	exists  bool
}

func managedFileContentState(content string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(content)))
}

func normalizeManagedFileDeliveryPath(relativePath string) string {
	return filepath.ToSlash(filepath.Clean(relativePath))
}

func managedFileDeliveryPathWithinProject(relativePath string) bool {
	if strings.TrimSpace(relativePath) == "" || filepath.IsAbs(relativePath) {
		return false
	}
	relativePath = normalizeManagedFileDeliveryPath(relativePath)
	return relativePath != "." &&
		relativePath != ".." &&
		!strings.HasPrefix(relativePath, "../")
}

func managedFileDeliverySnapshotFromInitRefresh(
	projectRoot string,
	changes []initRefreshFileChange,
) []managedFileDeliverySnapshot {
	snapshot := make([]managedFileDeliverySnapshot, 0, len(changes))
	paths := make([]string, 0, len(changes))
	for _, change := range changes {
		if change.result != instructionFileSkipped {
			paths = append(paths, change.relativePath)
		}
	}
	eligible := managedFileDeliveryEligiblePaths(projectRoot, paths)
	for _, change := range changes {
		relativePath := normalizeManagedFileDeliveryPath(change.relativePath)
		if change.result == instructionFileSkipped || !eligible[relativePath] {
			continue
		}

		preCommandState := managedFileAbsentState
		if change.result != instructionFileCreated {
			preCommandState = managedFileContentState(change.before)
		}
		resultState := managedFileContentState(change.after)
		if change.result == instructionFileRemoved {
			resultState = managedFileAbsentState
		}
		snapshot = append(snapshot, managedFileDeliverySnapshot{
			Path:            relativePath,
			Action:          dryRunActionLabel(change.result),
			PreCommandState: preCommandState,
			ResultState:     resultState,
		})
	}
	return snapshot
}

func appendManagedFileDeliveryTransition(
	snapshot []managedFileDeliverySnapshot,
	relativePath string,
	before string,
	beforeExists bool,
	after string,
	afterExists bool,
) []managedFileDeliverySnapshot {
	relativePath = normalizeManagedFileDeliveryPath(relativePath)
	if beforeExists == afterExists && before == after {
		return snapshot
	}

	action := "update"
	preCommandState := managedFileAbsentState
	resultState := managedFileAbsentState
	if beforeExists {
		preCommandState = managedFileContentState(before)
	} else {
		action = "create"
	}
	if afterExists {
		resultState = managedFileContentState(after)
	} else {
		action = "remove"
	}
	return append(snapshot, managedFileDeliverySnapshot{
		Path:            relativePath,
		Action:          action,
		PreCommandState: preCommandState,
		ResultState:     resultState,
	})
}

func captureManagedFileDeliveryBaseline(
	projectRoot string,
	relativePaths []string,
) (map[string]managedFileDeliveryBaselineEntry, error) {
	baseline := make(map[string]managedFileDeliveryBaselineEntry, len(relativePaths))
	for _, relativePath := range relativePaths {
		if !managedFileDeliveryPathWithinProject(relativePath) {
			continue
		}
		relativePath = normalizeManagedFileDeliveryPath(relativePath)
		if _, captured := baseline[relativePath]; captured {
			continue
		}
		content, exists, err := readManagedFileDeliveryState(
			filepath.Join(projectRoot, filepath.FromSlash(relativePath)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to snapshot %s before command: %w", relativePath, err)
		}
		baseline[relativePath] = managedFileDeliveryBaselineEntry{
			content: content,
			exists:  exists,
		}
	}
	return baseline, nil
}

func managedFileDeliverySnapshotFromBaseline(
	projectRoot string,
	baseline map[string]managedFileDeliveryBaselineEntry,
) ([]managedFileDeliverySnapshot, error) {
	snapshot := make([]managedFileDeliverySnapshot, 0, len(baseline))
	paths := make([]string, 0, len(baseline))
	for relativePath := range baseline {
		paths = append(paths, relativePath)
	}
	eligible := managedFileDeliveryEligiblePaths(projectRoot, paths)
	for relativePath, before := range baseline {
		if !eligible[normalizeManagedFileDeliveryPath(relativePath)] {
			continue
		}
		after, afterExists, err := readManagedFileDeliveryState(
			filepath.Join(projectRoot, filepath.FromSlash(relativePath)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to snapshot %s after command: %w", relativePath, err)
		}
		snapshot = appendManagedFileDeliveryTransition(
			snapshot,
			relativePath,
			before.content,
			before.exists,
			after,
			afterExists,
		)
	}
	return snapshot, nil
}

func mergeManagedFileDeliverySnapshots(
	primary []managedFileDeliverySnapshot,
	secondary []managedFileDeliverySnapshot,
) []managedFileDeliverySnapshot {
	merged := make(map[string]managedFileDeliverySnapshot, len(primary)+len(secondary))
	for _, change := range secondary {
		change.Path = normalizeManagedFileDeliveryPath(change.Path)
		merged[change.Path] = change
	}
	for _, change := range primary {
		change.Path = normalizeManagedFileDeliveryPath(change.Path)
		merged[change.Path] = change
	}

	snapshot := make([]managedFileDeliverySnapshot, 0, len(merged))
	for _, change := range merged {
		snapshot = append(snapshot, change)
	}
	return snapshot
}

func readManagedFileDeliveryState(path string) (string, bool, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(content), true, nil
}

func managedFileDeliveryPathEligible(projectRoot, relativePath string) bool {
	return managedFileDeliveryEligiblePaths(projectRoot, []string{relativePath})[normalizeManagedFileDeliveryPath(relativePath)]
}

// managedFileDeliveryEligiblePaths reports which paths belong in a delivery:
// inside the project, not secret-like, and not ignored by Git. It runs one
// `git rev-parse` and one batched `git check-ignore` for all paths.
func managedFileDeliveryEligiblePaths(projectRoot string, relativePaths []string) map[string]bool {
	eligible := map[string]bool{}
	var candidates []string
	for _, relativePath := range relativePaths {
		if !managedFileDeliveryPathWithinProject(relativePath) {
			continue
		}
		relativePath = normalizeManagedFileDeliveryPath(relativePath)
		base := strings.ToLower(filepath.Base(relativePath))
		if base == ".env" ||
			base == ".envrc" ||
			strings.HasPrefix(base, ".env.") ||
			strings.HasSuffix(base, ".pem") ||
			strings.HasSuffix(base, ".key") ||
			strings.Contains(base, "credentials") {
			continue
		}
		candidates = append(candidates, relativePath)
	}
	if len(candidates) == 0 {
		return eligible
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return eligible
	}
	output, err := exec.Command(gitPath, "-C", projectRoot, "rev-parse", "--is-inside-work-tree").Output()
	if err != nil {
		// Outside Git every candidate is eligible, unless Git metadata exists
		// but cannot be read.
		if _, statErr := os.Lstat(filepath.Join(projectRoot, ".git")); statErr == nil || !os.IsNotExist(statErr) {
			return eligible
		}
		for _, path := range candidates {
			eligible[path] = true
		}
		return eligible
	}
	if strings.TrimSpace(string(output)) != "true" {
		return eligible
	}

	cmd := exec.Command(gitPath, "-C", projectRoot, "check-ignore", "--stdin", "-z")
	cmd.Stdin = strings.NewReader(strings.Join(candidates, "\x00") + "\x00")
	ignoredOutput, err := cmd.Output()
	var exitErr *exec.ExitError
	if err != nil && (!errors.As(err, &exitErr) || exitErr.ExitCode() != 1) {
		return eligible // unknown ignore state: include nothing
	}
	ignored := map[string]bool{}
	for _, path := range strings.Split(string(ignoredOutput), "\x00") {
		if path != "" {
			ignored[filepath.ToSlash(path)] = true
		}
	}
	for _, path := range candidates {
		if !ignored[path] {
			eligible[path] = true
		}
	}
	return eligible
}

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
	for _, change := range changes {
		relativePath := normalizeManagedFileDeliveryPath(change.relativePath)
		if change.result == instructionFileSkipped ||
			!managedFileDeliveryPathEligible(projectRoot, relativePath) {
			continue
		}

		preCommandState := managedFileAbsentState
		if change.result != instructionFileCreated {
			preCommandState = managedFileContentState(change.before)
		}
		snapshot = append(snapshot, managedFileDeliverySnapshot{
			Path:            relativePath,
			Action:          dryRunActionLabel(change.result),
			PreCommandState: preCommandState,
			ResultState:     managedFileContentState(change.after),
		})
	}
	return snapshot
}

func managedFileDeliverySnapshotFromScaffold(
	projectRoot string,
	plans []instructionFileWritePlan,
	cleanupPlans []instructionRemovalPlan,
) ([]managedFileDeliverySnapshot, error) {
	snapshot := make([]managedFileDeliverySnapshot, 0, len(plans)+len(cleanupPlans))
	for _, plan := range plans {
		relativePath := normalizeManagedFileDeliveryPath(plan.relativePath)
		if plan.result == instructionFileSkipped ||
			!managedFileDeliveryPathEligible(projectRoot, relativePath) {
			continue
		}

		preCommandState := managedFileAbsentState
		if plan.result != instructionFileCreated {
			content, err := readManagedFileDeliveryContent(plan.absolutePath)
			if err != nil {
				return nil, fmt.Errorf("failed to snapshot %s before instruction scaffolding: %w", plan.relativePath, err)
			}
			preCommandState = managedFileContentState(content)
		}
		snapshot = append(snapshot, managedFileDeliverySnapshot{
			Path:            relativePath,
			Action:          dryRunActionLabel(plan.result),
			PreCommandState: preCommandState,
			ResultState:     managedFileContentState(plan.content),
		})
	}

	for _, plan := range cleanupPlans {
		relativePath := normalizeManagedFileDeliveryPath(plan.relativePath)
		if !managedFileDeliveryPathEligible(projectRoot, relativePath) {
			continue
		}
		content, err := readManagedFileDeliveryContent(plan.absolutePath)
		if err != nil {
			return nil, fmt.Errorf("failed to snapshot %s before instruction cleanup: %w", plan.relativePath, err)
		}
		snapshot = append(snapshot, managedFileDeliverySnapshot{
			Path:            relativePath,
			Action:          "remove",
			PreCommandState: managedFileContentState(content),
			ResultState:     managedFileAbsentState,
		})
	}
	return snapshot, nil
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
	for relativePath, before := range baseline {
		if !managedFileDeliveryPathEligible(projectRoot, relativePath) {
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

func readManagedFileDeliveryContent(path string) (string, error) {
	content, exists, err := readManagedFileDeliveryState(path)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", os.ErrNotExist
	}
	return content, nil
}

func managedFileDeliveryPathEligible(projectRoot, relativePath string) bool {
	if !managedFileDeliveryPathWithinProject(relativePath) {
		return false
	}
	relativePath = normalizeManagedFileDeliveryPath(relativePath)
	base := strings.ToLower(filepath.Base(relativePath))
	if base == ".env" ||
		base == ".envrc" ||
		strings.HasPrefix(base, ".env.") ||
		strings.HasSuffix(base, ".pem") ||
		strings.HasSuffix(base, ".key") ||
		strings.Contains(base, "credentials") {
		return false
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return false
	}
	worktreeCheck := exec.Command(gitPath, "-C", projectRoot, "rev-parse", "--is-inside-work-tree")
	output, err := worktreeCheck.Output()
	if err != nil {
		if _, statErr := os.Lstat(filepath.Join(projectRoot, ".git")); statErr == nil || !os.IsNotExist(statErr) {
			return false
		}
		return true
	}
	if strings.TrimSpace(string(output)) != "true" {
		return false
	}

	cmd := exec.Command(gitPath, "-C", projectRoot, "check-ignore", "--quiet", "--", relativePath)
	err = cmd.Run()
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == 1
}

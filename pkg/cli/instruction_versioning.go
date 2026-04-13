package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/instructions"
)

const instructionScaffoldVersionUnknown = instructions.UnknownVersion

type instructionRemovalPlan struct {
	relativePath string
	absolutePath string
}

func resolveInstructionScaffoldVersionFlag(raw int) (int, bool, error) {
	if raw == 0 {
		return 0, false, nil
	}
	if !config.IsInstructionScaffoldVersionSupported(raw) {
		return 0, false, fmt.Errorf("--version must be 1 or 2")
	}

	return raw, true, nil
}

func detectInstructionScaffoldVersion(projectRoot string, cfg *config.Config) int {
	return instructions.DetectVersion(projectRoot, cfg)
}

func instructionVersionChangeRequiresForce(currentVersion, targetVersion int) bool {
	return currentVersion != instructionScaffoldVersionUnknown && currentVersion != targetVersion
}

func planInstructionVersionCleanup(
	projectRoot string,
	currentVersion, targetVersion int,
) ([]instructionRemovalPlan, error) {
	if currentVersion != config.InstructionScaffoldVersionTOC ||
		targetVersion != config.InstructionScaffoldVersionVerbose {
		return nil, nil
	}

	managed := make(map[string]bool)
	for _, support := range instructions.SupportDocs(config.InstructionScaffoldVersionTOC) {
		managed[support.RelativePath] = true
	}

	for _, root := range []string{"docs/agents", "docs/references"} {
		absoluteRoot := filepath.Join(projectRoot, root)
		if !document.Exists(absoluteRoot) {
			continue
		}

		err := filepath.WalkDir(absoluteRoot, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if path == absoluteRoot {
				return nil
			}

			relativePath, err := filepath.Rel(projectRoot, path)
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return fmt.Errorf(
					"found unexpected directory `%s` under the v2 docs tree; move or remove it before downgrading to --version 1",
					relativePath,
				)
			}
			if !managed[filepath.ToSlash(relativePath)] {
				return fmt.Errorf(
					"found extra file `%s` under the v2 docs tree; move or remove it before downgrading to --version 1",
					relativePath,
				)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	var plans []instructionRemovalPlan
	for _, support := range instructions.SupportDocs(config.InstructionScaffoldVersionTOC) {
		absolutePath := filepath.Join(projectRoot, support.RelativePath)
		if document.Exists(absolutePath) {
			plans = append(plans, instructionRemovalPlan{
				relativePath: support.RelativePath,
				absolutePath: absolutePath,
			})
		}
	}

	return plans, nil
}

func applyInstructionVersionCleanup(projectRoot string, plans []instructionRemovalPlan) (int, error) {
	removed := 0
	for _, plan := range plans {
		if err := os.Remove(plan.absolutePath); err != nil && !os.IsNotExist(err) {
			return removed, fmt.Errorf("failed to remove %s: %w", plan.relativePath, err)
		}
		removed++
	}

	for _, root := range []string{"docs/agents", "docs/references", "docs"} {
		absoluteRoot := filepath.Join(projectRoot, root)
		if err := os.Remove(absoluteRoot); err != nil &&
			!os.IsNotExist(err) &&
			!isDirectoryNotEmpty(err) {
			return removed, fmt.Errorf("failed to remove %s: %w", root, err)
		}
	}

	return removed, nil
}

func isDirectoryNotEmpty(err error) bool {
	return errors.Is(err, syscall.ENOTEMPTY)
}

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
)

type initRefreshFileChange struct {
	relativePath string
	absolutePath string
	before       string
	after        string
	result       instructionFileWriteResult
}

func applyInitRefreshFileChangesAtomically(changes []initRefreshFileChange) error {
	return applyInitRefreshFileChangesAtomicallyWithRollback(changes, rollbackInitRefreshFileChange)
}

func applyInitRefreshFileChangesAtomicallyWithRollback(
	changes []initRefreshFileChange,
	rollback func(initRefreshFileChange) error,
) error {
	applied := make([]initRefreshFileChange, 0, len(changes))
	for _, change := range changes {
		if err := applyInitRefreshFileChange(change); err != nil {
			var rollbackErrors []string
			for i := len(applied) - 1; i >= 0; i-- {
				if rollbackErr := rollback(applied[i]); rollbackErr != nil {
					rollbackErrors = append(rollbackErrors, fmt.Sprintf("failed to roll back %s: %v", applied[i].relativePath, rollbackErr))
				}
			}
			if len(rollbackErrors) > 0 {
				return fmt.Errorf("%w; rollback failed: %s", err, strings.Join(rollbackErrors, "; "))
			}
			return err
		}
		if change.result != instructionFileSkipped {
			applied = append(applied, change)
		}
	}
	return nil
}

func rollbackInitRefreshFileChange(change initRefreshFileChange) error {
	if change.result == instructionFileCreated {
		return os.Remove(change.absolutePath)
	}
	return document.Write(change.absolutePath, change.before)
}

func newInitRefreshFileChange(
	projectRoot string,
	relativePath string,
	before string,
	after string,
	result instructionFileWriteResult,
) *initRefreshFileChange {
	relativePath = filepath.ToSlash(relativePath)
	return &initRefreshFileChange{
		relativePath: relativePath,
		absolutePath: filepath.Join(projectRoot, filepath.FromSlash(relativePath)),
		before:       before,
		after:        after,
		result:       result,
	}
}

func applyInitRefreshFileChange(change initRefreshFileChange) error {
	if change.result == instructionFileSkipped {
		return nil
	}
	if err := document.Write(change.absolutePath, change.after); err != nil {
		return fmt.Errorf("failed to write %s: %w", change.relativePath, err)
	}
	return nil
}

package cli

import (
	"fmt"
	"os"
	"path/filepath"

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
	applied := make([]initRefreshFileChange, 0, len(changes))
	for _, change := range changes {
		if err := applyInitRefreshFileChange(change); err != nil {
			for i := len(applied) - 1; i >= 0; i-- {
				_ = rollbackInitRefreshFileChange(applied[i])
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

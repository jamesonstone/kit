package cli

import (
	"fmt"
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

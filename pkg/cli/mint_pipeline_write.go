package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/v3/internal/config"
)

// Recheck the planned snapshot at the individual write boundary, after earlier
// batch changes. Existing files use the same staged conditional replacement as
// provider config: bytes and file identity are checked again before replacement.
func applyMintWorkflowChange(change initRefreshFileChange) error {
	if change.result == instructionFileSkipped {
		return nil
	}
	return withMintWorkflowLock(change.absolutePath, func() error { return applyMintWorkflowChangeLocked(change) })
}

func applyMintWorkflowChangeLocked(change initRefreshFileChange) error {
	if change.result == instructionFileSkipped {
		return nil
	}
	if err := validateMintWorkflowChanges([]initRefreshFileChange{change}); err != nil {
		return err
	}
	if change.result == instructionFileCreated {
		return createMintWorkflowExclusive(change.absolutePath, change.after, func(file *os.File, content string) error { _, err := file.WriteString(content); return err })
	}

	if change.result != instructionFileUpdated {
		return fmt.Errorf("unsupported Mint workflow change")
	}
	if err := config.ApplyIssueConfig(change.absolutePath, []byte(change.before), []byte(change.after)); err != nil {
		return fmt.Errorf("conditional Mint workflow write %s: %w", change.relativePath, err)
	}
	return nil
}

// Never overwrite a human edit during rollback of an earlier workflow either.
func rollbackMintWorkflowChange(change initRefreshFileChange) error {
	if change.result == instructionFileCreated {
		return withMintWorkflowLock(change.absolutePath, func() error {
			return fmt.Errorf("retained created Mint workflow for review after batch failure: %s", change.absolutePath)
		})
	}

	reversed := change
	reversed.before, reversed.after = change.after, change.before
	return applyMintWorkflowChange(reversed)
}

// Publish a fully written file with an exclusive hard link. Unlike rename, Link
// cannot replace a destination that appeared during staging. Failed writes and
// closes leave only the private temporary file, which is always removed.
func createMintWorkflowExclusive(path, content string, stage func(*os.File, string) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".kit-mint-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(temp.Name()) }()
	if err := temp.Chmod(0644); err != nil {
		_ = temp.Close()
		return err
	}
	stageErr := stage(temp, content)
	closeErr := temp.Close()
	if stageErr != nil {
		return stageErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Link(temp.Name(), path); err != nil {
		return fmt.Errorf("create Mint workflow without replacement: %w", err)
	}
	return nil
}

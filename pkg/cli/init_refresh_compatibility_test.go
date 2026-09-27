package cli

import (
	"strings"
	"testing"
)

func TestRunInit_DiffRequiresDryRun(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initRefresh = true
		initDiff = true

		err := runInit(initCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "--diff requires --dry-run") {
			t.Fatalf("expected --diff without --dry-run error, got %v", err)
		}
	})
}

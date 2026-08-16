package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestRunInitRefresh_DryRunDiffPrintsPlannedChangesWithoutWriting(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	existing := "source_env .custom\n"
	writeFile(t, filepath.Join(tempDir, envrcPath), existing)

	var output string
	withInitFlags(t, func() {
		initRefresh = true
		initDryRun = true
		initDiff = true
		initForce = true
		initRefreshFiles = []string{envrcPath}

		output = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	for _, check := range []string{
		"diff --git a/.envrc b/.envrc",
		"--- a/.envrc",
		"+++ b/.envrc",
		"-source_env .custom",
		"+dotenv_if_exists",
		"Dry run complete. Planned Created: 0, Updated: 1, Merged: 0, Skipped: 0",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected dry-run diff output to contain %q, got:\n%s", check, output)
		}
	}

	content, err := os.ReadFile(filepath.Join(tempDir, envrcPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", envrcPath, err)
	}
	if string(content) != existing {
		t.Fatalf("dry run wrote %s; content = %q, want %q", envrcPath, content, existing)
	}
}

func TestRunInitRefresh_DryRunDoesNotWritePlannedRefresh(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionVerbose
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, agentsMDPath), templates.LegacyAgentsMD)

	withInitFlags(t, func() {
		initRefresh = true
		initDryRun = true
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	unchanged, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if unchanged.InstructionScaffoldVersion != config.InstructionScaffoldVersionVerbose {
		t.Fatalf("dry run updated config version = %d", unchanged.InstructionScaffoldVersion)
	}

	agentsContent, err := os.ReadFile(filepath.Join(tempDir, agentsMDPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentsMDPath, err)
	}
	if string(agentsContent) != templates.LegacyAgentsMD {
		t.Fatalf("dry run updated %s", agentsMDPath)
	}
	assertFileDoesNotExist(t, filepath.Join(tempDir, "docs", "agents", "README.md"))
}

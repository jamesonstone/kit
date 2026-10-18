package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/spf13/cobra"
)

func TestSharedMemoryInitOptIn(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			root := t.TempDir()
			setupInitHome(t)
			setWorkingDirectory(t, root)
			cmd := &cobra.Command{}
			bindIssueProviderFlags(cmd)
			bindSharedMemoryFlags(cmd)
			if enabled {
				for flag, value := range map[string]string{"shared-memory": "true", "shared-memory-destination": "https://example.com/memory", "shared-memory-guide": "https://example.com/guide"} {
					if err := cmd.Flags().Set(flag, value); err != nil {
						t.Fatal(err)
					}
				}
			}
			withInitFlags(t, func() {
				initOutputOnly = true
				_ = captureStdout(t, func() {
					if err := runInitForTest(cmd, nil); err != nil {
						t.Fatal(err)
					}
				})
			})
			cfg, err := config.Load(root)
			if err != nil || cfg.SharedMemory.Enabled != enabled {
				t.Fatalf("%+v %v", cfg, err)
			}
			for _, path := range instructionFiles(cfg) {
				if strings.Contains(readFile(t, filepath.Join(root, path)), sharedMemoryBegin) != enabled {
					t.Fatal(path)
				}
			}
		})
	}
}

func TestSharedMemoryReconcileDryRunAndForceConfig(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	cfg := defaultInitConfig()
	cfg.SharedMemory = testSharedMemory()
	if err := config.Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	original := "# AGENTS\n\nProject-specific instructions.\n"
	path := filepath.Join(root, agentsMDPath)
	writeFile(t, path, original)
	beforeConfig := readFile(t, filepath.Join(root, config.ConfigFileName))
	_ = captureStdout(t, func() {
		_, err := runInitRefreshWithSnapshot(root, initRefreshOptions{dryRun: true, diff: true, files: []string{agentsMDPath}})
		if err != nil {
			t.Fatal(err)
		}
	})
	if readFile(t, path) != original || readFile(t, filepath.Join(root, config.ConfigFileName)) != beforeConfig {
		t.Fatal("dry-run changed files")
	}
	_ = captureStdout(t, func() {
		_, err := runInitRefreshWithSnapshot(root, initRefreshOptions{outputOnly: true, files: []string{agentsMDPath}})
		if err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(readFile(t, path), sharedMemoryBegin) {
		t.Fatal("did not apply guidance")
	}
	changes, _, _, err := planRefreshInitInstructionArtifacts(root, initRefreshOptions{}, cfg, map[string]bool{agentsMDPath: true})
	if err != nil || len(changes) != 1 || changes[0].result != instructionFileSkipped {
		t.Fatalf("repeat: %+v %v", changes, err)
	}
	forced, _, err := initRefreshConfig(root, initRefreshOptions{force: true}, map[string]bool{config.ConfigFileName: true})
	if err != nil || forced.SharedMemory != cfg.SharedMemory {
		t.Fatalf("force discarded opt-in: %+v %v", forced, err)
	}
}

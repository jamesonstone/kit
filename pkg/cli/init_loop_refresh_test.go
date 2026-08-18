package cli

import (
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestRunInitRefresh_PreservesCustomLoopAgentConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	cfg := config.Default()
	cfg.Loop.Agent = config.LoopAgentConfig{
		Command: "custom-agent",
		Args:    []string{"run", "--stdin"},
	}
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{config.ConfigFileName}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if updated.Loop.Agent.Command != "custom-agent" {
		t.Fatalf("Loop.Agent.Command = %q, want custom-agent", updated.Loop.Agent.Command)
	}
	if !stringSlicesEqual(updated.Loop.Agent.Args, []string{"run", "--stdin"}) {
		t.Fatalf("Loop.Agent.Args = %v, want custom args", updated.Loop.Agent.Args)
	}
}

func TestRunInitRefresh_PreservesCustomCodexModel(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	wantArgs := []string{
		"--ask-for-approval", "never", "exec", "--model", "custom-model",
		"--sandbox", "workspace-write", "--ignore-user-config", "--color", "never", "-",
	}
	cfg := config.Default()
	cfg.Loop.Agent = config.LoopAgentConfig{Command: "codex", Args: append([]string(nil), wantArgs...)}
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{config.ConfigFileName}
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if updated.Loop.Agent.Command != "codex" || !stringSlicesEqual(updated.Loop.Agent.Args, wantArgs) {
		t.Fatalf("custom Codex model changed: %#v", updated.Loop.Agent)
	}
}

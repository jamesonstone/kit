package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunInitRefresh_BackfillsLoopReviewAgentConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	cfg := config.Default()
	cfg.Loop.Agent = config.LoopAgentConfig{}
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
	assertDefaultInitLoopAgent(t, updated)

	content, err := os.ReadFile(filepath.Join(tempDir, config.ConfigFileName))
	if err != nil {
		t.Fatalf("failed to read %s: %v", config.ConfigFileName, err)
	}
	for _, check := range []string{"loop:", "command: codex", "- exec", "- --model", "- gpt-5.6", "- workspace-write", "- --ignore-user-config"} {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected refreshed config to contain %q, got:\n%s", check, content)
		}
	}
}

func TestRunInitRefresh_UpgradesGeneratedGPT55LoopAgentConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	cfg := config.Default()
	cfg.Loop.Agent = config.LoopAgentConfig{
		Command: "codex",
		Args: []string{
			"--ask-for-approval", "never", "exec", "--model", "gpt-5.5",
			"--sandbox", "workspace-write", "--ignore-user-config", "--color", "never", "-",
		},
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
	assertDefaultInitLoopAgent(t, updated)
}

func TestRunInitRefresh_ReplacesPlaceholderLoopAgentConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	writeFile(t, filepath.Join(tempDir, config.ConfigFileName), `goal_percentage: 95
specs_dir: docs/specs
skills_dir: .agents/skills
constitution_path: docs/CONSTITUTION.md
allow_out_of_order: false
loop:
  min_confidence: 95
  max_iterations: 10
  agent:
    command: your-agent
    args:
      - run
      - --stdin
agents:
  - AGENTS.md
instruction_scaffold_version: 2
feature_naming:
  numeric_width: 4
  separator: "-"
`)

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
	assertDefaultInitLoopAgent(t, updated)
}

func TestRunInitRefresh_UpgradesGeneratedLoopAgentConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	cfg := config.Default()
	cfg.Loop.MinConfidence = 95
	cfg.Loop.MaxIterations = 10
	cfg.Loop.Agent = config.LoopAgentConfig{
		Command: "codex",
		Args: []string{
			"--ask-for-approval",
			"never",
			"exec",
			"--sandbox",
			"workspace-write",
			"--color",
			"never",
			"-",
		},
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
	assertDefaultInitLoopAgent(t, updated)
}

func TestRunInitRefresh_UpgradesGeneratedLoopAgentConfigWithoutModel(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	cfg := config.Default()
	cfg.Loop.MinConfidence = 95
	cfg.Loop.MaxIterations = 10
	cfg.Loop.Agent = config.LoopAgentConfig{
		Command: "codex",
		Args: []string{
			"--ask-for-approval",
			"never",
			"exec",
			"--sandbox",
			"workspace-write",
			"--ignore-user-config",
			"--color",
			"never",
			"-",
		},
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
	assertDefaultInitLoopAgent(t, updated)
}

func TestRunInitRefresh_UpgradesMisorderedGeneratedLoopAgentConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	stubRulesetRegistry(t)

	cfg := config.Default()
	cfg.Loop.MinConfidence = 95
	cfg.Loop.MaxIterations = 10
	cfg.Loop.Agent = config.LoopAgentConfig{
		Command: "codex",
		Args: []string{
			"exec",
			"--sandbox",
			"workspace-write",
			"--ask-for-approval",
			"never",
			"--color",
			"never",
			"-",
		},
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
	assertDefaultInitLoopAgent(t, updated)
}

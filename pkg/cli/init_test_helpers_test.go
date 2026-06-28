package cli

import (
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/templates"
)

func initTestSupportFileContent(relativePath string) string {
	for _, file := range templates.InstructionSupportFiles(config.InstructionScaffoldVersionTOC) {
		if file.RelativePath == relativePath {
			return file.Content
		}
	}
	return ""
}

func assertDefaultInitLoopAgent(t *testing.T, cfg *config.Config) {
	t.Helper()

	want := defaultInitLoopAgentConfig()
	if cfg.Loop.Agent.Command != want.Command {
		t.Fatalf("Loop.Agent.Command = %q, want %q", cfg.Loop.Agent.Command, want.Command)
	}
	if !stringSlicesEqual(cfg.Loop.Agent.Args, want.Args) {
		t.Fatalf("Loop.Agent.Args = %v, want %v", cfg.Loop.Agent.Args, want.Args)
	}
	if cfg.Loop.MaxIterations != config.DefaultLoopMaxIterations {
		t.Fatalf("Loop.MaxIterations = %d, want %d", cfg.Loop.MaxIterations, config.DefaultLoopMaxIterations)
	}
}

func stringSlicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func withInitFlags(t *testing.T, run func()) {
	t.Helper()

	originalCopy := initCopy
	originalOutputOnly := initOutputOnly
	originalRefresh := initRefresh
	originalForce := initForce
	originalDryRun := initDryRun
	originalDiff := initDiff
	originalRefreshFiles := initRefreshFiles

	t.Cleanup(func() {
		initCopy = originalCopy
		initOutputOnly = originalOutputOnly
		initRefresh = originalRefresh
		initForce = originalForce
		initDryRun = originalDryRun
		initDiff = originalDiff
		initRefreshFiles = originalRefreshFiles
	})

	initCopy = false
	initOutputOnly = false
	initRefresh = false
	initForce = false
	initDryRun = false
	initDiff = false
	initRefreshFiles = nil

	run()
}

func setupInitHome(t *testing.T) string {
	t.Helper()

	stubRulesetRegistry(t)

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	return homeDir
}

func downstreamCapabilitiesUsageRulesetForTest() string {
	return `---
kind: ruleset
slug: kit-capabilities-usage
description: Downstream Kit command discovery guidance.
status: active
registry_scope: downstream
applies_to:
  - kit
  - cli
  - command-discovery
read_policy_default: conditional
---

# Ruleset: kit-capabilities-usage

## Purpose

- Use ` + "`kit capabilities`" + ` for downstream command discovery.

## Applies When

- A downstream project needs to choose a Kit command.

## Rules

- Use ` + "`kit capabilities`" + ` for command discovery.
- Prefer ` + "`kit capabilities <command> --json`" + ` after narrowing the command.
- Do not maintain Kit's internal command catalog from a downstream project.

## Anti-Patterns

- Do not tell downstream projects to edit ` + "`pkg/cli/capabilities_catalog.go`" + `.

## Verification

- ` + "`kit capabilities <command> --json`" + ` describes the selected command.

## Examples

` + "```bash" + `
kit capabilities dispatch --json
kit capabilities loop review --json
` + "```" + `
`
}

package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func withInitFlags(t *testing.T, run func()) {
	t.Helper()

	originalMint := initMint
	originalCopy := initCopy
	originalOutputOnly := initOutputOnly
	originalRefresh := initRefresh
	originalForce := initForce
	originalDryRun := initDryRun
	originalDiff := initDiff
	originalRefreshFiles := initRefreshFiles

	t.Cleanup(func() {
		initMint = originalMint
		initCopy = originalCopy
		initOutputOnly = originalOutputOnly
		initRefresh = originalRefresh
		initForce = originalForce
		initDryRun = originalDryRun
		initDiff = originalDiff
		initRefreshFiles = originalRefreshFiles
	})

	initMint = false
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

// Test-only knobs for exercising the shared refresh engine (the one `kit
// reconcile` applies) in place, now that `kit init --refresh` is gone.
var (
	initRefresh      bool
	initForce        bool
	initDryRun       bool
	initDiff         bool
	initRefreshFiles []string
)

func runInitForTest(cmd *cobra.Command, args []string) error {
	if !initRefresh {
		return runInit(cmd, args)
	}
	if initDiff && !initDryRun {
		return fmt.Errorf("--diff requires --dry-run")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	_, err = runInitRefreshWithSnapshot(cwd, initRefreshOptions{
		force:      initForce,
		dryRun:     initDryRun,
		diff:       initDiff,
		files:      initRefreshFiles,
		outputOnly: initOutputOnly,
	})
	return err
}

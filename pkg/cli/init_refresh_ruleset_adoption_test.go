package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func TestRunInitRefresh_AdoptsExistingStatusOnlyRulesetAsManaged(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	registry := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	stubRulesetRegistry(t, registry)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	local := strings.Replace(registry.Content, "status: active", "status: optional", 1)
	writeFile(t, filepath.Join(tempDir, rulesetTarget(registry.Slug)), local)

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{rulesetTarget(registry.Slug)}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content, err := os.ReadFile(filepath.Join(tempDir, rulesetTarget(registry.Slug)))
	if err != nil {
		t.Fatalf("failed to read ruleset: %v", err)
	}
	if !strings.Contains(string(content), "status: optional") {
		t.Fatalf("expected local status to remain optional, got:\n%s", content)
	}
	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	artifact, ok := updated.RegistryArtifact(rulesetKind, registry.Slug)
	if !ok {
		t.Fatalf("expected registry artifact for %s", registry.Slug)
	}
	if artifact.State != registryArtifactStateManaged || artifact.InstalledHash != registry.NormalizedHash {
		t.Fatalf("artifact = %#v, want managed hash %s", artifact, registry.NormalizedHash)
	}
}

func TestRunInitRefresh_InstallsDownstreamCapabilitiesUsageRuleNotMaintainerRule(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	usage := registryRulesetWithContentForTest("kit-capabilities-usage", downstreamCapabilitiesUsageRulesetForTest(), "test-usage-commit")
	maintainer := registryRulesetForTest("command-capabilities", []string{"kit", "cli", "capabilities"})
	maintainer.Metadata.RegistryScope = rulesetRegistryScopeKitMaintainer
	maintainer.Content = strings.Replace(
		maintainer.Content,
		"## Rules",
		"## Rules\n\n- Update `pkg/cli/capabilities_catalog.go`.",
		1,
	)
	stubRulesetRegistry(t, usage, maintainer)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	usageContent, err := os.ReadFile(filepath.Join(tempDir, rulesetTarget("kit-capabilities-usage")))
	if err != nil {
		t.Fatalf("expected downstream usage ruleset to be installed: %v", err)
	}
	for _, check := range []string{
		"slug: kit-capabilities-usage",
		"kit capabilities <command> --json",
		"Do not maintain Kit's internal command catalog from a downstream project",
	} {
		if !strings.Contains(string(usageContent), check) {
			t.Fatalf("expected downstream usage ruleset to contain %q, got:\n%s", check, usageContent)
		}
	}
	if document.Exists(filepath.Join(tempDir, rulesetTarget("command-capabilities"))) {
		t.Fatalf("maintainer-only command-capabilities ruleset should not be installed in downstream refresh")
	}

	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if _, ok := updated.RegistryArtifact(rulesetKind, "kit-capabilities-usage"); !ok {
		t.Fatalf("expected registry state for downstream usage ruleset")
	}
	if _, ok := updated.RegistryArtifact(rulesetKind, "command-capabilities"); ok {
		t.Fatalf("did not expect registry state for maintainer-only ruleset")
	}
}

func TestRunInitRefresh_DryRunDiffReportsDownstreamCapabilitiesUsageRuleAdoption(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	usage := registryRulesetWithContentForTest("kit-capabilities-usage", downstreamCapabilitiesUsageRulesetForTest(), "test-usage-commit")
	stubRulesetRegistry(t, usage)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	var output string
	withInitFlags(t, func() {
		initRefresh = true
		initDryRun = true
		initDiff = true
		initRefreshFiles = []string{rulesetTarget("kit-capabilities-usage")}

		output = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	for _, check := range []string{
		"diff --git a/docs/references/rules/kit-capabilities-usage.md b/docs/references/rules/kit-capabilities-usage.md",
		"+slug: kit-capabilities-usage",
		"+- Use `kit capabilities` for command discovery.",
		"Dry run complete. Planned Created:",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected dry-run diff output to contain %q, got:\n%s", check, output)
		}
	}
	assertFileDoesNotExist(t, filepath.Join(tempDir, rulesetTarget("kit-capabilities-usage")))
}

func TestRunInitRefresh_AdoptsExistingCustomRulesetWithoutOverwriting(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	registry := registryRulesetForTest("work-lane-gating", []string{"workflow"})
	stubRulesetRegistry(t, registry)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	local := strings.Replace(registry.Content, "## Examples", "Local custom guidance.\n\n## Examples", 1)
	writeFile(t, filepath.Join(tempDir, rulesetTarget(registry.Slug)), local)

	var output string
	withInitFlags(t, func() {
		initRefresh = true
		initRefreshFiles = []string{rulesetTarget(registry.Slug)}

		output = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content, err := os.ReadFile(filepath.Join(tempDir, rulesetTarget(registry.Slug)))
	if err != nil {
		t.Fatalf("failed to read ruleset: %v", err)
	}
	if string(content) != local {
		t.Fatalf("custom ruleset was overwritten:\n%s", content)
	}
	if !strings.Contains(output, "local custom content") {
		t.Fatalf("expected local-custom note, got:\n%s", output)
	}
	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	artifact, ok := updated.RegistryArtifact(rulesetKind, registry.Slug)
	if !ok || artifact.State != registryArtifactStateLocalCustom {
		t.Fatalf("artifact = %#v, want local-custom", artifact)
	}
}

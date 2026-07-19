package cli

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunRegistryStatusExplicitOptOutSkipsRegistry(t *testing.T) {
	projectRoot := t.TempDir()
	managed := false
	cfg := config.Default()
	cfg.Health = &config.HealthConfig{Managed: &managed}
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	setWorkingDirectory(t, projectRoot)

	registryCalls := 0
	stubRulesetRegistryFunc(t, func(_ context.Context) ([]registryRuleset, error) {
		registryCalls++
		return nil, errors.New("registry should not be called")
	})

	cmd := registryStatusCommandForTest(t, true)
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runRegistryStatus(cmd, nil); err != nil {
		t.Fatalf("runRegistryStatus() error = %v", err)
	}
	if registryCalls != 0 {
		t.Fatalf("registry calls = %d, want 0", registryCalls)
	}

	var report registryStatusReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, out.String())
	}
	if report.State != statusKitManagedStateDisabled || report.Managed {
		t.Fatalf("report = %#v, want disabled unmanaged project", report)
	}
}

func TestRunRegistryStatusReportsRefreshAvailable(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	setWorkingDirectory(t, projectRoot)
	stubRulesetRegistry(t, registryRulesetForTest("safety-guardrails", []string{"git"}))

	cmd := registryStatusCommandForTest(t, false)
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runRegistryStatus(cmd, nil); err != nil {
		t.Fatalf("runRegistryStatus() error = %v", err)
	}
	if !strings.HasPrefix(out.String(), statusKitManagedStateRefreshAvailable+" (") {
		t.Fatalf("output = %q, want compact refresh state", out.String())
	}
}

func TestRunRegistryStatusReportsCurrentAfterRefreshConverges(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	setWorkingDirectory(t, projectRoot)
	stubRulesetRegistry(t)

	plan, err := buildInitRefreshPlan(context.Background(), projectRoot, initRefreshOptions{outputOnly: true})
	if err != nil {
		t.Fatalf("buildInitRefreshPlan() error = %v", err)
	}
	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		t.Fatalf("applyInitRefreshFileChangesAtomically() error = %v", err)
	}

	cmd := registryStatusCommandForTest(t, false)
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runRegistryStatus(cmd, nil); err != nil {
		t.Fatalf("runRegistryStatus() error = %v", err)
	}
	if out.String() != statusKitManagedStateCurrent+"\n" {
		t.Fatalf("output = %q, want current", out.String())
	}
}

func TestRunRegistryStatusReportsLocalCustomAttention(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	setWorkingDirectory(t, projectRoot)
	ruleset := registryRulesetForTest("safety-guardrails", []string{"git"})
	local := strings.Replace(ruleset.Content, "## Examples", "Local guidance.\n\n## Examples", 1)
	recordRulesetRegistryState(cfg, ruleset, registryArtifactStateLocalCustom, "", local)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, rulesetTarget(ruleset.Slug), local)
	stubRulesetRegistry(t, ruleset)

	cmd := registryStatusCommandForTest(t, true)
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runRegistryStatus(cmd, nil); err != nil {
		t.Fatalf("runRegistryStatus() error = %v", err)
	}

	var report registryStatusReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.State != statusKitManagedStateAttentionNeeded || report.Registry.LocalCustom != 1 {
		t.Fatalf("report = %#v, want one local-custom attention item", report)
	}
}

func TestRunRegistryStatusReportsUnknownWithoutFailing(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	setWorkingDirectory(t, projectRoot)
	stubRulesetRegistryError(t, errors.New("registry offline"))

	cmd := registryStatusCommandForTest(t, true)
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runRegistryStatus(cmd, nil); err != nil {
		t.Fatalf("runRegistryStatus() error = %v", err)
	}

	var report registryStatusReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.State != statusKitManagedStateUnknown || !strings.Contains(report.CheckError, "registry offline") {
		t.Fatalf("report = %#v, want unknown state with registry cause", report)
	}
}

func registryStatusCommandForTest(t *testing.T, jsonOutput bool) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	if jsonOutput {
		if err := cmd.Flags().Set("json", "true"); err != nil {
			t.Fatalf("Flags().Set(json) error = %v", err)
		}
	}
	return cmd
}

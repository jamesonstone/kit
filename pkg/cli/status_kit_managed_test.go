package cli

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestRunStatusJSONIncludesLocalKitManagedRefreshSummary(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	t.Setenv("HOME", t.TempDir())
	setWorkingDirectory(t, projectRoot)
	stubRulesetRegistry(t)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("all", false, "")
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("Flags().Set(json) error = %v", err)
	}
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := runStatus(cmd, nil); err != nil {
		t.Fatalf("runStatus() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(out.String()), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, out.String())
	}
	kitManaged, ok := payload["kit_managed"].(map[string]any)
	if !ok {
		t.Fatalf("expected kit_managed JSON summary, got %#v", payload)
	}
	if got, exists := kitManaged["sync_checked"]; exists {
		t.Fatalf("sync_checked should be omitted from status JSON, got %v", got)
	}
	if got := kitManaged["state"]; got != statusKitManagedStateRefreshAvailable {
		t.Fatalf("state = %v, want %q", got, statusKitManagedStateRefreshAvailable)
	}
}

func TestStatusManagedSummaryReportsRemoteRegistryRefresh(t *testing.T) {
	projectRoot := t.TempDir()
	setupInitHome(t)
	base := registryRulesetForTest("sample-guardrails", []string{"git", "github"})
	stubRulesetRegistry(t, registryRulesetWithContentForTest(base.Slug, strings.Replace(base.Content, "## Verification", "- Remote registry addition.\n\n## Verification", 1), "new-commit"))

	cfg := config.Default()
	recordRulesetRegistryState(cfg, base, registryArtifactStateManaged, base.NormalizedHash)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, rulesetTarget(base.Slug)), base.Content)

	summary, err := buildStatusKitManagedSummary(projectRoot, cfg)
	if err != nil {
		t.Fatalf("buildStatusKitManagedSummary() error = %v", err)
	}
	if summary.Registry.Missing != 0 {
		t.Fatalf("Missing = %d, want 0; summary=%#v", summary.Registry.Missing, summary.Registry)
	}
	if summary.Registry.Managed != 1 {
		t.Fatalf("Managed = %d, want 1; summary=%#v", summary.Registry.Managed, summary.Registry)
	}
	if summary.Registry.Total != 1 {
		t.Fatalf("Total = %d, want 1; summary=%#v", summary.Registry.Total, summary.Registry)
	}
	if summary.ManagedFiles.Planned == 0 {
		t.Fatalf("Planned = 0, want remote registry refresh to be planned; summary=%#v", summary)
	}
	if summary.State != statusKitManagedStateRefreshAvailable {
		t.Fatalf("State = %q, want %q", summary.State, statusKitManagedStateRefreshAvailable)
	}
	var foundRulesetUpdate bool
	for _, item := range summary.Items {
		if item.Path == rulesetTarget(base.Slug) && item.State == string(instructionFileUpdated) {
			foundRulesetUpdate = true
		}
	}
	if !foundRulesetUpdate {
		t.Fatalf("expected ruleset update item for %s, got %#v", rulesetTarget(base.Slug), summary.Items)
	}
}

func TestStatusManagedSummaryFallsBackWhenRegistryUnavailable(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	stubRulesetRegistryError(t, errors.New("network unavailable"))

	summary, err := buildStatusKitManagedSummary(projectRoot, cfg)
	if err != nil {
		t.Fatalf("buildStatusKitManagedSummary() error = %v", err)
	}
	if summary.State != statusKitManagedStateUnknown {
		t.Fatalf("State = %q, want %q; summary=%#v", summary.State, statusKitManagedStateUnknown, summary)
	}
	if !summary.ManagedFiles.Unchecked {
		t.Fatalf("ManagedFiles.Unchecked = false, want true; summary=%#v", summary.ManagedFiles)
	}
	if !strings.Contains(summary.ManagedFiles.CheckError, "network unavailable") {
		t.Fatalf("CheckError = %q, want network cause", summary.ManagedFiles.CheckError)
	}
	if summary.ManagedFiles.Planned != 0 {
		t.Fatalf("Planned = %d, want 0 when freshness is unchecked", summary.ManagedFiles.Planned)
	}
	if !strings.Contains(strings.Join(summary.NextActions, "\n"), "rerun `kit status`") {
		t.Fatalf("NextActions = %#v, want status retry guidance", summary.NextActions)
	}
}

func TestStatusManagedSummaryUsesBoundedRegistryContext(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	observedDeadline := false
	stubRulesetRegistryFunc(t, func(ctx context.Context) ([]registryRuleset, error) {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("rulesetRegistryFetcher context has no deadline")
		}
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > statusKitManagedRefreshTimeout {
			t.Fatalf("deadline remaining = %v, want within %v", remaining, statusKitManagedRefreshTimeout)
		}
		observedDeadline = true
		return nil, context.DeadlineExceeded
	})

	summary, err := buildStatusKitManagedSummary(projectRoot, cfg)
	if err != nil {
		t.Fatalf("buildStatusKitManagedSummary() error = %v", err)
	}
	if !observedDeadline {
		t.Fatal("expected rulesetRegistryFetcher to observe a bounded context")
	}
	if summary.State != statusKitManagedStateUnknown {
		t.Fatalf("State = %q, want %q; summary=%#v", summary.State, statusKitManagedStateUnknown, summary)
	}
	if !summary.ManagedFiles.Unchecked {
		t.Fatalf("ManagedFiles.Unchecked = false, want true; summary=%#v", summary.ManagedFiles)
	}
	if !strings.Contains(summary.ManagedFiles.CheckError, context.DeadlineExceeded.Error()) {
		t.Fatalf("CheckError = %q, want deadline cause", summary.ManagedFiles.CheckError)
	}
}

func TestRunStatusJSONReportsUncheckedKitManagedFilesWhenRegistryUnavailable(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	t.Setenv("HOME", t.TempDir())
	setWorkingDirectory(t, projectRoot)
	stubRulesetRegistryError(t, errors.New("network unavailable"))

	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("all", false, "")
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("Flags().Set(json) error = %v", err)
	}
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := runStatus(cmd, nil); err != nil {
		t.Fatalf("runStatus() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(out.String()), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, out.String())
	}
	kitManaged, ok := payload["kit_managed"].(map[string]any)
	if !ok {
		t.Fatalf("expected kit_managed JSON summary, got %#v", payload)
	}
	if got := kitManaged["state"]; got != statusKitManagedStateUnknown {
		t.Fatalf("state = %v, want %q", got, statusKitManagedStateUnknown)
	}
	managedFiles, ok := kitManaged["managed_files"].(map[string]any)
	if !ok {
		t.Fatalf("expected managed_files JSON summary, got %#v", kitManaged)
	}
	if got := managedFiles["unchecked"]; got != true {
		t.Fatalf("managed_files.unchecked = %v, want true", got)
	}
	if got, ok := managedFiles["check_error"].(string); !ok || !strings.Contains(got, "network unavailable") {
		t.Fatalf("managed_files.check_error = %#v, want network cause", managedFiles["check_error"])
	}
}

func TestStatusManagedNextActionsPrioritizesReconcile(t *testing.T) {
	summary := &statusKitManagedSummary{
		ManagedFiles: statusManagedFilesSummary{Planned: 1},
		Registry:     statusRegistrySummary{Conflicts: 1},
	}

	actions := strings.Join(statusKitManagedNextActions(summary), "\n")
	for _, want := range []string{"kit reconcile --dry-run --diff", "kit reconcile --force --file"} {
		if !strings.Contains(actions, want) {
			t.Fatalf("actions missing %q:\n%s", want, actions)
		}
	}
	// Actions name only current flags.
	for _, stale := range []string{"--include-files", "--output-only"} {
		if strings.Contains(actions, stale) {
			t.Fatalf("actions use stale flag %s:\n%s", stale, actions)
		}
	}
}

func stubRulesetRegistryError(t *testing.T, err error) {
	t.Helper()
	stubRulesetRegistryFunc(t, func(_ context.Context) ([]registryRuleset, error) {
		return nil, err
	})
}

func stubRulesetRegistryFunc(t *testing.T, fetch func(context.Context) ([]registryRuleset, error)) {
	t.Helper()
	previous := rulesetRegistryFetcher
	t.Cleanup(func() {
		rulesetRegistryFetcher = previous
	})
	rulesetRegistryFetcher = fetch
}

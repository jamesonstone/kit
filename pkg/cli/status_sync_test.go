package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunStatusJSONIncludesKitManagedSummaryWithoutRegistryFetch(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	t.Setenv("HOME", t.TempDir())
	setWorkingDirectory(t, projectRoot)

	previousFetcher := rulesetRegistryFetcher
	t.Cleanup(func() {
		rulesetRegistryFetcher = previousFetcher
	})
	rulesetRegistryFetcher = func(_ context.Context) ([]registryRuleset, error) {
		t.Fatal("default kit status must not fetch the registry")
		return nil, nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("all", false, "")
	cmd.Flags().Bool("sync", false, "")
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
	if got := kitManaged["sync_checked"]; got != false {
		t.Fatalf("sync_checked = %v, want false", got)
	}
	if got := kitManaged["state"]; got != statusKitManagedStateUnsynced {
		t.Fatalf("state = %v, want %q", got, statusKitManagedStateUnsynced)
	}
}

func TestStatusSyncDetectsRegistryUpdateAvailable(t *testing.T) {
	projectRoot := t.TempDir()
	setupInitHome(t)
	base := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	remoteContent := strings.Replace(base.Content, "## Verification", "- Remote registry addition.\n\n## Verification", 1)
	remote := registryRulesetWithContentForTest(base.Slug, remoteContent, "new-commit")
	stubRulesetRegistry(t, remote)

	cfg := config.Default()
	recordRulesetRegistryState(cfg, base, registryArtifactStateManaged, base.NormalizedHash, base.Content)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(projectRoot, rulesetTarget(base.Slug)), base.Content)

	summary, err := buildStatusKitManagedSummary(context.Background(), projectRoot, cfg, true)
	if err != nil {
		t.Fatalf("buildStatusKitManagedSummary() error = %v", err)
	}
	if !summary.SyncChecked || !summary.Registry.Checked {
		t.Fatalf("expected registry sync check, got %#v", summary.Registry)
	}
	if summary.Registry.UpdateAvailable != 1 {
		t.Fatalf("UpdateAvailable = %d, want 1; summary=%#v", summary.Registry.UpdateAvailable, summary.Registry)
	}
	if summary.State != statusKitManagedStateStale {
		t.Fatalf("State = %q, want %q", summary.State, statusKitManagedStateStale)
	}
}

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

func TestRunStatusJSONIncludesLocalKitManagedRefreshSummary(t *testing.T) {
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

func TestStatusManagedSummaryUsesLocalRegistryStateOnly(t *testing.T) {
	projectRoot := t.TempDir()
	setupInitHome(t)
	base := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	stubRulesetRegistry(t, registryRulesetWithContentForTest(base.Slug, strings.Replace(base.Content, "## Verification", "- Remote registry addition.\n\n## Verification", 1), "new-commit"))

	cfg := config.Default()
	recordRulesetRegistryState(cfg, base, registryArtifactStateManaged, base.NormalizedHash, base.Content)
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
	for _, item := range summary.Items {
		if item.State == "update-available" {
			t.Fatalf("status should not report remote update availability; item=%#v", item)
		}
	}
}

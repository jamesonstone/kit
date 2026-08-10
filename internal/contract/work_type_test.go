package contract

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestImplementationDeliveryRequiresExplicitWorkType(t *testing.T) {
	resolved := resolveFeature(t, writeContractProject(t), Hints{
		Workflows: []string{"implementation-delivery"},
	})
	assertWorkTypeBlocked(t, resolved, "implementation-delivery or feature hints require explicit work type feature or maintenance")
}

func TestFeatureWorkRequiresFeatureHint(t *testing.T) {
	resolved := resolveFeature(t, writeContractProject(t), Hints{
		WorkType: WorkTypeFeature, Workflows: []string{"implementation-delivery"},
	})
	assertWorkTypeBlocked(t, resolved, "feature work requires a canonical feature hint")
}

func TestFeatureHintRequiresExplicitWorkType(t *testing.T) {
	resolved := resolveFeature(t, writeContractProject(t), Hints{
		Feature: "0059-example", Workflows: []string{"delivery"},
	})
	assertWorkTypeBlocked(t, resolved, "implementation-delivery or feature hints require explicit work type feature or maintenance")
}

func TestMaintenanceWorkIsExplicitFeatureSpecExemption(t *testing.T) {
	resolved := resolveFeature(t, writeContractProject(t), Hints{
		WorkType: WorkTypeMaintenance, Workflows: []string{"implementation-delivery"},
	})
	if resolved.State != "ready" || resolved.Hints.WorkType != WorkTypeMaintenance {
		t.Fatalf("resolved state = %s, hints = %#v", resolved.State, resolved.Hints)
	}
	if resolved.FeatureSpec != nil {
		t.Fatalf("maintenance feature spec = %#v", resolved.FeatureSpec)
	}
}

func TestMaintenanceWorkRejectsFeatureHint(t *testing.T) {
	resolved := resolveFeature(t, writeContractProject(t), Hints{
		WorkType: WorkTypeMaintenance, Feature: "0059-example",
		Workflows: []string{"implementation-delivery"},
	})
	assertWorkTypeBlocked(t, resolved, "maintenance work cannot include a feature hint")
	if resolved.FeatureSpec != nil {
		t.Fatalf("contradictory maintenance feature spec = %#v", resolved.FeatureSpec)
	}
}

func TestResolveRejectsUnknownWorkType(t *testing.T) {
	resolved := resolveFeature(t, writeContractProject(t), Hints{
		WorkType: "enhancement", Workflows: []string{"implementation-delivery"},
	})
	assertWorkTypeBlocked(t, resolved, "work type must be feature or maintenance")
}

func TestResolvedContractSchemaPublishesWorkType(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "schemas", "resolved-contract-v1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `"work_type": { "enum": ["feature", "maintenance"] }`) {
		t.Fatal("resolved-contract schema does not publish the work type enum")
	}
}

func assertWorkTypeBlocked(t *testing.T, resolved Resolved, diagnostic string) {
	t.Helper()
	if resolved.State != "blocked" || !slices.Contains(resolved.Diagnostics, diagnostic) {
		t.Fatalf("resolved state = %s, diagnostics = %v", resolved.State, resolved.Diagnostics)
	}
}

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The orchestration rule keeps invariants only; it must not regrow a
// mandatory planning lifecycle, manifests, or model-profile routing.
func TestAgentTeamOrchestrationRuleKeepsInvariantsWithoutLifecycle(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "rules", "agent-team-orchestration.md"))
	if err != nil {
		t.Fatal(err)
	}
	rule := strings.Join(strings.Fields(string(content)), " ")
	for _, invariant := range []string{
		"The primary agent owns scope, integration, validation, delivery, and the final report",
		"Delegated agents never create issues, branches, commits, pushes, pull requests, comments, or merges",
		"Never let two agents write overlapping files concurrently",
		"Report only real topology",
	} {
		if !strings.Contains(rule, invariant) {
			t.Errorf("agent-team-orchestration rule missing invariant %q", invariant)
		}
	}
	for _, retired := range []string{"CAPABILITY_NEGOTIATING", "Capability Manifest", "Lane Manifest", "single-lane, because", "PLAN_READY"} {
		if strings.Contains(rule, retired) {
			t.Errorf("agent-team-orchestration rule reintroduced retired ceremony %q", retired)
		}
	}
	if words := len(strings.Fields(string(content))); words > 700 {
		t.Errorf("agent-team-orchestration rule has %d words; keep it a short invariant rule", words)
	}
}

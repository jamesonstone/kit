package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeOrchestrationRulesStayContextAwareAndConcise(t *testing.T) {
	tests := map[string][]string{
		"github-pr-merge.md": {
			"Use one complete current-state snapshot per consequential mutation or wave",
			"nodes coupled through a base, service, environment, database, migration",
			"Prefer event-driven waits or bounded backoff",
		},
		"infrastructure-change-approval.md": {
			"Build the dependency graph and infrastructure outline during analysis",
			"Infrastructure deletion, destruction, purge, destructive replacement",
			"isolate them as a separate task",
			"are not infrastructure-approval batches",
			"never authorize deletion or removal",
		},
		"agent-team-orchestration.md": {
			"lower-cost or lower-capability configuration",
			"keep graph,",
			"repair, recovery, and acceptance decisions",
			"event-driven waits or bounded backoff",
		},
		"agent-completion-output.md": {
			"smallest evidence set that proves each",
			"terminal node",
			"Do not include a chronological command log",
		},
	}

	for name, required := range tests {
		content, err := os.ReadFile(filepath.Join(
			"..", "..", "docs", "references", "rules", name,
		))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, want := range required {
			if !strings.Contains(string(content), want) {
				t.Errorf("%s missing %q", name, want)
			}
		}
	}
}

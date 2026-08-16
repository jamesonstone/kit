package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedInContextWorkflowsMatchEmbeddedArtifacts(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts) != 7 {
		t.Fatalf("workflow count = %d, want 7", len(artifacts))
	}
	for _, artifact := range artifacts {
		path := filepath.Join("..", "..", filepath.FromSlash(artifact.Path))
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read checked-in %s: %v", artifact.Path, err)
		}
		if string(content) != artifact.Content {
			t.Fatalf("checked-in workflow %s differs from embedded artifact", artifact.Path)
		}
	}
}

func TestPRFeedbackWorkflowUsesCapabilityAwareOrchestration(t *testing.T) {
	artifacts, err := ContextWorkflowArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range artifacts {
		if artifact.Slug != "pr-feedback-repair" {
			continue
		}
		for _, want := range []string{
			"Negotiate host-confirmed agent controls",
			"actual agents from logical and omitted lanes",
			"requested from effective profiles",
			"confirmed from unconfirmed parallelism",
			"replacement rebriefs",
			"fresh independent read-only verifier",
			"supervisor self-review",
		} {
			if !strings.Contains(artifact.Content, want) {
				t.Fatalf("PR-feedback workflow missing %q", want)
			}
		}
		for _, retired := range []string{"at most three", "never more than four"} {
			if strings.Contains(artifact.Content, retired) {
				t.Fatalf("PR-feedback workflow retained fixed-cap policy %q", retired)
			}
		}
		return
	}
	t.Fatal("embedded pr-feedback-repair workflow not found")
}

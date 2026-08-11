package templates

import (
	"os"
	"path/filepath"
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

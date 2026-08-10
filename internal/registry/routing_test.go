package registry

import (
	"strings"
	"testing"
)

func TestRoutingRequiresFeatureSpecBeforeSourceEdits(t *testing.T) {
	content := RoutingContent("project-owned\n", []ArtifactRecord{
		{Kind: KindRuleset, Slug: "feature-specification"},
		{Kind: KindWorkflow, Slug: "implementation-delivery"},
	})
	for _, required := range []string{
		"--workflow implementation-delivery --work-type feature --feature <feature>",
		"--work-type maintenance",
		"living V3 spec", "re-resolve before source edits",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("routing block is missing %q", required)
		}
	}
	if !strings.HasPrefix(content, "project-owned\n") {
		t.Fatal("routing changed surrounding project content")
	}
}

package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSOPGuideMatchesCheckedInScaffold(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", SOPGuidePath))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != SOPGuide {
		t.Fatal("checked-in SOP guide differs from its embedded template")
	}
	for _, marker := range []string{"docs/sops/<slug>/SOP.md", "standard-operating-procedures.md", "preserves existing", "execution engine"} {
		if !strings.Contains(SOPGuide, marker) {
			t.Errorf("SOP guide lacks %q", marker)
		}
	}
}

func TestSOPTemplateIsDraftAndContainsOperationalSections(t *testing.T) {
	_, example, ok := strings.Cut(SOPGuide, "```markdown\n---\n")
	if !ok {
		t.Fatal("missing SOP Markdown template")
	}
	header, body, ok := strings.Cut(example, "\n---\n")
	if !ok {
		t.Fatal("unterminated template front matter")
	}
	var metadata map[string]string
	if err := yaml.Unmarshal([]byte(header), &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata["kind"] != "sop" || metadata["status"] != "draft" {
		t.Fatalf("new procedures must be draft SOPs: %#v", metadata)
	}
	for _, key := range []string{"slug", "description", "owner"} {
		if strings.TrimSpace(metadata[key]) == "" {
			t.Errorf("template lacks %s", key)
		}
	}
	for _, section := range []string{"Purpose", "Applicability", "Prerequisites", "Procedure", "Verification", "Recovery"} {
		if !strings.Contains(body, "\n## "+section+"\n") {
			t.Errorf("template lacks %s", section)
		}
	}
	if strings.Contains(body, "kit sop") || strings.Contains(body, "kit workflow") {
		t.Fatal("template must not invent a procedure execution command")
	}
}

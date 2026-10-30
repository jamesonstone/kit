package templates

import (
	"strings"
	"testing"
)

func TestCanonicalProjectGuidanceStructure(t *testing.T) {
	for _, path := range []string{DesignPath, SecurityPath} {
		content, ok := ProjectGuidance(path)
		if !ok || !strings.HasPrefix(content, "# ") || !strings.Contains(content, "\n## ") {
			t.Fatalf("%s lacks the canonical document structure", path)
		}
	}
	if !strings.Contains(Security, "\n## Vulnerability reporting\n") {
		t.Fatal("canonical SECURITY.md lacks vulnerability reporting")
	}
}

package cli

import (
	"strings"
	"testing"
)

func assertFinalResponseContractHeadings(t *testing.T, output string, headings ...string) {
	t.Helper()

	if !strings.Contains(output, "## Final Response Contract") {
		t.Fatalf("expected output to contain final response contract")
	}
	if !strings.Contains(output, "repo-relative paths") {
		t.Fatalf("expected final response contract to require repo-relative paths")
	}

	lastIndex := strings.Index(output, "## Final Response Contract")
	for _, heading := range headings {
		marker := "### " + heading
		index := strings.Index(output, marker)
		if index == -1 {
			t.Fatalf("expected final response contract to contain heading %q", marker)
		}
		if index < lastIndex {
			t.Fatalf("expected heading %q to appear after previous contract heading", marker)
		}
		lastIndex = index
	}
}

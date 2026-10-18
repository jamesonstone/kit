package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

// Ticket writes are native agent actions. These are contract fixtures, not API
// execution tests: they pin the instructions actually embedded for downstreams.
func TestEmbeddedIssueProviderContract(t *testing.T) {
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var delivery, cleanup string
	for _, rule := range registry {
		switch rule.Slug {
		case "delivery":
			delivery = rule.Content
		case "work-cleanup":
			cleanup = rule.Content
		}
	}
	for _, marker := range []string{
		"omitted or `github`", "`linear` uses Linear only", "Never inherit an issue target from global",
		"already-authorized Linear", "do not install a connector", "Missing configuration, permission errors",
		"Never choose the first name match", "A failed search is not an empty search", "uncertain write",
		"original ask", "scope and exclusions", "acceptance criteria", "expected verification",
		"complete Markdown as `description`", "human assignee ID", "Read back the returned identifier",
		"authentication, permissions, transport, rate limits, GraphQL errors", "Never infer success from HTTP status alone",
		"mark it completed only after", "verified workflow states", "ENG-123", "Refs: https://linear.app/",
		"Preserve it in the squash commit description", "GitHub closing references remain unchanged",
	} {
		if !strings.Contains(delivery, marker) {
			t.Errorf("embedded delivery missing %q", marker)
		}
	}
	if !strings.Contains(cleanup, "Never pass Linear IDs to `gh issue`") {
		t.Fatal("cleanup ignores provider")
	}
	for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
		contract := templates.InstructionEntryFile(path)
		if !strings.Contains(contract, "issues.provider") || !strings.Contains(contract, "ENG-123") {
			t.Fatalf("%s lacks provider guidance", path)
		}
	}
}

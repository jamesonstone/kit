package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/document"
)

func TestRunRulesAddCreatesRuleset(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runRulesAdd(cmd, []string{"frontend-ui"}); err != nil {
		t.Fatalf("runRulesAdd() error = %v", err)
	}

	path := filepath.Join(projectRoot, "docs", "references", "rules", "frontend-ui.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected ruleset file: %v", err)
	}
	for _, check := range []string{
		"kind: ruleset",
		"slug: frontend-ui",
		"- frontend",
		"## Purpose",
		"## Applies When",
		"## Rules",
		"## Anti-Patterns",
		"## Verification",
		"## Examples",
	} {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected ruleset content to contain %q, got:\n%s", check, content)
		}
	}
	if !strings.Contains(out.String(), "Created ruleset frontend-ui") {
		t.Fatalf("expected create output, got %q", out.String())
	}
}

func TestApplicationArchitectureRegistryRulesetsAreValid(t *testing.T) {
	tests := []struct {
		slug       string
		appliesTo  []string
		bodyChecks []string
	}{
		{
			slug:      "backend-service-architecture",
			appliesTo: []string{"backend", "api", "route", "controller", "service", "repository", "persistence"},
			bodyChecks: []string{
				"route → controller → service → repository",
				"Do not let routes or controllers call repositories directly",
				"The service chooses the business transaction boundary",
				"These names satisfy the rule when their responsibilities match",
			},
		},
		{
			slug:      "frontend-application-architecture",
			appliesTo: []string{"frontend", "web", "route", "page", "component", "state-management", "api-client"},
			bodyChecks: []string{
				"Prefer feature ownership and inward dependency direction",
				"Keep compile-time dependencies pointing inward",
				"Do not let shared components fetch feature data",
				"Apply the frontend prompt profile separately",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			path := filepath.Join("..", "..", "docs", "references", "rules", tt.slug+".md")
			ruleset, err := parseRulesetFile(path)
			if err != nil {
				t.Fatalf("parseRulesetFile() error = %v", err)
			}
			if issues := validateRulesetDocument(ruleset, tt.slug); len(issues) > 0 {
				t.Fatalf("%s ruleset issues = %#v", tt.slug, issues)
			}
			if ruleset.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
				t.Fatalf("registry_scope = %q, want downstream", ruleset.Metadata.RegistryScope)
			}
			if ruleset.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyMust {
				t.Fatalf("read_policy_default = %q, want must", ruleset.Metadata.ReadPolicyDefault)
			}
			for _, appliesTo := range tt.appliesTo {
				if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
					t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
				}
			}
			for _, check := range tt.bodyChecks {
				if !strings.Contains(ruleset.Body, check) {
					t.Errorf("expected %s ruleset to contain %q", tt.slug, check)
				}
			}
		})
	}
}

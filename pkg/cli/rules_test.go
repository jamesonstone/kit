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

func TestSafetyGuardrailsRegistryRulesetRequiresAutonomousRecovery(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "safety-guardrails.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "safety-guardrails"); len(issues) > 0 {
		t.Fatalf("safety-guardrails ruleset issues = %#v", issues)
	}
	if ruleset.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyMust {
		t.Fatalf("read_policy_default = %q, want must", ruleset.Metadata.ReadPolicyDefault)
	}
	for _, check := range []string{
		"retry autonomously",
		"including `gh`",
		"Outside explicit repo-local approval gates, ask permission only before large-scale deletion or deleting sensitive files",
		"follow `infrastructure-change-approval` before mutation",
		"one plan-level confirmation per complete batch",
		"always obtain post-outline confirmation for deletion or removal",
		"do not frame this as permission for a routine retry",
		"`~/worktrees/<owner>/<repository>/<lane>`",
		"exact uppercase `GH-<number>`",
		"exact uppercase `PR-<number>`",
		"native `git worktree` commands and ordinary filesystem operations",
		"must not require `git-wt`",
		"link the clone's primary checkout repository-root `.env` and `.envrc`",
		"accept already-matching symlinks during reuse",
		"omit both links when isolation is required",
		"Never copy environment contents",
		"direnv approval",
		"restore them if removal fails",
		"runtime services, databases, ports, Temporal state",
		"Remove only an exact registered path",
		"default to a new worklane without asking",
		"explicit same-scope direction to continue an existing lane",
		"exact existing-PR lifecycle target set as continuation",
		"coordination or corrective pull requests for scope-preserving work",
		"exact primary/root checkout as read-only",
		"Never edit the primary checkout with a plan to move the",
		"diff later",
		"default new-worklane route is a mandatory permission boundary",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected safety-guardrails ruleset to contain %q", check)
		}
	}
	for _, forbidden := range []string{
		"Do not retry with mutation",
		"Surface the failure to the user and await instruction",
		"Do not create or use git worktrees for agent work",
		"`--no-link-env`",
		"Let GitWT",
	} {
		if strings.Contains(ruleset.Body, forbidden) {
			t.Fatalf("expected safety-guardrails ruleset to omit blanket stop behavior %q", forbidden)
		}
	}
}

func TestConstitutionCurationRegistryRulesetIsValid(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "constitution-curation.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "constitution-curation"); len(issues) > 0 {
		t.Fatalf("constitution-curation ruleset issues = %#v", issues)
	}
	if ruleset.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatalf("registry_scope = %q, want downstream", ruleset.Metadata.RegistryScope)
	}
	if ruleset.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyMust {
		t.Fatalf("read_policy_default = %q, want must", ruleset.Metadata.ReadPolicyDefault)
	}
	for _, check := range []string{
		"Treat the exact generated Constitution starter as a valid bootstrap state",
		"When no project-wide truth changed, leave the Constitution unchanged",
		"Resolve `kit context resolve --workflow repository-maintenance --json` when a broader semantic review is needed",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected constitution-curation ruleset to contain %q", check)
		}
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

func TestTestingAndEnvironmentValidationRegistryRulesetIsValid(t *testing.T) {
	const slug = "testing-and-environment-validation"
	path := filepath.Join("..", "..", "docs", "references", "rules", slug+".md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, slug); len(issues) > 0 {
		t.Fatalf("%s ruleset issues = %#v", slug, issues)
	}
	if ruleset.Metadata.RegistryScope != rulesetRegistryScopeDownstream {
		t.Fatalf("registry_scope = %q, want downstream", ruleset.Metadata.RegistryScope)
	}
	if ruleset.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyMust {
		t.Fatalf("read_policy_default = %q, want must", ruleset.Metadata.ReadPolicyDefault)
	}
	for _, appliesTo := range []string{
		"implementation",
		"testing",
		"validation",
		"ci",
		"deployment",
		"local",
		"production",
		"end-to-end",
		"live-integration",
		"browser-automation",
		"browser-testing",
	} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}
	normalizedBody := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"Confidence, Not Certainty",
		"supplement code-level tests and never",
		"end-to-end/",
		"live-integration/",
		"tmp/<UTC-date>/<stable-test-id>/<positive-run-number>/",
		"`output.txt`",
		"`result.json`",
		"`PASS`, `FAIL`, `PARTIAL`, `BLOCKED`",
		"tests/RUN_STATUS.md",
		"one current row per suite and environment",
		"kit-e2e-<project>-<environment>-<run-id>-<resource>[-<ordinal>]",
		"Cleanup must select both the `kit-e2e-` marker and exact run ID",
		"report `PARTIAL`, not complete end-to-end validation",
		"Never use customer data",
		"### Browser Automation Lifecycle",
		"For interactive browser work in Codex",
		"use `@Browser`; do not use `@Chrome`",
		"Playwright, Selenium, Cypress, or browser MCP tools unless the user explicitly requests it",
		"report the limitation instead of silently falling back",
		"explicit authorization for an external browser as bounded",
		"isolated automation-managed browser",
		"one uniquely named browser session per task or test run",
		"existing user-owned browser or another task's automation",
		"detach from the automation connection without closing the browser",
		"`finally`, `defer`, teardown hooks",
		"Cleanup must run on success, failure, cancellation, timeout",
		"A final `close` command at the end of a linear script is insufficient",
		"`pkill Chrome`, `killall`, “close all",
		"shared temporary directories",
		"Before claiming browser validation complete",
		"session, browser process, and automation daemon have exited",
		"cleanup or exit-verification failure as a validation failure",
		"Do not disable or weaken macOS code-signing protections",
		"Do not routinely delete Chrome code-sign clone directories",
	} {
		if !strings.Contains(normalizedBody, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}

	index, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "README.md"))
	if err != nil {
		t.Fatalf("read references index: %v", err)
	}
	if !strings.Contains(string(index), "| `testing-and-environment-validation` |") {
		t.Fatal("references index does not list testing-and-environment-validation")
	}
}

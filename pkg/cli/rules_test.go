package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/document"
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
		"Ask permission only before large-scale deletion or deleting sensitive files",
		"do not frame this as permission for a routine retry",
		"`~/worktrees/<owner>/<repository>/<lane>`",
		"exact uppercase `GH-<number>`",
		"exact uppercase `PR-<number>`",
		"native `git worktree` commands and ordinary filesystem operations",
		"must not require `git-wt`",
		"link the clone's primary checkout repository-root `.env`",
		"accept an already-matching symlink during reuse",
		"omit the link when isolation is required",
		"Never copy `.env` contents",
		"automatically share `.envrc`",
		"restore it if removal fails",
		"runtime services, databases, ports, Temporal state",
		"Remove only an exact registered path",
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
		"Treat project-refresh cadence as a trigger for reviewed semantic analysis",
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
	} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}
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
	} {
		if !strings.Contains(ruleset.Body, check) {
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

func TestGitHubPRDeliveryRulesetUsesAutonomousRecovery(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "github-pr-delivery.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "github-pr-delivery"); len(issues) > 0 {
		t.Fatalf("github-pr-delivery ruleset issues = %#v", issues)
	}
	for _, check := range []string{
		"retry autonomously",
		"another supported authenticated path such as `gh`",
		"without requesting routine retry permission",
		"Verify that no duplicate issue or PR was created",
		"Project-Oriented Worktree Delivery",
		"`~/worktrees/<owner>/<repository>/<lane>`",
		"native `git worktree` commands as the portable authority",
		"rules and reconciled guidance must not depend on them",
		"git worktree add -b",
		"Writable review repair must use the pull request's same-repository head branch",
		"symlink the clone's primary checkout repository-root `.env` by default",
		"never automatically share `.envrc`",
		"restore it if removal fails",
		"application startup, databases, port allocation, Temporal state",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected github-pr-delivery ruleset to contain %q", check)
		}
	}
	if strings.Contains(ruleset.Body, "stop and do not mutate to retry") {
		t.Fatal("expected github-pr-delivery ruleset to omit blanket mutation retry prohibition")
	}
	for _, forbidden := range []string{"`--no-link-env`", "`git wt repair", "`git wt path", "GitWT"} {
		if strings.Contains(ruleset.Body, forbidden) {
			t.Fatalf("github-pr-delivery must not depend on wrapper-specific policy %q", forbidden)
		}
	}
}

func TestGitHubPRDeliveryRulesetPreservesAdditionalScopeLane(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "github-pr-delivery.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "github-pr-delivery"); len(issues) > 0 {
		t.Fatalf("github-pr-delivery ruleset issues = %#v", issues)
	}
	for _, check := range []string{
		"Create or reuse a separate GitHub issue for the additional scope",
		"Keep the existing pull request head branch. Do not create a second branch or pull request",
		"Scope every new commit for the additional work to its own issue number",
		"append a separate `Closes #123` line",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected github-pr-delivery ruleset to contain %q", check)
		}
	}
}

func TestGitHubPRDeliveryRulesetSkipsCIForDocumentationOnlySquashMerges(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "references", "rules", "github-pr-delivery.md")
	ruleset, err := parseRulesetFile(path)
	if err != nil {
		t.Fatalf("parseRulesetFile() error = %v", err)
	}
	if issues := validateRulesetDocument(ruleset, "github-pr-delivery"); len(issues) > 0 {
		t.Fatalf("github-pr-delivery ruleset issues = %#v", issues)
	}
	for _, check := range []string{
		"complete branch and pull request diff",
		"`kit reconcile` is a candidate signal, not proof of eligibility",
		"documentation-only follow-up on a mixed pull request is not eligible",
		"inspect branch protection, repository rulesets, and other required-check policy",
		"Append the literal suffix `[skip ci]` to every qualifying commit title",
		"Append the same literal `[skip ci]` suffix to the pull request title",
		"apply only to workflows triggered by `push` and `pull_request`",
		"They do not suppress `pull_request_target` or other events",
		"Before selecting `Confirm squash and merge`",
		"generated squash commit title and body",
		"confirm the pull request title and HEAD commit message contain `[skip ci]`",
		"gh run list --commit \"$SQUASH_SHA\"",
	} {
		if !strings.Contains(ruleset.Body, check) {
			t.Fatalf("expected github-pr-delivery ruleset to contain %q", check)
		}
	}
}

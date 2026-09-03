package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/document"
)

func TestDeadlineModeRegistryRulesetIsValid(t *testing.T) {
	const slug = "deadline-mode"
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
	if ruleset.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyConditional {
		t.Fatalf("read_policy_default = %q, want conditional", ruleset.Metadata.ReadPolicyDefault)
	}
	for _, appliesTo := range []string{"coding-agent", "workflow", "testing", "implementation", "prioritization"} {
		if !slices.Contains(ruleset.Metadata.AppliesTo, appliesTo) {
			t.Errorf("applies_to = %#v, want %q", ruleset.Metadata.AppliesTo, appliesTo)
		}
	}

	normalized := strings.Join(strings.Fields(ruleset.Body), " ")
	for _, check := range []string{
		"Load this ruleset only when the user explicitly signals a real deadline",
		"Never proactively suggest or enter deadline mode unprompted",
		"Do not substitute repeated broad suites for focused evidence",
		"required merge readiness",
		"required destructive-effect fencing",
		"independent final review",
		"required post-deployment tests",
		"one final UI verification after every result",
		"Skip UI, browser, and operator walkthrough verification until every result",
		"Then run one final UI verification",
		"Deadline mode is an explicit, recorded, invariant-preserving supersession",
		"use `PARTIAL` or `SKIPPED`, never `PASS`",
		"Ask again, or fall back to the ordinary",
	} {
		if !strings.Contains(normalized, check) {
			t.Errorf("expected %s ruleset to contain %q", slug, check)
		}
	}
}

func TestDeadlineModeIsIntegratedWithRelatedRules(t *testing.T) {
	checks := map[string][]string{
		"docs/references/README.md": {
			"Use `rules/deadline-mode.md`",
			"| `deadline-mode` |",
		},
		"docs/agents/RLM.md": {
			"Load `docs/references/rules/deadline-mode.md` only when the user explicitly signals a real time constraint or deadline in-thread; never infer or proactively suggest deadline mode",
		},
		"docs/references/rules/testing-and-environment-validation.md": {
			"superseded only by an active, explicitly recorded",
			"`docs/references/rules/deadline-mode.md`",
			"until every result is delivered, then requires one final",
			"skip UI and browser walkthrough verification until",
		},
		"docs/references/rules/github-pr-merge.md": {
			"without interleaving UI or browser walkthrough verification",
			"run one final UI verification",
		},
		"docs/references/workflows/implementation-delivery.md": {
			"slug: deadline-mode",
			"required: false",
		},
	}
	for path, required := range checks {
		content, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		normalized := strings.Join(strings.Fields(string(content)), " ")
		for _, check := range required {
			if !strings.Contains(normalized, check) {
				t.Errorf("expected %s to contain %q", path, check)
			}
		}
	}
}

func TestDeadlineModeDoesNotAddHardGateOrConstitutionRoute(t *testing.T) {
	for _, path := range []string{
		"CLAUDE.md",
		"AGENTS.md",
		".github/copilot-instructions.md",
		"docs/agents/GUARDRAILS.md",
		"docs/CONSTITUTION.md",
	} {
		content, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(strings.ToLower(string(content)), "deadline-mode") {
			t.Errorf("%s unexpectedly references deadline-mode; this ruleset must stay conditional and pointer-loaded only", path)
		}
	}
}

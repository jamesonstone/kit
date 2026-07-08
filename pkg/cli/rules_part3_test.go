package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
	"github.com/spf13/cobra"
)

func TestRunRulesViewPrefersLocalRuleset(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	stubRulesetRegistry(t, registryRulesetForTest("safety-guardrails", []string{"git", "github"}))
	local := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	local.Content = strings.Replace(local.Content, "Description for safety-guardrails", "Local description", 1)
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "safety-guardrails.md"), local.Content)

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runRulesView(cmd, []string{"safety-guardrails"}); err != nil {
		t.Fatalf("runRulesView() error = %v", err)
	}

	for _, check := range []string{
		"Source: docs/references/rules/safety-guardrails.md",
		"Local description",
	} {
		if !strings.Contains(out.String(), check) {
			t.Fatalf("expected local view output to contain %q, got:\n%s", check, out.String())
		}
	}
}

func TestProjectRulesetRegistryFiltersMaintainerOnlyRules(t *testing.T) {
	usage := registryRulesetForTest("kit-capabilities-usage", []string{"kit", "cli"})
	maintainer := registryRulesetForTest("command-capabilities", []string{"kit", "cli"})
	maintainer.Metadata.RegistryScope = rulesetRegistryScopeKitMaintainer

	filtered := projectRulesetRegistry([]registryRuleset{usage, maintainer})
	if len(filtered) != 1 || filtered[0].Slug != "kit-capabilities-usage" {
		t.Fatalf("filtered registry = %#v, want only downstream usage rule", filtered)
	}
}

func TestRunRulesAddRegistrySelectorDeactivatesExistingRuleset(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	registry := registryRulesetForTest("work-lane-gating", []string{"workflow"})
	stubRulesetRegistry(t, registry)
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "work-lane-gating.md"), registry.Content)

	output := withStdin(t, "1\n", func() string {
		return captureStdout(t, func() {
			if err := runRulesAdd(&cobra.Command{}, nil); err != nil {
				t.Fatalf("runRulesAdd() error = %v", err)
			}
		})
	})

	content, err := os.ReadFile(filepath.Join(projectRoot, "docs", "references", "rules", "work-lane-gating.md"))
	if err != nil {
		t.Fatalf("expected local ruleset file: %v", err)
	}
	if !strings.Contains(string(content), "status: optional") {
		t.Fatalf("expected deactivated optional status, got:\n%s", content)
	}
	if !strings.Contains(output, "Deactivated: 1") {
		t.Fatalf("expected deactivate summary, got:\n%s", output)
	}
}

func TestRunRulesAddRegistrySelectorReactivatesModifiedRulesetWithoutOverwriting(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	registry := registryRulesetForTest("github-pr-delivery", []string{"github"})
	stubRulesetRegistry(t, registry)

	local := strings.Replace(registry.Content, "status: active", "status: optional", 1)
	local = strings.Replace(local, "## Examples", "Custom local guidance.\n\n## Examples", 1)
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "github-pr-delivery.md"), local)

	output := withStdin(t, "1\n", func() string {
		return captureStdout(t, func() {
			if err := runRulesAdd(&cobra.Command{}, nil); err != nil {
				t.Fatalf("runRulesAdd() error = %v", err)
			}
		})
	})

	content, err := os.ReadFile(filepath.Join(projectRoot, "docs", "references", "rules", "github-pr-delivery.md"))
	if err != nil {
		t.Fatalf("expected local ruleset file: %v", err)
	}
	for _, check := range []string{"status: active", "Custom local guidance."} {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected local content to contain %q, got:\n%s", check, content)
		}
	}
	if !strings.Contains(output, "LOCAL-CUSTOM") || !strings.Contains(output, "Activated: 1") {
		t.Fatalf("expected modified activation output, got:\n%s", output)
	}
}

func TestNormalizedRulesetHashIgnoresStatusOnlyChanges(t *testing.T) {
	registry := registryRulesetForTest("status-only", []string{"git"})
	local := strings.Replace(registry.Content, "status: active", "status: optional", 1)

	registryHash, err := normalizedRulesetContentHash(registry.Content, registry.Metadata.Status)
	if err != nil {
		t.Fatalf("registry hash error: %v", err)
	}
	localHash, err := normalizedRulesetContentHash(local, registry.Metadata.Status)
	if err != nil {
		t.Fatalf("local hash error: %v", err)
	}
	if registryHash != localHash {
		t.Fatalf("status-only hash drift: registry %s local %s", registryHash, localHash)
	}
}

func TestRulesetRegistrySectionArtifactsUseHeadingPathsAndSkipFences(t *testing.T) {
	registry := registryRulesetForTest("section-keys", []string{"git"})
	content := registry.Content + "\n## Parent One\n\n### Duplicate\n\none\n\n## Parent Two\n\n### Duplicate\n\ntwo\n\n```bash\n# not a heading\n## also not a heading\n```\n"

	sections := rulesetRegistrySectionArtifacts(content, registry.Metadata.Status)
	var keys []string
	for _, section := range sections {
		keys = append(keys, section.Key)
	}
	for _, want := range []string{
		"# ruleset: section-keys",
		"# ruleset: section-keys > ## purpose",
		"# ruleset: section-keys > ## rules",
		"# ruleset: section-keys > ## parent one > ### duplicate",
		"# ruleset: section-keys > ## parent two > ### duplicate",
	} {
		if !slices.Contains(keys, want) {
			t.Fatalf("expected section key %q in %#v", want, keys)
		}
	}
	for _, unwanted := range []string{"# not a heading", "## also not a heading"} {
		if slices.Contains(keys, unwanted) {
			t.Fatalf("unexpected fenced-code section key %q in %#v", unwanted, keys)
		}
	}
}

func TestFormatRulesetStateTokenUsesColorWhenEnabled(t *testing.T) {
	rendered := formatRulesetStateToken(humanOutputStyle{enabled: true}, "ACTIVE", plan)
	if !strings.Contains(rendered, "\033[") || !strings.Contains(rendered, "ACTIVE") {
		t.Fatalf("expected colored ACTIVE token, got %q", rendered)
	}
}

func TestRulesCommandSupportsRuleAlias(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"rule", "list"})
	if err != nil {
		t.Fatalf("rootCmd.Find(rule list) error = %v", err)
	}
	if cmd != rulesListCmd {
		t.Fatalf("expected rule list to resolve to rules list command, got %q", cmd.Name())
	}
}

func TestRunRulesLinkPreservesFrontMatterAndAvoidsDuplicates(t *testing.T) {
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-alpha")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-alpha"))
	writeFile(t, filepath.Join(projectRoot, "docs", "references", "rules", "api-conventions.md"), templates.BuildRuleset("api-conventions", []string{"api"}))

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	rulesLinkReadPolicy = document.ReferenceReadPolicyMust
	if err := runRulesLink(cmd, []string{"alpha", "api-conventions"}); err != nil {
		t.Fatalf("runRulesLink() error = %v", err)
	}
	if err := runRulesLink(cmd, []string{"alpha", "api-conventions"}); err != nil {
		t.Fatalf("second runRulesLink() error = %v", err)
	}

	doc, err := document.ParseFile(filepath.Join(featurePath, "SPEC.md"), document.TypeSpec)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if doc.Metadata == nil || doc.Metadata.Feature.Slug != "alpha" {
		t.Fatalf("expected feature front matter to be preserved, got %#v", doc.Metadata)
	}
	var count int
	for _, reference := range doc.References() {
		if reference.ID == "ruleset-api-conventions" {
			count++
			if reference.ReadPolicy != document.ReferenceReadPolicyMust {
				t.Fatalf("ReadPolicy = %q, want must", reference.ReadPolicy)
			}
			if reference.Target != "docs/references/rules/api-conventions.md" {
				t.Fatalf("Target = %q", reference.Target)
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected one ruleset reference, got %d in %#v", count, doc.References())
	}
}

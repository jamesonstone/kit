package releaseprompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderIsDeterministicCompleteAndAuthorityAware(t *testing.T) {
	config := goldenConfig()
	first, err := Render(config)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Render(config)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("rendered prompt is not deterministic")
	}
	for _, forbidden := range []string{"{{", "}}", "ghp_example", "token="} {
		if strings.Contains(first, forbidden) {
			t.Fatalf("rendered prompt contains forbidden value %q", forbidden)
		}
	}
	for _, required := range []string{
		"Mandatory Global Release Graph",
		"pull-request-merge",
		"direct merge request or an accepted bounded merge plan",
		"authorized `MERGE_READY` frontier",
		"Pending, missing, stale-head",
		"Report partial waves literally",
		"Destructive or Replacing Changes",
		"Failure and PR Remediation Loop",
		"push to the same branch without rebasing, force-pushing, or retargeting",
		"Create a reviewed replacement PR and first-class graph node only when",
		"Final Integrated-System Verification",
		"VERIFIED\nINFERRED\nNOT_APPLICABLE\nUNRESOLVED",
	} {
		if !strings.Contains(first, required) {
			t.Errorf("rendered prompt missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"Convert failures requiring source changes into normal corrective PRs",
		"Create a normal reviewed corrective PR.",
	} {
		if strings.Contains(first, forbidden) {
			t.Errorf("rendered prompt retained replacement-first remediation %q", forbidden)
		}
	}
}

func TestRenderGolden(t *testing.T) {
	prompt, err := Render(goldenConfig())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("testdata", "release_prompt.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(prompt), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if prompt != string(want) {
		t.Fatal("release prompt differs from golden; run UPDATE_GOLDEN=1 go test ./internal/releaseprompt")
	}
}

func TestRenderDryRunIncludesResolvedYAMLAndPrompt(t *testing.T) {
	config := goldenConfig()
	prompt, err := Render(config)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := RenderDryRun(config, prompt)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"# Release Orchestration Dry Run",
		"source: explicit",
		"github: acme/service-a",
		"# Coding-Agent Release Orchestration",
	} {
		if !strings.Contains(bundle, required) {
			t.Errorf("dry-run bundle missing %q", required)
		}
	}
}

func TestRenderDryRunEscapesResolvedYAMLCodeFences(t *testing.T) {
	config := goldenConfig()
	config.FeatureContext = "```unsafe fence```"
	bundle, err := RenderDryRun(config, "generated prompt")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(bundle, "```unsafe fence```") || !strings.Contains(bundle, "`` `unsafe fence`` `") {
		t.Fatalf("resolved YAML code fence was not neutralized:\n%s", bundle)
	}
}

func TestRenderSanitizesOptionalFreeText(t *testing.T) {
	config := goldenConfig()
	config.AdditionalHardRules = "preserve rules\n# injected `heading`"
	config.FinalReportRequirements = "report evidence\n# injected `report`"
	prompt, err := Render(config)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"\n# injected `heading`", "\n# injected `report`"} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("optional free text changed prompt structure with %q", forbidden)
		}
	}
	for _, required := range []string{"preserve rules # injected 'heading'", "report evidence # injected 'report'"} {
		if !strings.Contains(prompt, required) {
			t.Errorf("sanitized prompt missing %q", required)
		}
	}
}

func TestRenderEscapesDynamicCodeFences(t *testing.T) {
	config := goldenConfig()
	config.FeatureContext = "line one\n```unsafe fence```\nline two"
	prompt, err := Render(config)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt, "```unsafe fence```") || !strings.Contains(prompt, "`` `unsafe fence`` `") {
		t.Fatalf("dynamic code fence was not neutralized:\n%s", prompt)
	}
}

func goldenConfig() Config {
	return Config{
		Project:        "example-platform",
		ScopeExpansion: "related",
		Organization:   "acme",
		FeatureContext: "Authentication and workflow authoring release",
		Repositories: []Repository{
			{Path: "/workspace/service-a", Name: "service-a", GitHub: "acme/service-a", DefaultBranch: "main", KitManaged: true, ReleaseWorkflowPresent: true, VerificationHint: "script:scripts/verify-production.sh"},
			{Path: "/workspace/service-b", Name: "service-b", GitHub: "acme/service-b", DefaultBranch: "main", KitManaged: true},
		},
		SourceControl: SourceControl{Provider: "github", CLI: "gh"},
		Infrastructure: Infrastructure{
			Mode: "direct", Provider: "aws", CLI: "aws",
			IdentityCheck: "Run kit aws verify and confirm account, ARN, region, and environment.",
			Policy:        "Additive changes require an approved batch; destructive changes require explicit manual approval.",
		},
		Production:              Production{Environment: "production", Verification: "command:make verify-production"},
		IntegrationSuite:        "script:tests/end-to-end/production/run.sh",
		DeploymentContext:       "Use repository-owned deployment workflows and verify exact artifact identity.",
		ReviewSystems:           "Human review and CodeRabbit",
		RequiredChecks:          "Repository-required build, lint, test, contract, and security checks",
		DatabaseMigrationPolicy: "Use expand-and-contract.",
		AdditionalHardRules:     "Preserve repository-local safety and delivery rules.",
		FinalReportRequirements: "Report evidence literally.",
		Resolution: []FieldResolution{
			{Field: "project", Value: "example-platform", Source: SourceExplicit},
			{Field: "organization", Value: "acme", Source: SourceDiscovered},
		},
	}
}

package registry

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestParseCatalogRejectsInvalidGraphsAndPaths(t *testing.T) {
	valid := Catalog{
		SchemaVersion: CatalogSchemaVersion,
		Artifacts: []CatalogArtifact{
			testCatalogArtifact(KindRuleset, "base", nil),
			testCatalogArtifact(KindWorkflow, "delivery", []string{"ruleset/base"}),
		},
	}
	if err := ValidateCatalog(valid); err != nil {
		t.Fatalf("valid catalog: %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*Catalog)
		message string
	}{
		{"duplicate", func(c *Catalog) { c.Artifacts = append(c.Artifacts, c.Artifacts[0]) }, "duplicated"},
		{"missing dependency", func(c *Catalog) { c.Artifacts[1].Dependencies = []string{"ruleset/absent"} }, "depends on missing"},
		{"cycle", func(c *Catalog) { c.Artifacts[0].Dependencies = []string{"workflow/delivery"} }, "cycle"},
		{"source escape", func(c *Catalog) { c.Artifacts[0].SourcePath = "../secret" }, "stay inside"},
		{"target escape", func(c *Catalog) { c.Artifacts[0].TargetPath = "/tmp/rule.md" }, "stay inside"},
		{"bad digest", func(c *Catalog) { c.Artifacts[0].Digest = "sha256:nope" }, "sha256"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog := cloneCatalog(valid)
			test.mutate(&catalog)
			err := ValidateCatalog(catalog)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("error = %v, want containing %q", err, test.message)
			}
		})
	}
}

func TestParseCatalogSortsArtifacts(t *testing.T) {
	content := []byte(`schema_version: 1
artifacts:
  - kind: workflow
    slug: zeta
    description: Zeta workflow
    visibility: downstream
    source_path: workflows/zeta.md
    target_path: docs/references/workflows/zeta.md
    version: 1
    digest: sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
    read_policy: conditional
  - kind: ruleset
    slug: alpha
    description: Alpha rules
    visibility: downstream
    source_path: rules/alpha.md
    target_path: docs/references/rules/alpha.md
    version: 1
    digest: sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
    read_policy: must
`)
	catalog, err := ParseCatalog(content)
	if err != nil {
		t.Fatal(err)
	}
	if got := ArtifactKey(catalog.Artifacts[0].Kind, catalog.Artifacts[0].Slug); got != "ruleset/alpha" {
		t.Fatalf("first artifact = %q", got)
	}
}

func TestRepositoryCatalogIncludesPRFeedbackRepairDependencies(t *testing.T) {
	content, err := os.ReadFile("../../registry/catalog.yaml")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := ParseCatalog(content)
	if err != nil {
		t.Fatal(err)
	}
	artifact, found := FindArtifact(catalog, KindWorkflow, "pr-feedback-repair")
	if !found {
		t.Fatal("workflow/pr-feedback-repair is missing")
	}
	want := []string{
		"ruleset/agent-team-orchestration", "ruleset/github-pr-delivery",
		"ruleset/safety-guardrails", "ruleset/source-file-size",
		"ruleset/testing-and-environment-validation", "ruleset/work-lane-gating",
	}
	if !sameStrings(artifact.Dependencies, want) {
		t.Fatalf("dependencies = %v, want %v", artifact.Dependencies, want)
	}
	workflow, err := os.ReadFile("../../" + artifact.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if artifact.Digest != HashContent(string(workflow)) {
		t.Fatalf("catalog digest %s does not match workflow", artifact.Digest)
	}
	doc, err := ParseMarkdown(string(workflow))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDocument(doc, artifact); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(doc.Metadata.PRFeedback.Modes, "collect") {
		t.Fatal("one-shot collect mode is missing")
	}
}

func TestRepositoryCatalogIncludesRepositoryBootstrapDependencies(t *testing.T) {
	content, err := os.ReadFile("../../registry/catalog.yaml")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := ParseCatalog(content)
	if err != nil {
		t.Fatal(err)
	}
	artifact, found := FindArtifact(catalog, KindWorkflow, "repository-bootstrap")
	if !found {
		t.Fatal("workflow/repository-bootstrap is missing")
	}
	want := []string{
		"ruleset/constitution-curation", "ruleset/readme-header-tagline",
		"ruleset/safety-guardrails", "ruleset/source-file-size",
		"ruleset/testing-and-environment-validation",
	}
	if !sameStrings(artifact.Dependencies, want) {
		t.Fatalf("dependencies = %v, want %v", artifact.Dependencies, want)
	}
	workflow, err := os.ReadFile("../../" + artifact.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if artifact.Digest != HashContent(string(workflow)) {
		t.Fatalf("catalog digest %s does not match workflow", artifact.Digest)
	}
	doc, err := ParseMarkdown(string(workflow))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDocument(doc, artifact); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryCatalogRequiresFeatureSpecificationForDelivery(t *testing.T) {
	content, err := os.ReadFile("../../registry/catalog.yaml")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := ParseCatalog(content)
	if err != nil {
		t.Fatal(err)
	}
	rule, found := FindArtifact(catalog, KindRuleset, "feature-specification")
	if !found || rule.ReadPolicy != "must" {
		t.Fatalf("mandatory feature-specification rule = %#v, found = %t", rule, found)
	}
	if _, retired := FindArtifact(catalog, KindRuleset, "feature-notes"); retired {
		t.Fatal("retired feature-notes ruleset remains active")
	}
	workflow, found := FindArtifact(catalog, KindWorkflow, "implementation-delivery")
	if !found || !slices.Contains(workflow.Dependencies, "ruleset/feature-specification") {
		t.Fatalf("implementation-delivery dependencies = %v", workflow.Dependencies)
	}
	for _, artifact := range []CatalogArtifact{rule, workflow} {
		document, readErr := os.ReadFile("../../" + artifact.SourcePath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if artifact.Digest != HashContent(string(document)) {
			t.Fatalf("catalog digest for %s is stale", ArtifactKey(artifact.Kind, artifact.Slug))
		}
		doc, parseErr := ParseMarkdown(string(document))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if validateErr := ValidateDocument(doc, artifact); validateErr != nil {
			t.Fatal(validateErr)
		}
	}
}

func testCatalogArtifact(kind, slug string, dependencies []string) CatalogArtifact {
	directory := "rules"
	target := "docs/references/rules/"
	if kind == KindWorkflow {
		directory, target = "workflows", "docs/references/workflows/"
	}
	return CatalogArtifact{
		Kind: kind, Slug: slug, Description: slug + " artifact", Visibility: "downstream",
		SourcePath: directory + "/" + slug + ".md", TargetPath: target + slug + ".md",
		Version: 1, Digest: "sha256:" + strings.Repeat("a", 64), ReadPolicy: "conditional",
		Dependencies: append([]string(nil), dependencies...),
	}
}

func cloneCatalog(catalog Catalog) Catalog {
	result := Catalog{SchemaVersion: catalog.SchemaVersion, Artifacts: append([]CatalogArtifact(nil), catalog.Artifacts...)}
	for index := range result.Artifacts {
		result.Artifacts[index].Dependencies = append([]string(nil), result.Artifacts[index].Dependencies...)
	}
	return result
}

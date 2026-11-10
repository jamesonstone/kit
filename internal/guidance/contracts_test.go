package guidance

import (
	"strings"
	"testing"
)

func TestMetadataContracts(t *testing.T) {
	for _, status := range []string{"active", "optional", "stale"} {
		data := strings.Replace(rule("sample", status), "description: Sample rule\n", "", 1)
		if _, err := parseMetadata([]byte(data), "ruleset"); err != nil {
			t.Fatalf("%s: %v", status, err)
		}
	}
	for _, status := range []string{"draft", "active", "retired"} {
		data := "---\nkind: sop\nslug: sample\ndescription: Procedure\nowner: Maintainers\nstatus: " + status + "\n---\n"
		if _, err := parseMetadata([]byte(data), "sop"); err != nil {
			t.Fatalf("%s: %v", status, err)
		}
	}
	for _, data := range []string{rule("sample", "draft"), strings.Replace(rule("sample", "active"), "read_policy_default: conditional", "read_policy_default: mystery", 1), strings.Replace(rule("sample", "active"), "slug: sample", "slug: sample\nslug: duplicate", 1)} {
		if _, err := parseMetadata([]byte(data), "ruleset"); err == nil {
			t.Fatalf("accepted invalid metadata %s", data)
		}
	}
}
func TestInspectSlugMismatch(t *testing.T) {
	root := project(t)
	write(t, root, "docs/references/rules/wrong.md", rule("different", "active"))
	write(t, root, "docs/sops/wrong/SOP.md", "---\nkind: sop\nslug: different\ndescription: Procedure\nowner: Maintainers\nstatus: active\n---\n")
	report, err := Inspect(root)
	if err == nil {
		t.Fatal("mismatch accepted")
	}
	for _, path := range []string{"docs/references/rules/wrong.md", "docs/sops/wrong/SOP.md"} {
		e := find(t, report, path)
		if e.Active || e.State != "invalid" || !strings.Contains(e.Error, "does not match") {
			t.Fatalf("%+v", e)
		}
	}
}
func TestInspectDeduplicatesRegistryGuidance(t *testing.T) {
	root := project(t)
	write(t, root, ".kit.yaml", "schema_version: 2\nregistry:\n  artifacts:\n    - kind: project-guidance\n      slug: DESIGN.md\n      path: ./DESIGN.md\n")
	write(t, root, "DESIGN.md", "# Design")
	report, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, e := range report.Entries {
		if e.Path == "DESIGN.md" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("got %d design entries", count)
	}
}

func TestInspectReportsMissingRegistryGuidance(t *testing.T) {
	root := project(t)
	write(t, root, ".kit.yaml", "schema_version: 2\nregistry:\n  artifacts:\n    - kind: project-guidance\n      slug: DESIGN.md\n      path: DESIGN.md\n")
	report, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if e := find(t, report, "DESIGN.md"); e.State != "missing" {
		t.Fatalf("%+v", e)
	}
}

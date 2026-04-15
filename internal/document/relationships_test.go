package document

import "testing"

func TestParseRelationshipsSection_AcceptsNone(t *testing.T) {
	got, err := ParseRelationshipsSection(&Section{
		Name:    "RELATIONSHIPS",
		Content: "none\n",
	})
	if err != nil {
		t.Fatalf("ParseRelationshipsSection() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("ParseRelationshipsSection() len = %d, want 0", len(got))
	}
}

func TestParseRelationshipsSection_ParsesExplicitEdges(t *testing.T) {
	got, err := ParseRelationshipsSection(&Section{
		Name: "RELATIONSHIPS",
		Content: `- builds on: 0007-catchup-command
- depends on: 0009-spec-skills-discovery
- related to: 0011-handoff-document-sync`,
	})
	if err != nil {
		t.Fatalf("ParseRelationshipsSection() error = %v", err)
	}

	want := []Relationship{
		{Type: "builds on", Target: "0007-catchup-command"},
		{Type: "depends on", Target: "0009-spec-skills-discovery"},
		{Type: "related to", Target: "0011-handoff-document-sync"},
	}

	if len(got) != len(want) {
		t.Fatalf("ParseRelationshipsSection() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseRelationshipsSection()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseRelationshipsSection_NormalizesInlineCodeTargets(t *testing.T) {
	got, err := ParseRelationshipsSection(&Section{
		Name: "RELATIONSHIPS",
		Content: `- builds on: ` + "`0007-catchup-command`" + `
- depends on:   ` + "`0009-spec-skills-discovery`" + `  `,
	})
	if err != nil {
		t.Fatalf("ParseRelationshipsSection() error = %v", err)
	}

	want := []Relationship{
		{Type: "builds on", Target: "0007-catchup-command"},
		{Type: "depends on", Target: "0009-spec-skills-discovery"},
	}

	if len(got) != len(want) {
		t.Fatalf("ParseRelationshipsSection() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseRelationshipsSection()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseRelationshipsSection_RejectsInvalidSyntax(t *testing.T) {
	_, err := ParseRelationshipsSection(&Section{
		Name:    "RELATIONSHIPS",
		Content: "- follows: 0001-example-feature\n",
	})
	if err == nil {
		t.Fatal("ParseRelationshipsSection() error = nil, want error")
	}
}

func TestParseRelationshipsSectionRelaxed_SkipsInvalidLinesAndKeepsValidEdges(t *testing.T) {
	got, warnings := ParseRelationshipsSectionRelaxed(&Section{
		Name: "RELATIONSHIPS",
		Content: `- builds on: ` + "`0007-catchup-command`" + `
- follows: 0001-example-feature
- depends on: 0009-spec-skills-discovery`,
	})

	want := []Relationship{
		{Type: "builds on", Target: "0007-catchup-command"},
		{Type: "depends on", Target: "0009-spec-skills-discovery"},
	}
	if len(got) != len(want) {
		t.Fatalf("ParseRelationshipsSectionRelaxed() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseRelationshipsSectionRelaxed()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
	if len(warnings) != 1 {
		t.Fatalf("ParseRelationshipsSectionRelaxed() warnings len = %d, want 1", len(warnings))
	}
	if warnings[0].Line != "- follows: 0001-example-feature" {
		t.Fatalf("ParseRelationshipsSectionRelaxed() warning line = %q, want invalid input", warnings[0].Line)
	}
}

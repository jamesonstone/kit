package registry

import (
	"strings"
	"testing"
)

func TestMergeSectionsPreservesDisjointEdits(t *testing.T) {
	base := testRuleDocument("Original purpose.", "Original rule.")
	local := testRuleDocument("Local purpose.", "Original rule.")
	remote := testRuleDocument("Original purpose.", "Registry rule.")
	hashes, err := HashSections(base)
	if err != nil {
		t.Fatal(err)
	}
	merged, conflicts, err := MergeSections(hashes, local, remote)
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) != 0 {
		t.Fatalf("conflicts = %v", conflicts)
	}
	if !strings.Contains(merged, "Local purpose.") || !strings.Contains(merged, "Registry rule.") {
		t.Fatalf("merge lost an edit:\n%s", merged)
	}
}

func TestMergeSectionsBlocksSameSectionDivergence(t *testing.T) {
	base := testRuleDocument("Original purpose.", "Original rule.")
	hashes, err := HashSections(base)
	if err != nil {
		t.Fatal(err)
	}
	_, conflicts, err := MergeSections(
		hashes,
		testRuleDocument("Local purpose.", "Original rule."),
		testRuleDocument("Registry purpose.", "Original rule."),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) != 1 || conflicts[0] != "purpose" {
		t.Fatalf("conflicts = %v", conflicts)
	}
}

func TestMergeSectionsHandlesAddedAndRemovedSections(t *testing.T) {
	base := testRuleDocument("Purpose.", "Rule.")
	hashes, err := HashSections(base)
	if err != nil {
		t.Fatal(err)
	}
	local := strings.Replace(base, "## Rules\n\nRule.\n", "", 1)
	remote := base + "\n## Verification\n\nRegistry check.\n"
	merged, conflicts, err := MergeSections(hashes, local, remote)
	if err != nil || len(conflicts) != 0 {
		t.Fatalf("merge error = %v, conflicts = %v", err, conflicts)
	}
	if strings.Contains(merged, "## Rules") || !strings.Contains(merged, "Registry check.") {
		t.Fatalf("unexpected merge:\n%s", merged)
	}
}

func testRuleDocument(purpose, rule string) string {
	return `---
kind: ruleset
slug: example
description: Example rule
status: active
registry_scope: downstream
read_policy_default: conditional
---

# Ruleset: Example

## Purpose

` + purpose + `

## Rules

` + rule + "\n"
}

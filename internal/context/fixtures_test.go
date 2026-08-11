package context

import "fmt"

func workflowDocument(slug string, dependencies, rules, evidence []string) string {
	document := "---\nkind: workflow\nslug: " + slug + "\ndescription: test workflow\n"
	if len(dependencies) == 0 {
		document += "dependencies: []\n"
	} else {
		document += "dependencies:\n"
		for _, dependency := range dependencies {
			document += "  - " + dependency + "\n"
		}
	}
	if len(rules) == 0 {
		document += "rules: []\n"
	} else {
		document += "rules:\n"
		for _, rule := range rules {
			document += "  - slug: " + rule + "\n    required: true\n"
		}
	}
	if len(evidence) == 0 {
		document += "evidence: []\n"
	} else {
		document += "evidence:\n"
		for _, path := range evidence {
			document += "  - kind: reference\n    path: " + path + "\n    required: true\n"
		}
	}
	return document + "---\n# Workflow\n"
}

func rulesetDocument(slug string) string {
	return "---\nkind: ruleset\nslug: " + slug + "\nstatus: active\n---\n# Rule\n"
}

func v3Spec(id, slug, dir, relationships, references string) string {
	frontMatter := fmt.Sprintf(`---
kit_metadata_version: 1
artifact: spec
workflow_version: 3
phase: implementation
feature:
  id: %s
  slug: %s
  dir: %s
`, id, slug, dir)
	if relationships != "" {
		frontMatter += relationships + "\n"
	}
	if references != "" {
		frontMatter += references + "\n"
	}
	return frontMatter + `---
# SPEC

## PURPOSE

Purpose.

## CONTEXT

Context.

## REQUIREMENTS

Requirements.

## ACCEPTED PLAN

Plan.

## DECISIONS

None.

## DISCOVERIES

None.

## VALIDATION

Pending.

## OUTCOME

Pending.

## REPOSITORY MEMORY

Pending.
`
}

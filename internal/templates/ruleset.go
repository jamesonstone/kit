package templates

import (
	"strings"
)

type RulesetOptions struct {
	Slug              string
	AppliesTo         []string
	ReadPolicyDefault string
	Context           string
}

// BuildRuleset returns the default durable ruleset artifact. Rulesets are
// pointer-loaded reference docs, not always-loaded instruction content.
func BuildRuleset(slug string, appliesTo []string) string {
	return BuildRulesetWithOptions(RulesetOptions{
		Slug:              slug,
		AppliesTo:         appliesTo,
		ReadPolicyDefault: "conditional",
	})
}

func BuildRulesetWithOptions(opts RulesetOptions) string {
	slug := strings.TrimSpace(opts.Slug)
	appliesTo := opts.AppliesTo
	if len(appliesTo) == 0 {
		appliesTo = []string{slug}
	}
	readPolicyDefault := strings.TrimSpace(opts.ReadPolicyDefault)
	if readPolicyDefault == "" {
		readPolicyDefault = "conditional"
	}
	context := strings.TrimSpace(opts.Context)

	var builder strings.Builder
	builder.WriteString("---\n")
	builder.WriteString("kind: ruleset\n")
	builder.WriteString("slug: " + slug + "\n")
	builder.WriteString("status: active\n")
	builder.WriteString("applies_to:\n")
	for _, entry := range appliesTo {
		builder.WriteString("  - " + strings.TrimSpace(entry) + "\n")
	}
	builder.WriteString("read_policy_default: " + readPolicyDefault + "\n")
	builder.WriteString("---\n\n")
	builder.WriteString("# Ruleset: " + slug + "\n\n")
	builder.WriteString("## Purpose\n\n")
	if context != "" {
		builder.WriteString(context + "\n\n")
	}
	builder.WriteString("- Capture durable project guidance for " + slug + " work.\n")
	builder.WriteString("- Keep this file pointer-loaded: agents should load only the sections relevant to the current decision.\n\n")
	builder.WriteString("## Applies When\n\n")
	builder.WriteString("- A feature references this ruleset with `read_policy: must` or `read_policy: conditional`.\n")
	builder.WriteString("- The current task touches code, docs, tests, or workflows covered by `applies_to`.\n\n")
	builder.WriteString("## Rules\n\n")
	builder.WriteString("- Add stable project rules here once they are durable enough to reuse across features.\n")
	builder.WriteString("- Prefer specific, testable rules over broad style preferences.\n\n")
	builder.WriteString("## Anti-Patterns\n\n")
	builder.WriteString("- Do not inline this ruleset into AGENTS.md, CLAUDE.md, copilot instructions, or prompt bodies by default.\n")
	builder.WriteString("- Do not load unrelated ruleset sections just because the file exists.\n\n")
	builder.WriteString("## Verification\n\n")
	builder.WriteString("- Confirm the referenced feature docs declare whether this ruleset is `must` or `conditional`.\n")
	builder.WriteString("- Run the feature's declared checks after applying rules that affect behavior.\n\n")
	builder.WriteString("## Examples\n\n")
	builder.WriteString("- `kit rules link <feature> " + slug + " --read-policy conditional`\n")
	return builder.String()
}

package cli

// retiredRulesets lists rules that earlier Kit releases installed and this
// release no longer ships, mapped to what replaced them. Historical specs may
// still reference them, and legacy projects may still carry their files.
var retiredRulesets = map[string]string{
	"agent-completion-output":     "universal contract (Reporting)",
	"work-lane-gating":            "delivery",
	"github-pr-delivery":          "delivery",
	"safety-guardrails":           "delivery",
	"human-authorship":            "delivery",
	"source-file-size":            "`.kit.yaml` source_file_line_limit enforced by `kit check --project`",
	"coding-agent-context-usage":  "removed with `kit context resolve`",
	"kit-capabilities-usage":      "removed with `kit capabilities`",
	"command-capabilities":        "removed with `kit capabilities`",
	"codex-thread-initialization": "removed in GH-213",
	"feature-notes":               "removed in GH-137",
}

func isRetiredRuleset(slug string) bool {
	_, ok := retiredRulesets[slug]
	return ok
}

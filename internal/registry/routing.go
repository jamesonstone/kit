package registry

import (
	"fmt"
	"strings"
)

const (
	routingStart = "<!-- BEGIN KIT AGENT CONTRACT -->"
	routingEnd   = "<!-- END KIT AGENT CONTRACT -->"
)

var routingPaths = []string{
	"AGENTS.md",
	"CLAUDE.md",
	".github/copilot-instructions.md",
	"docs/agents/README.md",
}

func PlanRouting(root string, records []ArtifactRecord) ([]Change, error) {
	block := routingBlock(records)
	var changes []Change
	for _, relativePath := range routingPaths {
		before, _, err := ReadOptional(root, relativePath)
		if err != nil {
			return nil, err
		}
		after := upsertManagedBlock(before, block)
		if before != after {
			action := "update"
			if before == "" {
				action = "create"
			}
			changes = append(changes, Change{Path: relativePath, Action: action, Before: before, After: after})
		}
	}
	return changes, nil
}

// RoutingContent applies Kit's bounded contract block while preserving all
// surrounding project-owned text.
func RoutingContent(content string, records []ArtifactRecord) string {
	return upsertManagedBlock(content, routingBlock(records))
}

func routingBlock(records []ArtifactRecord) string {
	rules, workflows := 0, 0
	for _, record := range records {
		switch record.Kind {
		case KindRuleset:
			rules++
		case KindWorkflow:
			workflows++
		}
	}
	return fmt.Sprintf(`%s
## Kit Agent Contract

- Run `+"`kit contract resolve --json`"+` before implementation and whenever task scope materially changes.
- Classify implementation delivery explicitly: use `+"`--workflow implementation-delivery --work-type feature --feature <feature>`"+` for features or `+"`--work-type maintenance`"+` only for genuinely mechanical maintenance.
- For feature work, author the reported living V3 spec and re-resolve before source edits.
- Treat repository-local rulesets and workflows returned by the resolver as the authoritative contract.
- The resolver is local-only and read-only; use `+"`kit registry status`"+` for remote freshness and `+"`kit reconcile`"+` to preview drift.
- Installed contract: %d ruleset(s), %d workflow(s).
%s`, routingStart, rules, workflows, routingEnd)
}

func upsertManagedBlock(content, block string) string {
	content = strings.TrimRight(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	start := strings.Index(content, routingStart)
	end := strings.Index(content, routingEnd)
	if start >= 0 && end >= start {
		end += len(routingEnd)
		prefix := strings.TrimRight(content[:start], "\n")
		suffix := strings.TrimLeft(content[end:], "\n")
		result := ""
		if prefix != "" {
			result = prefix + "\n\n"
		}
		result += block
		if suffix != "" {
			result += "\n\n" + suffix
		}
		return strings.TrimRight(result, "\n") + "\n"
	}
	if content == "" {
		return block + "\n"
	}
	return content + "\n\n" + block + "\n"
}

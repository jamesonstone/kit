package cli

import (
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

// workflowActionPinPattern matches a GitHub Actions `uses:` YAML key, not
// commented text, pinned to a semantic major tag such as
// `actions/github-script@v9` or `@v9.1.0`. It rewrites pins in Kit's own
// template, which has no `uses:` text inside scalars.
var workflowActionPinPattern = regexp.MustCompile(`(?m)^([ \t]*(?:-[ \t]+)?uses:[ \t]*)([A-Za-z0-9_.-]+/[A-Za-z0-9_./-]+)@(v(\d+)(?:\.[0-9]+)*)\b`)

// workflowActionPinValuePattern matches the value of a parsed `uses` key.
var workflowActionPinValuePattern = regexp.MustCompile(`^([A-Za-z0-9_.-]+/[A-Za-z0-9_./-]+)@(v(\d+)(?:\.[0-9]+)*)$`)

type workflowActionPin struct {
	ref   string
	major int
}

// preserveNewerWorkflowActionPins keeps an existing workflow's action pin when
// it names the same action at a newer major version than the Kit template, so
// reconcile never downgrades a Dependabot or human bump. Template pins that are
// equal or newer, and pins that are not semantic major tags, are unchanged.
func preserveNewerWorkflowActionPins(existing, template string) string {
	pins := existingWorkflowActionPins(existing)
	if len(pins) == 0 {
		return template
	}
	return workflowActionPinPattern.ReplaceAllStringFunc(template, func(use string) string {
		match := workflowActionPinPattern.FindStringSubmatch(use)
		major, err := strconv.Atoi(match[4])
		if err != nil {
			return use
		}
		pin, ok := pins[match[2]]
		if !ok || pin.major <= major {
			return use
		}
		return match[1] + match[2] + "@" + pin.ref
	})
}

// existingWorkflowActionPins collects the highest semantic-major pin per action
// from real YAML `uses` keys. Comments and `uses:` text inside scalars such as
// `run: |` blocks are ignored. Unparseable YAML yields no pins.
func existingWorkflowActionPins(existing string) map[string]workflowActionPin {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(existing), &document); err != nil {
		return nil
	}
	pins := map[string]workflowActionPin{}
	var visit func(*yaml.Node)
	visit = func(node *yaml.Node) {
		switch node.Kind {
		case yaml.DocumentNode, yaml.SequenceNode:
			for _, child := range node.Content {
				visit(child)
			}
		case yaml.MappingNode:
			for i := 0; i+1 < len(node.Content); i += 2 {
				key, value := node.Content[i], node.Content[i+1]
				if key.Value == "uses" && value.Kind == yaml.ScalarNode {
					addWorkflowActionPin(pins, value.Value)
				}
				visit(value)
			}
		}
	}
	visit(&document)
	return pins
}

func addWorkflowActionPin(pins map[string]workflowActionPin, value string) {
	match := workflowActionPinValuePattern.FindStringSubmatch(value)
	if match == nil {
		return
	}
	major, err := strconv.Atoi(match[3])
	if err != nil {
		return
	}
	if current, ok := pins[match[1]]; !ok || major > current.major {
		pins[match[1]] = workflowActionPin{ref: match[2], major: major}
	}
}

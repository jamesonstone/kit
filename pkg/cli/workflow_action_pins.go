package cli

import (
	"regexp"
	"strconv"
)

// workflowActionPinPattern matches a GitHub Actions `uses:` reference pinned
// to a semantic major tag such as `actions/github-script@v9` or `@v9.1.0`.
var workflowActionPinPattern = regexp.MustCompile(`(uses:[ \t]*)([A-Za-z0-9_.-]+/[A-Za-z0-9_./-]+)@(v(\d+)(?:\.[0-9]+)*)\b`)

type workflowActionPin struct {
	ref   string
	major int
}

// preserveNewerWorkflowActionPins keeps an existing workflow's action pin when
// it names the same action at a newer major version than the Kit template, so
// reconcile never downgrades a Dependabot or human bump. Template pins that are
// equal or newer, and pins that are not semantic major tags, are unchanged.
func preserveNewerWorkflowActionPins(existing, template string) string {
	pins := map[string]workflowActionPin{}
	for _, match := range workflowActionPinPattern.FindAllStringSubmatch(existing, -1) {
		major, err := strconv.Atoi(match[4])
		if err != nil {
			continue
		}
		if current, ok := pins[match[2]]; !ok || major > current.major {
			pins[match[2]] = workflowActionPin{ref: match[3], major: major}
		}
	}
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

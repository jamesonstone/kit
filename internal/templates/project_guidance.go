package templates

import _ "embed"

const DesignPath = "DESIGN.md"
const SecurityPath = "SECURITY.md"

// Design contains the canonical, project-customizable design defaults.
//
//go:embed design.md
var Design string

// Security contains engineering defaults and project-owned reporting placeholders.
//
//go:embed security.md
var Security string

// ProjectGuidance returns the canonical defaults for a supported root document.
func ProjectGuidance(path string) (string, bool) {
	switch path {
	case DesignPath:
		return Design, true
	case SecurityPath:
		return Security, true
	default:
		return "", false
	}
}

package templates

import (
	"path"
	"strings"
)

// InstructionEntryTitle returns the heading Kit writes above the managed
// contract block in an agent entry file.
func InstructionEntryTitle(relativePath string) string {
	base := path.Base(strings.ReplaceAll(relativePath, "\\", "/"))
	if base == "copilot-instructions.md" {
		return "GitHub Copilot Repository Instructions"
	}
	return strings.TrimSuffix(base, ".md")
}

// InstructionEntryFile returns the generated content of an agent entry file:
// its title and the universal contract block.
func InstructionEntryFile(relativePath string) string {
	return renderUniversalContract(InstructionEntryTitle(relativePath))
}

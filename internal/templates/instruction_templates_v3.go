package templates

import "strings"

func memoryInstructionSupportContent(relativePath string) string {
	if relativePath == "docs/references/testing.md" {
		return strings.ReplaceAll(referencesTesting, "Validation Map and Evidence sections", "VALIDATION and OUTCOME sections")
	}
	return ""
}

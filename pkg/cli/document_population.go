package cli

import (
	"fmt"
	"strings"
)

func appendNonEmptySectionRules(sb *strings.Builder, documentName string) {
	sb.WriteString(fmt.Sprintf(
		"- no section in %s may remain empty or contain only an HTML TODO comment\n",
		documentName,
	))
	sb.WriteString(
		"- if a section has no additional detail, remove the placeholder comment and write `not applicable`, `not required`, or `no additional information required`\n",
	)
}

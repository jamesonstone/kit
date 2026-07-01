package cli

import "strings"

const readmeMaintainersSection = `## Maintainers

Maintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (` + "`jamesonstone`" + `).
`

func upsertReadmeMaintainersSection(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return strings.TrimRight(readmeMaintainersSection, "\n") + "\n"
	}
	withoutMaintainers := removeReadmeMaintainersSections(content)
	return joinReadmeParts(withoutMaintainers, readmeMaintainersSection, "")
}

func removeReadmeMaintainersSections(content string) string {
	lines := strings.SplitAfter(strings.TrimRight(content, "\n"), "\n")
	var kept []string
	for i := 0; i < len(lines); {
		if readmeMaintainersHeading(lines[i]) {
			i++
			for i < len(lines) && !readmeH2Heading(lines[i]) {
				i++
			}
			continue
		}
		kept = append(kept, lines[i])
		i++
	}
	return strings.TrimRight(strings.Join(kept, ""), "\n")
}

func readmeMaintainersHeading(line string) bool {
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "## maintainer", "## maintainers":
		return true
	default:
		return false
	}
}

func readmeH2Heading(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### ")
}

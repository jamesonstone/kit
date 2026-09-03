package cli

import "strings"

const readmeMaintainersFallback = "## Maintainers\n\nMaintained by this project's maintainers.\n"

// defaultReadmeMaintainersSection derives a generic, non-personal maintainers
// credit from the project's own GitHub owner instead of hardcoding Kit's
// author. Downstream projects are not Kit's own repository, so the generated
// default must not name Kit's maintainer.
func defaultReadmeMaintainersSection(owner string) string {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return readmeMaintainersFallback
	}
	return "## Maintainers\n\nMaintained by the [" + owner + "](https://github.com/" + owner + ") team.\n"
}

func upsertReadmeMaintainersSection(content, owner string) string {
	section := defaultReadmeMaintainersSection(owner)
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return strings.TrimRight(section, "\n") + "\n"
	}
	withoutMaintainers := removeReadmeMaintainersSections(content)
	return joinReadmeParts(withoutMaintainers, section, "")
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

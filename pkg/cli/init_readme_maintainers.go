package cli

import (
	"regexp"
	"strings"
)

const (
	readmeMaintainersHeadingLine = "## Maintainers"
	readmeMaintainersFallback    = "## Maintainers\n\nMaintained by this project's maintainers.\n"
	readmeOwnerTypeOrganization  = "Organization"
)

// readmeGeneratedMaintainersPattern matches every maintainers body Kit
// generates. Only these bodies are replaced on refresh; any other text is
// treated as human-authored and preserved.
var readmeGeneratedMaintainersPattern = regexp.MustCompile(
	`^(Maintained by this project's maintainers\.|Maintained by (the )?\[([^\]\s]+)\]\(https://github\.com/([^)\s]+)\)( team)?\.)$`,
)

// defaultReadmeMaintainersSection derives a generic, non-personal maintainers
// credit from the project's own GitHub owner instead of hardcoding Kit's
// author. Downstream projects are not Kit's own repository, so the generated
// default must not name Kit's maintainer. Only a confirmed organization owner
// gets "team" wording; personal accounts and unknown owner types do not.
func defaultReadmeMaintainersSection(owner, ownerType string) string {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return readmeMaintainersFallback
	}
	link := "[" + owner + "](https://github.com/" + owner + ")"
	if strings.EqualFold(strings.TrimSpace(ownerType), readmeOwnerTypeOrganization) {
		return readmeMaintainersHeadingLine + "\n\nMaintained by the " + link + " team.\n"
	}
	return readmeMaintainersHeadingLine + "\n\nMaintained by " + link + ".\n"
}

func upsertReadmeMaintainersSection(content, owner, ownerType string) string {
	section := defaultReadmeMaintainersSection(owner, ownerType)
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return strings.TrimRight(section, "\n") + "\n"
	}
	withoutMaintainers, customBody := removeReadmeMaintainersSections(content)
	if customBody != "" {
		section = readmeMaintainersHeadingLine + "\n\n" + customBody + "\n"
	}
	return joinReadmeParts(withoutMaintainers, section, "")
}

// removeReadmeMaintainersSections strips every Maintainers section and returns
// the remaining content plus the first human-authored section body, if any.
func removeReadmeMaintainersSections(content string) (string, string) {
	lines := strings.SplitAfter(strings.TrimRight(content, "\n"), "\n")
	var kept []string
	customBody := ""
	for i := 0; i < len(lines); {
		if readmeMaintainersHeading(lines[i]) {
			i++
			start := i
			for i < len(lines) && !readmeH2Heading(lines[i]) {
				i++
			}
			body := strings.TrimSpace(strings.Join(lines[start:i], ""))
			if customBody == "" && !readmeGeneratedMaintainersBody(body) {
				customBody = body
			}
			continue
		}
		kept = append(kept, lines[i])
		i++
	}
	return strings.TrimRight(strings.Join(kept, ""), "\n"), customBody
}

func readmeGeneratedMaintainersBody(body string) bool {
	if body == "" {
		return true
	}
	match := readmeGeneratedMaintainersPattern.FindStringSubmatch(body)
	if match == nil {
		return false
	}
	// The link text and target must name the same owner for the generated forms.
	return match[3] == "" || match[3] == match[4]
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

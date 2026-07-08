package cli

import "strings"

func reconcileTaskFields(content string) map[string]bool {
	fields := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		match := reconcileTaskFieldPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		fields[strings.ToUpper(strings.TrimSpace(match[1]))] = true
	}
	return fields
}

func nextReconcileSection(content string, start int) int {
	matches := reconcileSectionPattern.FindAllStringIndex(content[start:], -1)
	if len(matches) == 0 {
		return -1
	}
	return start + matches[0][0]
}

func missingExecutableFields(fields map[string]bool) []string {
	required := []string{"VERIFY", "EXPECTED FILES", "RISK", "ROLLBACK"}
	var missing []string
	for _, field := range required {
		if !fields[field] {
			missing = append(missing, field)
		}
	}
	return missing
}

package cli

import (
	"regexp"
	"sort"
	"strings"
)

var ciSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(token|secret|password|passwd|api[_-]?key|access[_-]?key|private[_-]?key)=([^\s]+)`),
	regexp.MustCompile(`(?i)(ghp|github_pat|gho|ghu|ghs|ghr)_[A-Za-z0-9_]+`),
	regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`),
}

func extractRelevantCILogExcerpt(raw string, maxLines int) ([]string, bool) {
	lines := splitNonEmptyLines(redactCILog(raw))
	if len(lines) <= maxLines {
		return lines, false
	}

	firstMatch := -1
	include := map[int]bool{}
	for i, line := range lines {
		if ciFailureLinePattern.MatchString(line) {
			if firstMatch < 0 {
				firstMatch = i
			}
			start := max(0, i-5)
			end := min(len(lines), i+15)
			for j := start; j < end; j++ {
				include[j] = true
			}
		}
	}
	if len(include) == 0 {
		return append([]string{}, lines[len(lines)-maxLines:]...), true
	}

	indexes := make([]int, 0, len(include))
	for index := range include {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	if len(indexes) > maxLines && firstMatch >= 0 {
		start := max(0, firstMatch-(maxLines/2))
		end := min(len(lines), start+maxLines)
		if end-start < maxLines {
			start = max(0, end-maxLines)
		}
		return append([]string{}, lines[start:end]...), true
	}
	excerpt := make([]string, 0, min(len(indexes), maxLines))
	for _, index := range indexes {
		excerpt = append(excerpt, lines[index])
		if len(excerpt) >= maxLines {
			break
		}
	}
	return excerpt, true
}

func splitNonEmptyLines(raw string) []string {
	rawLines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func redactCILog(raw string) string {
	redacted := raw
	for _, pattern := range ciSecretPatterns {
		redacted = pattern.ReplaceAllStringFunc(redacted, redactSecretMatch)
	}
	return redacted
}

func redactSecretMatch(match string) string {
	if strings.Contains(match, "=") {
		parts := strings.SplitN(match, "=", 2)
		return parts[0] + "=[REDACTED]"
	}
	return "[REDACTED]"
}

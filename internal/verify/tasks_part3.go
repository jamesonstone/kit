package verify

import (
	"fmt"
	"strings"
)

func stripPlanLinks(value string) string {
	parts := strings.Fields(value)
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "[PLAN-") || strings.HasPrefix(part, "[SPEC-") {
			continue
		}
		kept = append(kept, part)
	}
	return strings.Join(kept, " ")
}

func splitDependencies(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	})
	deps := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			deps = append(deps, part)
		}
	}
	return deps
}

func handoffNeeded(bundle TaskBundle) bool {
	risk := strings.ToLower(bundle.Risk)
	return strings.Contains(risk, "medium") || strings.Contains(risk, "high") || len(bundle.Dependencies) > 1
}

func hasShellSyntax(command string) bool {
	syntax := []string{"&&", "||", ";", "|", "<", ">", "$(", "${", "\n"}
	for _, item := range syntax {
		if strings.Contains(command, item) {
			return true
		}
	}
	return false
}

func shellArgv(command string) []string {
	return []string{"sh", "-c", command}
}

func splitCommandLine(command string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	escaped := false

	for _, r := range command {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			current.WriteRune(r)
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == ' ' || r == '\t' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(r)
	}
	if escaped {
		current.WriteRune('\\')
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote in command")
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args, nil
}

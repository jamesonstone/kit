package cli

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Keep every existing byte, including comments and custom expressions. Only
// insert ordinary block steps at the reviewed anchor and sequence boundary.
func insertMintStages(data []byte, steps *yaml.Node, index int, publish bool) (string, error) {
	lines := strings.Split(string(data), "\n")
	first := steps.Content[0].Line - 1
	start := steps.Content[index].Line - 1
	if first < 0 || start < 0 || start >= len(lines) {
		return "", fmt.Errorf("cannot locate Mint insertion point")
	}
	prefix := lines[first][:len(lines[first])-len(strings.TrimLeft(lines[first], " "))]
	if !strings.HasPrefix(strings.TrimSpace(lines[first]), "-") || !strings.HasPrefix(strings.TrimSpace(lines[start]), "-") {
		return "", fmt.Errorf("mint integration requires block-style steps")
	}
	end := first + 1
	for end < len(lines) {
		trim := strings.TrimSpace(lines[end])
		if trim != "" && !strings.HasPrefix(trim, "#") && len(lines[end])-len(strings.TrimLeft(lines[end], " ")) < len(prefix) {
			break
		}
		end++
	}
	prep, err := mintIndentedSteps(mintReleasePreparationSteps(), prefix)
	if err != nil {
		return "", err
	}
	tail := ""
	if publish {
		step := mintReleasePublisherStep()
		step["if"] = mintPublishGuard
		tail, err = mintIndentedSteps([]any{step}, prefix)
		if err != nil {
			return "", err
		}
	}
	join := func(part []string) string {
		s := strings.Join(part, "\n")
		if len(part) > 0 && !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		return s
	}
	content := join(lines[:start]) + prep + join(lines[start:end]) + tail + strings.Join(lines[end:], "\n")
	var check yaml.Node
	if err := yaml.Unmarshal([]byte(content), &check); err != nil {
		return "", fmt.Errorf("validate Mint workflow insertion: %w", err)
	}
	return content, nil
}
func mintIndentedSteps(steps []any, prefix string) (string, error) {
	data, err := yaml.Marshal(steps)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		out.WriteString(prefix + line + "\n")
	}
	return out.String(), nil
}

package plugin

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

// Apply inserts blocks immediately after the universal-contract block.
// With no blocks and no plugin markers it returns content unchanged.
// Removing the last block also removes the blank line Kit added and leaves
// every other byte in place.
func Apply(content string, blocks []Block) (string, error) {
	if err := validateBlocks(blocks); err != nil {
		return "", err
	}
	hasMarkers := strings.Contains(content, pluginBeginPrefix) || strings.Contains(content, pluginEndPrefix)
	if len(blocks) == 0 && !hasMarkers {
		return content, nil
	}
	begin := templates.UniversalContractBeginMarker
	end := templates.UniversalContractEndMarker
	if strings.Count(content, begin) != 1 || strings.Count(content, end) != 1 {
		return "", fmt.Errorf("plugin blocks require exactly one Kit-managed contract block")
	}
	markerAt := strings.Index(content, end)
	lineEnd := markerAt + len(end)
	if lineEnd < len(content) && content[lineEnd] == '\n' {
		lineEnd++
	}
	prefix, suffix := content[:lineEnd], content[lineEnd:]
	if strings.Contains(prefix, pluginBeginPrefix) || strings.Contains(prefix, pluginEndPrefix) {
		return "", fmt.Errorf("plugin markers must sit immediately after the Kit-managed contract")
	}
	project, err := stripPluginRegion(suffix)
	if err != nil {
		return "", err
	}
	if len(blocks) == 0 {
		return prefix + project, nil
	}
	return prefix + renderBlocks(blocks) + project, nil
}

func validateBlocks(blocks []Block) error {
	seen := map[string]bool{}
	for _, block := range blocks {
		if err := ValidateID(block.ID); err != nil {
			return err
		}
		if seen[block.ID] {
			return fmt.Errorf("plugin %s is listed more than once", block.ID)
		}
		seen[block.ID] = true
		body := strings.TrimRight(block.Body, "\n")
		if strings.TrimSpace(body) == "" {
			return fmt.Errorf("plugin %s instructions are empty", block.ID)
		}
		if strings.Contains(body, "\r") || strings.Contains(block.ConfigPath, "\r") {
			return fmt.Errorf("plugin %s instructions must use LF line endings", block.ID)
		}
		if strings.Contains(body, "KIT-PLUGIN") || strings.Contains(body, "KIT-MANAGED CONTRACT") {
			return fmt.Errorf("plugin %s instructions contain a managed marker", block.ID)
		}
		if strings.Contains(block.ConfigPath, "\n") || strings.Contains(block.ConfigPath, "-->") {
			return fmt.Errorf("plugin %s config path is invalid", block.ID)
		}
	}
	return nil
}

func renderBlocks(blocks []Block) string {
	var out strings.Builder
	for _, block := range blocks {
		body := strings.TrimRight(block.Body, "\n")
		out.WriteString("\n")
		out.WriteString(pluginBeginMarker(block.ID))
		out.WriteString("\n")
		if block.ConfigPath != "" {
			out.WriteString(pluginConfigComment(block.ConfigPath))
			out.WriteString("\n")
		}
		out.WriteString(body)
		out.WriteString("\n")
		out.WriteString(pluginEndMarker(block.ID))
		out.WriteString("\n")
	}
	return out.String()
}

func stripPluginRegion(suffix string) (string, error) {
	if !strings.Contains(suffix, pluginBeginPrefix) && !strings.Contains(suffix, pluginEndPrefix) {
		return suffix, nil
	}
	rest := suffix
	if strings.HasPrefix(rest, "\n") {
		if _, ok := pluginIDFromBegin(firstLine(rest[1:])); ok {
			rest = rest[1:]
		}
	}
	if _, ok := pluginIDFromBegin(firstLine(rest)); !ok {
		return "", fmt.Errorf("plugin markers must sit immediately after the Kit-managed contract")
	}
	for {
		var err error
		rest, err = consumePluginBlock(rest)
		if err != nil {
			return "", err
		}
		if _, ok := pluginIDFromBegin(firstLine(rest)); ok {
			continue
		}
		if strings.HasPrefix(rest, "\n") {
			if _, ok := pluginIDFromBegin(firstLine(rest[1:])); ok {
				rest = rest[1:]
				continue
			}
		}
		break
	}
	if strings.Contains(rest, pluginBeginPrefix) || strings.Contains(rest, pluginEndPrefix) {
		return "", fmt.Errorf("plugin markers must sit immediately after the Kit-managed contract")
	}
	return rest, nil
}

func consumePluginBlock(rest string) (string, error) {
	line, after, found := strings.Cut(rest, "\n")
	id, ok := pluginIDFromBegin(line)
	if !ok || !found {
		return "", fmt.Errorf("plugin block is incomplete")
	}
	end := pluginEndMarker(id)
	for {
		var bodyLine string
		bodyLine, after, found = strings.Cut(after, "\n")
		if bodyLine == end {
			if !found {
				return "", nil
			}
			return after, nil
		}
		if !found {
			return "", fmt.Errorf("plugin block %s is missing its end marker", id)
		}
		if strings.Contains(bodyLine, "KIT-PLUGIN") || strings.Contains(bodyLine, "KIT-MANAGED CONTRACT") {
			return "", fmt.Errorf("plugin block %s contains a managed marker", id)
		}
	}
}

func firstLine(value string) string {
	line, _, _ := strings.Cut(value, "\n")
	return line
}

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
	prefixEnd := markerAt + len(end)
	prefix, suffix := content[:prefixEnd], content[prefixEnd:]
	if strings.Contains(prefix, pluginBeginPrefix) || strings.Contains(prefix, pluginEndPrefix) {
		return "", fmt.Errorf("plugin markers must sit immediately after the Kit-managed contract")
	}
	kept, err := stripOwnedPluginRegion(suffix)
	if err != nil {
		return "", err
	}
	if len(blocks) == 0 {
		return prefix + kept, nil
	}
	// A newline that already ends the contract line stays in kept. The region
	// supplies its own leading newline, which is the blank line after that
	// ending, or the ending itself when the file stopped on the marker.
	region := renderBlocks(blocks)
	if strings.HasPrefix(kept, "\n") {
		return prefix + kept[:1] + region + kept[1:], nil
	}
	return prefix + region + kept, nil
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
		if strings.Contains(block.ConfigPath, "\n") || strings.Contains(block.ConfigPath, "-->") || strings.Contains(block.ConfigPath, "KIT-PLUGIN") || strings.Contains(block.ConfigPath, "KIT-MANAGED CONTRACT") {
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

// stripOwnedPluginRegion removes the plugin region Kit inserted after the
// contract marker text. A newline that ended the contract line stays. A
// newline Kit added because the file ended on the marker is removed with the
// region.
func stripOwnedPluginRegion(suffix string) (string, error) {
	if !strings.Contains(suffix, pluginBeginPrefix) && !strings.Contains(suffix, pluginEndPrefix) {
		return suffix, nil
	}
	if strings.HasPrefix(suffix, "\n") {
		afterNewline := suffix[1:]
		if _, ok := pluginIDFromBegin(firstLine(afterNewline)); ok {
			rest, err := consumePluginBlocks(afterNewline)
			return rest, err
		}
		if strings.HasPrefix(afterNewline, "\n") {
			if _, ok := pluginIDFromBegin(firstLine(afterNewline[1:])); ok {
				rest, err := consumePluginBlocks(afterNewline[1:])
				if err != nil {
					return "", err
				}
				return "\n" + rest, nil
			}
		}
	}
	if _, ok := pluginIDFromBegin(firstLine(suffix)); ok {
		return consumePluginBlocks(suffix)
	}
	return "", fmt.Errorf("plugin markers must sit immediately after the Kit-managed contract")
}

func consumePluginBlocks(rest string) (string, error) {
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
		if strings.HasPrefix(bodyLine, "<!-- Plugin config: ") && strings.HasSuffix(bodyLine, " -->") {
			continue
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

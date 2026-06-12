package cli

import (
	"fmt"
	"io"
	"strings"
)

func renderRegistryRulesetSelector(out io.Writer, entries []registrySelectorEntry, cursor int) {
	style := styleForWriter(out)
	if cursor >= 0 {
		_, _ = fmt.Fprint(out, "\x1b[H\x1b[2J")
	}
	tableWidth := registrySelectorTableWidth(out)

	_, _ = fmt.Fprintln(out, style.selectionTitle("Select registry rulesets"))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.muted(truncateString("Source: "+rulesetRegistrySourceDescription(), tableWidth)))
	if cursor >= 0 {
		_, _ = fmt.Fprintln(out, style.muted(truncateString("Keys: Tab/Down/j move | Shift+Tab/Up/k move | Space toggles | v previews | Enter applies | q cancels", tableWidth)))
	} else {
		_, _ = fmt.Fprintln(out, style.muted(truncateString("Input: type rule numbers to toggle, then Enter applies. Preview full text with `kit rules view <slug>`.", tableWidth)))
	}
	_, _ = fmt.Fprintln(out)

	widths := registrySelectorColumnWidths(tableWidth, entries)
	renderRegistrySelectorBorder(out, style, "┌", "┬", "┐", widths)
	renderRegistrySelectorRow(out, style, []registrySelectorCell{
		{text: "No", width: widths[0], color: whiteBold},
		{text: "Use", width: widths[1], color: whiteBold},
		{text: "Ruleset", width: widths[2], color: whiteBold},
		{text: "State", width: widths[3], color: whiteBold},
		{text: "Source", width: widths[4], color: whiteBold},
		{text: "Description", width: widths[5], color: whiteBold},
	})
	renderRegistrySelectorBorder(out, style, "├", "┼", "┤", widths)
	for i := range entries {
		number := fmt.Sprintf("%d", i+1)
		if i == cursor {
			number = ">" + number
		}
		checkbox := "[ ]"
		checkboxColor := dim
		if entries[i].DesiredActive {
			checkbox = "[x]"
			checkboxColor = plan
		}
		stateLabel, stateColor := registrySelectorState(entries[i])
		sourceLabel, sourceColor := registrySelectorSource(entries[i])
		renderRegistrySelectorRow(out, style, []registrySelectorCell{
			{text: number, width: widths[0], color: registrySelectorCursorColor(i, cursor), alignRight: true},
			{text: checkbox, width: widths[1], color: checkboxColor},
			{text: entries[i].Registry.Slug, width: widths[2], color: registrySelectorSlugColor(entries[i], i == cursor)},
			{text: stateLabel, width: widths[3], color: stateColor},
			{text: sourceLabel, width: widths[4], color: sourceColor},
			{text: selectorRulesetDescription(entries[i]), width: widths[5]},
		})
	}
	renderRegistrySelectorBorder(out, style, "└", "┴", "┘", widths)
}

type registrySelectorCell struct {
	text       string
	width      int
	color      string
	alignRight bool
}

func registrySelectorTableWidth(out io.Writer) int {
	width := terminalWidthOrDefault(out, registrySelectorDefaultTableWidth)
	if width < registrySelectorMinimumTableWidth {
		return registrySelectorMinimumTableWidth
	}
	return width
}

func registrySelectorColumnWidths(tableWidth int, entries []registrySelectorEntry) []int {
	numberWidth := len(fmt.Sprintf("%d", len(entries))) + 1
	if numberWidth < 2 {
		numberWidth = 2
	}

	slugWidth := len("Ruleset")
	for _, entry := range entries {
		if width := len([]rune(entry.Registry.Slug)); width > slugWidth {
			slugWidth = width
		}
	}
	if slugWidth < registrySelectorMinimumSlugWidth {
		slugWidth = registrySelectorMinimumSlugWidth
	}
	if slugWidth > registrySelectorMaximumSlugWidth {
		slugWidth = registrySelectorMaximumSlugWidth
	}

	widths := []int{
		numberWidth,
		3,
		slugWidth,
		9,
		16,
		registrySelectorMinimumDescWidth,
	}

	for {
		descWidth := tableWidth - registrySelectorTableOverhead(len(widths))
		for _, width := range widths[:len(widths)-1] {
			descWidth -= width
		}
		if descWidth >= registrySelectorMinimumDescWidth {
			widths[len(widths)-1] = descWidth
			return widths
		}
		if widths[2] > registrySelectorMinimumSlugWidth {
			widths[2]--
			continue
		}
		widths[len(widths)-1] = registrySelectorMinimumDescWidth
		return widths
	}
}

func registrySelectorTableOverhead(columnCount int) int {
	return 3*columnCount + 1
}

func renderRegistrySelectorBorder(out io.Writer, style humanOutputStyle, left, middle, right string, widths []int) {
	var builder strings.Builder
	builder.WriteString(left)
	for i, width := range widths {
		if i > 0 {
			builder.WriteString(middle)
		}
		builder.WriteString(strings.Repeat("─", width+2))
	}
	builder.WriteString(right)
	_, _ = fmt.Fprintln(out, style.muted(builder.String()))
}

func renderRegistrySelectorRow(out io.Writer, style humanOutputStyle, cells []registrySelectorCell) {
	_, _ = fmt.Fprint(out, style.muted("│"))
	for _, cell := range cells {
		text := truncateString(strings.Join(strings.Fields(cell.text), " "), cell.width)
		rendered := statusMatrixField(style, text, cell.width, cell.color, cell.alignRight)
		_, _ = fmt.Fprintf(out, " %s %s", rendered, style.muted("│"))
	}
	_, _ = fmt.Fprintln(out)
}

func registrySelectorCursorColor(index, cursor int) string {
	if index == cursor {
		return spec
	}
	return dim
}

func registrySelectorSlugColor(entry registrySelectorEntry, selected bool) string {
	if selected {
		return whiteBold
	}
	if entry.Installed {
		return brainstorm
	}
	return ""
}

func renderRegistryRulesetPreview(out io.Writer, entry registrySelectorEntry) {
	style := styleForWriter(out)
	content := entry.Registry.Content
	source := rulesetRegistryRulesetURL(entry.Registry.Slug)
	if entry.Installed {
		content = entry.LocalContent
		source = rulesetTarget(entry.Registry.Slug)
	}
	_, _ = fmt.Fprint(out, "\x1b[H\x1b[2J")
	_, _ = fmt.Fprintln(out, style.selectionTitle("Preview: "+entry.Registry.Slug))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.muted("Source: "+source))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprint(out, ensureTrailingNewline(content))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.muted("Press any key to return."))
}

func selectorRulesetDescription(entry registrySelectorEntry) string {
	if entry.Local != nil {
		if description := strings.TrimSpace(entry.Local.Metadata.Description); description != "" {
			return description
		}
	}
	if description := strings.TrimSpace(entry.Registry.Metadata.Description); description != "" {
		return description
	}
	return "No description provided."
}

func registrySelectorState(entry registrySelectorEntry) (string, string) {
	switch {
	case entry.DesiredActive:
		return "ACTIVE", plan
	case entry.Installed:
		return "INACTIVE", implement
	default:
		return "AVAILABLE", dim
	}
}

func registrySelectorSource(entry registrySelectorEntry) (string, string) {
	switch entry.RegistryState {
	case "update-available":
		return "UPDATE AVAILABLE", spec
	case registryArtifactStateConflict:
		return "CONFLICT", constitution
	case registryArtifactStateLocalCustom:
		return "LOCAL-CUSTOM", constitution
	case registryArtifactStateManaged:
		return "MANAGED", plan
	}
	switch {
	case entry.Modified:
		return "MODIFIED", constitution
	case entry.Installed:
		return "LOCAL", brainstorm
	default:
		return "REGISTRY", dim
	}
}

func formatRulesetStateToken(style humanOutputStyle, label string, color string) string {
	if !style.enabled {
		return label
	}
	return color + whiteBold + label + reset
}

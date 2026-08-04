package worktree

import (
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	selectorDefaultWidth      = 120
	selectorStateWidth        = 8
	selectorHeadWidth         = 24
	selectorPRWidth           = 6
	selectorTitleHeaderWidth  = len("TITLE")
	selectorUpdatedWidth      = 18
	selectorFixedColumnsWidth = 1 + 6 + selectorStateWidth + selectorHeadWidth + selectorPRWidth + selectorUpdatedWidth
	selectorMinimumHeight     = 4
)

func renderWorktreeSelector(output *os.File, entries []worktreeEntry, selected int) (int, error) {
	width, height := selectorDefaultWidth, len(entries)+3
	if terminalWidth, terminalHeight, err := term.GetSize(int(output.Fd())); err == nil {
		if terminalWidth > 0 {
			width = terminalWidth
		}
		if terminalHeight >= selectorMinimumHeight {
			height = terminalHeight
		}
	}
	return renderWorktreeSelectorAtSize(output, entries, selected, width, height)
}

func renderWorktreeSelectorAtSize(
	output io.Writer,
	entries []worktreeEntry,
	selected int,
	width int,
	height int,
) (int, error) {
	visibleCount := min(len(entries), max(height-3, 1))
	start := max(selected-visibleCount+1, 0)
	if start+visibleCount > len(entries) {
		start = max(len(entries)-visibleCount, 0)
	}
	end := min(start+visibleCount, len(entries))
	titleWidth := selectorTitleWidth(entries, width)

	titleText := fmt.Sprintf("Select a worktree (%d/%d)  arrows/Tab: move  Enter: open  h: home  q: cancel", selected+1, len(entries))
	title := truncateTerminalLine(titleText, width)
	header := fmt.Sprintf(
		"  %-8s %-24s %-6s %-*s %-18s %s",
		"STATE", "HEAD", "PR#", titleWidth, "TITLE", "LAST UPDATED", "PATH",
	)
	if _, err := fmt.Fprintf(output, "%s%s%s\r\n", colorBold, title, colorReset); err != nil {
		return 0, fmt.Errorf("render worktree selector title: %w", err)
	}
	if _, err := fmt.Fprintf(output, "%s%s%s\r\n", colorBold, header, colorReset); err != nil {
		return 0, fmt.Errorf("render worktree selector header: %w", err)
	}
	renderedLines := selectorDisplayRows(title, width) + selectorDisplayRows(header, width)
	for i := start; i < end; i++ {
		entry := entries[i]
		pointer := " "
		state := truncateTerminalLine(sanitizeTerminalField(entry.state), selectorStateWidth)
		head := truncateTerminalLine(sanitizeTerminalField(selectorDisplayHead(entry)), selectorHeadWidth)
		pr := truncateTerminalLine(sanitizeTerminalField(entry.prText), selectorPRWidth)
		prTitle := truncateTerminalLine(sanitizeTerminalField(entry.prTitle), titleWidth)
		updated := truncateTerminalLine(sanitizeTerminalField(entry.updatedText), selectorUpdatedWidth)
		path := sanitizeTerminalField(entry.path)
		color := selectorEntryColor(entry, state, i == selected)
		if i == selected {
			pointer = ">"
		}
		line := fmt.Sprintf(
			"%s %-8s %-24s %-6s %-*s %-18s %s",
			pointer, state, head, pr, titleWidth, prTitle, updated, path,
		)
		if _, err := fmt.Fprintf(output, "%s%s%s\r\n", color, line, colorReset); err != nil {
			return 0, fmt.Errorf("render worktree selector entry: %w", err)
		}
		renderedLines += selectorDisplayRows(line, width)
	}
	return renderedLines, nil
}

func selectorTitleWidth(entries []worktreeEntry, terminalWidth int) int {
	maxTitleWidth := selectorTitleHeaderWidth
	maxPathWidth := len("PATH")
	for _, entry := range entries {
		maxTitleWidth = max(
			maxTitleWidth,
			utf8.RuneCountInString(sanitizeTerminalField(entry.prTitle)),
		)
		maxPathWidth = max(
			maxPathWidth,
			utf8.RuneCountInString(sanitizeTerminalField(entry.path)),
		)
	}
	available := terminalWidth - selectorFixedColumnsWidth - maxPathWidth
	return min(maxTitleWidth, max(available, selectorTitleHeaderWidth))
}

func selectorDisplayRows(value string, terminalWidth int) int {
	if terminalWidth <= 0 {
		return 1
	}
	return max((utf8.RuneCountInString(value)+terminalWidth-1)/terminalWidth, 1)
}

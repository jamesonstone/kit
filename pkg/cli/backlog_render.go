package cli

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"
)

const (
	backlogDefaultTableWidth     = 100
	backlogMinimumTableWidth     = 72
	backlogMinimumFeatureWidth   = 18
	backlogMaximumFeatureWidth   = 32
	backlogMinimumDescriptionWid = 40
)

func printBacklogTable(w io.Writer, entries []backlogEntry) error {
	style := styleForWriter(w)
	featureWidth, descriptionWidth := backlogColumnWidths(w, entries)

	header := statusMatrixField(style, "Feature", featureWidth, whiteBold, false) + "  " +
		statusMatrixField(style, "Description", descriptionWidth, whiteBold, false)
	if _, err := fmt.Fprintln(w, header); err != nil {
		return err
	}

	separator := strings.Repeat("-", featureWidth) + "  " + strings.Repeat("-", descriptionWidth)
	if _, err := fmt.Fprintln(w, style.muted(separator)); err != nil {
		return err
	}

	for _, entry := range entries {
		descriptionLines := wrapBacklogText(entry.Description, descriptionWidth)
		for i, line := range descriptionLines {
			featureText := ""
			featureColor := ""
			if i == 0 {
				featureText = truncateString(entry.Feature.Slug, featureWidth)
				featureColor = brainstorm
			}

			row := statusMatrixField(style, featureText, featureWidth, featureColor, false) + "  " +
				statusMatrixField(style, line, descriptionWidth, "", false)
			if _, err := fmt.Fprintln(w, row); err != nil {
				return err
			}
		}
	}

	return nil
}

func backlogColumnWidths(w io.Writer, entries []backlogEntry) (int, int) {
	tableWidth := terminalWidthOrDefault(w, backlogDefaultTableWidth)
	if tableWidth < backlogMinimumTableWidth {
		tableWidth = backlogMinimumTableWidth
	}

	featureWidth := len("Feature")
	for _, entry := range entries {
		if width := len([]rune(entry.Feature.Slug)); width > featureWidth {
			featureWidth = width
		}
	}

	if featureWidth < backlogMinimumFeatureWidth {
		featureWidth = backlogMinimumFeatureWidth
	}
	if featureWidth > backlogMaximumFeatureWidth {
		featureWidth = backlogMaximumFeatureWidth
	}

	descriptionWidth := tableWidth - featureWidth - 2
	if descriptionWidth < backlogMinimumDescriptionWid {
		featureWidth = tableWidth - backlogMinimumDescriptionWid - 2
		if featureWidth < backlogMinimumFeatureWidth {
			featureWidth = backlogMinimumFeatureWidth
		}
		descriptionWidth = tableWidth - featureWidth - 2
	}
	if descriptionWidth < backlogMinimumDescriptionWid {
		descriptionWidth = backlogMinimumDescriptionWid
	}

	return featureWidth, descriptionWidth
}

func wrapBacklogText(text string, width int) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return []string{"(no description)"}
	}

	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return []string{"(no description)"}
	}

	lines := make([]string, 0, 2)
	current := words[0]

	for _, word := range words[1:] {
		candidate := current + " " + word
		if len([]rune(candidate)) <= width {
			current = candidate
			continue
		}
		lines = append(lines, truncateString(current, width))
		current = word
	}

	lines = append(lines, truncateString(current, width))
	return lines
}

func terminalWidthOrDefault(w io.Writer, fallback int) int {
	fileLike, ok := w.(interface{ Fd() uintptr })
	if !ok {
		return fallback
	}

	width, _, err := term.GetSize(int(fileLike.Fd()))
	if err != nil || width <= 0 {
		return fallback
	}

	return width
}

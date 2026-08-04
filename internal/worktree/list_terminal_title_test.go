package worktree

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSelectorTitleWidthReservesCompleteLongestPath(t *testing.T) {
	t.Parallel()
	entries := []worktreeEntry{
		{prTitle: "A title that must be truncated", path: strings.Repeat("p", 30)},
		{prTitle: "Short", path: "/tmp/short"},
	}
	if got, want := selectorTitleWidth(entries, 100), 7; got != want {
		t.Fatalf("title width = %d, want %d", got, want)
	}
	if got, want := selectorTitleWidth(entries, 50), len("TITLE"); got != want {
		t.Fatalf("narrow title width = %d, want header width %d", got, want)
	}
}

func TestRenderWorktreeSelectorTruncatesTitleWithoutTruncatingPath(t *testing.T) {
	t.Parallel()
	const (
		width     = 120
		fullPath  = "/Users/example/worktrees/project/GH-119"
		fullTitle = "Add a responsive pull request title column"
	)
	entries := []worktreeEntry{{
		branch:      "a-branch-name-that-exceeds-the-fixed-head-width",
		state:       "clean",
		prText:      "120",
		prTitle:     fullTitle,
		updatedText: "Aug 04, 2026 12:34",
		path:        fullPath,
	}}
	var output bytes.Buffer
	if _, err := renderWorktreeSelectorAtSize(&output, entries, 0, width, 4); err != nil {
		t.Fatal(err)
	}
	rendered := stripSelectorColors(output.String())
	if !strings.Contains(rendered, fullPath) {
		t.Fatalf("selector truncated PATH %q:\n%s", fullPath, rendered)
	}
	wantTitle := truncateTerminalLine(fullTitle, selectorTitleWidth(entries, width))
	if strings.Contains(rendered, fullTitle) || !strings.Contains(rendered, wantTitle) {
		t.Fatalf("selector did not truncate TITLE:\n%s", rendered)
	}
	if strings.Contains(rendered, entries[0].branch) {
		t.Fatalf("selector did not constrain the fixed HEAD column:\n%s", rendered)
	}
	lines := strings.Split(strings.TrimSpace(rendered), "\r\n")
	header := lines[len(lines)-2]
	prIndex := strings.Index(header, "PR#")
	titleIndex := strings.Index(header, "TITLE")
	updatedIndex := strings.Index(header, "LAST UPDATED")
	pathIndex := strings.Index(header, "PATH")
	if prIndex < 0 || prIndex >= titleIndex || titleIndex >= updatedIndex || updatedIndex >= pathIndex {
		t.Fatalf("selector header is missing TITLE ordering:\n%s", rendered)
	}
	if got := utf8.RuneCountInString(lines[len(lines)-1]); got > width {
		t.Fatalf("entry width = %d, want at most %d:\n%s", got, width, rendered)
	}

	var narrow bytes.Buffer
	lineCount, err := renderWorktreeSelectorAtSize(&narrow, entries, 0, 50, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(narrow.String(), fullPath) {
		t.Fatalf("narrow selector truncated PATH %q:\n%s", fullPath, narrow.String())
	}
	if lineCount <= len(entries)+2 {
		t.Fatalf("narrow selector line count = %d, want wrapped physical rows", lineCount)
	}
}

func stripSelectorColors(value string) string {
	for _, sequence := range []string{
		colorReset,
		colorBold,
		colorBrightCyan,
		colorBrightMagenta,
		colorGreen,
		colorYellow,
		colorRed,
	} {
		value = strings.ReplaceAll(value, sequence, "")
	}
	return value
}

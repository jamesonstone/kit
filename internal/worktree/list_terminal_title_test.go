package worktree

import (
	"bytes"
	"strings"
	"testing"
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
	if got := terminalDisplayWidth(lines[len(lines)-1]); got > width {
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

func TestSelectorTerminalWidthsHandleWideAndCombiningCharacters(t *testing.T) {
	t.Parallel()

	if got, want := terminalDisplayWidth("界"), 2; got != want {
		t.Fatalf("wide character width = %d, want %d", got, want)
	}
	if got, want := terminalDisplayWidth("e\u0301"), 1; got != want {
		t.Fatalf("combining character width = %d, want %d", got, want)
	}
	if got, want := selectorTitleWidth([]worktreeEntry{{prTitle: "界界界", path: "/tmp"}}, 100), 6; got != want {
		t.Fatalf("wide title width = %d, want %d", got, want)
	}
	if got, want := truncateTerminalLine("界界界", 5), "界..."; got != want {
		t.Fatalf("truncated wide title = %q, want %q", got, want)
	}
	if got, want := truncateTerminalLine(strings.Repeat("e\u0301", 6), 5), strings.Repeat("e\u0301", 2)+"..."; got != want {
		t.Fatalf("truncated combining title = %q, want %q", got, want)
	}
	if got, want := terminalDisplayWidth(padTerminalLine("界", 4)), 4; got != want {
		t.Fatalf("padded wide title width = %d, want %d", got, want)
	}
	if got, want := selectorDisplayRows("界界界", 5), 2; got != want {
		t.Fatalf("wide title rows = %d, want %d", got, want)
	}
	if got, want := selectorDisplayRows(strings.Repeat("e\u0301", 5), 5), 1; got != want {
		t.Fatalf("combining title rows = %d, want %d", got, want)
	}
}

func TestRenderWorktreeSelectorKeepsUnicodePathOnOnePhysicalRow(t *testing.T) {
	t.Parallel()
	const (
		width    = 90
		fullPath = "/tmp/界/e\u0301"
	)
	entries := []worktreeEntry{{
		branch:      "unicode-width",
		state:       "clean",
		prText:      "120",
		prTitle:     strings.Repeat("界", 8) + strings.Repeat("e\u0301", 8),
		updatedText: "Aug 04, 2026 12:34",
		path:        fullPath,
	}}
	var output bytes.Buffer
	lineCount, err := renderWorktreeSelectorAtSize(&output, entries, 0, width, 4)
	if err != nil {
		t.Fatal(err)
	}
	rendered := stripSelectorColors(output.String())
	if !strings.Contains(rendered, fullPath) {
		t.Fatalf("selector truncated Unicode PATH %q:\n%s", fullPath, rendered)
	}
	if got, want := lineCount, len(entries)+2; got != want {
		t.Fatalf("selector line count = %d, want %d:\n%s", got, want, rendered)
	}
	lines := strings.Split(strings.TrimSpace(rendered), "\r\n")
	if got := terminalDisplayWidth(lines[len(lines)-1]); got > width {
		t.Fatalf("Unicode entry width = %d, want at most %d:\n%s", got, width, rendered)
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

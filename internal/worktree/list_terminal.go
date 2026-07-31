package worktree

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	colorReset         = "\x1b[0m"
	colorBold          = "\x1b[1m"
	colorBrightCyan    = "\x1b[1;36m"
	colorBrightMagenta = "\x1b[1;95m"
	colorGreen         = "\x1b[32m"
	colorYellow        = "\x1b[33m"
	colorRed           = "\x1b[31m"
	hideCursor         = "\x1b[?25l"
	showCursor         = "\x1b[?25h"
)

func newListInteraction(out io.Writer) (func() bool, listSelectorFunc) {
	output, outputIsFile := out.(*os.File)
	isTerminal := func() bool {
		return outputIsFile &&
			term.IsTerminal(int(os.Stdin.Fd())) &&
			term.IsTerminal(int(output.Fd()))
	}
	selector := func(ctx context.Context, entries []worktreeEntry) (worktreeEntry, bool, error) {
		if !outputIsFile {
			return worktreeEntry{}, false, errors.New("interactive list output is not a terminal")
		}
		return selectWorktreeTerminal(ctx, os.Stdin, output, entries)
	}
	return isTerminal, selector
}

func selectWorktreeTerminal(ctx context.Context, input, output *os.File, entries []worktreeEntry) (selected worktreeEntry, ok bool, err error) {
	if len(entries) == 0 {
		return worktreeEntry{}, false, errors.New("no worktrees are registered")
	}
	inputFD := int(input.Fd())
	previousState, err := term.MakeRaw(inputFD)
	if err != nil {
		return worktreeEntry{}, false, fmt.Errorf("enable interactive worktree selection: %w", err)
	}
	defer func() {
		if restoreErr := term.Restore(inputFD, previousState); err == nil && restoreErr != nil {
			err = fmt.Errorf("restore terminal: %w", restoreErr)
		}
	}()

	if _, err := fmt.Fprint(output, hideCursor); err != nil {
		return worktreeEntry{}, false, fmt.Errorf("hide terminal cursor: %w", err)
	}
	lineCount := 0
	defer func() {
		if lineCount > 0 {
			if clearErr := clearWorktreeSelector(output, lineCount); err == nil && clearErr != nil {
				err = clearErr
			}
		}
		if _, cursorErr := fmt.Fprint(output, showCursor); err == nil && cursorErr != nil {
			err = fmt.Errorf("show terminal cursor: %w", cursorErr)
		}
	}()
	lineCount, err = renderWorktreeSelector(output, entries, 0)
	if err != nil {
		return worktreeEntry{}, false, err
	}

	selectorInput := io.Reader(input)
	if ctx.Done() != nil {
		var restoreInput func() error
		selectorInput, restoreInput, err = newContextTerminalReader(ctx, input)
		if err != nil {
			return worktreeEntry{}, false, fmt.Errorf("enable cancellable terminal input: %w", err)
		}
		defer func() {
			if restoreErr := restoreInput(); err == nil && restoreErr != nil {
				err = fmt.Errorf("restore terminal input: %w", restoreErr)
			}
		}()
	}

	current := 0
	for {
		select {
		case <-ctx.Done():
			return worktreeEntry{}, false, ctx.Err()
		default:
		}

		key, readErr := readSelectorKey(selectorInput)
		if readErr != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return worktreeEntry{}, false, ctxErr
			}
			return worktreeEntry{}, false, fmt.Errorf("read worktree selection: %w", readErr)
		}
		switch key {
		case selectorNext:
			current = (current + 1) % len(entries)
		case selectorPrevious:
			current = (current - 1 + len(entries)) % len(entries)
		case selectorChoose:
			return entries[current], true, nil
		case selectorHome:
			if home, found := primaryListEntry(entries); found {
				return home, true, nil
			}
			continue
		case selectorCancel:
			return worktreeEntry{}, false, nil
		default:
			continue
		}
		if err := clearWorktreeSelector(output, lineCount); err != nil {
			return worktreeEntry{}, false, err
		}
		lineCount, err = renderWorktreeSelector(output, entries, current)
		if err != nil {
			return worktreeEntry{}, false, err
		}
	}
}

func renderWorktreeSelector(output *os.File, entries []worktreeEntry, selected int) (int, error) {
	width, height := 120, len(entries)+3
	if terminalWidth, terminalHeight, err := term.GetSize(int(output.Fd())); err == nil {
		if terminalWidth > 0 {
			width = terminalWidth
		}
		if terminalHeight > 3 {
			height = terminalHeight
		}
	}
	visibleCount := min(len(entries), max(height-3, 1))
	start := max(selected-visibleCount+1, 0)
	if start+visibleCount > len(entries) {
		start = max(len(entries)-visibleCount, 0)
	}
	end := min(start+visibleCount, len(entries))

	titleText := fmt.Sprintf("Select a worktree (%d/%d)  arrows/Tab: move  Enter: open  h: home  q: cancel", selected+1, len(entries))
	title := truncateTerminalLine(titleText, width)
	headerText := fmt.Sprintf("  %-8s %-24s %-6s %-18s %s", "STATE", "HEAD", "PR#", "LAST UPDATED", "PATH")
	header := truncateTerminalLine(headerText, width)
	if _, err := fmt.Fprintf(output, "%s%s%s\r\n", colorBold, title, colorReset); err != nil {
		return 0, fmt.Errorf("render worktree selector title: %w", err)
	}
	if _, err := fmt.Fprintf(output, "%s%s%s\r\n", colorBold, header, colorReset); err != nil {
		return 0, fmt.Errorf("render worktree selector header: %w", err)
	}
	for i := start; i < end; i++ {
		entry := entries[i]
		pointer := " "
		state := sanitizeTerminalField(entry.state)
		head := sanitizeTerminalField(selectorDisplayHead(entry))
		pr := sanitizeTerminalField(entry.prText)
		updated := sanitizeTerminalField(entry.updatedText)
		path := sanitizeTerminalField(entry.path)
		color := selectorEntryColor(entry, state, i == selected)
		if i == selected {
			pointer = ">"
		}
		line := fmt.Sprintf("%s %-8s %-24s %-6s %-18s %s", pointer, state, head, pr, updated, path)
		if _, err := fmt.Fprintf(output, "%s%s%s\r\n", color, truncateTerminalLine(line, width), colorReset); err != nil {
			return 0, fmt.Errorf("render worktree selector entry: %w", err)
		}
	}
	return end - start + 2, nil
}

func primaryListEntry(entries []worktreeEntry) (worktreeEntry, bool) {
	for _, entry := range entries {
		if entry.primary {
			return entry, true
		}
	}
	return worktreeEntry{}, false
}

func selectorDisplayHead(entry worktreeEntry) string {
	head := displayHead(entry)
	if entry.primary {
		return head + " [home]"
	}
	if entry.branch == "main" {
		return head + " [main]"
	}
	return head
}

func selectorEntryColor(entry worktreeEntry, state string, selected bool) string {
	if entry.primary || entry.branch == "main" {
		return colorBrightMagenta
	}
	if selected {
		return colorBrightCyan
	}
	return selectorStateColor(state)
}

func selectorStateColor(state string) string {
	switch state {
	case "clean":
		return colorGreen
	case "dirty":
		return colorYellow
	default:
		return colorRed
	}
}

func clearWorktreeSelector(output *os.File, lineCount int) error {
	if _, err := fmt.Fprintf(output, "\x1b[%dA", lineCount); err != nil {
		return fmt.Errorf("position worktree selector: %w", err)
	}
	for i := 0; i < lineCount; i++ {
		if _, err := fmt.Fprint(output, "\r\x1b[2K"); err != nil {
			return fmt.Errorf("clear worktree selector: %w", err)
		}
		if i < lineCount-1 {
			if _, err := fmt.Fprint(output, "\n"); err != nil {
				return fmt.Errorf("clear worktree selector: %w", err)
			}
		}
	}
	if lineCount > 1 {
		if _, err := fmt.Fprintf(output, "\x1b[%dA", lineCount-1); err != nil {
			return fmt.Errorf("position worktree selector: %w", err)
		}
	}
	if _, err := fmt.Fprint(output, "\r"); err != nil {
		return fmt.Errorf("position worktree selector: %w", err)
	}
	return nil
}

func sanitizeTerminalField(value string) string {
	var sanitized strings.Builder
	sanitized.Grow(len(value))
	for len(value) > 0 {
		char, size := utf8.DecodeRuneInString(value)
		if char == utf8.RuneError && size == 1 {
			if value[0] < 0x20 || value[0] >= 0x7f && value[0] <= 0x9f {
				value = value[1:]
				continue
			}
			sanitized.WriteRune(utf8.RuneError)
			value = value[1:]
			continue
		}
		if !unicode.IsControl(char) {
			sanitized.WriteString(value[:size])
		}
		value = value[size:]
	}
	return sanitized.String()
}

func truncateTerminalLine(value string, width int) string {
	if width <= 3 {
		return strings.Repeat(".", max(width, 0))
	}
	if utf8.RuneCountInString(value) <= width {
		return value
	}
	runes := []rune(value)
	return string(runes[:width-3]) + "..."
}

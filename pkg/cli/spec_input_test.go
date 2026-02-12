package cli

import (
	"testing"

	"github.com/chzyer/readline"
)

func TestSpecInputRuneFilter_CtrlJToNewline(t *testing.T) {
	got, ok := specInputRuneFilter(readline.CharCtrlJ)
	if !ok {
		t.Fatalf("expected rune to be processed")
	}
	if got != '\n' {
		t.Fatalf("expected newline rune, got %q", got)
	}
}

func TestSpecInputRuneFilter_EnterUnchanged(t *testing.T) {
	got, ok := specInputRuneFilter(readline.CharEnter)
	if !ok {
		t.Fatalf("expected rune to be processed")
	}
	if got != readline.CharEnter {
		t.Fatalf("expected enter rune %q, got %q", readline.CharEnter, got)
	}
}

func TestNormalizeSpecAnswer_TrimsOuterWhitespace(t *testing.T) {
	raw := "  first line\nsecond line  \n"
	got := normalizeSpecAnswer(raw)
	want := "first line\nsecond line"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNormalizeSpecAnswer_WhitespaceOnlyBecomesEmpty(t *testing.T) {
	got := normalizeSpecAnswer(" \n\t ")
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

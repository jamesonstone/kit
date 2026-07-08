package cli

import (
	"bytes"
	"io"
	"testing"

	"github.com/chzyer/readline"
)

func TestMultilineInputRuneFilter_CtrlJToSentinel(t *testing.T) {
	got, ok := multilineInputRuneFilter(readline.CharCtrlJ)
	if !ok {
		t.Fatalf("expected rune to be processed")
	}
	if got != multilineInputSentinel {
		t.Fatalf("expected sentinel rune, got %q", got)
	}
}

func TestMultilineInputRuneFilter_EnterUnchanged(t *testing.T) {
	got, ok := multilineInputRuneFilter(readline.CharEnter)
	if !ok {
		t.Fatalf("expected rune to be processed")
	}
	if got != readline.CharEnter {
		t.Fatalf("expected enter rune %q, got %q", readline.CharEnter, got)
	}
}

func TestMultilineInputListener_ReplacesSentinelWithNewline(t *testing.T) {
	line := []rune{'f', 'o', multilineInputSentinel, multilineInputSentinel, 'b', 'a', 'r'}

	got, pos, ok := multilineInputListener(line, 4, multilineInputSentinel)
	if !ok {
		t.Fatalf("expected listener update")
	}
	if pos != 4 {
		t.Fatalf("expected cursor position 4, got %d", pos)
	}

	want := []rune{'f', 'o', '\n', '\n', 'b', 'a', 'r'}
	if string(got) != string(want) {
		t.Fatalf("expected %q, got %q", string(want), string(got))
	}
}

func TestTranslateMultilineEscapeSequence_ShiftEnterToSentinel(t *testing.T) {
	got, ok := translateMultilineEscapeSequence([]byte("\x1b[13;2u"))
	if !ok {
		t.Fatalf("expected shift+enter escape to be translated")
	}

	want := []byte(string(multilineInputSentinel))
	if !bytes.Equal(got, want) {
		t.Fatalf("expected %q, got %q", string(want), string(got))
	}
}

func TestTranslateMultilineEscapeSequence_NonShiftEnterUnchanged(t *testing.T) {
	if _, ok := translateMultilineEscapeSequence([]byte("\x1b[13;3u")); ok {
		t.Fatalf("expected alt+enter escape to remain unchanged")
	}
}

func TestMultilineInputReader_TranslatesShiftEnterSequence(t *testing.T) {
	reader := newMultilineInputReader(bytes.NewBufferString("one\x1b[13;2utwo"))
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("expected read to succeed: %v", err)
	}

	want := append([]byte("one"), []byte(string(multilineInputSentinel))...)
	want = append(want, []byte("two")...)
	if !bytes.Equal(got, want) {
		t.Fatalf("expected %q, got %q", string(want), string(got))
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

func TestNormalizeSpecDeliveryIntentAcceptsNumberedDefaults(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty defaults to issue branch PR later", raw: "", want: specDeliveryIntentIssueBranchPRLater},
		{name: "one selects issue branch PR later", raw: "1", want: specDeliveryIntentIssueBranchPRLater},
		{name: "yes alias selects issue branch PR later", raw: "yes", want: specDeliveryIntentIssueBranchPRLater},
		{name: "y alias selects issue branch PR later", raw: "y", want: specDeliveryIntentIssueBranchPRLater},
		{name: "two selects idea only", raw: "2", want: specDeliveryIntentIdeaOnly},
		{name: "no alias selects idea only", raw: "no", want: specDeliveryIntentIdeaOnly},
		{name: "three selects continue current", raw: "3", want: specDeliveryIntentContinueCurrent},
		{name: "continue alias selects continue current", raw: "continue", want: specDeliveryIntentContinueCurrent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeSpecDeliveryIntent(tt.raw)
			if err != nil {
				t.Fatalf("normalizeSpecDeliveryIntent(%q) error = %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeSpecDeliveryIntent(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNormalizeSpecDeliveryIntentRejectsInvalidNumber(t *testing.T) {
	if got, err := normalizeSpecDeliveryIntent("4"); err == nil {
		t.Fatalf("normalizeSpecDeliveryIntent(4) = %q, want error", got)
	}
}

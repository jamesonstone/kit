package cli

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

const (
	multilineInputSentinel  rune = '\ue000'
	keyboardProtocolEnable       = "\x1b[>1u"
	keyboardProtocolRestore      = "\x1b[<u"
	keyboardEscape          byte = 0x1b
)

func newMultilineReadline() (*readline.Instance, error) {
	enableKeyboardProtocol(os.Stdout)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:              whiteBold + "   > " + reset,
		InterruptPrompt:     "^C",
		EOFPrompt:           "",
		Stdin:               io.NopCloser(newMultilineInputReader(os.Stdin)),
		Stdout:              os.Stdout,
		Stderr:              os.Stderr,
		FuncFilterInputRune: multilineInputRuneFilter,
		Listener:            readline.FuncListener(multilineInputListener),
	})
	if err != nil {
		restoreKeyboardProtocol(os.Stdout)
		return nil, err
	}

	return rl, nil
}

func closeMultilineReadline(rl *readline.Instance) {
	restoreKeyboardProtocol(os.Stdout)
	_ = rl.Close()
}

func enableKeyboardProtocol(w io.Writer) {
	_, _ = io.WriteString(w, keyboardProtocolEnable)
}

func restoreKeyboardProtocol(w io.Writer) {
	_, _ = io.WriteString(w, keyboardProtocolRestore)
}

func multilineInputRuneFilter(r rune) (rune, bool) {
	if r == readline.CharCtrlJ {
		return multilineInputSentinel, true
	}
	return r, true
}

func multilineInputListener(line []rune, pos int, key rune) ([]rune, int, bool) {
	if key != multilineInputSentinel {
		return nil, 0, false
	}

	updated := false
	for i, r := range line {
		if r == multilineInputSentinel {
			line[i] = '\n'
			updated = true
		}
	}

	if !updated {
		return nil, 0, false
	}

	return line, pos, true
}

type multilineInputReader struct {
	source  io.Reader
	pending []byte
}

func newMultilineInputReader(source io.Reader) *multilineInputReader {
	return &multilineInputReader{source: source}
}

func (r *multilineInputReader) Read(p []byte) (int, error) {
	if len(r.pending) == 0 {
		if err := r.fillPending(); err != nil {
			return 0, err
		}
	}

	n := copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

func (r *multilineInputReader) fillPending() error {
	first, err := r.readByte()
	if err != nil {
		return err
	}

	if first != keyboardEscape {
		r.pending = append(r.pending, first)
		return nil
	}

	sequence, sequenceErr := r.readEscapeSequence(first)
	if translated, ok := translateMultilineEscapeSequence(sequence); ok {
		r.pending = append(r.pending, translated...)
	} else {
		r.pending = append(r.pending, sequence...)
	}

	if len(r.pending) > 0 {
		return nil
	}

	return sequenceErr
}

func (r *multilineInputReader) readEscapeSequence(first byte) ([]byte, error) {
	second, err := r.readByte()
	if err != nil {
		return []byte{first}, err
	}

	sequence := []byte{first, second}
	if second != '[' {
		return sequence, nil
	}

	for len(sequence) < 32 {
		next, err := r.readByte()
		if err != nil {
			return sequence, err
		}

		sequence = append(sequence, next)
		if next >= 0x40 && next <= 0x7e {
			return sequence, nil
		}
	}

	return sequence, nil
}

func (r *multilineInputReader) readByte() (byte, error) {
	buf := make([]byte, 1)
	if _, err := r.source.Read(buf); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func translateMultilineEscapeSequence(sequence []byte) ([]byte, bool) {
	if len(sequence) < 4 || sequence[0] != keyboardEscape || sequence[1] != '[' {
		return nil, false
	}
	if sequence[len(sequence)-1] != 'u' {
		return nil, false
	}

	body := string(sequence[2 : len(sequence)-1])
	parts := strings.SplitN(body, ";", 2)
	if len(parts) != 2 || parts[0] != "13" {
		return nil, false
	}

	modifierField := strings.SplitN(parts[1], ":", 2)[0]
	modifier, err := strconv.Atoi(modifierField)
	if err != nil || modifier < 1 {
		return nil, false
	}

	if (modifier-1)&1 == 0 {
		return nil, false
	}

	return []byte(string(multilineInputSentinel)), true
}

package worktree

import (
	"io"
)

type selectorKey int

const (
	selectorIgnore selectorKey = iota
	selectorNext
	selectorPrevious
	selectorChoose
	selectorHome
	selectorCancel
)

func readSelectorKey(input io.Reader) (selectorKey, error) {
	var first [1]byte
	if _, err := io.ReadFull(input, first[:]); err != nil {
		return selectorIgnore, err
	}
	switch first[0] {
	case '\r', '\n':
		return selectorChoose, nil
	case '\t', 'j', 'J':
		return selectorNext, nil
	case 'k', 'K':
		return selectorPrevious, nil
	case 'h', 'H':
		return selectorHome, nil
	case 'q', 'Q', 3:
		return selectorCancel, nil
	case 0x1b:
		return readEscapeSequence(input)
	default:
		return selectorIgnore, nil
	}
}

func readEscapeSequence(input io.Reader) (selectorKey, error) {
	var sequence [2]byte
	if _, err := io.ReadFull(input, sequence[:]); err != nil {
		return selectorIgnore, err
	}
	if sequence[0] != '[' && sequence[0] != 'O' {
		return selectorIgnore, nil
	}
	switch sequence[1] {
	case 'A', 'D', 'Z':
		return selectorPrevious, nil
	case 'B', 'C':
		return selectorNext, nil
	default:
		return selectorIgnore, nil
	}
}

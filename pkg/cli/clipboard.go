package cli

import (
	"os/exec"
	"strings"
)

var clipboardReadFunc = readFromClipboard

func readFromClipboard() (string, error) {
	output, err := exec.Command("pbpaste").Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

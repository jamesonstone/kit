package agentcli

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

var clipboardCopyFunc = copyToClipboard

func copyToClipboard(content string) error {
	name, args := clipboardCommand()
	if name == "" {
		return fmt.Errorf("no supported clipboard command is available; use --output-only")
	}
	// #nosec G204 -- name and args come only from clipboardCommand's fixed allowlist.
	command := exec.Command(name, args...)
	command.Stdin = strings.NewReader(content)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func clipboardCommand() (string, []string) {
	candidates := []struct {
		name string
		args []string
	}{{"wl-copy", nil}, {"xclip", []string{"-selection", "clipboard"}}, {"xsel", []string{"--clipboard", "--input"}}}
	if runtime.GOOS == "darwin" {
		candidates = append([]struct {
			name string
			args []string
		}{{"pbcopy", nil}}, candidates...)
	}
	if runtime.GOOS == "windows" {
		candidates = append([]struct {
			name string
			args []string
		}{{"clip", nil}}, candidates...)
	}
	for _, candidate := range candidates {
		if _, err := exec.LookPath(candidate.name); err == nil {
			return candidate.name, candidate.args
		}
	}
	return "", nil
}

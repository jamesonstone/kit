package agentcli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var prFixEditor = editPRFixTasks

func shouldEditPRFix(options prFixOptions) bool {
	return options.Edit || options.UseVim || strings.TrimSpace(options.Editor) != ""
}

func editPRFixTasks(command *cobra.Command, options prFixOptions, initial string) (string, error) {
	name, args, err := selectPREditor(options)
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp("", "kit-pr-fix-*.md")
	if err != nil {
		return "", err
	}
	path := file.Name()
	defer func() { _ = os.Remove(path) }()
	if _, err := file.WriteString(initial + "\n"); err != nil {
		_ = file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	// #nosec G204 -- the user explicitly selected this editor; no shell is used.
	editor := exec.CommandContext(command.Context(), name, append(args, path)...)
	editor.Stdin = command.InOrStdin()
	editor.Stdout = command.ErrOrStderr()
	editor.Stderr = command.ErrOrStderr()
	if err := editor.Run(); err != nil {
		return "", fmt.Errorf("edit PR feedback: %w", err)
	}
	// #nosec G304 -- path is the exact file returned by os.CreateTemp above.
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(content)) == "" {
		return "", fmt.Errorf("edited PR feedback cannot be empty")
	}
	return strings.TrimSpace(string(content)), nil
}

func selectPREditor(options prFixOptions) (string, []string, error) {
	value := strings.TrimSpace(options.Editor)
	if options.UseVim {
		for _, candidate := range []string{"nvim", "vim"} {
			if path, err := exec.LookPath(candidate); err == nil {
				return path, nil, nil
			}
		}
		return "", nil, fmt.Errorf("neither nvim nor vim is available")
	}
	if options.Edit {
		value = strings.TrimSpace(os.Getenv("VISUAL"))
		if value == "" {
			value = strings.TrimSpace(os.Getenv("EDITOR"))
		}
	}
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("no editor configured; set VISUAL or EDITOR, or pass --editor")
	}
	path, err := exec.LookPath(parts[0])
	if err != nil {
		return "", nil, fmt.Errorf("find editor %q: %w", parts[0], err)
	}
	return path, parts[1:], nil
}

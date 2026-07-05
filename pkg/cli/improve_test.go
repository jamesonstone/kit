package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestImproveCommandOverview(t *testing.T) {
	cmd := newImproveCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("kit improve error = %v", err)
	}
	content := out.String()
	for _, want := range []string{
		"Kit improve",
		"kit improve run --suite default",
		"kit improve mine --from .kit/improve/latest",
		"kit improve validate --candidate <path>",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected overview to contain %q, got %q", want, content)
		}
	}
}

func TestImproveCommandRegistersSubcommands(t *testing.T) {
	for _, commandPath := range [][]string{
		{"improve", "run"},
		{"improve", "mine"},
		{"improve", "propose"},
		{"improve", "validate"},
		{"improve", "report"},
		{"improve", "pr"},
	} {
		cmd, _, err := rootCmd.Find(commandPath)
		if err != nil {
			t.Fatalf("rootCmd.Find(%v) error = %v", commandPath, err)
		}
		if cmd == nil {
			t.Fatalf("expected command %v to be registered", commandPath)
		}
	}
}

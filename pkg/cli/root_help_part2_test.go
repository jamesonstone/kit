package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRemovedCompatibilityCommandsAreNotRegistered(t *testing.T) {
	removed := [][]string{
		{"update"},
		{"skills"},
		{"skills", "mine"},
		{"catchup"},
		{"rollup"},
		{"review-loop"},
		{"brainstorm"},
		{"plan"},
		{"tasks"},
		{"implement"},
		{"reflect"},
		{"verify"},
	}

	for _, args := range removed {
		if cmd, _, err := rootCmd.Find(args); err == nil && cmd != nil && cmd.CommandPath() == "kit "+strings.Join(args, " ") {
			t.Fatalf("expected %q to be removed", strings.Join(args, " "))
		}
	}
}

func TestLegacyNamespaceCommandsAreRegistered(t *testing.T) {
	for _, args := range [][]string{
		{"legacy", "brainstorm"},
		{"legacy", "plan"},
		{"legacy", "tasks"},
		{"legacy", "implement"},
		{"legacy", "reflect"},
		{"legacy", "verify"},
	} {
		cmd, _, err := rootCmd.Find(args)
		if err != nil {
			t.Fatalf("rootCmd.Find(%v) error = %v", args, err)
		}
		if cmd == nil || cmd.CommandPath() != "kit "+strings.Join(args, " ") {
			t.Fatalf("expected %q to be registered, got %#v", strings.Join(args, " "), cmd)
		}
	}
}

func TestScaffoldAgentsRootCommandIsNotRegistered(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "scaffold-agents" {
			t.Fatal("expected scaffold-agents root command to be removed")
		}
	}

	cmd, _, err := rootCmd.Find([]string{"scaffold", "agents"})
	if err != nil {
		t.Fatalf("rootCmd.Find(scaffold agents) error = %v", err)
	}
	if cmd != scaffoldAgentsCmd {
		t.Fatalf("expected scaffold agents to resolve to scaffoldAgentsCmd, got %q", cmd.CommandPath())
	}
	if cmd.Hidden {
		t.Fatal("expected scaffold agents to be visible")
	}
}

func TestPromptLibraryCommandsRemainRegisteredAndDiscoverable(t *testing.T) {
	tests := [][]string{
		{"prompt"},
		{"prompt", "list"},
		{"set"},
		{"set", "prompt"},
	}

	for _, args := range tests {
		cmd, _, err := rootCmd.Find(args)
		if err != nil {
			t.Fatalf("rootCmd.Find(%v) error = %v", args, err)
		}
		if cmd == nil {
			t.Fatalf("expected command %v to be registered", args)
		}
		if cmd.Hidden {
			t.Fatalf("expected command %v to be visible", args)
		}
	}
}

func TestRemoveCommandSupportsRmAndRemove(t *testing.T) {
	rmCommand, _, err := rootCmd.Find([]string{"rm"})
	if err != nil {
		t.Fatalf("rootCmd.Find(rm) error = %v", err)
	}
	if rmCommand == nil {
		t.Fatal("expected rm command to be registered")
	}
	if rmCommand.Name() != "rm" {
		t.Fatalf("expected rm command name, got %q", rmCommand.Name())
	}

	removeCommand, _, err := rootCmd.Find([]string{"remove"})
	if err != nil {
		t.Fatalf("rootCmd.Find(remove) error = %v", err)
	}
	if removeCommand != rmCommand {
		t.Fatalf("expected remove alias to resolve to rm command")
	}
	if flag := rmCommand.Flags().Lookup("notes"); flag == nil {
		t.Fatal("expected rm command to expose --notes")
	}
}

func TestPromptLibraryHelpShowsSupportedFlagsOnly(t *testing.T) {
	promptHelp := executeHelp(t, []string{"prompt", "--help"})
	for _, check := range []string{"list", "--copy", "--output-only"} {
		if !strings.Contains(promptHelp, check) {
			t.Fatalf("expected prompt help to contain %q, got %q", check, promptHelp)
		}
	}
	for _, unsupported := range []string{"--source", "--no-copy"} {
		if strings.Contains(promptHelp, unsupported) {
			t.Fatalf("expected prompt help to omit %q, got %q", unsupported, promptHelp)
		}
	}

	setPromptHelp := executeHelp(t, []string{"set", "prompt", "--help"})
	for _, check := range []string{"--local", "--global"} {
		if !strings.Contains(setPromptHelp, check) {
			t.Fatalf("expected set prompt help to contain %q, got %q", check, setPromptHelp)
		}
	}
	for _, unsupported := range []string{"--file", "--source", "--no-copy"} {
		if strings.Contains(setPromptHelp, unsupported) {
			t.Fatalf("expected set prompt help to omit %q, got %q", unsupported, setPromptHelp)
		}
	}
}

func executeHelp(t *testing.T, args []string) string {
	t.Helper()

	previousOut := rootCmd.OutOrStdout()
	previousErr := rootCmd.ErrOrStderr()
	defer func() {
		rootCmd.SetOut(previousOut)
		rootCmd.SetErr(previousErr)
		rootCmd.SetArgs(nil)
	}()

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute(%v) error = %v", args, err)
	}
	return out.String()
}

func TestBrainstormPickupFlagIsRemoved(t *testing.T) {
	if flag := brainstormCmd.Flags().Lookup("pickup"); flag != nil {
		t.Fatalf("expected brainstorm pickup flag to be removed, got %#v", flag)
	}
}

func stripANSI(input string) string {
	return ansiEscapeRE.ReplaceAllString(input, "")
}

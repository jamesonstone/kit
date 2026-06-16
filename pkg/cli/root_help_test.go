package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpGroupsCanonicalCommands(t *testing.T) {
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
	rootCmd.SetArgs([]string{"--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	content := out.String()
	checks := []string{
		"Setup",
		"Workflow",
		"Inspect & Repair",
		"Prompt Utilities",
		"Utilities",
		"scaffold",
		"prompt",
		"set",
		"resume",
		"status",
		"capabilities",
		"ci",
		"rules",
		"loop",
		"\n  rm ",
		"upgrade",
		"skill",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Fatalf("expected root help to contain %q, got %q", check, content)
		}
	}

	hiddenChecks := []string{
		"\n  update ",
		"\n  skills ",
		"\n  catchup ",
		"\n  review-loop ",
		"\n  scaffold-agents ",
		"\n  rollup ",
	}
	for _, check := range hiddenChecks {
		if strings.Contains(content, check) {
			t.Fatalf("expected root help to omit hidden command %q, got %q", check, content)
		}
	}
}

func TestDeprecatedCommandsRemainRegisteredAndHidden(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "update", want: "kit upgrade"},
		{name: "skills", want: "kit skill mine"},
		{name: "catchup", want: "kit resume"},
		{name: "rollup", want: "maintenance command"},
	}

	for _, tt := range tests {
		cmd, _, err := rootCmd.Find([]string{tt.name})
		if err != nil {
			t.Fatalf("rootCmd.Find(%q) error = %v", tt.name, err)
		}
		if !cmd.Hidden {
			t.Fatalf("expected %q to be hidden", tt.name)
		}
		if !strings.Contains(cmd.Deprecated, tt.want) {
			t.Fatalf("expected %q deprecation to contain %q, got %q", tt.name, tt.want, cmd.Deprecated)
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

func TestSkillsMineAliasCarriesDeprecationGuidance(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"skills", "mine"})
	if err != nil {
		t.Fatalf("rootCmd.Find(skills mine) error = %v", err)
	}
	if cmd.Name() != "mine" {
		t.Fatalf("expected mine subcommand, got %q", cmd.Name())
	}
	if !strings.Contains(cmd.Deprecated, "kit skill mine") {
		t.Fatalf("expected skills mine deprecation guidance, got %q", cmd.Deprecated)
	}
}

func TestBrainstormPickupFlagIsHiddenAndDeprecated(t *testing.T) {
	flag := brainstormCmd.Flags().Lookup("pickup")
	if flag == nil {
		t.Fatal("expected brainstorm pickup flag to exist")
	}
	if !flag.Hidden {
		t.Fatal("expected brainstorm pickup flag to be hidden")
	}
	if !strings.Contains(flag.Deprecated, "kit resume <feature>") {
		t.Fatalf("expected pickup deprecation guidance, got %q", flag.Deprecated)
	}
}

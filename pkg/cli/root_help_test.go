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
		"resume",
		"status",
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
		"\n  scaffold ",
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
		{name: "scaffold", want: "kit brainstorm"},
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

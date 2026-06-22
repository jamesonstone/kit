package cli

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"
)

var ansiEscapeRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

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

	raw := out.String()
	if ansiEscapeRE.MatchString(raw) {
		t.Fatalf("expected non-terminal root help to omit ANSI escapes, got %q", raw)
	}

	content := stripANSI(raw)
	checks := []string{
		"Setup",
		"Workflow",
		"Inspect & Repair",
		"Prompt Utilities",
		"Utilities",
		"V2 Feature Workflow",
		"kit spec <feature>",
		"Idea / input",
		"Clarifying Loop",
		"source map",
		"binary acceptance criteria",
		"Supervisor + Agent Team Plan",
		"Subagent Implementation",
		"Subagent Reflection",
		"Subagent Validation / Verification",
		"Evidence + Delivery Gate",
		"Durable Artifacts",
		"v2 feature artifact",
		"legacy v1 artifacts",
		"legacy",
		"List deprecated v1 staged workflow commands",
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
		"\n  brainstorm ",
		"\n  plan ",
		"\n  tasks ",
		"\n  implement ",
		"\n  reflect ",
		"\n  verify ",
	}
	for _, check := range hiddenChecks {
		if strings.Contains(content, check) {
			t.Fatalf("expected root help to omit hidden command %q, got %q", check, content)
		}
	}

	staleWorkflowChecks := []string{
		"Optional Research Step",
		"Artifact Pipeline",
		"Specification │ ─▶ │ Plan",
		"Tasks │ ─▶ │ Implementation",
		"Reflection     — verify correctness",
	}
	for _, check := range staleWorkflowChecks {
		if strings.Contains(content, check) {
			t.Fatalf("expected root help to omit stale v1 workflow text %q, got %q", check, content)
		}
	}
}

func TestRootNoCommandShowsV2WorkflowDiagram(t *testing.T) {
	content := stripANSI(executeHelp(t, []string{}))

	checks := []string{
		"Kit v2 Thought-Work Harness",
		"Idea / input",
		"kit spec <feature> creates/updates one durable SPEC.md",
		"Clarifying Loop",
		"Subagent Implementation",
		"Subagent Reflection",
		"Subagent Validation / Verification",
		"Evidence + Delivery Gate",
		"Usage",
		"Available Commands",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Fatalf("expected no-command root help to contain %q, got %q", check, content)
		}
	}
}

func TestSpecHelpUsesReadableV2InstructionStructure(t *testing.T) {
	content := stripANSI(executeHelp(t, []string{"spec", "--help"}))

	checks := []string{
		"Start or resume Kit v2 feature work from one durable SPEC.md.",
		"🧭 Human flow",
		"Pick or provide a feature slug/name.",
		"🧠 Agent workflow",
		"idea → clarification loop → agent-team implementation → reflection",
		"📦 What Kit writes",
		"🔁 Modes",
		"🚫 Git/GitHub safety",
		"kit spec records delivery intent only",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Fatalf("expected spec help to contain %q, got %q", check, content)
		}
	}
}

func TestSpecHelpUsesColorForTerminalOutput(t *testing.T) {
	previousCheck := terminalWriterCheck
	terminalWriterCheck = func(_ io.Writer) bool { return true }
	defer func() { terminalWriterCheck = previousCheck }()

	output := executeHelp(t, []string{"spec", "--help"})
	if !ansiEscapeRE.MatchString(output) {
		t.Fatalf("expected terminal spec help to include ANSI color, got %q", output)
	}

	content := stripANSI(output)
	for _, check := range []string{
		"🧭 Human flow",
		"🧠 Agent workflow",
		"🚫 Git/GitHub safety",
		"🚀 Usage",
		"⚙️ Flags",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected colored spec help to contain %q, got %q", check, content)
		}
	}
}

func TestRootHelpUsesColorForTerminalOutput(t *testing.T) {
	previousCheck := terminalWriterCheck
	terminalWriterCheck = func(_ io.Writer) bool { return true }
	defer func() { terminalWriterCheck = previousCheck }()

	output := executeHelp(t, []string{"--help"})
	if !ansiEscapeRE.MatchString(output) {
		t.Fatalf("expected terminal root help to include ANSI color, got %q", output)
	}

	content := stripANSI(output)
	for _, check := range []string{
		"Idea / input",
		"Clarifying Loop",
		"Subagent Implementation",
		"Subagent Reflection",
		"Subagent Validation / Verification",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected colored root help to contain %q, got %q", check, content)
		}
	}
}

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

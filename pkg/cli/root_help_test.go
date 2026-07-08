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
		"notes",
		"project",
		"status",
		"capabilities",
		"ci",
		"pr",
		"improve",
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
		"🧱 Setup gate",
		"copy the kit init prompt and stop",
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

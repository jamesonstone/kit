package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/threadtitle"
)

func executeTitleCommand(args ...string) (string, error) {
	cmd := newInstructionsTitleCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs(args)
	err := cmd.Execute()
	return output.String(), err
}

func TestInstructionsTitleOutput(t *testing.T) {
	base := []string{"--scope=kit", "--domain=agents", "--objective=thread naming"}
	got, err := executeTitleCommand(base...)
	if err != nil {
		t.Fatal(err)
	}
	if got != "[kit] agents / thread naming\n" {
		t.Fatalf("unexpected output %q", got)
	}
	got, err = executeTitleCommand(append(base, "--current=[kit] agents / thread naming")...)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("unchanged output = %q", got)
	}
	got, err = executeTitleCommand(append(base, "--current=[kit] agents / naming policy", "--accurate", "--json")...)
	if err != nil {
		t.Fatal(err)
	}
	var decision threadtitle.Decision
	if err := json.Unmarshal([]byte(got), &decision); err != nil {
		t.Fatal(err)
	}
	if decision.Changed || decision.Title != "[kit] agents / naming policy" || decision.Reason != "ownership_unchanged" {
		t.Fatalf("unexpected decision %+v", decision)
	}
}

func TestInstructionsTitleInvalidInput(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"--scope=kit", "--domain=agents"},
		{"--scope=kit,labcore", "--domain=agents", "--objective=naming"},
		{"--scope=kit", "--domain=agents", "--objective=naming\npolicy"},
		{"unexpected"},
	} {
		out, err := executeTitleCommand(args...)
		if err == nil {
			t.Fatalf("accepted %v", args)
		}
		if out != "" {
			t.Fatalf("invalid input printed %q", out)
		}
	}
}

func TestInstructionsTitleHelpExplainsReadOnlyOutput(t *testing.T) {
	out, err := executeTitleCommand("--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{"empty stdout", "does not infer intent", "--accurate", "--scope"} {
		if !strings.Contains(out, phrase) {
			t.Fatalf("help missing %q", phrase)
		}
	}
}

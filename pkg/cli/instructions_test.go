package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/instructions"
)

func TestInstructionsCommandPrintsCurrentVersionByDefault(t *testing.T) {
	want, err := instructions.AgentInstructions(instructions.CurrentAgentVersion)
	if err != nil {
		t.Fatalf("AgentInstructions(%q) error = %v", instructions.CurrentAgentVersion, err)
	}

	got, err := executeInstructionsCommand()
	if err != nil {
		t.Fatalf("kit instructions error = %v", err)
	}
	if got != want {
		t.Fatalf("kit instructions output does not match current version\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestInstructionsCommandPrintsExplicitVersion(t *testing.T) {
	want, err := instructions.AgentInstructions("v1")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v1\") error = %v", err)
	}

	got, err := executeInstructionsCommand("--version=v1")
	if err != nil {
		t.Fatalf("kit instructions --version=v1 error = %v", err)
	}
	if got != want {
		t.Fatal("kit instructions --version=v1 output does not match embedded v1")
	}
}

func TestInstructionsCommandRejectsInvalidVersions(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{name: "empty", arg: "--version=", want: "--version cannot be empty; available versions: v1"},
		{name: "malformed", arg: "--version=1", want: `unsupported instructions version "1"`},
		{name: "unavailable", arg: "--version=v2", want: `unsupported instructions version "v2"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := executeInstructionsCommand(test.arg)
			if err == nil {
				t.Fatalf("kit instructions %s expected an error", test.arg)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("kit instructions %s error = %q, want %q", test.arg, err, test.want)
			}
		})
	}
}

func TestInstructionsCommandIsProjectIndependent(t *testing.T) {
	if !skipAutomaticConfigCheck(instructionsCmd) {
		t.Fatal("kit instructions must skip automatic project config inspection")
	}
}

func TestInstructionsCapabilityIsReadOnlyAndComplete(t *testing.T) {
	record, ok := capabilityByCommandPath("instructions")
	if !ok {
		t.Fatal("instructions capability is missing")
	}
	detail := record.detail()
	if detail.MutationLevel != mutationNone || detail.NetworkUse.Summary != "none" || detail.FileWrites.Summary != "none" || detail.GitMutation.Summary != "none" {
		t.Fatalf("instructions capability is not read-only: %#v", detail)
	}
	if len(detail.ImportantFlags) != 1 || detail.ImportantFlags[0].Name != "--version" {
		t.Fatalf("instructions capability flags = %#v, want --version", detail.ImportantFlags)
	}
	if len(detail.WhenToUse) == 0 || len(detail.WhenNotToUse) == 0 || len(detail.Examples) < 2 || len(detail.Caveats) == 0 {
		t.Fatalf("instructions capability guidance is incomplete: %#v", detail)
	}
}

func executeInstructionsCommand(args ...string) (string, error) {
	cmd := newInstructionsCommand()
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return output.String(), err
}

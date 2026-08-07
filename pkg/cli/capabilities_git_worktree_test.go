package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCapabilitiesDescribeGitWTList(t *testing.T) {
	output, err := executeCapabilitiesCommand("--json", "git", "wt", "list")
	if err != nil {
		t.Fatalf("kit capabilities git wt list --json error = %v", err)
	}
	var payload capabilityDetailPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal(git wt list) error = %v", err)
	}
	if payload.Command.Command != "git wt list" {
		t.Fatalf("command = %q, want git wt list", payload.Command.Command)
	}
	if payload.Command.MutationLevel != mutationNetwork {
		t.Fatalf("mutation level = %q, want %q", payload.Command.MutationLevel, mutationNetwork)
	}
	if !strings.Contains(payload.Command.Summary, "primary checkout pinned") {
		t.Fatalf("expected list capability summary to document default ordering, got %q", payload.Command.Summary)
	}
	var sortSummary, rootPositionSummary string
	for _, flag := range payload.Command.ImportantFlags {
		if flag.Name == "--sort" {
			sortSummary = flag.Summary
		}
		if flag.Name == "--root-position" {
			rootPositionSummary = flag.Summary
		}
	}
	if !strings.Contains(sortSummary, "updated (default), state, head, or path") {
		t.Fatalf("expected --sort capability to document default and alternate orderings, got %q", sortSummary)
	}
	if !strings.Contains(rootPositionSummary, "top (default) or bottom") {
		t.Fatalf("expected --root-position capability to document primary pinning, got %q", rootPositionSummary)
	}
	combinedParts := append([]string{}, payload.Command.Examples...)
	combinedParts = append(combinedParts, payload.Command.Caveats...)
	combinedParts = append(combinedParts, payload.Command.WhenToUse...)
	combinedParts = append(combinedParts, payload.Command.WhenNotToUse...)
	combined := strings.Join(combinedParts, " ") +
		" " + payload.Command.NetworkUse.Summary +
		" " + payload.Command.NetworkUse.FlagDependent +
		" " + payload.Command.GitMutation.Summary
	for _, want := range []string{"--plain", "--root-position bottom", "arrow keys", "press h", "child shell", "bright magenta", "--sort state", "--sort head", "--sort path", "display-only", "local timezone", "HH:MM", "without seconds", "PR#", "TITLE", "complete PATH", "two-second", "NG", "RL", "TO", "??", "no local mutation"} {
		if !strings.Contains(combined, want) {
			t.Fatalf("expected list capability to mention %q, got %#v", want, payload.Command)
		}
	}
}

func TestCapabilitiesDescribeGitWTHome(t *testing.T) {
	output, err := executeCapabilitiesCommand("--json", "git", "wt", "home")
	if err != nil {
		t.Fatalf("kit capabilities git wt home --json error = %v", err)
	}
	var payload capabilityDetailPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal(git wt home) error = %v", err)
	}
	command := payload.Command
	if command.Command != "git wt home" {
		t.Fatalf("command = %q, want git wt home", command.Command)
	}
	if command.MutationLevel != mutationNone {
		t.Fatalf("mutation level = %q, want %q", command.MutationLevel, mutationNone)
	}
	combinedParts := append([]string{}, command.WhenToUse...)
	combinedParts = append(combinedParts, command.WhenNotToUse...)
	combinedParts = append(combinedParts, command.Caveats...)
	combined := strings.Join(combinedParts, " ")
	for _, want := range []string{"primary checkout", "child shell", "configured `$SHELL`"} {
		if !strings.Contains(combined, want) {
			t.Fatalf("expected home capability to mention %q, got %#v", want, command)
		}
	}
}

func TestCapabilitiesDescribeGitWTSync(t *testing.T) {
	output, err := executeCapabilitiesCommand("--json", "git", "wt", "sync")
	if err != nil {
		t.Fatalf("kit capabilities git wt sync --json error = %v", err)
	}
	var payload capabilityDetailPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal(git wt sync) error = %v", err)
	}
	command := payload.Command
	if command.Command != "git wt sync" {
		t.Fatalf("command = %q, want git wt sync", command.Command)
	}
	if command.MutationLevel != mutationGit {
		t.Fatalf("mutation level = %q, want %q", command.MutationLevel, mutationGit)
	}
	combined := strings.Join(append(
		append(
			append([]string{}, command.Examples...),
			command.Caveats...,
		),
		command.WhenToUse...,
	), " ")
	combined += " " + command.NetworkUse.Summary +
		" " + command.NetworkUse.FlagDependent +
		" " + command.FileWrites.Summary +
		" " + command.FileWrites.FlagDependent +
		" " + command.GitMutation.Summary +
		" " + command.GitMutation.FlagDependent
	for _, want := range []string{
		"--dry-run",
		"--json",
		"GitHub",
		"origin only",
		"exact PR-head OID",
		"compare-and-swap",
		"ignored root bin/",
		"squash merges",
		"nonzero",
		"never stashes",
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("expected sync capability to mention %q, got %#v", want, command)
		}
	}
}

func TestCapabilitiesDescribeGitWTPath(t *testing.T) {
	output, err := executeCapabilitiesCommand("--json", "git", "wt", "path")
	if err != nil {
		t.Fatalf("kit capabilities git wt path --json error = %v", err)
	}
	var payload capabilityDetailPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal(git wt path) error = %v", err)
	}
	if payload.Command.Command != "git wt path" {
		t.Fatalf("command = %q, want git wt path", payload.Command.Command)
	}
	if payload.Command.MutationLevel != mutationNone {
		t.Fatalf("mutation level = %q, want %q", payload.Command.MutationLevel, mutationNone)
	}
	if !strings.Contains(strings.Join(payload.Command.Examples, " "), `cd "$(git wt path GH-101)"`) {
		t.Fatalf("expected navigation example, got %#v", payload.Command.Examples)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "optional manual convenience") {
		t.Fatalf("expected optional-wrapper caveat, got %#v", payload.Command.Caveats)
	}
}

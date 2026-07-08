package cli

import (
	"strings"
	"testing"
)

func assertCapabilitiesDispatchTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "dispatch", "dispatch")
	if payload.Command.Command != "dispatch" {
		t.Fatalf("command = %q, want dispatch", payload.Command.Command)
	}
	if payload.Command.MutationLevel != mutationNetwork {
		t.Fatalf("expected dispatch mutation level to reflect optional network mutation, got %q", payload.Command.MutationLevel)
	}
	if !strings.Contains(payload.Command.Summary, "CodeRabbit prompt-prep intake") {
		t.Fatalf("expected dispatch summary to describe CodeRabbit prompt-prep intake, got %q", payload.Command.Summary)
	}
	if !strings.Contains(payload.Command.NetworkUse.FlagDependent, "unresolved, non-outdated") {
		t.Fatalf("expected dispatch network notes to describe review-thread filtering, got %#v", payload.Command.NetworkUse)
	}
	prFlag := findDetailedFlag(payload.Command.DetailedFlagBehavior, "--pr")
	if prFlag == nil || !strings.Contains(prFlag.Summary, "unresolved, non-outdated PR review threads") {
		t.Fatalf("expected --pr flag to describe filtered review-thread intake, got %#v", prFlag)
	}
	coderabbitFlag := findDetailedFlag(payload.Command.DetailedFlagBehavior, "--coderabbit")
	if coderabbitFlag == nil || !strings.Contains(coderabbitFlag.Summary, "Prompt for AI Agents") {
		t.Fatalf("expected --coderabbit flag to describe CodeRabbit prompt extraction, got %#v", coderabbitFlag)
	}
	resolveFlag := findDetailedFlag(payload.Command.DetailedFlagBehavior, "--resolve")
	if resolveFlag == nil || !strings.Contains(resolveFlag.Safety, "requires --yes") {
		t.Fatalf("expected --resolve flag to describe explicit mutation boundary, got %#v", resolveFlag)
	}
	yesFlag := findDetailedFlag(payload.Command.DetailedFlagBehavior, "--yes")
	if yesFlag == nil || !strings.Contains(yesFlag.Summary, "confirm --resolve") {
		t.Fatalf("expected --yes flag to document resolve confirmation, got %#v", yesFlag)
	}
	dispatchMaxFlag := findDetailedFlag(payload.Command.DetailedFlagBehavior, "--max-subagents")
	if dispatchMaxFlag == nil || !strings.Contains(dispatchMaxFlag.Summary, "default 3") || !strings.Contains(dispatchMaxFlag.Summary, "hard ceiling 4") {
		t.Fatalf("expected dispatch --max-subagents to document default and ceiling, got %#v", dispatchMaxFlag)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "Agent Team Plan") {
		t.Fatalf("expected dispatch caveats to document Agent Team Plan, got %#v", payload.Command.Caveats)
	}
}

func assertCapabilitiesImproveTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "improve", "improve")
	if payload.Command.Command != "improve" {
		t.Fatalf("command = %q, want improve", payload.Command.Command)
	}
	if payload.Command.MutationLevel != mutationExecutesCommands {
		t.Fatalf("improve mutation_level = %q, want %q", payload.Command.MutationLevel, mutationExecutesCommands)
	}
	if !strings.Contains(payload.Command.FileWrites.Summary, ".kit/improve/runs") {
		t.Fatalf("expected improve file writes to document artifacts, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "does not embed a model runtime") {
		t.Fatalf("expected improve caveats to document deterministic V1 boundary, got %#v", payload.Command.Caveats)
	}
}

func assertCapabilitiesLoopPromptTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "loop prompt", "loop", "prompt")
	if payload.Command.Command != "loop prompt" {
		t.Fatalf("command = %q, want loop prompt", payload.Command.Command)
	}
	if payload.Command.MutationLevel != mutationNone {
		t.Fatalf("expected loop prompt to be prompt-only, got %#v", payload.Command)
	}
	if !strings.Contains(payload.Command.FileWrites.Summary, "none") {
		t.Fatalf("expected loop prompt to document no file writes, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(payload.Command.GitMutation.Summary, "none") {
		t.Fatalf("expected loop prompt to document no git mutation, got %#v", payload.Command.GitMutation)
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--output-only") == nil {
		t.Fatalf("expected loop prompt to document --output-only")
	}
	if !strings.Contains(strings.Join(payload.Command.WhenToUse, " "), "ad hoc") {
		t.Fatalf("expected loop prompt guidance to document ad hoc usage, got %#v", payload.Command.WhenToUse)
	}
}

func assertCapabilitiesLoopWorkflowTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "loop workflow", "loop", "workflow")
	if payload.Command.Command != "loop workflow" {
		t.Fatalf("command = %q, want loop workflow", payload.Command.Command)
	}
	if !strings.Contains(payload.Command.FileWrites.Summary, "REFLECT.json") {
		t.Fatalf("expected loop workflow file writes to mention REFLECT.json, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "raw command, git, and diff evidence") {
		t.Fatalf("expected loop workflow caveats to document raw reflect evidence, got %#v", payload.Command.Caveats)
	}
}

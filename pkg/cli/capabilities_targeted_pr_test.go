package cli

import (
	"strings"
	"testing"
)

func assertCapabilitiesLoopReviewTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "loop review", "loop", "review")
	if payload.Command.Command != "loop review" {
		t.Fatalf("command = %q, want loop review", payload.Command.Command)
	}
	if payload.Command.MutationLevel != mutationExecutesCommands {
		t.Fatalf("expected loop review to execute configured agent, got %#v", payload.Command)
	}
	if !strings.Contains(payload.Command.NetworkUse.FlagDependent, "--pr") {
		t.Fatalf("expected loop review network use to document --pr, got %#v", payload.Command.NetworkUse)
	}
	if !strings.Contains(payload.Command.GitMutation.Summary, "none") {
		t.Fatalf("expected loop review to forbid git mutation, got %#v", payload.Command.GitMutation)
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--wait-for-coderabbit") == nil {
		t.Fatalf("expected loop review to document --wait-for-coderabbit")
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--subagents") == nil {
		t.Fatalf("expected loop review to document --subagents")
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "one agent by default") || !strings.Contains(strings.Join(payload.Command.Caveats, " "), "hard ceiling 4") {
		t.Fatalf("expected loop review caveats to document subagent orchestration, got %#v", payload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "--ignore-user-config") {
		t.Fatalf("expected loop review caveats to document generated Codex config isolation, got %#v", payload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "gpt-5.5") {
		t.Fatalf("expected loop review caveats to document generated Codex model pinning, got %#v", payload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "stderr") {
		t.Fatalf("expected loop review caveats to document progress streaming, got %#v", payload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "stop immediately") {
		t.Fatalf("expected loop review caveats to document agent setup failures, got %#v", payload.Command.Caveats)
	}
}

func assertCapabilitiesPRFixTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "pr fix", "pr", "fix")
	if payload.Command.Command != "pr fix" {
		t.Fatalf("command = %q, want pr fix", payload.Command.Command)
	}
	if payload.Command.MutationLevel != mutationNetwork {
		t.Fatalf("expected pr fix to fetch PR feedback for prompt generation, got %#v", payload.Command)
	}
	if !strings.Contains(payload.Command.NetworkUse.Summary, "gh pr list") {
		t.Fatalf("expected pr fix to document open-PR selector network use, got %#v", payload.Command.NetworkUse)
	}
	if !strings.Contains(payload.Command.GitMutation.Summary, "none") {
		t.Fatalf("expected pr fix to document no git mutation, got %#v", payload.Command.GitMutation)
	}
	if !strings.Contains(payload.Command.NetworkUse.FlagDependent, "human and CodeRabbit review threads") {
		t.Fatalf("expected pr fix to document human and CodeRabbit review-thread intake, got %#v", payload.Command.NetworkUse)
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--pr") == nil {
		t.Fatalf("expected pr fix to document --pr")
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--output-only") == nil {
		t.Fatalf("expected pr fix to document --output-only")
	}
	prFixMaxFlag := findDetailedFlag(payload.Command.DetailedFlagBehavior, "--max-subagents")
	if prFixMaxFlag == nil || !strings.Contains(prFixMaxFlag.Summary, "default 3") || !strings.Contains(prFixMaxFlag.Summary, "hard ceiling 4") {
		t.Fatalf("expected pr fix --max-subagents to document default and ceiling, got %#v", prFixMaxFlag)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "does not run the loop agent") {
		t.Fatalf("expected pr fix caveats to document prompt-only dispatch behavior, got %#v", payload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "Agent Team Plan") {
		t.Fatalf("expected pr fix caveats to document Agent Team Plan, got %#v", payload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "post-push reflection") {
		t.Fatalf("expected pr fix caveats to document post-push reflection, got %#v", payload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(payload.Command.Caveats, " "), "kit dispatch --pr <target> --resolve --yes") {
		t.Fatalf("expected pr fix caveats to document explicit resolution path, got %#v", payload.Command.Caveats)
	}
}

func assertCapabilitiesProjectRefreshTarget(t *testing.T) {
	t.Helper()

	payload := mustCapabilityDetail(t, "project refresh", "project", "refresh")
	if payload.Command.Command != "project refresh" {
		t.Fatalf("command = %q, want project refresh", payload.Command.Command)
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--now") == nil {
		t.Fatalf("expected project refresh to document --now")
	}
	if !strings.Contains(payload.Command.FileWrites.FlagDependent, ".kit.yaml") {
		t.Fatalf("expected project refresh file writes to document .kit.yaml cadence state, got %#v", payload.Command.FileWrites)
	}
	if !strings.Contains(strings.Join(payload.Command.WhenNotToUse, " "), "automatic changelog") {
		t.Fatalf("expected project refresh guidance to reject automatic changelog usage, got %#v", payload.Command.WhenNotToUse)
	}
}

func assertCapabilitiesRemovedReviewLoopTarget(t *testing.T) {
	t.Helper()

	if _, err := executeCapabilitiesCommand("--json", "review-loop"); err == nil || !strings.Contains(err.Error(), "unknown Kit command path") {
		t.Fatalf("expected review-loop lookup to fail as removed, got %v", err)
	}
}

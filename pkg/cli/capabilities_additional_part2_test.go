package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCapabilitiesTargetedJSONLoopAndPR(t *testing.T) {
	loopPromptOutput, err := executeCapabilitiesCommand("--json", "loop", "prompt")
	if err != nil {
		t.Fatalf("kit capabilities loop prompt --json error = %v", err)
	}
	var loopPromptPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(loopPromptOutput), &loopPromptPayload); err != nil {
		t.Fatalf("json.Unmarshal(loop prompt) error = %v", err)
	}
	if loopPromptPayload.Command.Command != "loop prompt" {
		t.Fatalf("command = %q, want loop prompt", loopPromptPayload.Command.Command)
	}
	if loopPromptPayload.Command.MutationLevel != mutationNone {
		t.Fatalf("expected loop prompt to be prompt-only, got %#v", loopPromptPayload.Command)
	}
	if !strings.Contains(loopPromptPayload.Command.FileWrites.Summary, "none") {
		t.Fatalf("expected loop prompt to document no file writes, got %#v", loopPromptPayload.Command.FileWrites)
	}
	if !strings.Contains(loopPromptPayload.Command.GitMutation.Summary, "none") {
		t.Fatalf("expected loop prompt to document no git mutation, got %#v", loopPromptPayload.Command.GitMutation)
	}
	if findDetailedFlag(loopPromptPayload.Command.DetailedFlagBehavior, "--output-only") == nil {
		t.Fatalf("expected loop prompt to document --output-only")
	}
	if !strings.Contains(strings.Join(loopPromptPayload.Command.WhenToUse, " "), "ad hoc") {
		t.Fatalf("expected loop prompt guidance to document ad hoc usage, got %#v", loopPromptPayload.Command.WhenToUse)
	}

	loopWorkflowOutput, err := executeCapabilitiesCommand("--json", "loop", "workflow")
	if err != nil {
		t.Fatalf("kit capabilities loop workflow --json error = %v", err)
	}
	var loopWorkflowPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(loopWorkflowOutput), &loopWorkflowPayload); err != nil {
		t.Fatalf("json.Unmarshal(loop workflow) error = %v", err)
	}
	if !loopWorkflowPayload.Command.Deprecated || !strings.Contains(loopWorkflowPayload.Command.DeprecationNote, "workflow_version 2") {
		t.Fatalf("expected loop workflow compatibility deprecation, got %#v", loopWorkflowPayload.Command)
	}
	if !strings.Contains(strings.Join(loopWorkflowPayload.Command.Caveats, " "), "V3 specs are rejected") {
		t.Fatalf("expected V3 rejection guidance, got %#v", loopWorkflowPayload.Command.Caveats)
	}

	loopReviewOutput, err := executeCapabilitiesCommand("--json", "loop", "review")
	if err != nil {
		t.Fatalf("kit capabilities loop review --json error = %v", err)
	}
	var loopReviewPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(loopReviewOutput), &loopReviewPayload); err != nil {
		t.Fatalf("json.Unmarshal(loop review) error = %v", err)
	}
	if loopReviewPayload.Command.Command != "loop review" {
		t.Fatalf("command = %q, want loop review", loopReviewPayload.Command.Command)
	}
	if loopReviewPayload.Command.MutationLevel != mutationExecutesCommands {
		t.Fatalf("expected loop review to execute configured agent, got %#v", loopReviewPayload.Command)
	}
	if !strings.Contains(loopReviewPayload.Command.NetworkUse.FlagDependent, "--pr") {
		t.Fatalf("expected loop review network use to document --pr, got %#v", loopReviewPayload.Command.NetworkUse)
	}
	if !strings.Contains(loopReviewPayload.Command.GitMutation.Summary, "none") {
		t.Fatalf("expected loop review to forbid git mutation, got %#v", loopReviewPayload.Command.GitMutation)
	}
	if findDetailedFlag(loopReviewPayload.Command.DetailedFlagBehavior, "--wait-for-coderabbit") == nil {
		t.Fatalf("expected loop review to document --wait-for-coderabbit")
	}
	if findDetailedFlag(loopReviewPayload.Command.DetailedFlagBehavior, "--subagents") == nil {
		t.Fatalf("expected loop review to document --subagents")
	}
	if !strings.Contains(strings.Join(loopReviewPayload.Command.Caveats, " "), "one agent by default") || !strings.Contains(strings.Join(loopReviewPayload.Command.Caveats, " "), "hard ceiling 4") {
		t.Fatalf("expected loop review caveats to document subagent orchestration, got %#v", loopReviewPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(loopReviewPayload.Command.Caveats, " "), "--ignore-user-config") {
		t.Fatalf("expected loop review caveats to document generated Codex config isolation, got %#v", loopReviewPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(loopReviewPayload.Command.Caveats, " "), "gpt-5.6") {
		t.Fatalf("expected loop review caveats to document generated Codex model pinning, got %#v", loopReviewPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(loopReviewPayload.Command.Caveats, " "), "stderr") {
		t.Fatalf("expected loop review caveats to document progress streaming, got %#v", loopReviewPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(loopReviewPayload.Command.Caveats, " "), "stop immediately") {
		t.Fatalf("expected loop review caveats to document agent setup failures, got %#v", loopReviewPayload.Command.Caveats)
	}

	prFixOutput, err := executeCapabilitiesCommand("--json", "pr", "fix")
	if err != nil {
		t.Fatalf("kit capabilities pr fix --json error = %v", err)
	}
	var prFixPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(prFixOutput), &prFixPayload); err != nil {
		t.Fatalf("json.Unmarshal(pr fix) error = %v", err)
	}
	if prFixPayload.Command.Command != "pr fix" {
		t.Fatalf("command = %q, want pr fix", prFixPayload.Command.Command)
	}
	if prFixPayload.Command.MutationLevel != mutationGit {
		t.Fatalf("expected pr fix to disclose conditional worktree mutation, got %#v", prFixPayload.Command)
	}
	if !strings.Contains(prFixPayload.Command.NetworkUse.Summary, "gh pr list") {
		t.Fatalf("expected pr fix to document open-PR selector network use, got %#v", prFixPayload.Command.NetworkUse)
	}
	if !strings.Contains(prFixPayload.Command.GitMutation.Summary, "add or attach") {
		t.Fatalf("expected pr fix to document exact worktree preparation, got %#v", prFixPayload.Command.GitMutation)
	}
	if !strings.Contains(prFixPayload.Command.NetworkUse.FlagDependent, "human and CodeRabbit review threads") {
		t.Fatalf("expected pr fix to document human and CodeRabbit review-thread intake, got %#v", prFixPayload.Command.NetworkUse)
	}
	if findDetailedFlag(prFixPayload.Command.DetailedFlagBehavior, "--pr") == nil {
		t.Fatalf("expected pr fix to document --pr")
	}
	if findDetailedFlag(prFixPayload.Command.DetailedFlagBehavior, "--output-only") == nil {
		t.Fatalf("expected pr fix to document --output-only")
	}
	if findDetailedFlag(prFixPayload.Command.DetailedFlagBehavior, "--edit") == nil {
		t.Fatalf("expected pr fix to document opt-in --edit")
	}
	if !strings.Contains(prFixPayload.Command.FileWrites.FlagDependent, "only with --edit") {
		t.Fatalf("expected pr fix to document opt-in editor writes, got %#v", prFixPayload.Command.FileWrites)
	}
	prFixMaxFlag := findDetailedFlag(prFixPayload.Command.DetailedFlagBehavior, "--max-subagents")
	if prFixMaxFlag == nil || !strings.Contains(prFixMaxFlag.Summary, "default 3") || !strings.Contains(prFixMaxFlag.Summary, "hard ceiling 4") {
		t.Fatalf("expected pr fix --max-subagents to document default and ceiling, got %#v", prFixMaxFlag)
	}
	if !strings.Contains(strings.Join(prFixPayload.Command.Caveats, " "), "does not run the loop agent") {
		t.Fatalf("expected pr fix caveats to document prompt-only dispatch behavior, got %#v", prFixPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(prFixPayload.Command.Caveats, " "), "dirty") {
		t.Fatalf("expected pr fix caveats to document dirty-lane confirmation, got %#v", prFixPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(prFixPayload.Command.Caveats, " "), "Agent Team Plan") {
		t.Fatalf("expected pr fix caveats to document Agent Team Plan, got %#v", prFixPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(prFixPayload.Command.Caveats, " "), "post-push reflection") {
		t.Fatalf("expected pr fix caveats to document post-push reflection, got %#v", prFixPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(prFixPayload.Command.Caveats, " "), "kit dispatch --pr <target> --resolve --yes") {
		t.Fatalf("expected pr fix caveats to document explicit resolution path, got %#v", prFixPayload.Command.Caveats)
	}

	projectRefreshOutput, err := executeCapabilitiesCommand("--json", "project", "refresh")
	if err != nil {
		t.Fatalf("kit capabilities project refresh --json error = %v", err)
	}
	var projectRefreshPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(projectRefreshOutput), &projectRefreshPayload); err != nil {
		t.Fatalf("json.Unmarshal(project refresh) error = %v", err)
	}
	if projectRefreshPayload.Command.Command != "project refresh" {
		t.Fatalf("command = %q, want project refresh", projectRefreshPayload.Command.Command)
	}
	if findDetailedFlag(projectRefreshPayload.Command.DetailedFlagBehavior, "--now") == nil {
		t.Fatalf("expected project refresh to document --now")
	}
	if !strings.Contains(projectRefreshPayload.Command.FileWrites.FlagDependent, ".kit.yaml") {
		t.Fatalf("expected project refresh file writes to document .kit.yaml cadence state, got %#v", projectRefreshPayload.Command.FileWrites)
	}
	if !strings.Contains(strings.Join(projectRefreshPayload.Command.WhenNotToUse, " "), "automatic changelog") {
		t.Fatalf("expected project refresh guidance to reject automatic changelog usage, got %#v", projectRefreshPayload.Command.WhenNotToUse)
	}
	if !strings.Contains(strings.Join(projectRefreshPayload.Command.WhenNotToUse, " "), "kit reconcile --all --include-files") {
		t.Fatalf("expected project refresh guidance to point structural refreshes at reconcile include-files, got %#v", projectRefreshPayload.Command.WhenNotToUse)
	}

	if _, err := executeCapabilitiesCommand("--json", "review-loop"); err == nil || !strings.Contains(err.Error(), "unknown Kit command path") {
		t.Fatalf("expected review-loop lookup to fail as removed, got %v", err)
	}
}

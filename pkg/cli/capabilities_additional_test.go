package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCapabilitiesTargetedJSONAdditional(t *testing.T) {
	output, err := executeCapabilitiesCommand("--json", "legacy", "verify")
	if err != nil {
		t.Fatalf("kit capabilities legacy verify --json error = %v", err)
	}

	var payload capabilityDetailPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, output)
	}
	if payload.Kind != "capability_detail" {
		t.Fatalf("kind = %q, want capability_detail", payload.Kind)
	}
	if payload.Command.Command != "legacy verify" {
		t.Fatalf("command = %q, want legacy verify", payload.Command.Command)
	}
	if len(payload.Command.WhenToUse) == 0 || len(payload.Command.WhenNotToUse) == 0 || len(payload.Command.Examples) == 0 {
		t.Fatalf("expected detailed guidance fields to be populated: %#v", payload.Command)
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--dry-run") == nil {
		t.Fatalf("expected verify detailed flags to include --dry-run")
	}
	if findDetailedFlag(payload.Command.DetailedFlagBehavior, "--allow-shell") == nil {
		t.Fatalf("expected verify detailed flags to include --allow-shell")
	}

	ciOutput, err := executeCapabilitiesCommand("--json", "ci")
	if err != nil {
		t.Fatalf("kit capabilities ci --json error = %v", err)
	}
	var ciPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(ciOutput), &ciPayload); err != nil {
		t.Fatalf("json.Unmarshal(ci) error = %v", err)
	}
	if ciPayload.Command.Command != "ci" {
		t.Fatalf("command = %q, want ci", ciPayload.Command.Command)
	}
	if ciPayload.Command.NetworkUse.Summary == "none" {
		t.Fatalf("expected ci targeted detail to describe network use")
	}
	if !strings.Contains(ciPayload.Command.NetworkUse.Summary, "git/gh") {
		t.Fatalf("expected ci targeted detail to describe git/gh subprocess use, got %#v", ciPayload.Command.NetworkUse)
	}
	if !strings.Contains(ciPayload.Command.NetworkUse.FlagDependent, "--copilot") {
		t.Fatalf("expected ci targeted detail to describe optional copilot behavior, got %#v", ciPayload.Command.NetworkUse)
	}
	if !strings.Contains(ciPayload.Command.FileWrites.Summary, ".kit.yaml") {
		t.Fatalf("expected ci targeted detail to describe .kit.yaml cache writes, got %#v", ciPayload.Command.FileWrites)
	}

	dispatchOutput, err := executeCapabilitiesCommand("--json", "dispatch")
	if err != nil {
		t.Fatalf("kit capabilities dispatch --json error = %v", err)
	}
	var dispatchPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(dispatchOutput), &dispatchPayload); err != nil {
		t.Fatalf("json.Unmarshal(dispatch) error = %v", err)
	}
	if dispatchPayload.Command.Command != "dispatch" {
		t.Fatalf("command = %q, want dispatch", dispatchPayload.Command.Command)
	}
	if dispatchPayload.Command.MutationLevel != mutationGit {
		t.Fatalf("expected dispatch mutation level to reflect optional worktree mutation, got %q", dispatchPayload.Command.MutationLevel)
	}
	if !strings.Contains(dispatchPayload.Command.Summary, "CodeRabbit prompt-prep intake") {
		t.Fatalf("expected dispatch summary to describe CodeRabbit prompt-prep intake, got %q", dispatchPayload.Command.Summary)
	}
	if !strings.Contains(dispatchPayload.Command.NetworkUse.FlagDependent, "unresolved, non-outdated") {
		t.Fatalf("expected dispatch network notes to describe review-thread filtering, got %#v", dispatchPayload.Command.NetworkUse)
	}
	prFlag := findDetailedFlag(dispatchPayload.Command.DetailedFlagBehavior, "--pr")
	if prFlag == nil || !strings.Contains(prFlag.Summary, "unresolved, non-outdated PR review threads") {
		t.Fatalf("expected --pr flag to describe filtered review-thread intake, got %#v", prFlag)
	}
	coderabbitFlag := findDetailedFlag(dispatchPayload.Command.DetailedFlagBehavior, "--coderabbit")
	if coderabbitFlag == nil || !strings.Contains(coderabbitFlag.Summary, "Prompt for AI Agents") {
		t.Fatalf("expected --coderabbit flag to describe CodeRabbit prompt extraction, got %#v", coderabbitFlag)
	}
	resolveFlag := findDetailedFlag(dispatchPayload.Command.DetailedFlagBehavior, "--resolve")
	if resolveFlag == nil || !strings.Contains(resolveFlag.Safety, "requires --yes") {
		t.Fatalf("expected --resolve flag to describe explicit mutation boundary, got %#v", resolveFlag)
	}
	yesFlag := findDetailedFlag(dispatchPayload.Command.DetailedFlagBehavior, "--yes")
	if yesFlag == nil || !strings.Contains(yesFlag.Summary, "confirm --resolve") {
		t.Fatalf("expected --yes flag to document resolve confirmation, got %#v", yesFlag)
	}
	dispatchMaxFlag := findDetailedFlag(dispatchPayload.Command.DetailedFlagBehavior, "--max-subagents")
	if dispatchMaxFlag == nil || !strings.Contains(dispatchMaxFlag.Summary, "default 3") || !strings.Contains(dispatchMaxFlag.Summary, "hard ceiling 4") {
		t.Fatalf("expected dispatch --max-subagents to document default and ceiling, got %#v", dispatchMaxFlag)
	}
	if !strings.Contains(strings.Join(dispatchPayload.Command.Caveats, " "), "Agent Team Plan") {
		t.Fatalf("expected dispatch caveats to document Agent Team Plan, got %#v", dispatchPayload.Command.Caveats)
	}

	improveOutput, err := executeCapabilitiesCommand("--json", "improve")
	if err != nil {
		t.Fatalf("kit capabilities improve --json error = %v", err)
	}
	var improvePayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(improveOutput), &improvePayload); err != nil {
		t.Fatalf("json.Unmarshal(improve) error = %v", err)
	}
	if improvePayload.Command.Command != "improve" {
		t.Fatalf("command = %q, want improve", improvePayload.Command.Command)
	}
	if improvePayload.Command.MutationLevel != mutationExecutesCommands {
		t.Fatalf("improve mutation_level = %q, want %q", improvePayload.Command.MutationLevel, mutationExecutesCommands)
	}
	if !strings.Contains(improvePayload.Command.FileWrites.Summary, ".kit/improve/runs") {
		t.Fatalf("expected improve file writes to document artifacts, got %#v", improvePayload.Command.FileWrites)
	}
	if !strings.Contains(strings.Join(improvePayload.Command.Caveats, " "), "does not embed a model runtime") {
		t.Fatalf("expected improve caveats to document deterministic V1 boundary, got %#v", improvePayload.Command.Caveats)
	}
	improveRunOutput, err := executeCapabilitiesCommand("--json", "improve", "run")
	if err != nil {
		t.Fatalf("kit capabilities improve run --json error = %v", err)
	}
	var improveRunPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(improveRunOutput), &improveRunPayload); err != nil {
		t.Fatalf("json.Unmarshal(improve run) error = %v", err)
	}
	improveRunCaveats := strings.Join(improveRunPayload.Command.Caveats, " ")
	for _, want := range []string{"redacted", "workspace-normalized", "200-line", "rather than raw command output"} {
		if !strings.Contains(improveRunCaveats, want) {
			t.Fatalf("expected improve run caveats to contain %q, got %#v", want, improveRunPayload.Command.Caveats)
		}
	}
}

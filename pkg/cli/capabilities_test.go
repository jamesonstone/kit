package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCapabilitiesIndexJSON(t *testing.T) {
	output, err := executeCapabilitiesCommand("--json")
	if err != nil {
		t.Fatalf("kit capabilities --json error = %v", err)
	}

	var payload capabilitiesIndexPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, output)
	}

	if payload.SchemaVersion != capabilitiesSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", payload.SchemaVersion, capabilitiesSchemaVersion)
	}
	if payload.Kind != "capabilities_index" {
		t.Fatalf("kind = %q, want capabilities_index", payload.Kind)
	}
	if payload.GeneratedBy != "kit capabilities" {
		t.Fatalf("generated_by = %q, want kit capabilities", payload.GeneratedBy)
	}

	for _, command := range []string{"capabilities", "ci", "pr fix", "legacy verify", "loop review", "dispatch", "rules add", "skill mine"} {
		if findCompactCapability(payload.Commands, command) == nil {
			t.Fatalf("expected compact capabilities to include %q", command)
		}
	}
	for _, command := range []string{"update", "skills", "catchup", "rollup", "review-loop"} {
		if findCompactCapability(payload.Commands, command) != nil {
			t.Fatalf("expected compact capabilities to omit removed command %q", command)
		}
	}

	verify := findCompactCapability(payload.Commands, "legacy verify")
	if verify == nil {
		t.Fatal("expected legacy verify capability")
	}
	if verify.MutationLevel != mutationExecutesCommands {
		t.Fatalf("legacy verify mutation_level = %q, want %q", verify.MutationLevel, mutationExecutesCommands)
	}
	if !strings.Contains(verify.FileWrites.FlagDependent, "--no-write") {
		t.Fatalf("expected verify file write behavior to mention --no-write, got %#v", verify.FileWrites)
	}

	ci := findCompactCapability(payload.Commands, "ci")
	if ci == nil {
		t.Fatal("expected ci capability")
	}
	if ci.NetworkUse.Summary == "none" {
		t.Fatalf("expected ci network behavior to be documented, got %#v", ci.NetworkUse)
	}
}

func TestCapabilitiesTargetedJSON(t *testing.T) {
	initOutput, err := executeCapabilitiesCommand("--json", "init")
	if err != nil {
		t.Fatalf("kit capabilities init --json error = %v", err)
	}
	var initPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(initOutput), &initPayload); err != nil {
		t.Fatalf("json.Unmarshal(init) error = %v", err)
	}
	if initPayload.Command.Command != "init" {
		t.Fatalf("command = %q, want init", initPayload.Command.Command)
	}
	if !strings.Contains(initPayload.Command.NetworkUse.FlagDependent, "--refresh") {
		t.Fatalf("expected init network use to document refresh registry fetch, got %#v", initPayload.Command.NetworkUse)
	}
	if !strings.Contains(initPayload.Command.FileWrites.FlagDependent, "--dry-run") {
		t.Fatalf("expected init file writes to document dry-run, got %#v", initPayload.Command.FileWrites)
	}
	refreshFlag := findDetailedFlag(initPayload.Command.DetailedFlagBehavior, "--refresh")
	if refreshFlag == nil || !strings.Contains(refreshFlag.Summary, "loop.agent.command") {
		t.Fatalf("expected --refresh flag to document loop agent backfill, got %#v", refreshFlag)
	}
	if findDetailedFlag(initPayload.Command.DetailedFlagBehavior, "--diff") == nil {
		t.Fatalf("expected init detailed flags to include --diff")
	}

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
	if dispatchPayload.Command.MutationLevel != mutationNetwork {
		t.Fatalf("expected dispatch mutation level to reflect optional network mutation, got %q", dispatchPayload.Command.MutationLevel)
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
	if !strings.Contains(strings.Join(loopReviewPayload.Command.Caveats, " "), "one agent by default") {
		t.Fatalf("expected loop review caveats to document subagent orchestration, got %#v", loopReviewPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(loopReviewPayload.Command.Caveats, " "), "--ignore-user-config") {
		t.Fatalf("expected loop review caveats to document generated Codex config isolation, got %#v", loopReviewPayload.Command.Caveats)
	}
	if !strings.Contains(strings.Join(loopReviewPayload.Command.Caveats, " "), "gpt-5.5") {
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
	if prFixPayload.Command.MutationLevel != mutationExecutesCommands {
		t.Fatalf("expected pr fix to execute the configured repair agent, got %#v", prFixPayload.Command)
	}
	if !strings.Contains(prFixPayload.Command.NetworkUse.Summary, "gh pr list") {
		t.Fatalf("expected pr fix to document open-PR selector network use, got %#v", prFixPayload.Command.NetworkUse)
	}
	if !strings.Contains(prFixPayload.Command.GitMutation.Summary, "forbid staging") {
		t.Fatalf("expected pr fix to forbid git mutation, got %#v", prFixPayload.Command.GitMutation)
	}
	if findDetailedFlag(prFixPayload.Command.DetailedFlagBehavior, "--pr") == nil {
		t.Fatalf("expected pr fix to document --pr")
	}
	if findDetailedFlag(prFixPayload.Command.DetailedFlagBehavior, "--json") == nil {
		t.Fatalf("expected pr fix to document --json")
	}
	if !strings.Contains(strings.Join(prFixPayload.Command.Caveats, " "), "does not push") {
		t.Fatalf("expected pr fix caveats to document push boundary, got %#v", prFixPayload.Command.Caveats)
	}

	if _, err := executeCapabilitiesCommand("--json", "review-loop"); err == nil || !strings.Contains(err.Error(), "unknown Kit command path") {
		t.Fatalf("expected review-loop lookup to fail as removed, got %v", err)
	}
}

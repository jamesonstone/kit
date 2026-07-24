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

	for _, command := range []string{"capabilities", "config", "config check", "aws", "aws verify", "ci", "pr fix", "legacy verify", "loop prompt", "loop review", "project refresh", "improve", "improve run", "dispatch", "rules add", "skill mine", "git wt path"} {
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
	if !strings.Contains(initPayload.Command.NetworkUse.FlagDependent, "gh repo visibility") {
		t.Fatalf("expected init network use to document README badge visibility lookup, got %#v", initPayload.Command.NetworkUse)
	}
	if !strings.Contains(initPayload.Command.FileWrites.FlagDependent, "--dry-run") {
		t.Fatalf("expected init file writes to document dry-run, got %#v", initPayload.Command.FileWrites)
	}
	if !strings.Contains(initPayload.Command.FileWrites.Summary, "auto-assign.yml") {
		t.Fatalf("expected init file writes to document auto-assign workflow, got %#v", initPayload.Command.FileWrites)
	}
	if !strings.Contains(initPayload.Command.FileWrites.Summary, "README.md managed status badges and final Maintainers section") {
		t.Fatalf("expected init file writes to document README badge management, got %#v", initPayload.Command.FileWrites)
	}
	if !strings.Contains(initPayload.Command.FileWrites.FlagDependent, "github.default_assignees") {
		t.Fatalf("expected init file writes to document auto-assignee config fallback, got %#v", initPayload.Command.FileWrites)
	}
	if !strings.Contains(initPayload.Command.FileWrites.FlagDependent, "github.repository") {
		t.Fatalf("expected init file writes to document README badge repository source, got %#v", initPayload.Command.FileWrites)
	}
	if !strings.Contains(initPayload.Command.FileWrites.FlagDependent, "private repositories skip public Shields") {
		t.Fatalf("expected init file writes to document private README badge behavior, got %#v", initPayload.Command.FileWrites)
	}
	if !strings.Contains(initPayload.Command.FileWrites.FlagDependent, "jamesonstone attribution") {
		t.Fatalf("expected init file writes to document README Maintainers attribution, got %#v", initPayload.Command.FileWrites)
	}
	refreshFlag := findDetailedFlag(initPayload.Command.DetailedFlagBehavior, "--refresh")
	if refreshFlag == nil || !strings.Contains(refreshFlag.Summary, "loop.agent.command") || !strings.Contains(refreshFlag.Summary, "auto-assignment workflow") || !strings.Contains(refreshFlag.Summary, "README.md managed badges and Maintainers section") {
		t.Fatalf("expected --refresh flag to document loop agent, README badge, Maintainers, and auto-assignment workflow backfill, got %#v", refreshFlag)
	}
	if findDetailedFlag(initPayload.Command.DetailedFlagBehavior, "--diff") == nil {
		t.Fatalf("expected init detailed flags to include --diff")
	}
	forceFlag := findDetailedFlag(initPayload.Command.DetailedFlagBehavior, "--force")
	if forceFlag == nil || !strings.Contains(forceFlag.Summary, "replace existing generated files") {
		t.Fatalf("expected --force flag to document generated file replacement, got %#v", forceFlag)
	}

	reconcileOutput, err := executeCapabilitiesCommand("--json", "reconcile")
	if err != nil {
		t.Fatalf("kit capabilities reconcile --json error = %v", err)
	}
	var reconcilePayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(reconcileOutput), &reconcilePayload); err != nil {
		t.Fatalf("json.Unmarshal(reconcile) error = %v", err)
	}
	if reconcilePayload.Command.Command != "reconcile" {
		t.Fatalf("command = %q, want reconcile", reconcilePayload.Command.Command)
	}
	if reconcilePayload.Command.MutationLevel != mutationWritesFiles {
		t.Fatalf("expected reconcile mutation level to reflect included managed-file refreshes, got %q", reconcilePayload.Command.MutationLevel)
	}
	if !strings.Contains(reconcilePayload.Command.NetworkUse.Summary, "ruleset registry") {
		t.Fatalf("expected reconcile network use to document registry fetch, got %#v", reconcilePayload.Command.NetworkUse)
	}
	if !strings.Contains(strings.Join(reconcilePayload.Command.WhenToUse, " "), "include files?") {
		t.Fatalf("expected reconcile guidance to document interactive menu, got %#v", reconcilePayload.Command.WhenToUse)
	}
	for _, flagName := range []string{"--include-files", "--all", "--force", "--dry-run", "--diff", "--file"} {
		if findDetailedFlag(reconcilePayload.Command.DetailedFlagBehavior, flagName) == nil {
			t.Fatalf("expected reconcile detailed flags to include %s", flagName)
		}
	}
	if !strings.Contains(strings.Join(reconcilePayload.Command.Examples, " "), "kit reconcile --all --include-files") {
		t.Fatalf("expected reconcile examples to document whole-project file refresh, got %#v", reconcilePayload.Command.Examples)
	}

	specOutput, err := executeCapabilitiesCommand("--json", "spec")
	if err != nil {
		t.Fatalf("kit capabilities spec --json error = %v", err)
	}
	var specPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(specOutput), &specPayload); err != nil {
		t.Fatalf("json.Unmarshal(spec) error = %v", err)
	}
	if !strings.Contains(specPayload.Command.FileWrites.Summary, "workflow_version 3") {
		t.Fatalf("expected spec file writes to document V3 scaffolding, got %#v", specPayload.Command.FileWrites)
	}
	if !strings.Contains(strings.Join(specPayload.Command.Caveats, " "), "does not ingest agent transcripts") {
		t.Fatalf("expected spec caveats to document semantic plan translation, got %#v", specPayload.Command.Caveats)
	}

	statusOutput, err := executeCapabilitiesCommand("--json", "status")
	if err != nil {
		t.Fatalf("kit capabilities status --json error = %v", err)
	}
	var statusPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(statusOutput), &statusPayload); err != nil {
		t.Fatalf("json.Unmarshal(status) error = %v", err)
	}
	if !strings.Contains(statusPayload.Command.NetworkUse.Summary, "30s timeout") {
		t.Fatalf("expected status network use to document registry timeout, got %#v", statusPayload.Command.NetworkUse)
	}
	if !strings.Contains(statusPayload.Command.NetworkUse.FlagDependent, "unchecked/unknown") {
		t.Fatalf("expected status network use to document registry fallback, got %#v", statusPayload.Command.NetworkUse)
	}
	statusCaveats := strings.Join(statusPayload.Command.Caveats, " ")
	if !strings.Contains(statusCaveats, "deadline expiry") || !strings.Contains(statusCaveats, "managed_files.unchecked") {
		t.Fatalf("expected status caveats to document unchecked managed files, got %#v", statusPayload.Command.Caveats)
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
	if prFixPayload.Command.MutationLevel != mutationNetwork {
		t.Fatalf("expected pr fix to fetch PR feedback for prompt generation, got %#v", prFixPayload.Command)
	}
	if !strings.Contains(prFixPayload.Command.NetworkUse.Summary, "gh pr list") {
		t.Fatalf("expected pr fix to document open-PR selector network use, got %#v", prFixPayload.Command.NetworkUse)
	}
	if !strings.Contains(prFixPayload.Command.GitMutation.Summary, "none") {
		t.Fatalf("expected pr fix to document no git mutation, got %#v", prFixPayload.Command.GitMutation)
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

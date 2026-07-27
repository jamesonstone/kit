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

	for _, command := range []string{"capabilities", "config", "config check", "aws", "aws verify", "ci", "pr fix", "legacy verify", "loop prompt", "loop review", "project refresh", "improve", "improve run", "dispatch", "rules add", "skill mine", "git wt list", "git wt sync", "git wt home", "git wt cd", "git wt path"} {
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
	if ci.MutationLevel != mutationGit {
		t.Fatalf("ci mutation_level = %q, want conditional git mutation", ci.MutationLevel)
	}
}

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
	if payload.Command.MutationLevel != mutationNone {
		t.Fatalf("mutation level = %q, want %q", payload.Command.MutationLevel, mutationNone)
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
	combined := strings.Join(combinedParts, " ")
	for _, want := range []string{"--plain", "--root-position bottom", "arrow keys", "press h", "child shell", "bright green", "--sort state", "--sort head", "--sort path", "display-only", "local timezone", "HH:MM", "without seconds"} {
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
		"git branch -d",
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

}

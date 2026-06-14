package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

	for _, command := range []string{"capabilities", "ci", "verify", "review-loop", "dispatch", "rules add", "skill mine"} {
		if findCompactCapability(payload.Commands, command) == nil {
			t.Fatalf("expected compact capabilities to include %q", command)
		}
	}
	for _, command := range []string{"update", "skills", "catchup", "rollup"} {
		if findCompactCapability(payload.Commands, command) != nil {
			t.Fatalf("expected compact capabilities to omit hidden/deprecated command %q", command)
		}
	}

	verify := findCompactCapability(payload.Commands, "verify")
	if verify == nil {
		t.Fatal("expected verify capability")
	}
	if verify.MutationLevel != mutationExecutesCommands {
		t.Fatalf("verify mutation_level = %q, want %q", verify.MutationLevel, mutationExecutesCommands)
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
	output, err := executeCapabilitiesCommand("--json", "verify")
	if err != nil {
		t.Fatalf("kit capabilities verify --json error = %v", err)
	}

	var payload capabilityDetailPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, output)
	}
	if payload.Kind != "capability_detail" {
		t.Fatalf("kind = %q, want capability_detail", payload.Kind)
	}
	if payload.Command.Command != "verify" {
		t.Fatalf("command = %q, want verify", payload.Command.Command)
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
	if !strings.Contains(dispatchPayload.Command.Summary, "CodeRabbit review-loop intake") {
		t.Fatalf("expected dispatch summary to describe review-loop intake, got %q", dispatchPayload.Command.Summary)
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

	reviewLoopOutput, err := executeCapabilitiesCommand("--json", "review-loop")
	if err != nil {
		t.Fatalf("kit capabilities review-loop --json error = %v", err)
	}
	var reviewLoopPayload capabilityDetailPayload
	if err := json.Unmarshal([]byte(reviewLoopOutput), &reviewLoopPayload); err != nil {
		t.Fatalf("json.Unmarshal(review-loop) error = %v", err)
	}
	if reviewLoopPayload.Command.Command != "review-loop" {
		t.Fatalf("command = %q, want review-loop", reviewLoopPayload.Command.Command)
	}
	if reviewLoopPayload.Command.MutationLevel != mutationNone || reviewLoopPayload.Command.GitMutation.Summary != "none" {
		t.Fatalf("expected review-loop to be read-only, got %#v", reviewLoopPayload.Command)
	}
	if !strings.Contains(reviewLoopPayload.Command.NetworkUse.Summary, "PR metadata") {
		t.Fatalf("expected review-loop network use to mention PR metadata, got %#v", reviewLoopPayload.Command.NetworkUse)
	}
	if findDetailedFlag(reviewLoopPayload.Command.DetailedFlagBehavior, "--watch") == nil {
		t.Fatalf("expected review-loop to document --watch")
	}
}

func TestCapabilitiesHumanDetailIncludesAgentGuidance(t *testing.T) {
	output, err := executeCapabilitiesCommand("dispatch")
	if err != nil {
		t.Fatalf("kit capabilities dispatch error = %v", err)
	}

	for _, want := range []string{
		"When to use:",
		"When not to use:",
		"Examples:",
		"Important flags:",
		"--resolve: with --pr, resolve matching unresolved review threads",
		"[GitHub mutation; requires --yes]",
		"Related commands:",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("human detail output missing %q:\n%s", want, output)
		}
	}
}

func TestCapabilitiesUnknownCommandIsActionable(t *testing.T) {
	_, err := executeCapabilitiesCommand("--json", "does-not-exist")
	if err == nil {
		t.Fatal("expected unknown command to fail")
	}
	if !strings.Contains(err.Error(), "unknown Kit command path") {
		t.Fatalf("expected actionable unknown-command error, got %v", err)
	}
	if !strings.Contains(err.Error(), "kit capabilities --json") {
		t.Fatalf("expected unknown-command error to recommend listing commands, got %v", err)
	}
}

func TestCapabilitiesNestedCommandPaths(t *testing.T) {
	for _, commandPath := range [][]string{
		{"scaffold", "agents"},
		{"prompt", "list"},
		{"set", "prompt"},
		{"skill", "mine"},
		{"rules", "add"},
		{"rules", "list"},
		{"rules", "view"},
		{"rules", "link"},
	} {
		args := append([]string{"--json"}, commandPath...)
		output, err := executeCapabilitiesCommand(args...)
		if err != nil {
			t.Fatalf("kit capabilities %s --json error = %v", strings.Join(commandPath, " "), err)
		}
		var payload capabilityDetailPayload
		if err := json.Unmarshal([]byte(output), &payload); err != nil {
			t.Fatalf("json.Unmarshal(%s) error = %v", strings.Join(commandPath, " "), err)
		}
		if payload.Command.Command != strings.Join(commandPath, " ") {
			t.Fatalf("command = %q, want %q", payload.Command.Command, strings.Join(commandPath, " "))
		}
	}
}

func TestCapabilitiesFullAndHiddenPolicy(t *testing.T) {
	fullOutput, err := executeCapabilitiesCommand("--full", "--json")
	if err != nil {
		t.Fatalf("kit capabilities --full --json error = %v", err)
	}
	var fullPayload capabilitiesFullPayload
	if err := json.Unmarshal([]byte(fullOutput), &fullPayload); err != nil {
		t.Fatalf("json.Unmarshal(full) error = %v", err)
	}
	update := findDetailCapability(fullPayload.Commands, "update")
	if update == nil {
		t.Fatal("expected full capabilities to include hidden update")
	}
	if !update.Hidden || !update.Deprecated {
		t.Fatalf("expected update to be hidden and deprecated, got hidden=%v deprecated=%v", update.Hidden, update.Deprecated)
	}

	targetedOutput, err := executeCapabilitiesCommand("--json", "update")
	if err != nil {
		t.Fatalf("kit capabilities update --json error = %v", err)
	}
	var targeted capabilityDetailPayload
	if err := json.Unmarshal([]byte(targetedOutput), &targeted); err != nil {
		t.Fatalf("json.Unmarshal(targeted hidden) error = %v", err)
	}
	if targeted.Command.Command != "update" || !targeted.Command.Hidden || !targeted.Command.Deprecated {
		t.Fatalf("expected targeted hidden lookup to return labeled update, got %#v", targeted.Command)
	}

	searchOutput, err := executeCapabilitiesCommand("--search", "update", "--json")
	if err != nil {
		t.Fatalf("kit capabilities --search update --json error = %v", err)
	}
	var search capabilitiesSearchPayload
	if err := json.Unmarshal([]byte(searchOutput), &search); err != nil {
		t.Fatalf("json.Unmarshal(search) error = %v", err)
	}
	if findCompactCapability(search.Commands, "update") != nil {
		t.Fatalf("expected search to omit hidden deprecated update")
	}
}

func TestCapabilitiesSearchJSON(t *testing.T) {
	output, err := executeCapabilitiesCommand("--search", "dispatch", "--json")
	if err != nil {
		t.Fatalf("kit capabilities --search dispatch --json error = %v", err)
	}
	var payload capabilitiesSearchPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal(search) error = %v", err)
	}
	if payload.Kind != "capabilities_search" {
		t.Fatalf("kind = %q, want capabilities_search", payload.Kind)
	}
	if payload.Query != "dispatch" {
		t.Fatalf("query = %q, want dispatch", payload.Query)
	}
	if findCompactCapability(payload.Commands, "dispatch") == nil {
		t.Fatalf("expected search results to include dispatch")
	}

	reviewLoopOutput, err := executeCapabilitiesCommand("--search", "review-loop", "--json")
	if err != nil {
		t.Fatalf("kit capabilities --search review-loop --json error = %v", err)
	}
	var reviewLoopSearch capabilitiesSearchPayload
	if err := json.Unmarshal([]byte(reviewLoopOutput), &reviewLoopSearch); err != nil {
		t.Fatalf("json.Unmarshal(review-loop search) error = %v", err)
	}
	if findCompactCapability(reviewLoopSearch.Commands, "review-loop") == nil {
		t.Fatalf("expected review-loop search results to include review-loop")
	}
	if findCompactCapability(reviewLoopSearch.Commands, "dispatch") == nil {
		t.Fatalf("expected review-loop search results to include dispatch alias metadata")
	}

	emptyOutput, err := executeCapabilitiesCommand("--search", "no-such-capability-token", "--json")
	if err != nil {
		t.Fatalf("kit capabilities --search no-such-capability-token --json error = %v", err)
	}
	var emptyPayload capabilitiesSearchPayload
	if err := json.Unmarshal([]byte(emptyOutput), &emptyPayload); err != nil {
		t.Fatalf("json.Unmarshal(empty search) error = %v", err)
	}
	if len(emptyPayload.Commands) != 0 {
		t.Fatalf("expected zero search matches, got %#v", emptyPayload.Commands)
	}
}

func TestCapabilitiesRejectsInvalidCombinationsAndSuggests(t *testing.T) {
	if _, err := executeCapabilitiesCommand("--search", "verify", "verify"); err == nil || !strings.Contains(err.Error(), "--search cannot be combined") {
		t.Fatalf("expected --search plus command path to fail actionably, got %v", err)
	}
	if _, err := executeCapabilitiesCommand("--full", "verify"); err == nil || !strings.Contains(err.Error(), "--full cannot be combined") {
		t.Fatalf("expected --full plus command path to fail actionably, got %v", err)
	}
	if _, err := executeCapabilitiesCommand("verif"); err == nil || !strings.Contains(err.Error(), "verify") {
		t.Fatalf("expected unknown command to suggest verify, got %v", err)
	}
}

func TestCapabilitiesDoesNotRequireProjectRootOrWriteFiles(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	paths := []string{
		".kit.yaml",
		filepath.Join(".kit", "state.json"),
		filepath.Join(".kit", "runs", "existing.json"),
		filepath.Join(".kit", "loops", "existing.json"),
		filepath.Join("docs", "specs", "sample", "TASKS.md"),
		filepath.Join("docs", "PROJECT_PROGRESS_SUMMARY.md"),
	}
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte("before:"+path), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", path, err)
		}
	}
	before := snapshotFiles(t, tmp)

	for _, args := range [][]string{
		{"--json"},
		{"--full", "--json"},
		{"--search", "verify", "--json"},
		{"--json", "ci"},
	} {
		if _, err := executeCapabilitiesCommand(args...); err != nil {
			t.Fatalf("kit capabilities %v error = %v", args, err)
		}
	}

	after := snapshotFiles(t, tmp)
	if len(after) != len(before) {
		t.Fatalf("file count changed: before=%d after=%d before=%v after=%v", len(before), len(after), before, after)
	}
	for path, beforeContent := range before {
		if after[path] != beforeContent {
			t.Fatalf("file %q changed: before %q after %q", path, beforeContent, after[path])
		}
	}
}

func TestCapabilityCatalogCoversVisibleRootCommands(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Hidden || cmd.Deprecated != "" {
			continue
		}
		if cmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		if _, ok := capabilityByCommandPath(cmd.Name()); !ok {
			t.Fatalf("visible root command %q is missing from capability catalog", cmd.Name())
		}
	}
}

func TestCapabilityCatalogRecordsIncludeAgentGuidance(t *testing.T) {
	for _, record := range capabilityCatalog() {
		if record.Hidden || record.Deprecated {
			continue
		}
		detail := record.detail()
		if strings.TrimSpace(detail.Summary) == "" {
			t.Fatalf("%s missing summary", detail.Command)
		}
		if strings.TrimSpace(detail.MutationLevel) == "" {
			t.Fatalf("%s missing mutation level", detail.Command)
		}
		if strings.TrimSpace(detail.NetworkUse.Summary) == "" {
			t.Fatalf("%s missing network behavior", detail.Command)
		}
		if strings.TrimSpace(detail.FileWrites.Summary) == "" {
			t.Fatalf("%s missing file-write behavior", detail.Command)
		}
		if strings.TrimSpace(detail.GitMutation.Summary) == "" {
			t.Fatalf("%s missing git mutation behavior", detail.Command)
		}
		if len(detail.WhenToUse) == 0 {
			t.Fatalf("%s missing when_to_use guidance", detail.Command)
		}
		if len(detail.WhenNotToUse) == 0 {
			t.Fatalf("%s missing when_not_to_use guidance", detail.Command)
		}
		if len(detail.Examples) == 0 {
			t.Fatalf("%s missing examples", detail.Command)
		}
	}
}

func TestCapabilityCatalogNestedCommandsAreRegistered(t *testing.T) {
	for _, commandPath := range [][]string{
		{"scaffold", "agents"},
		{"prompt", "list"},
		{"set", "prompt"},
		{"skill", "mine"},
		{"rules", "add"},
		{"rules", "list"},
		{"rules", "view"},
		{"rules", "link"},
	} {
		commandName := strings.Join(commandPath, " ")
		cmd, _, err := rootCmd.Find(commandPath)
		if err != nil {
			t.Fatalf("rootCmd.Find(%s) error = %v", commandName, err)
		}
		if cmd == nil {
			t.Fatalf("expected %s to be registered", commandName)
		}
		if _, ok := capabilityByCommandPath(commandName); !ok {
			t.Fatalf("registered nested command %q is missing from capability catalog", commandName)
		}
	}
}

func executeCapabilitiesCommand(args ...string) (string, error) {
	cmd := newCapabilitiesCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func findCompactCapability(records []capabilityCompactRecord, command string) *capabilityCompactRecord {
	for i := range records {
		if records[i].Command == command {
			return &records[i]
		}
	}
	return nil
}

func findDetailCapability(records []capabilityDetailRecord, command string) *capabilityDetailRecord {
	for i := range records {
		if records[i].Command == command {
			return &records[i]
		}
	}
	return nil
}

func findDetailedFlag(flags []capabilityFlag, name string) *capabilityFlag {
	for i := range flags {
		if flags[i].Name == name {
			return &flags[i]
		}
	}
	return nil
}

func snapshotFiles(t *testing.T, root string) map[string]string {
	t.Helper()

	files := map[string]string{}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[relative] = string(content)
		return nil
	}); err != nil {
		t.Fatalf("WalkDir(%q) error = %v", root, err)
	}
	return files
}

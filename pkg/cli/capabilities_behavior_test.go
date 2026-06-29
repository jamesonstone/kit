package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestCapabilitiesSelfGuidanceDistinguishesMaintainersAndDownstreamProjects(t *testing.T) {
	output, err := executeCapabilitiesCommand("--json", "capabilities")
	if err != nil {
		t.Fatalf("kit capabilities capabilities --json error = %v", err)
	}

	var payload capabilityDetailPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, output)
	}
	var parts []string
	parts = append(parts, payload.Command.WhenToUse...)
	parts = append(parts, payload.Command.WhenNotToUse...)
	parts = append(parts, payload.Command.Caveats...)
	combined := strings.Join(parts, "\n")
	for _, want := range []string{
		"Inside the Kit source repository",
		"Do not maintain Kit's internal command catalog from a downstream project",
		"downstream projects should use it for discovery",
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("expected capabilities self-guidance to contain %q, got:\n%s", want, combined)
		}
	}
}

func TestCapabilitiesIncludesNotesCommandGuidance(t *testing.T) {
	output, err := executeCapabilitiesCommand("--json", "notes")
	if err != nil {
		t.Fatalf("kit capabilities notes --json error = %v", err)
	}

	var payload capabilityDetailPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, output)
	}
	combined := strings.Join(append(payload.Command.WhenToUse, payload.Command.Caveats...), "\n")
	for _, want := range []string{
		"Slack excerpts",
		"private notes",
		"feature-notes.md",
		"ignore .gitkeep",
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("expected notes capabilities guidance to contain %q, got:\n%s", want, combined)
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
		{"legacy", "brainstorm"},
		{"legacy", "plan"},
		{"legacy", "tasks"},
		{"legacy", "implement"},
		{"legacy", "reflect"},
		{"legacy", "verify"},
		{"scaffold", "agents"},
		{"prompt", "list"},
		{"set", "prompt"},
		{"skill", "mine"},
		{"pr", "fix"},
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
	for _, command := range []string{"update", "skills", "skills mine", "catchup", "rollup", "review-loop", "brainstorm", "plan", "tasks", "implement", "reflect", "verify"} {
		if found := findDetailCapability(fullPayload.Commands, command); found != nil {
			t.Fatalf("expected full capabilities to omit removed command %q, got %#v", command, found)
		}
	}

	if _, err := executeCapabilitiesCommand("--json", "update"); err == nil || !strings.Contains(err.Error(), "unknown Kit command path") {
		t.Fatalf("expected targeted update lookup to fail as removed, got %v", err)
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
	if findCompactCapability(reviewLoopSearch.Commands, "loop review") == nil {
		t.Fatalf("expected review-loop search results to include loop review")
	}
	if findCompactCapability(reviewLoopSearch.Commands, "dispatch") == nil {
		t.Fatalf("expected review-loop search results to include dispatch alias metadata")
	}

	kitSpecPromptOutput, err := executeCapabilitiesCommand("--search", "kit spec prompt", "--json")
	if err != nil {
		t.Fatalf("kit capabilities --search kit spec prompt --json error = %v", err)
	}
	var kitSpecPromptSearch capabilitiesSearchPayload
	if err := json.Unmarshal([]byte(kitSpecPromptOutput), &kitSpecPromptSearch); err != nil {
		t.Fatalf("json.Unmarshal(kit spec prompt search) error = %v", err)
	}
	if findCompactCapability(kitSpecPromptSearch.Commands, "prompt") == nil {
		t.Fatalf("expected kit spec prompt search results to include prompt")
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
	if _, err := executeCapabilitiesCommand("--search", "verify", "legacy", "verify"); err == nil || !strings.Contains(err.Error(), "--search cannot be combined") {
		t.Fatalf("expected --search plus command path to fail actionably, got %v", err)
	}
	if _, err := executeCapabilitiesCommand("--full", "legacy", "verify"); err == nil || !strings.Contains(err.Error(), "--full cannot be combined") {
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
		{"--json", "legacy", "verify"},
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

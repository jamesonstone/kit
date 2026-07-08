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

	for _, command := range []string{"capabilities", "ci", "pr fix", "legacy verify", "loop prompt", "loop review", "project refresh", "improve", "improve run", "dispatch", "rules add", "skill mine"} {
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

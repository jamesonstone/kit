package cli

import (
	"bytes"
	"encoding/json"
	stdreflect "reflect"
	"sort"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/commandset"
)

func TestCapabilityCatalogMatchesSupportedKitCommandPaths(t *testing.T) {
	seen := map[string]bool{}
	for _, path := range commandset.ProtectedPaths() {
		fields := strings.Fields(path)
		for index := range fields {
			seen[strings.Join(fields[:index+1], " ")] = true
		}
	}
	want := make([]string, 0, len(seen))
	for path := range seen {
		want = append(want, path)
	}
	sort.Strings(want)
	var got []string
	for _, record := range capabilityCatalog() {
		if strings.HasPrefix(record.Command, "git wt") || record.Command == "help" {
			continue
		}
		got = append(got, record.Command)
	}
	sort.Strings(got)
	if !stdreflect.DeepEqual(got, want) {
		t.Fatalf("capability paths mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCapabilitiesIndexContainsNoRemovedCommands(t *testing.T) {
	cmd := newCapabilitiesCommand()
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{"--full", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var payload capabilitiesFullPayload
	if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	for _, removed := range []string{"notes", "project refresh", "loop", "ci", "scaffold", "prompt", "legacy", "skill"} {
		for _, record := range payload.Commands {
			if record.Command == removed || strings.HasPrefix(record.Command, removed+" ") {
				t.Fatalf("removed capability %q remains", record.Command)
			}
		}
	}
}

func TestContextAndUsageCapabilitiesExposeSafetyBoundaries(t *testing.T) {
	contextRecord, ok := capabilityByCommandPath("context resolve")
	if !ok || contextRecord.MutationLevel != mutationNone || contextRecord.NetworkUse.Summary != "none" {
		t.Fatalf("context capability = %#v", contextRecord)
	}
	usageRecord, ok := capabilityByCommandPath("usage")
	if !ok || usageRecord.NetworkUse.Summary != "none" || !strings.Contains(strings.Join(usageRecord.Caveats, " "), "secrets") {
		t.Fatalf("usage capability = %#v", usageRecord)
	}
}

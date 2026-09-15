package instructions

import (
	"strings"
	"testing"
)

func TestV16NamingPreservesOtherInstructions(t *testing.T) {
	prior, err := AgentInstructions("v15")
	if err != nil {
		t.Fatal(err)
	}
	current, err := AgentInstructions("v16")
	if err != nil {
		t.Fatal(err)
	}
	if CurrentAgentVersion != "v16" {
		t.Fatal("v16 must be current")
	}
	const tail = "# Multi-agent orchestration evaluation gate"
	if prior[strings.Index(prior, tail):] != current[strings.Index(current, tail):] {
		t.Fatal("changed unrelated instructions")
	}
	for _, want := range []string{"[scope] domain / objective", "kit instructions naming", "ownership materially changes", "leaving\nthe parent unchanged"} {
		if !strings.Contains(current, want) {
			t.Fatalf("missing %q", want)
		}
	}
	for _, stale := range []string{"at most four words", "[<project>] <description>"} {
		if strings.Contains(current, stale) {
			t.Fatalf("old naming rule %q", stale)
		}
	}
}

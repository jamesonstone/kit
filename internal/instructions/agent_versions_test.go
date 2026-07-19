package instructions

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestAgentInstructionsDefaultsToCurrentVersion(t *testing.T) {
	current, err := AgentInstructions(CurrentAgentVersion)
	if err != nil {
		t.Fatalf("AgentInstructions(%q) error = %v", CurrentAgentVersion, err)
	}

	got, err := AgentInstructions("")
	if err != nil {
		t.Fatalf("AgentInstructions(\"\") error = %v", err)
	}
	if got != current {
		t.Fatal("default agent instructions do not match the current version")
	}
}

func TestAgentInstructionsV1IsImmutable(t *testing.T) {
	content, err := AgentInstructions("v1")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v1\") error = %v", err)
	}

	if got, want := fmt.Sprintf("%x", sha256.Sum256([]byte(content))), "50cbfd80732e7b1912dc65f160cbf8555d2da95cb79079f33d7131cd51a86be5"; got != want {
		t.Fatalf("v1 content SHA-256 = %s, want %s", got, want)
	}
	if !strings.HasSuffix(content, "\n") || strings.HasSuffix(content, "\n\n") {
		t.Fatal("v1 content must end with exactly one newline")
	}
}

func TestAgentInstructionsRejectsUnavailableVersion(t *testing.T) {
	_, err := AgentInstructions("v2")
	if err == nil {
		t.Fatal("AgentInstructions(\"v2\") expected an error")
	}
	for _, want := range []string{`unsupported instructions version "v2"`, "available versions: v1"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("AgentInstructions(\"v2\") error = %q, want %q", err, want)
		}
	}
}

func TestAgentInstructionVersionsReturnsCopy(t *testing.T) {
	versions := AgentInstructionVersions()
	if len(versions) != 1 || versions[0] != CurrentAgentVersion {
		t.Fatalf("AgentInstructionVersions() = %v, want [%s]", versions, CurrentAgentVersion)
	}

	versions[0] = "changed"
	if got := AgentInstructionVersions()[0]; got != CurrentAgentVersion {
		t.Fatalf("AgentInstructionVersions() exposed registry mutation: got %q", got)
	}
}

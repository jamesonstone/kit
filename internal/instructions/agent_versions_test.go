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

func TestAgentInstructionVersionsAreImmutable(t *testing.T) {
	tests := []struct {
		version string
		sha256  string
	}{
		{version: "v1", sha256: "50cbfd80732e7b1912dc65f160cbf8555d2da95cb79079f33d7131cd51a86be5"},
		{version: "v2", sha256: "1050ce514a49f9bef9446c9a9166bae168e79ed9bf454a8a09ff94f4c2feb59a"},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			content, err := AgentInstructions(test.version)
			if err != nil {
				t.Fatalf("AgentInstructions(%q) error = %v", test.version, err)
			}

			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(content))); got != test.sha256 {
				t.Fatalf("%s content SHA-256 = %s, want %s", test.version, got, test.sha256)
			}
			if !strings.HasSuffix(content, "\n") || strings.HasSuffix(content, "\n\n") {
				t.Fatalf("%s content must end with exactly one newline", test.version)
			}
		})
	}
}

func TestAgentInstructionsV2EncodesLaneAllocationPolicy(t *testing.T) {
	content, err := AgentInstructions("v2")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v2\") error = %v", err)
	}

	for _, want := range []string{
		"Do not ask whether to create a new issue, branch, and pull request or continue existing work.",
		"Create a branch named `GH-<issue-number>` from the refreshed remote default branch.",
		"create a new GitHub Issue for the additional work, keep the existing branch and pull request",
		"scope the new commits to the new issue",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v2 instructions do not contain %q", want)
		}
	}
}

func TestAgentInstructionsRejectsUnavailableVersion(t *testing.T) {
	_, err := AgentInstructions("v3")
	if err == nil {
		t.Fatal("AgentInstructions(\"v3\") expected an error")
	}
	for _, want := range []string{`unsupported instructions version "v3"`, "available versions: v1, v2"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("AgentInstructions(\"v3\") error = %q, want %q", err, want)
		}
	}
}

func TestAgentInstructionVersionsReturnsCopy(t *testing.T) {
	versions := AgentInstructionVersions()
	if len(versions) != 2 || versions[0] != "v1" || versions[1] != "v2" {
		t.Fatalf("AgentInstructionVersions() = %v, want [v1 v2]", versions)
	}
	if versions[len(versions)-1] != CurrentAgentVersion {
		t.Fatalf("last available version = %q, want current %q", versions[len(versions)-1], CurrentAgentVersion)
	}

	versions[0] = "changed"
	if got := AgentInstructionVersions()[0]; got != "v1" {
		t.Fatalf("AgentInstructionVersions() exposed registry mutation: got %q", got)
	}
}

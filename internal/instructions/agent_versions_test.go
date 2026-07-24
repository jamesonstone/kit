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
		{version: "v2", sha256: "811842c5c87a1b8c7f82831c7c76739071921583c44b0ab9c5dc62cbc08b27fc"},
		{version: "v3", sha256: "970eead03113cbd0e576894f83098f28cacb2fadb8b15aeb17acc57a240098d3"},
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

func TestAgentInstructionsV3EncodesLaneAllocationPolicy(t *testing.T) {
	content, err := AgentInstructions("v3")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v3\") error = %v", err)
	}

	for _, want := range []string{
		"Do not ask whether to create a new issue, branch, and pull request or continue existing work.",
		"Create a branch named `GH-<issue-number>` from the refreshed remote default branch.",
		"create or reuse a separate GitHub Issue for the additional work, keep the existing branch and pull request",
		"scope the new commits to that issue",
		"open a pull request for review when none exists; otherwise update the existing pull request",
		"`~/worktrees/<owner>/<repository>/<lane>`",
		"uppercase detached `PR-<number>`",
		"Use native `git worktree` commands and ordinary filesystem operations as the portable authority",
		"do not require `git-wt`, an alias, or another wrapper",
		"Optional wrappers are manual conveniences only",
		"Keep the root checkout on the protected default branch",
		"Link the primary checkout's `.env` into writable lanes by default when it exists",
		"omit the link when isolation is required",
		"Never copy `.env` contents or automatically share `.envrc`",
		"worktree tooling does not manage runtime services, databases, ports, Temporal state, processes, or sibling repositories",
		"load `docs/references/rules/constitution-curation.md` when present",
		"`Repository Memory`, `Decision`, `Rationale`, and `Artifacts`",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v3 instructions do not contain %q", want)
		}
	}
	for _, forbidden := range []string{"Use the Kit-owned `git wt`", "`--no-link-env`", "Let GitWT"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v3 instructions must not depend on wrapper-specific policy %q", forbidden)
		}
	}
}

func TestAgentInstructionsRejectsUnavailableVersion(t *testing.T) {
	_, err := AgentInstructions("v4")
	if err == nil {
		t.Fatal("AgentInstructions(\"v4\") expected an error")
	}
	for _, want := range []string{`unsupported instructions version "v4"`, "available versions: v1, v2, v3"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("AgentInstructions(\"v4\") error = %q, want %q", err, want)
		}
	}
}

func TestAgentInstructionVersionsReturnsCopy(t *testing.T) {
	versions := AgentInstructionVersions()
	if len(versions) != 3 || versions[0] != "v1" || versions[1] != "v2" || versions[2] != "v3" {
		t.Fatalf("AgentInstructionVersions() = %v, want [v1 v2 v3]", versions)
	}
	if versions[len(versions)-1] != CurrentAgentVersion {
		t.Fatalf("last available version = %q, want current %q", versions[len(versions)-1], CurrentAgentVersion)
	}

	versions[0] = "changed"
	if got := AgentInstructionVersions()[0]; got != "v1" {
		t.Fatalf("AgentInstructionVersions() exposed registry mutation: got %q", got)
	}
}

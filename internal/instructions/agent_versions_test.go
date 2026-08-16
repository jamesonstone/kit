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
		{version: "v3", sha256: "a75fb2b02d37a7fbdc5926b9c71130210c6e929366b09707b410ab2f5b90792f"},
		{version: "v4", sha256: "96fc2b3bbd4f458ef55ae32910d737dd1ea35110d6443d6ee8e03d389d851986"},
		{version: "v5", sha256: "67122fe42ce3bcb65a9b1f355271395ebe4c65d43fef5b1b9632580cacf5e3d6"},
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
		"Link the primary checkout's `.env` and `.envrc` into writable lanes by default when each exists",
		"omit both links when isolation is required",
		"preserve a repository- or user-supplied `.envrc`",
		"direnv approval remains path-specific",
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

func TestAgentInstructionsV4RequiresExplicitPullRequestLane(t *testing.T) {
	content, err := AgentInstructions("v4")
	if err != nil {
		t.Fatalf("AgentInstructions(\"v4\") error = %v", err)
	}

	for _, want := range []string{
		"blocking pre-response gate",
		"before the first commentary message",
		"After the title action resolves, invoke the available thread-pin operation",
		"Verify each result from returned host state when that state is available",
		"unsupported or unavailable capability",
		"operation failure",
		"Thread initialization: rename <status>; pin <status>.",
		"continue the user's substantive request",
		"must not deadlock unrelated work",
		"working-directory context supplied by the host",
		"Do not inspect the repository",
		"Start repository work at `docs/agents/README.md` when present",
		"Before every repository mutation in a newly created or resumed session",
		"Load `docs/agents/GUARDRAILS.md`",
		"`docs/references/rules/work-lane-gating.md`",
		"when each is present and relevant",
		"Before I make any repository changes, should I create a new GitHub issue",
		"canonical worktree, and pull request for this work",
		"Record or verify a complete Pull-Request Landing Plan before every",
		"Pull-Request Landing Plan",
		"A Pull-Request Landing Plan proves a delivery lane; it does not authorize",
		"direct user request or an accepted bounded merge",
		"exact pull-request set",
		"resolve `pull-request-merge`",
		"`docs/references/rules/github-pr-merge.md`",
		"Only exact current `MERGE_READY` nodes may merge",
		"`kit capabilities context resolve --json`",
		"`kit context resolve --workflow <slug> --json` with relevant",
		"`--feature` and `--path` hints",
		"every selected artifact that is required",
		"Use native planning for research, clarification, design, and the accepted",
		"decide whether material rationale requires creating or adopting",
		"`docs/specs/<feature>/SPEC.md`",
		"capture the accepted native",
		"plan there before editing implementation files",
		"`docs/references/rules/testing-and-environment-validation.md`",
		"the project's `docs/references/testing.md`.",
		"`docs/references/rules/source-file-size.md`.",
		"300 physical lines or fewer",
		"complete affected scope before delivery",
		"Do not infer the choice from a clean default branch",
		"every other repository file",
		"issue, branch, worktree, staging, commit, push, pull-request",
		"Treat the clone's primary/root checkout as read-only",
		"Never edit, generate, stage, commit, or switch branches there",
		"one planned ready pull",
		"Do not stage, commit, push, stash, reset, clean, discard, or silently transfer",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("v4 instructions do not contain %q", want)
		}
	}
	for _, order := range []struct {
		before string
		after  string
	}{
		{before: "Invoke the available thread-title operation", after: "After the title action resolves, invoke the available thread-pin operation"},
		{before: "After the title action resolves, invoke the available thread-pin operation", after: "Verify each result from returned host state"},
		{before: "Start repository work at `docs/agents/README.md`", after: "Before every repository mutation"},
		{before: "Before every repository mutation", after: "Load `docs/agents/GUARDRAILS.md`"},
		{before: "`docs/agents/GUARDRAILS.md`", after: "`docs/references/rules/work-lane-gating.md`"},
		{before: "`kit capabilities context resolve --json`", after: "`kit context resolve --workflow <slug> --json`"},
		{before: "`kit context resolve --workflow <slug> --json`", after: "Use native planning for research, clarification, design, and the accepted"},
		{before: "every selected artifact that is required", after: "`docs/references/rules/testing-and-environment-validation.md`"},
		{before: "`docs/references/rules/testing-and-environment-validation.md`", after: "`docs/references/rules/source-file-size.md`"},
		{before: "`docs/references/rules/source-file-size.md`", after: "Own the requested outcome"},
	} {
		beforeIndex := strings.Index(content, order.before)
		afterIndex := strings.Index(content, order.after)
		if beforeIndex < 0 || afterIndex < 0 || beforeIndex >= afterIndex {
			t.Fatalf("v4 instructions must place %q before %q", order.before, order.after)
		}
	}
	for _, forbidden := range []string{
		"Do not ask whether to create a new issue",
		"automatic clean-preflight",
		"Never stop or delay work if renaming or pinning is unsupported or fails",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("v4 instructions contain forbidden automatic-lane policy %q", forbidden)
		}
	}
}

func TestAgentInstructionsRejectsUnavailableVersion(t *testing.T) {
	_, err := AgentInstructions("v6")
	if err == nil {
		t.Fatal("AgentInstructions(\"v6\") expected an error")
	}
	for _, want := range []string{`unsupported instructions version "v6"`, "available versions: v1, v2, v3, v4, v5"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("AgentInstructions(\"v6\") error = %q, want %q", err, want)
		}
	}
}

func TestAgentInstructionVersionsReturnsCopy(t *testing.T) {
	versions := AgentInstructionVersions()
	if len(versions) != 5 || versions[0] != "v1" || versions[1] != "v2" || versions[2] != "v3" || versions[3] != "v4" || versions[4] != "v5" {
		t.Fatalf("AgentInstructionVersions() = %v, want [v1 v2 v3 v4 v5]", versions)
	}
	if versions[len(versions)-1] != CurrentAgentVersion {
		t.Fatalf("last available version = %q, want current %q", versions[len(versions)-1], CurrentAgentVersion)
	}

	versions[0] = "changed"
	if got := AgentInstructionVersions()[0]; got != "v1" {
		t.Fatalf("AgentInstructionVersions() exposed registry mutation: got %q", got)
	}
}

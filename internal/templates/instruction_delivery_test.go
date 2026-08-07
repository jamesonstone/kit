package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestMemoryCopilotInstructionsPreserveMutationRouting(t *testing.T) {
	for _, want := range []string{
		"Start with `docs/agents/README.md`",
		"load `docs/references/rules/constitution-curation.md`",
		"Before Git, GitHub, or AWS mutations, load `docs/agents/GUARDRAILS.md` and relevant `docs/references/rules/*`",
		"Repo-local Kit rules outrank generic defaults",
	} {
		if !strings.Contains(MemoryCopilotInstructionsMD, want) {
			t.Fatalf("expected V3 Copilot instructions to contain %q", want)
		}
	}

	checkedIn, err := os.ReadFile(filepath.Join("..", "..", ".github", "copilot-instructions.md"))
	if err != nil {
		t.Fatalf("read checked-in Copilot instructions: %v", err)
	}
	if string(checkedIn) != MemoryCopilotInstructionsMD {
		t.Fatal("checked-in Copilot instructions are not aligned with the V3 generator")
	}
}

func TestMemoryGuardrailsPreserveAutonomousRecovery(t *testing.T) {
	generated := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/agents/GUARDRAILS.md",
	)
	for _, want := range []string{
		"Resolve all in-scope issues autonomously and continue until the goal is fully complete",
		"including authenticated `gh`",
		"Outside explicit repo-local approval gates, ask permission only before large-scale deletion or deleting sensitive files",
		"not as routine retry-permission requests",
		"Use native `git worktree` commands and ordinary filesystem operations as the portable authority",
		"do not require a wrapper, alias, or plugin",
		"omit both links when isolation is required",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("expected V3 guardrails to contain %q", want)
		}
	}
	if strings.Contains(generated, "Do not run `git add` or `git commit` without explicit approval") {
		t.Fatal("expected V3 guardrails to omit routine git approval requirement")
	}
	for _, forbidden := range []string{"`--no-link-env`", "Let GitWT"} {
		if strings.Contains(generated, forbidden) {
			t.Fatalf("V3 guardrails must not depend on wrapper-specific policy %q", forbidden)
		}
	}

	checkedIn, err := os.ReadFile(filepath.Join("..", "..", "docs", "agents", "GUARDRAILS.md"))
	if err != nil {
		t.Fatalf("read checked-in guardrails: %v", err)
	}
	if string(checkedIn) != generated {
		t.Fatal("checked-in guardrails are not aligned with the V3 generator")
	}
}

func TestMemoryInstructionsPreserveProjectOrientedWorktrees(t *testing.T) {
	tooling := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/agents/TOOLING.md",
	)
	for _, want := range []string{
		"`~/worktrees/<owner>/<repository>/<lane>`",
		"uppercase detached `PR-<number>`",
		"never edit the detached `PR-<number>` view",
		"Use native `git worktree` commands as the portable authority",
		"do not require `git-wt`, an alias, or another wrapper",
		"Optional wrappers are manual conveniences only",
		"Keep the root checkout on the protected default branch",
		"Link the primary checkout's `.env` and `.envrc` into writable lanes by default when each exists",
		"omit both links when isolation is required",
		"preserve a repository- or user-supplied `.envrc`",
		"direnv approval remains path-specific",
		"worktree tooling does not manage runtime services, databases, ports, Temporal state, processes, or sibling repositories",
		"refs, remotes, objects, configuration, and stash state are shared",
	} {
		if !strings.Contains(tooling, want) {
			t.Fatalf("expected V3 tooling to contain %q", want)
		}
	}
	for _, forbidden := range []string{"Use the Kit-owned `git wt`", "`--no-link-env`", "Let GitWT"} {
		if strings.Contains(tooling, forbidden) {
			t.Fatalf("V3 tooling must not depend on wrapper-specific policy %q", forbidden)
		}
	}

	checkedIn, err := os.ReadFile(filepath.Join("..", "..", "docs", "agents", "TOOLING.md"))
	if err != nil {
		t.Fatalf("read checked-in tooling: %v", err)
	}
	if string(checkedIn) != tooling {
		t.Fatal("checked-in tooling is not aligned with the V3 generator")
	}
}

func TestMemoryWorktreeReferenceIsManagedAndNative(t *testing.T) {
	generated := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/references/worktrees.md",
	)
	for _, want := range []string{
		"Native `git worktree` commands and ordinary filesystem operations define this",
		"The clone's primary checkout owns the shared repository-root `.env`",
		"The user does not need to",
		"`include` makes the existing diff part of the full repair review",
		"`git worktree remove",
		"Runtime services, databases, ports",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("expected V3 worktree reference to contain %q", want)
		}
	}
	for _, forbidden := range []string{"git wt issue", "--no-link-env"} {
		if strings.Contains(generated, forbidden) {
			t.Fatalf("V3 worktree reference must not require optional wrapper syntax %q", forbidden)
		}
	}

	checkedIn, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "worktrees.md"))
	if err != nil {
		t.Fatalf("read checked-in worktree reference: %v", err)
	}
	if string(checkedIn) != generated {
		t.Fatal("checked-in worktree reference is not aligned with the V3 generator")
	}
}

func TestMemoryRepositoryInstructionsRouteConstitutionCuration(t *testing.T) {
	for relativePath, generated := range map[string]string{
		"AGENTS.md": MemoryAgentsMD,
		"CLAUDE.md": MemoryClaudeMD,
	} {
		if !strings.Contains(generated, "load `docs/references/rules/constitution-curation.md`") {
			t.Fatalf("expected V3 %s to route Constitution curation", relativePath)
		}
		checkedIn, err := os.ReadFile(filepath.Join("..", "..", relativePath))
		if err != nil {
			t.Fatalf("read checked-in %s: %v", relativePath, err)
		}
		if string(checkedIn) != generated {
			t.Fatalf("checked-in %s is not aligned with the V3 generator", relativePath)
		}
	}
}

func TestMemoryRepositoryInstructionsRouteApplicationArchitecture(t *testing.T) {
	routes := []string{
		"load `docs/references/rules/backend-service-architecture.md`",
		"load `docs/references/rules/frontend-application-architecture.md`",
		"responsibility boundaries rather than mandatory directory names",
	}
	for name, content := range map[string]string{
		"AGENTS.md":                       MemoryAgentsMD,
		"CLAUDE.md":                       MemoryClaudeMD,
		".github/copilot-instructions.md": MemoryCopilotInstructionsMD,
	} {
		for _, route := range routes {
			if !strings.Contains(content, route) {
				t.Errorf("expected V3 %s to contain architecture route %q", name, route)
			}
		}
	}

	generatedRLM := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/agents/RLM.md",
	)
	for _, rulePath := range []string{
		"`docs/references/rules/backend-service-architecture.md`",
		"`docs/references/rules/frontend-application-architecture.md`",
	} {
		if !strings.Contains(generatedRLM, rulePath) {
			t.Errorf("expected generated RLM guidance to contain %q", rulePath)
		}
	}
	checkedInRLM, err := os.ReadFile(filepath.Join("..", "..", "docs", "agents", "RLM.md"))
	if err != nil {
		t.Fatalf("read checked-in RLM guidance: %v", err)
	}
	if string(checkedInRLM) != generatedRLM {
		t.Fatal("checked-in RLM guidance is not aligned with the V3 generator")
	}
}

func TestInstructionTemplatesRouteTestingAndEnvironmentValidation(t *testing.T) {
	routes := []string{
		"Before implementation or validation, including browser automation and browser testing, load `docs/references/rules/testing-and-environment-validation.md` and the project's `docs/references/testing.md`",
		"end-to-end and live-integration suites supplement rather than replace them",
	}
	for name, content := range map[string]string{
		"V2 AGENTS.md":                       AgentsMD,
		"V2 CLAUDE.md":                       ClaudeMD,
		"V2 .github/copilot-instructions.md": CopilotInstructionsMD,
		"V3 AGENTS.md":                       MemoryAgentsMD,
		"V3 CLAUDE.md":                       MemoryClaudeMD,
		"V3 .github/copilot-instructions.md": MemoryCopilotInstructionsMD,
	} {
		for _, route := range routes {
			if !strings.Contains(content, route) {
				t.Errorf("expected %s to contain testing route %q", name, route)
			}
		}
	}

	for _, version := range []int{
		config.InstructionScaffoldVersionTOC,
		config.InstructionScaffoldVersionMemory,
	} {
		generatedRLM := fileContentByPath(
			InstructionSupportFiles(version),
			"docs/agents/RLM.md",
		)
		for _, route := range []string{
			"Load `docs/references/rules/testing-and-environment-validation.md` and `docs/references/testing.md` before implementation or validation, including browser automation and browser testing",
		} {
			if !strings.Contains(generatedRLM, route) {
				t.Errorf("expected version %d RLM guidance to contain %q", version, route)
			}
		}
	}

	generatedTesting := fileContentByPath(
		InstructionSupportFiles(config.InstructionScaffoldVersionMemory),
		"docs/references/testing.md",
	)
	for _, check := range []string{
		"rules/testing-and-environment-validation.md",
		"## Code-Level Validation",
		"## High-Level Suites",
		"## Environment Preflights",
		"## Credentials And Test Data",
		"## Evidence And Retention",
		"`tests/RUN_STATUS.md`",
		"## Automation And Fallbacks",
		"## Known Gaps",
	} {
		if !strings.Contains(generatedTesting, check) {
			t.Errorf("expected generated testing reference to contain %q", check)
		}
	}
	checkedInTesting, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "testing.md"))
	if err != nil {
		t.Fatalf("read checked-in testing reference: %v", err)
	}
	if string(checkedInTesting) != generatedTesting {
		t.Fatal("checked-in testing reference is not aligned with the V3 generator")
	}
}

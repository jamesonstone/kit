package cli

import (
	"strings"
	"testing"
)

func TestBuiltInKitPromptSourceCatalog(t *testing.T) {
	source := builtInKitPromptSource()
	if source.Kind != "builtin" {
		t.Fatalf("source.Kind = %q, want builtin", source.Kind)
	}
	if source.Location != builtInPromptLocation {
		t.Fatalf("source.Location = %q, want %q", source.Location, builtInPromptLocation)
	}

	prompts := make(map[string]bool, len(source.Prompts))
	for _, prompt := range source.Prompts {
		if prompt.Description == "" {
			t.Fatalf("prompt %q has empty description", prompt.Identity.CommandName())
		}
		if prompt.Render == nil {
			t.Fatalf("prompt %q has nil render adapter", prompt.Identity.CommandName())
		}
		prompts[prompt.Identity.CommandName()] = true
	}

	expected := []string{
		"workflow spec",
		"kit spec",
		"support resume",
		"support handoff",
		"support summarize",
		"support reconcile",
		"support dispatch",
		"support code-review",
		"skill mine",
		"project init",
		"project refresh",
	}
	for _, command := range expected {
		if !prompts[command] {
			t.Fatalf("missing built-in prompt %q", command)
		}
	}
	removed := []string{
		"workflow brainstorm",
		"workflow plan",
		"workflow tasks",
		"workflow implement",
		"workflow reflect",
	}
	for _, command := range removed {
		if prompts[command] {
			t.Fatalf("removed v1 workflow prompt %q is still registered", command)
		}
	}
}

func TestBuiltInPromptSourcesIncludesToolboxAndKitCatalogs(t *testing.T) {
	sources := builtInPromptSources()
	if len(sources) != 2 {
		t.Fatalf("len(builtInPromptSources()) = %d, want 2", len(sources))
	}
	if sources[0].Prompts[0].Identity.CommandName() != "coding-agent short" {
		t.Fatalf("expected toolbox source first, got %q", sources[0].Prompts[0].Identity.CommandName())
	}
	if sources[1].Prompts[0].Identity.CommandName() != "workflow spec" {
		t.Fatalf("expected Kit source second, got %q", sources[1].Prompts[0].Identity.CommandName())
	}
}

func TestBuiltInKitStaticRenderAdapters(t *testing.T) {
	tests := map[string]string{
		"support summarize":   "## Context Summarization Instructions",
		"support code-review": "## Code Review",
		"project init":        "This document will drive the \"rules for development\" going forward.",
	}

	for _, prompt := range builtInKitPromptSource().Prompts {
		check, ok := tests[prompt.Identity.CommandName()]
		if !ok {
			continue
		}
		rendered, err := prompt.Render()
		if err != nil {
			t.Fatalf("%s Render() error = %v", prompt.Identity.CommandName(), err)
		}
		if !strings.Contains(rendered, check) {
			t.Fatalf("expected %q render output to contain %q", prompt.Identity.CommandName(), check)
		}
		delete(tests, prompt.Identity.CommandName())
	}

	if len(tests) != 0 {
		t.Fatalf("static render adapter tests did not run for %v", tests)
	}
}

func TestBuildProjectRefreshPrompt(t *testing.T) {
	projectRoot := t.TempDir()
	prompt := buildProjectRefreshPrompt(projectRoot, defaultInitConfig())

	checks := []string{
		"## Project Refresh",
		"Only update project-level documentation; do not modify product code, tests, runtime config, generated artifacts, or implementation files.",
		"docs/CONSTITUTION.md",
		"kit reconcile --all",
		"PROJECT_PROGRESS_SUMMARY.md",
		"kit check --project",
		"`Findings`",
		"`Updates`",
		"`Verification`",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected project refresh prompt to contain %q, got %q", check, prompt)
		}
	}
	if strings.Contains(prompt, "/plan") {
		t.Fatalf("expected project refresh prompt to avoid native plan-mode triggers, got %q", prompt)
	}
}

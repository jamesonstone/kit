package cli

import (
	"strings"
	"testing"
)

func TestBuiltInToolboxPromptSource(t *testing.T) {
	source := builtInToolboxPromptSource()
	if source.Kind != "builtin" {
		t.Fatalf("source.Kind = %q, want builtin", source.Kind)
	}
	if source.Location != builtInPromptLocation {
		t.Fatalf("source.Location = %q, want %q", source.Location, builtInPromptLocation)
	}

	prompts := make(map[string]string, len(source.Prompts))
	for _, prompt := range source.Prompts {
		if prompt.Description == "" {
			t.Fatalf("prompt %q has empty description", prompt.Identity.CommandName())
		}
		prompts[prompt.Identity.CommandName()] = prompt.Content
	}

	expectedCommands := []string{
		"coding-agent short",
		"coding-agent long",
		"coding-agent instructions",
	}
	for _, command := range expectedCommands {
		if prompts[command] == "" {
			t.Fatalf("missing built-in prompt %q", command)
		}
	}

	if !strings.Contains(prompts["coding-agent short"], "Clarify before implementing") {
		t.Fatalf("short prompt does not preserve planning payload")
	}
	if !strings.Contains(prompts["coding-agent long"], "Ask clarifying questions until you have >=95% confidence") {
		t.Fatalf("long prompt does not preserve planning payload")
	}
	if !strings.Contains(prompts["coding-agent instructions"], "Files to Create / Modify") {
		t.Fatalf("instructions prompt does not preserve instruction payload")
	}
}

func TestBuiltInToolboxPromptsExcludeShellAutomation(t *testing.T) {
	blocked := []string{
		"osascript",
		"pbcopy",
		"pbpaste",
		"sleep ",
		"old_clipboard",
		"keystroke",
	}

	for _, prompt := range builtInToolboxPromptSource().Prompts {
		for _, value := range blocked {
			if strings.Contains(prompt.Content, value) {
				t.Fatalf("prompt %q includes shell automation value %q", prompt.Identity.CommandName(), value)
			}
		}
	}
}

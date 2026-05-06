package promptlib

import (
	"strings"
	"testing"
)

func TestMergeUsesLocalGlobalBuiltinPrecedence(t *testing.T) {
	effective, err := Merge(
		Source{
			Kind:     SourceBuiltin,
			Location: "builtin",
			Prompts: []Prompt{{
				Identity:    Identity{Noun: "coding-agent", Verb: "short"},
				Content:     "builtin",
				Description: "builtin prompt",
			}},
		},
		Source{
			Kind:     SourceGlobal,
			Location: "global",
			Prompts: []Prompt{{
				Identity:    Identity{Noun: "coding-agent", Verb: "short"},
				Content:     "global",
				Description: "global prompt",
			}},
		},
		Source{
			Kind:     SourceLocal,
			Location: "local",
			Prompts: []Prompt{{
				Identity:    Identity{Noun: "coding-agent", Verb: "short"},
				Content:     "local",
				Description: "local prompt",
			}},
		},
	)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if len(effective) != 1 {
		t.Fatalf("len(effective) = %d, want 1", len(effective))
	}

	got := effective[0]
	if got.Kind != SourceLocal {
		t.Fatalf("Kind = %q, want %q", got.Kind, SourceLocal)
	}
	if got.Prompt.Content != "local" {
		t.Fatalf("Content = %q, want local", got.Prompt.Content)
	}
	if got.ShadowSummary() != "local overrides global, builtin" {
		t.Fatalf("ShadowSummary() = %q", got.ShadowSummary())
	}
}

func TestMergeUsesGlobalOverBuiltinWhenLocalAbsent(t *testing.T) {
	effective, err := Merge(
		Source{
			Kind:     SourceBuiltin,
			Location: "builtin",
			Prompts: []Prompt{{
				Identity: Identity{Noun: "coding-agent", Verb: "short"},
				Content:  "builtin",
			}},
		},
		Source{
			Kind:     SourceGlobal,
			Location: "global",
			Prompts: []Prompt{{
				Identity: Identity{Noun: "coding-agent", Verb: "short"},
				Content:  "global",
			}},
		},
	)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if len(effective) != 1 {
		t.Fatalf("len(effective) = %d, want 1", len(effective))
	}
	if effective[0].Kind != SourceGlobal {
		t.Fatalf("Kind = %q, want %q", effective[0].Kind, SourceGlobal)
	}
	if effective[0].Prompt.Content != "global" {
		t.Fatalf("Content = %q, want global", effective[0].Prompt.Content)
	}
	if effective[0].ShadowSummary() != "global overrides builtin" {
		t.Fatalf("ShadowSummary() = %q", effective[0].ShadowSummary())
	}
}

func TestMergeSortsEffectivePrompts(t *testing.T) {
	effective, err := Merge(Source{
		Kind:     SourceBuiltin,
		Location: "builtin",
		Prompts: []Prompt{
			{Identity: Identity{Noun: "workflow", Verb: "plan"}, Content: "workflow plan"},
			{Identity: Identity{Noun: "coding-agent", Verb: "short"}, Content: "coding short"},
			{Identity: Identity{Noun: "coding-agent", Verb: "long"}, Content: "coding long"},
		},
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	got := []string{
		effective[0].CommandName(),
		effective[1].CommandName(),
		effective[2].CommandName(),
	}
	want := []string{"coding-agent long", "coding-agent short", "workflow plan"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sorted command %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestMergeRejectsNormalizedCollisionsWithinSource(t *testing.T) {
	_, err := Merge(Source{
		Kind:     SourceBuiltin,
		Location: "builtin",
		Prompts: []Prompt{
			{Identity: Identity{Noun: "Coding Agent", Verb: "short"}, Content: "first"},
			{Identity: Identity{Noun: "coding-agent", Verb: "short"}, Content: "second"},
		},
	})
	if err == nil {
		t.Fatal("expected normalized collision to fail")
	}
	if !strings.Contains(err.Error(), "duplicate builtin prompt") {
		t.Fatalf("error = %q, want duplicate builtin prompt", err)
	}
}

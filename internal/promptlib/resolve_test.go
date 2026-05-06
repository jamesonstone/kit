package promptlib

import (
	"errors"
	"strings"
	"testing"
)

func TestResolveFindsNormalizedPrompt(t *testing.T) {
	effective, err := Merge(Source{
		Kind:     SourceBuiltin,
		Location: "builtin",
		Prompts: []Prompt{{
			Identity: Identity{Noun: "coding-agent", Verb: "short"},
			Content:  "prompt",
		}},
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	got, err := Resolve(effective, "Coding Agent", "SHORT")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Prompt.Content != "prompt" {
		t.Fatalf("Content = %q, want prompt", got.Prompt.Content)
	}
}

func TestResolveReturnsNoMatchWithSuggestions(t *testing.T) {
	effective, err := Merge(Source{
		Kind:     SourceBuiltin,
		Location: "builtin",
		Prompts: []Prompt{{
			Identity: Identity{Noun: "coding-agent", Verb: "short"},
			Content:  "short",
		}},
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	_, err = Resolve(effective, "coding", "short")
	if err == nil {
		t.Fatal("expected no-match error")
	}

	var noMatch NoMatchError
	if !errors.As(err, &noMatch) {
		t.Fatalf("error = %T, want NoMatchError", err)
	}
	if len(noMatch.Nouns) != 1 || noMatch.Nouns[0] != "coding-agent" {
		t.Fatalf("Nouns = %v, want [coding-agent]", noMatch.Nouns)
	}
}

func TestResolveNoMatchSuggestsNearestVerbs(t *testing.T) {
	effective, err := Merge(Source{
		Kind:     SourceBuiltin,
		Location: "builtin",
		Prompts: []Prompt{
			{Identity: Identity{Noun: "coding-agent", Verb: "instructions"}, Content: "instructions"},
			{Identity: Identity{Noun: "coding-agent", Verb: "long"}, Content: "long"},
			{Identity: Identity{Noun: "coding-agent", Verb: "short"}, Content: "short"},
		},
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	_, err = Resolve(effective, "coding-agent", "shrt")
	if err == nil {
		t.Fatal("expected no-match error")
	}
	if !strings.Contains(err.Error(), `nearest verbs for "coding-agent": short`) {
		t.Fatalf("error = %q, want nearest short suggestion", err)
	}
}

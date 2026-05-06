package promptlib

import (
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestSourceFromConfigConvertsPrompts(t *testing.T) {
	cfg := &config.Config{
		Prompts: map[string]map[string]config.Prompt{
			"coding-agent": {
				"short": {
					Content:     "clarify first",
					Description: "short prompt",
				},
			},
		},
	}

	source := SourceFromConfig(SourceGlobal, "global", cfg)
	if source.Kind != SourceGlobal {
		t.Fatalf("Kind = %q, want %q", source.Kind, SourceGlobal)
	}
	if len(source.Prompts) != 1 {
		t.Fatalf("len(Prompts) = %d, want 1", len(source.Prompts))
	}
	got := source.Prompts[0]
	if got.Identity != (Identity{Noun: "coding-agent", Verb: "short"}) {
		t.Fatalf("Identity = %#v", got.Identity)
	}
	if got.Content != "clarify first" {
		t.Fatalf("Content = %q, want clarify first", got.Content)
	}
	if got.Description != "short prompt" {
		t.Fatalf("Description = %q, want short prompt", got.Description)
	}
}

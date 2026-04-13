package config

import (
	"reflect"
	"testing"
)

func TestDefaultIncludesAllRepositoryInstructionFiles(t *testing.T) {
	cfg := Default()

	want := []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"}
	if !reflect.DeepEqual(cfg.Agents, want) {
		t.Fatalf("Default().Agents = %v, want %v", cfg.Agents, want)
	}
}

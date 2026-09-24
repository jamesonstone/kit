package instructions

import (
	"strings"
	"testing"
)

func TestV17RemovesThreadNamingAndInitialization(t *testing.T) {
	prior, err := AgentInstructions("v16")
	if err != nil {
		t.Fatal(err)
	}
	current, err := AgentInstructions("v17")
	if err != nil {
		t.Fatal(err)
	}
	if CurrentAgentVersion != "v17" {
		t.Fatal("v17 must be current")
	}
	const tail = "# Multi-agent orchestration evaluation gate"
	priorTail := strings.Index(prior, tail)
	currentTail := strings.Index(current, tail)
	if priorTail < 0 || currentTail < 0 {
		t.Fatal("both versions must keep the multi-agent gate")
	}
	if prior[priorTail:] != current[currentTail:] {
		t.Fatal("changed unrelated instructions")
	}
	if !strings.HasPrefix(current, "# Principles\n") {
		t.Fatal("v17 must start with principles")
	}
	for _, stale := range []string{
		"# Conversation naming and initialization",
		"kit instructions naming",
		"[scope] domain / objective",
		"before pinning",
		"thread-pin",
		"set_thread_title",
		"set_thread_pinned",
		"# Coding session initialization",
	} {
		if strings.Contains(current, stale) {
			t.Fatalf("v17 still contains retired thread lifecycle %q", stale)
		}
	}
}

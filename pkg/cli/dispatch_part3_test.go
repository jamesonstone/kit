package cli

import "testing"

func TestValidateDispatchMaxSubagents(t *testing.T) {
	if err := validateDispatchMaxSubagents(1); err != nil {
		t.Fatalf("expected positive max-subagents to be valid: %v", err)
	}
	if err := validateDispatchMaxSubagents(hardDispatchMaxSubagents); err != nil {
		t.Fatalf("expected hard ceiling max-subagents to be valid: %v", err)
	}

	if err := validateDispatchMaxSubagents(0); err == nil {
		t.Fatalf("expected max-subagents validation to fail for zero")
	}
	if err := validateDispatchMaxSubagents(hardDispatchMaxSubagents + 1); err == nil {
		t.Fatalf("expected max-subagents validation to fail above hard ceiling")
	}
}

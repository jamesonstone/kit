package cli

import "testing"

func TestFormatAgentInstructionBlock(t *testing.T) {
	block := formatAgentInstructionBlock("line one\nline two\n")
	want := "---\nline one\nline two\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

func TestFormatAgentInstructionBlockEmptyInput(t *testing.T) {
	block := formatAgentInstructionBlock("")
	want := "---\n\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

func TestFormatAgentInstructionBlockWhitespaceOnly(t *testing.T) {
	block := formatAgentInstructionBlock("   \t\n")
	want := "---\n   \t\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

func TestFormatAgentInstructionBlockMultipleNewlines(t *testing.T) {
	block := formatAgentInstructionBlock("a\n\nb\n")
	want := "---\na\n\nb\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

func TestFormatAgentInstructionBlockAddsTrailingNewline(t *testing.T) {
	block := formatAgentInstructionBlock("line one")
	want := "---\nline one\n---\n"

	if block != want {
		t.Fatalf("formatAgentInstructionBlock() = %q, want %q", block, want)
	}
}

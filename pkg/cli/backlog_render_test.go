package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/feature"
)

func TestPrintBacklogTableUsesFixedWidthLayout(t *testing.T) {
	out := &bytes.Buffer{}
	entries := []backlogEntry{
		{
			Feature: feature.Feature{Slug: "refactor-to-blobstore"},
			Description: "internal/models/artifact.go, internal/repositories/assistant/" +
				"artifacts.go, and internal/services/assistant/messages_prompt.go show " +
				"that active assistant artifacts are live markdown documents",
		},
	}

	if err := printBacklogTable(out, entries); err != nil {
		t.Fatalf("printBacklogTable() error = %v", err)
	}

	content := out.String()
	checks := []string{
		"Feature",
		"Description",
		"refactor-to-blobstore",
		"internal/models/artifact.go",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Fatalf("expected output to contain %q, got %q", check, content)
		}
	}
	if strings.Contains(content, "| feature | description |") {
		t.Fatalf("expected fixed-width rendering instead of markdown, got %q", content)
	}
}

func TestPrintBacklogTableUsesANSIColorWhenTerminalEnabled(t *testing.T) {
	previousCheck := terminalWriterCheck
	terminalWriterCheck = func(_ io.Writer) bool { return true }
	defer func() { terminalWriterCheck = previousCheck }()

	out := &bytes.Buffer{}
	entries := []backlogEntry{
		{
			Feature:     feature.Feature{Slug: "refactor-to-blobstore"},
			Description: "move active artifact editing to pointer-backed storage",
		},
	}

	if err := printBacklogTable(out, entries); err != nil {
		t.Fatalf("printBacklogTable() error = %v", err)
	}

	content := out.String()
	if !strings.Contains(content, "\033[") {
		t.Fatalf("expected ANSI color sequences in terminal output, got %q", content)
	}
	if !strings.Contains(content, "refactor-to-blobstore") {
		t.Fatalf("expected feature slug in colored output, got %q", content)
	}
}

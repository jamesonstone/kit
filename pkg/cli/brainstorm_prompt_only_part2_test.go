package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestOutputExistingBrainstormPrompt_RejectsOutputFile(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-sample", "BRAINSTORM.md"), `# BRAINSTORM

## SUMMARY

summary

## USER THESIS

thesis

## CODEBASE FINDINGS

findings

## AFFECTED FILES

files

## DEPENDENCIES

| Dependency | Type | Location | Used For | Status |
| ---------- | ---- | -------- | -------- | ------ |
| none | n/a | n/a | no phase dependencies recorded yet | active |

## QUESTIONS

questions

## OPTIONS

options

## RECOMMENDED STRATEGY

strategy

## NEXT STEP

next
`)

	cfg := config.Default()
	previousOutput := brainstormOutput
	brainstormOutput = filepath.Join(projectRoot, "prompt.txt")
	defer func() {
		brainstormOutput = previousOutput
	}()

	err := outputExistingBrainstormPrompt([]string{"sample"}, projectRoot, cfg, true)
	if err == nil || !strings.Contains(err.Error(), "--prompt-only cannot be used with --output") {
		t.Fatalf("expected --output rejection, got %v", err)
	}
	if _, statErr := os.Stat(brainstormOutput); !os.IsNotExist(statErr) {
		t.Fatalf("expected prompt output file to remain absent, got %v", statErr)
	}
}

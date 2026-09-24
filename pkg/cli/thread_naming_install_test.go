package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitDoesNotInstallThreadNamingPolicy(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	path := filepath.Join(tempDir, "docs", "references", "thread-naming.md")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("init installed retired thread-naming.md: %v", err)
	}
	agents := readFile(t, filepath.Join(tempDir, "AGENTS.md"))
	for _, stale := range []string{
		"## Conversation Naming",
		"kit instructions naming",
		"## Codex Thread Initialization Hard Gate",
		"set_thread_title",
		"set_thread_pinned",
	} {
		if strings.Contains(agents, stale) {
			t.Fatalf("init AGENTS.md still contains %q", stale)
		}
	}
}

func TestInstructionsRejectsNamingAndTitleSubcommands(t *testing.T) {
	for _, sub := range []string{"naming", "title"} {
		_, err := executeInstructionsCommand(sub)
		if err == nil {
			t.Fatalf("kit instructions %s succeeded", sub)
		}
	}
}

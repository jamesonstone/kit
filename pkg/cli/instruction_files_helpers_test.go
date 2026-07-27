package cli

import (
	"os"
	"slices"
	"testing"
)

func TestScaffoldAgentsCmd_UsesScaffoldNamespace(t *testing.T) {
	if scaffoldAgentsCmd.Use != "agents" {
		t.Fatalf("expected scaffold agents subcommand use, got %q", scaffoldAgentsCmd.Use)
	}
	if slices.Contains(scaffoldAgentsCmd.Aliases, "scaffold-agent") {
		t.Fatal("expected legacy scaffold-agent alias to be removed")
	}
}

func withScaffoldAgentFlags(t *testing.T, run func()) {
	t.Helper()

	originalForce := scaffoldAgentsForce
	originalCopilot := scaffoldAgentsCopilot
	originalClaude := scaffoldAgentsClaude
	originalAgentsMD := scaffoldAgentsAgentsMD
	originalYes := scaffoldAgentsYes
	originalAppendOnly := scaffoldAgentsAppendOnly
	originalVersion := scaffoldAgentsVersion

	t.Cleanup(func() {
		scaffoldAgentsForce = originalForce
		scaffoldAgentsCopilot = originalCopilot
		scaffoldAgentsClaude = originalClaude
		scaffoldAgentsAgentsMD = originalAgentsMD
		scaffoldAgentsYes = originalYes
		scaffoldAgentsAppendOnly = originalAppendOnly
		scaffoldAgentsVersion = originalVersion
	})

	scaffoldAgentsForce = false
	scaffoldAgentsCopilot = false
	scaffoldAgentsClaude = false
	scaffoldAgentsAgentsMD = false
	scaffoldAgentsYes = false
	scaffoldAgentsAppendOnly = false
	scaffoldAgentsVersion = 0

	run()
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertFileDoesNotExist(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to not exist, got err = %v", path, err)
	}
}

func setWorkingDirectory(t *testing.T, dir string) {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})
}

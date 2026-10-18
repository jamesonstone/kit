package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
	"github.com/jamesonstone/kit/v3/internal/worktreeprep"
)

func testSharedMemory() config.SharedMemoryConfig {
	return config.SharedMemoryConfig{Enabled: true, Provider: "notion", Destination: "https://example.com/memory", Guide: "https://example.com/guide"}
}

func TestSharedMemoryInstructionLifecycle(t *testing.T) {
	root := t.TempDir()
	cfg := defaultInitConfig()
	cfg.SharedMemory = testSharedMemory()
	mcp := filepath.Join(root, ".mcp.json")
	writeFile(t, mcp, "{\"mcpServers\":{\"existing\":{}}}\n")
	for _, path := range instructionFiles(cfg) {
		custom := "\n## Project guidance\n\nPreserve this text exactly.\n"
		original := templates.InstructionEntryFile(path) + custom
		writeFile(t, filepath.Join(root, path), original)
		for _, enabled := range []bool{false, true, true, false, false} {
			cfg.SharedMemory.Enabled = enabled
			plan, err := planInstructionArtifactWrite(root, path, instructionFileWriteModeConverge, false, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if plan.result != instructionFileSkipped {
				writeFile(t, plan.absolutePath, plan.content)
			}
			got := readFile(t, filepath.Join(root, path))
			if strings.Contains(got, sharedMemoryBegin) != enabled || !strings.Contains(got, custom) {
				t.Fatalf("enabled=%v: %s", enabled, got)
			}
			if enabled && (strings.Count(got, sharedMemoryBegin) != 1 || !strings.Contains(got, cfg.SharedMemory.Guide)) {
				t.Fatal("missing or duplicate guidance")
			}
			repeat, err := planInstructionArtifactWrite(root, path, instructionFileWriteModeConverge, false, cfg)
			if err != nil || repeat.result != instructionFileSkipped {
				t.Fatalf("not idempotent: %+v %v", repeat, err)
			}
		}
	}
	if got := readFile(t, mcp); got != "{\"mcpServers\":{\"existing\":{}}}\n" {
		t.Fatal("changed MCP config")
	}
}

func TestSharedMemoryMissingAndMalformedEntries(t *testing.T) {
	cfg := defaultInitConfig()
	cfg.SharedMemory = testSharedMemory()
	root := t.TempDir()
	changes, _, _, err := planRefreshInitInstructionArtifacts(root, initRefreshOptions{}, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range changes {
		if change.relativePath != templates.TestingReferencePath && !strings.Contains(change.after, sharedMemoryBegin) {
			t.Fatal(change.relativePath)
		}
		if _, err := os.Stat(filepath.Join(root, change.relativePath)); !os.IsNotExist(err) {
			t.Fatal("planner wrote files")
		}
	}
	for _, markers := range []string{sharedMemoryBegin, sharedMemoryEnd, sharedMemoryEnd + sharedMemoryBegin, sharedMemoryBegin + sharedMemoryBegin + sharedMemoryEnd} {
		path := filepath.Join(root, agentsMDPath)
		original := templates.InstructionEntryFile(agentsMDPath) + "\n" + markers + "\ncustom\n"
		writeFile(t, path, original)
		if _, err := planInstructionArtifactWrite(root, agentsMDPath, instructionFileWriteModeConverge, true, cfg); err == nil {
			t.Fatal("accepted ambiguous markers")
		}
		if readFile(t, path) != original {
			t.Fatal("mutated malformed entry")
		}
	}
}

func TestSharedMemoryConfigSet(t *testing.T) {
	root := t.TempDir()
	setWorkingDirectory(t, root)
	stubReconcileWorktreeInspection(t, worktreeprep.Location{})
	original := "# keep\nschema_version: 2\ncustom: {keep: yes}\nissues: {provider: github}\n"
	path := filepath.Join(root, config.ConfigFileName)
	writeFile(t, path, original)
	args := []string{"--shared-memory", "--shared-memory-destination", "https://example.com/memory", "--shared-memory-guide", "https://example.com/guide"}
	report, err := executeConfigSet(t, append(args, "--dry-run")...)
	if err != nil || !report.Changed || !report.SharedMemory.Enabled {
		t.Fatalf("%+v %v", report, err)
	}
	if readFile(t, path) != original {
		t.Fatal("dry-run wrote")
	}
	report, err = executeConfigSet(t, args...)
	if err != nil || !report.SharedMemory.Enabled {
		t.Fatalf("%+v %v", report, err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "# keep") || !strings.Contains(got, "custom: {keep: yes}") {
		t.Fatal(got)
	}
	report, err = executeConfigSet(t, args...)
	if err != nil || report.Changed {
		t.Fatalf("repeat: %+v %v", report, err)
	}
	report, err = executeConfigSet(t, "--shared-memory=false")
	if err != nil || report.SharedMemory.Enabled || report.SharedMemory.Guide == "" {
		t.Fatalf("disable: %+v %v", report, err)
	}
	before := readFile(t, path)
	if _, err := executeConfigSet(t, "--shared-memory", "--shared-memory-guide", "https://example.com/?token=secret"); err == nil {
		t.Fatal("accepted credential query")
	}
	if readFile(t, path) != before {
		t.Fatal("invalid config wrote")
	}
}

package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCapabilitiesDoesNotRequireProjectRootOrWriteFiles(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	paths := []string{
		".kit.yaml",
		filepath.Join(".kit", "state.json"),
		filepath.Join(".kit", "runs", "existing.json"),
		filepath.Join(".kit", "loops", "existing.json"),
		filepath.Join("docs", "specs", "sample", "TASKS.md"),
		filepath.Join("docs", "PROJECT_PROGRESS_SUMMARY.md"),
	}
	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte("before:"+path), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", path, err)
		}
	}
	before := snapshotFiles(t, tmp)

	for _, args := range [][]string{
		{"--json"},
		{"--full", "--json"},
		{"--search", "verify", "--json"},
		{"--json", "ci"},
		{"--json", "legacy", "verify"},
	} {
		if _, err := executeCapabilitiesCommand(args...); err != nil {
			t.Fatalf("kit capabilities %v error = %v", args, err)
		}
	}

	after := snapshotFiles(t, tmp)
	if len(after) != len(before) {
		t.Fatalf("file count changed: before=%d after=%d before=%v after=%v", len(before), len(after), before, after)
	}
	for path, beforeContent := range before {
		if after[path] != beforeContent {
			t.Fatalf("file %q changed: before %q after %q", path, beforeContent, after[path])
		}
	}
}

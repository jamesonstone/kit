package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func executeCapabilitiesCommand(args ...string) (string, error) {
	cmd := newCapabilitiesCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func findCompactCapability(records []capabilityCompactRecord, command string) *capabilityCompactRecord {
	for i := range records {
		if records[i].Command == command {
			return &records[i]
		}
	}
	return nil
}

func findDetailCapability(records []capabilityDetailRecord, command string) *capabilityDetailRecord {
	for i := range records {
		if records[i].Command == command {
			return &records[i]
		}
	}
	return nil
}

func findDetailedFlag(flags []capabilityFlag, name string) *capabilityFlag {
	for i := range flags {
		if flags[i].Name == name {
			return &flags[i]
		}
	}
	return nil
}

func snapshotFiles(t *testing.T, root string) map[string]string {
	t.Helper()

	files := map[string]string{}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[relative] = string(content)
		return nil
	}); err != nil {
		t.Fatalf("WalkDir(%q) error = %v", root, err)
	}
	return files
}

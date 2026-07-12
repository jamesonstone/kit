package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunInitCreatesCurrentConfigSchema(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)

	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	_, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if inspection.SchemaState != config.SchemaStateCurrent {
		t.Fatalf("SchemaState = %q, want current", inspection.SchemaState)
	}
}

func TestRunInitRefreshMigratesUnversionedConfig(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	content := "goal_percentage: 95\ninstruction_scaffold_version: 2\n"
	if err := os.WriteFile(filepath.Join(root, config.ConfigFileName), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{config.ConfigFileName}
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	_, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if inspection.SchemaState != config.SchemaStateCurrent {
		t.Fatalf("SchemaState = %q, want current", inspection.SchemaState)
	}
}

func TestRunInitRefreshForcePreservesAWSContext(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.DefaultInstructionScaffoldVersion
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "012345678901"}
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initForce = true
		initOutputOnly = true
		initRefreshFiles = []string{config.ConfigFileName}
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	updated, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if updated.AWS == nil || updated.AWS.Profile != "dev" || updated.AWS.AccountID != "012345678901" {
		t.Fatalf("AWS = %#v, want preserved dev context", updated.AWS)
	}
}

func TestRunInitRejectsNewerConfigWithoutWriting(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	path := filepath.Join(root, config.ConfigFileName)
	content := []byte("schema_version: 2\ngoal_percentage: automatic\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true
		err := runInit(initCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "upgrade Kit") {
			t.Fatalf("runInit() error = %v, want upgrade guidance", err)
		}
	})
	assertFileContent(t, path, content)
}

func TestRunInitRefreshRejectsNewerConfigWithoutWriting(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	path := filepath.Join(root, config.ConfigFileName)
	content := []byte("schema_version: 2\ngoal_percentage: automatic\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{config.ConfigFileName}
		err := runInit(initCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "upgrade Kit") {
			t.Fatalf("runInit() error = %v, want upgrade guidance", err)
		}
	})
	assertFileContent(t, path, content)
}

func assertFileContent(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("file changed:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

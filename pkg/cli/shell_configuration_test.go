package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/spf13/cobra"
)

func TestConfigCheckReportsShellSourceWithoutExecutingOrEchoingArguments(t *testing.T) {
	for _, command := range [][]string{nil, {"/not-an-installed-shell", "private-argument"}, {""}} {
		root := t.TempDir()
		cfg := config.Default()
		cfg.Shell.Command = command
		if err := config.Save(root, cfg); err != nil {
			t.Fatal(err)
		}
		loaded, inspection, err := config.LoadWithInspection(root)
		if err != nil {
			t.Fatal(err)
		}
		report := buildConfigCheckReport(loaded, inspection)
		want := "user"
		if len(command) != 0 {
			want = "command"
		}
		if report.Shell != want || report.Valid != (cfg.Shell.Validate() == nil) {
			t.Fatalf("unexpected report: %+v", report)
		}
		var out bytes.Buffer
		printConfigCheckReport(&out, report)
		data, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		for _, output := range []string{out.String(), string(data)} {
			if strings.Contains(output, "private-argument") || strings.Contains(output, "/not-an-installed-shell") {
				t.Fatal("config report leaked command arguments")
			}
		}
		if !strings.Contains(out.String(), "resolved by execution host") {
			t.Fatal("config report claimed live shell resolution")
		}
	}
}

func TestShellConfigurationSurvivesReconcile(t *testing.T) {
	setupMigrationEnvironment(t)
	for _, generation := range append([]string{"current"}, migrationFixtures...) {
		t.Run(generation, func(t *testing.T) {
			var root string
			if generation == "current" {
				root = freshInitProject(t)
			} else {
				root = copyMigrationFixture(t, generation, nil)
			}
			file := filepath.Join(root, config.ConfigFileName)
			before := readFile(t, file)
			writeFile(t, file, before+"\nshell:\n  command: [zsh, -lic]\n")
			commitMigrationFixture(t, root)
			migrateProject(t, root, false)
			cfg := loadMigratedConfig(t, root)
			if !slices.Equal(cfg.Shell.Command, []string{"zsh", "-lic"}) {
				t.Fatalf("reconcile lost shell override: %+v", cfg.Shell)
			}
			if changes := plannedChanges(t, root); len(changes) != 0 {
				t.Fatalf("second reconcile is not idempotent: %v", changes)
			}
		})
	}
}

func TestConfigCheckShellJSONIsReadOnly(t *testing.T) {
	previous := configCheckJSON
	configCheckJSON = true
	t.Cleanup(func() { configCheckJSON = previous })
	for _, command := range [][]string{nil, {"zsh", "-lic"}, {""}} {
		root := t.TempDir()
		cfg := config.Default()
		cfg.Shell.Command = command
		if err := config.Save(root, cfg); err != nil {
			t.Fatal(err)
		}
		setWorkingDirectory(t, root)
		path := filepath.Join(root, config.ConfigFileName)
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		cmd := &cobra.Command{}
		cmd.SetOut(&out)
		err = runConfigCheck(cmd, nil)
		if (err != nil) != (cfg.Shell.Validate() != nil) {
			t.Fatalf("config check error = %v", err)
		}
		var report configCheckReport
		if err := json.Unmarshal(out.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if report.Shell != cfg.Shell.Source() {
			t.Fatalf("source = %q, want %q", report.Shell, cfg.Shell.Source())
		}
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("config check changed the file: %v", err)
		}
	}
}

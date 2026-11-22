package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/spf13/cobra"
)

func TestConfigCheckWarnsAboutLegacyGlobalConfig(t *testing.T) {
	home := setupInitHome(t)
	project := t.TempDir()
	setWorkingDirectory(t, project)
	legacy := filepath.Join(home, ".kit", config.ConfigFileName)
	canonical := filepath.Join(home, ".config", "kit", config.ConfigFileName)
	writeFile(t, filepath.Join(project, config.ConfigFileName), "schema_version: 2\n")
	writeFile(t, canonical, "specs_dir: docs/specs\n")
	const legacyContents = "source_file_line_limit: 280\n"
	writeFile(t, legacy, legacyContents)

	original := configCheckJSON
	t.Cleanup(func() { configCheckJSON = original })
	configCheckJSON = true
	out := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(out)
	if err := runConfigCheck(cmd, nil); err != nil {
		t.Fatal(err)
	}
	var report configCheckReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if !report.Valid || len(report.Findings) != 1 {
		t.Fatalf("expected one nonblocking finding, got %+v", report)
	}
	warning := report.Findings[0]
	if warning.Field != "global_config" || warning.Severity != config.FindingWarning ||
		!strings.Contains(warning.Message, legacy) || !strings.Contains(warning.Message, canonical) {
		t.Fatalf("unexpected warning: %+v", warning)
	}
	if got := readFile(t, legacy); got != legacyContents || readFile(t, canonical) != "specs_dir: docs/specs\n" {
		t.Fatal("config check rewrote a global file")
	}

	if err := os.Rename(legacy, legacy+".bak"); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := runConfigCheck(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out.Bytes(), &report); err != nil || !report.Valid || len(report.Findings) != 0 {
		t.Fatalf("warning remained after migration: %+v, %v", report, err)
	}
}

func TestInitPreflightWarnsAboutLegacyConfigOutsideProject(t *testing.T) {
	home := setupInitHome(t)
	setWorkingDirectory(t, t.TempDir())
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	legacy := filepath.Join(home, ".kit", config.ConfigFileName)
	oldDefault := filepath.Join(home, ".config", "kit", config.ConfigFileName)
	writeFile(t, legacy, "{}\n")
	writeFile(t, oldDefault, "{}\n")

	root := &cobra.Command{Use: "kit", PersistentPreRunE: runAutomaticConfigCheck}
	root.AddCommand(&cobra.Command{Use: "init", Run: func(*cobra.Command, []string) {}})
	errOut := new(bytes.Buffer)
	root.SetErr(errOut)
	root.SetArgs([]string{"init"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got := errOut.String()
	if strings.Count(got, "Warning:") != 2 || !strings.Contains(got, legacy) ||
		!strings.Contains(got, oldDefault) || !strings.Contains(got, filepath.Join(xdg, "kit", config.ConfigFileName)) {
		t.Fatalf("unexpected stderr: %q", got)
	}
}

package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	reflectpkg "reflect"
	"sort"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/commandset"
	"github.com/jamesonstone/kit/v3/internal/guide"
	"github.com/spf13/cobra"
)

func TestGuideMatchesInstalledTree(t *testing.T) {
	rootCmd.InitDefaultCompletionCmd()
	pruneCommandTree(rootCmd, "")
	d := guide.Build(rootCmd)
	var got []string
	for _, c := range d.Commands {
		if c.Path == "kit" || c.Path == "kit help" {
			continue
		}
		got = append(got, strings.TrimPrefix(c.Path, "kit "))
	}
	sort.Strings(got)
	if want := commandPaths(rootCmd, ""); !reflectpkg.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
	for _, retired := range []string{"capabilities", "context", "dispatch", "plan", "instructions"} {
		for _, c := range got {
			if c == retired || strings.HasPrefix(c, retired+" ") {
				t.Errorf("retired command %s", c)
			}
		}
	}
	if commandset.IsTelemetryPath("guide") {
		t.Fatal("guide must not record usage")
	}
}

func TestGuideOutsideProjectNeverRunsPreflightOrWrites(t *testing.T) {
	for _, invalidConfig := range []bool{false, true} {
		t.Run(map[bool]string{false: "outside", true: "invalid config"}[invalidConfig], func(t *testing.T) {
			dir := t.TempDir()
			setWorkingDirectory(t, dir)
			t.Setenv("HOME", dir)
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
			t.Setenv("XDG_STATE_HOME", filepath.Join(dir, "state"))
			if invalidConfig {
				if err := os.WriteFile(".kit.yaml", []byte("invalid: ["), 0600); err != nil {
					t.Fatal(err)
				}
			}
			root := &cobra.Command{Use: "kit", PersistentPreRunE: func(*cobra.Command, []string) error { t.Fatal("preflight executed"); return nil }}
			root.AddCommand(newGuideCommand())
			var b bytes.Buffer
			root.SetOut(&b)
			root.SetErr(&b)
			root.SetArgs([]string{"guide", "--json"})
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			var d guide.Document
			if err := json.Unmarshal(b.Bytes(), &d); err != nil {
				t.Fatal(err)
			}
			if d.SchemaVersion != 1 || len(d.Commands) < 2 {
				t.Fatal("empty guide")
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			want := 0
			if invalidConfig {
				want = 1
				data, err := os.ReadFile(".kit.yaml")
				if err != nil || string(data) != "invalid: [" {
					t.Fatal("config changed")
				}
			}
			if len(entries) != want {
				t.Fatalf("unexpected writes: %v", entries)
			}
		})
	}
}

func TestGuideDiscoveryRoutes(t *testing.T) {
	var out bytes.Buffer
	previous := rootCmd.OutOrStdout()
	rootCmd.SetOut(&out)
	t.Cleanup(func() { rootCmd.SetOut(previous) })
	if err := renderRootHelp(rootCmd); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "guide") {
		t.Fatal("root help missing guide")
	}
	if !strings.Contains(buildProjectInitPrompt("/repo", "/repo/docs/CONSTITUTION.md"), "kit guide") {
		t.Fatal("onboarding missing guide")
	}
	for _, path := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md", "internal/templates/universal_contract.md"} {
		data, err := os.ReadFile(filepath.Join("../..", path))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "kit guide") {
			t.Errorf("%s missing route", path)
		}
	}
}

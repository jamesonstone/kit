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

func TestGuideRealRootBypassesConfigPreflight(t *testing.T) {
	dir := t.TempDir()
	setWorkingDirectory(t, dir)
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(dir, "state"))
	if err := os.WriteFile(".kit.yaml", []byte("invalid: ["), 0600); err != nil {
		t.Fatal(err)
	}
	if rootCmd.PersistentPreRunE == nil {
		t.Fatal("real root has no config preflight")
	}
	guideCmd, _, err := rootCmd.Find([]string{"guide"})
	if err != nil || guideCmd.PersistentPreRunE == nil {
		t.Fatal("guide must override the root preflight")
	}
	var b bytes.Buffer
	previousOut, previousErr := rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()
	rootCmd.SetOut(&b)
	rootCmd.SetErr(&b)
	rootCmd.SetArgs([]string{"guide", "--json"})
	t.Cleanup(func() {
		rootCmd.SetOut(previousOut)
		rootCmd.SetErr(previousErr)
		rootCmd.SetArgs(nil)
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(".kit.yaml")
	if err != nil || string(data) != "invalid: [" {
		t.Fatal("preflight changed config")
	}
	if !strings.Contains(b.String(), `"path": "kit guide"`) || !strings.Contains(b.String(), `"path": "kit inspect"`) {
		t.Fatalf("real root guide missing live commands:\n%s", b.String())
	}
}

func TestGuideRetrievesExactVisibleCommand(t *testing.T) {
	run := func(args ...string) (string, error) {
		root := &cobra.Command{Use: "kit"}
		group := &cobra.Command{Use: "future", Short: "Manage future tasks"}
		leaf := &cobra.Command{
			Use:     "show <target>",
			Short:   "Show a target",
			Long:    "Display target details",
			Example: "kit future show example",
			Run:     func(*cobra.Command, []string) { t.Fatal("guide executed the command") },
		}
		leaf.Flags().String("format", "text", "Select output format")
		group.AddCommand(leaf, &cobra.Command{
			Use:    "hidden",
			Hidden: true,
			Run:    func(*cobra.Command, []string) { t.Fatal("guide executed the command") },
		})
		root.AddCommand(newGuideCommand(), group)
		var b bytes.Buffer
		root.SetOut(&b)
		root.SetErr(&b)
		root.SetArgs(append([]string{"guide"}, args...))
		err := root.Execute()
		return b.String(), err
	}

	out, err := run("future", "show")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# kit future show", "Display target details", "kit future show <target>", "kit future show example", "--format"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "Manage future tasks") || strings.Contains(out, "Repository artifacts") {
		t.Fatalf("focused output includes unrelated guidance: %s", out)
	}

	out, err = run("kit", "future", "show", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var d guide.Document
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		t.Fatal(err)
	}
	if d.SchemaVersion != 1 || len(d.Artifacts) != 0 || len(d.Commands) != 1 || d.Commands[0].Path != "kit future show" {
		t.Fatalf("unexpected scoped guide: %+v", d)
	}

	for _, name := range []string{"sh", "hidden"} {
		_, err := run("future", name)
		if err == nil || !strings.Contains(err.Error(), "Run kit guide to list command paths") {
			t.Errorf("expected actionable error for %q, got %v", name, err)
		}
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
	if !strings.Contains(out.String(), "Coding agents: run `kit guide` to choose a command") {
		t.Fatal("root help does not point agents at kit guide")
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

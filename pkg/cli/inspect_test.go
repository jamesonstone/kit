package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/commandset"
	"github.com/jamesonstone/kit/v3/internal/guidance"
	"github.com/spf13/cobra"
)

func TestInspectCommandExplicitPathBypassesCwdRepair(t *testing.T) {
	cwd := t.TempDir()
	target := t.TempDir()
	bad := []byte("schema_version: [")
	if err := os.WriteFile(filepath.Join(cwd, ".kit.yaml"), bad, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, ".kit.yaml"), []byte("schema_version: 2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)
	root := &cobra.Command{Use: "kit", PersistentPreRunE: runAutomaticConfigCheck}
	cmd := newInspectCommand()
	root.AddCommand(cmd)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"inspect", "--path", target, "--json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var report guidance.Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("%v: %s", err, out.String())
	}
	want, _ := filepath.EvalSymlinks(target)
	if report.ProjectRoot != want {
		t.Fatalf("wrong root %s", report.ProjectRoot)
	}
	if strings.Contains(out.String(), `"findings": null`) || !strings.Contains(out.String(), `"findings": []`) {
		t.Fatalf("empty findings must be an array:\n%s", out.String())
	}
	after, _ := os.ReadFile(filepath.Join(cwd, ".kit.yaml"))
	if !bytes.Equal(after, bad) {
		t.Fatal("cwd config repaired")
	}
	entries, _ := os.ReadDir(target)
	if len(entries) != 1 {
		t.Fatal("inspect wrote project files")
	}
	if commandset.IsTelemetryPath("inspect") {
		t.Fatal("inspect must not record cwd telemetry")
	}
}
func TestInspectCommandSurfaceAndErrors(t *testing.T) {
	if !commandset.IsProtectedOrParent("inspect") {
		t.Fatal("inspect pruned")
	}
	if findVisibleSubcommand(rootCmd, "inspect") == nil {
		t.Fatal("inspect not registered")
	}
	found := false
	for _, s := range rootCommandSections {
		for _, c := range s.commands {
			if c == "inspect" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("inspect missing from root help")
	}
	cmd := newInspectCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--path", t.TempDir(), "--json"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("non-project succeeded")
	}
	var report guidance.Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Errors) != 1 || !strings.Contains(report.Errors[0], "not a kit project") {
		t.Fatalf("%+v", report)
	}
}

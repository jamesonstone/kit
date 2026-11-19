package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/plugin"
	"github.com/jamesonstone/kit/v3/internal/templates"
	"github.com/jamesonstone/kit/v3/internal/worktreeprep"
	"github.com/spf13/pflag"
)

func writeFixturePlugin(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, plugin.RootDir, "fixture")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := "schema_version: 1\nid: fixture\ninstructions: instructions.md\nconfig: config.yaml\nname: ignored\n"
	writeFile(t, filepath.Join(dir, plugin.ManifestName), manifest)
	writeFile(t, filepath.Join(dir, "instructions.md"), body)
}

func TestReconcilePluginBlockAddUpdateAndRemove(t *testing.T) {
	setupMigrationEnvironment(t)
	root := freshInitProject(t)
	stubReconcileWorktreeInspection(t, worktreeprep.Location{})
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range instructionFiles(cfg) {
		got := readFile(t, filepath.Join(root, path))
		if got != templates.InstructionEntryFile(path) {
			t.Fatalf("%s changed before any plugin", path)
		}
	}
	runManagedReconcileForWorktreeTest(t)
	for _, path := range instructionFiles(cfg) {
		if readFile(t, filepath.Join(root, path)) != templates.InstructionEntryFile(path) {
			t.Fatalf("reconcile changed %s with no plugin", path)
		}
	}

	agents := filepath.Join(root, "AGENTS.md")
	writeFile(t, agents, readFile(t, agents)+"\n## Team\n\nkeep  \n")
	before := map[string]string{}
	for _, path := range instructionFiles(cfg) {
		before[path] = readFile(t, filepath.Join(root, path))
	}
	runManagedReconcileForWorktreeTest(t)
	for path, content := range before {
		if readFile(t, filepath.Join(root, path)) != content {
			t.Fatalf("no-plugin reconcile changed %s", path)
		}
	}

	writeFixturePlugin(t, root, "Remember the roster.\n")
	report, err := executeConfigSet(t, "--plugin", "fixture", "--enable")
	if err != nil || !report.Changed || report.Plugin != "fixture" || report.PluginEnabled == nil || !*report.PluginEnabled {
		t.Fatalf("%#v %v", report, err)
	}
	if !strings.Contains(readFile(t, filepath.Join(root, config.ConfigFileName)), "fixture") {
		t.Fatal("config did not enable the plugin")
	}
	runManagedReconcileForWorktreeTest(t)
	for _, path := range instructionFiles(cfg) {
		content := readFile(t, filepath.Join(root, path))
		for _, want := range []string{
			templates.UniversalContractBlock(),
			"<!-- BEGIN KIT-PLUGIN fixture -->",
			"<!-- Plugin config: .kit/plugins/fixture/config.yaml -->",
			"Remember the roster.",
			"<!-- END KIT-PLUGIN fixture -->",
		} {
			if !strings.Contains(content, want) {
				t.Fatalf("%s missing %q\n%s", path, want, content)
			}
		}
		if strings.Count(content, templates.UniversalContractBeginMarker) != 1 {
			t.Fatalf("%s contract count", path)
		}
	}
	if !strings.Contains(readFile(t, agents), "## Team\n\nkeep  \n") {
		t.Fatal("project text was dropped")
	}
	enabled := map[string]string{}
	for _, path := range instructionFiles(cfg) {
		enabled[path] = readFile(t, filepath.Join(root, path))
	}
	runManagedReconcileForWorktreeTest(t)
	for path, content := range enabled {
		if readFile(t, filepath.Join(root, path)) != content {
			t.Fatalf("second reconcile changed %s", path)
		}
	}

	writeFixturePlugin(t, root, "Updated roster.\n")
	runManagedReconcileForWorktreeTest(t)
	updated := readFile(t, agents)
	if strings.Contains(updated, "Remember the roster.") || !strings.Contains(updated, "Updated roster.") {
		t.Fatalf("update:\n%s", updated)
	}

	report, err = executeConfigSet(t, "--plugin", "fixture", "--disable")
	if err != nil || !report.Changed || report.PluginEnabled == nil || *report.PluginEnabled {
		t.Fatalf("disable %#v %v", report, err)
	}
	if strings.Contains(readFile(t, filepath.Join(root, config.ConfigFileName)), "plugins:") {
		t.Fatal("disabled plugin left a plugins key")
	}
	runManagedReconcileForWorktreeTest(t)
	for path, content := range before {
		got := readFile(t, filepath.Join(root, path))
		if got != content || strings.Contains(got, "KIT-PLUGIN") {
			t.Fatalf("disable left a trace in %s\n%s", path, got)
		}
	}
}

func TestConfigSetPluginPreservesIssueSettings(t *testing.T) {
	root := t.TempDir()
	setWorkingDirectory(t, root)
	stubReconcileWorktreeInspection(t, worktreeprep.Location{})
	original := "# keep comment\nschema_version: 2\nissues:\n  provider: linear\n  linear:\n    workspace: acme\n    team_id: 11111111-1111-4111-8111-111111111111\n"
	writeFile(t, filepath.Join(root, config.ConfigFileName), original)
	if _, err := executeConfigSet(t, "--plugin", "fixture", "--enable"); err == nil {
		t.Fatal("enabled a missing plugin")
	}
	if readFile(t, filepath.Join(root, config.ConfigFileName)) != original {
		t.Fatal("failed enable wrote config")
	}
	writeFixturePlugin(t, root, "Roster.\n")
	if _, err := executeConfigSet(t, "--plugin", "fixture", "--enable", "--dry-run"); err != nil {
		t.Fatal(err)
	}
	if readFile(t, filepath.Join(root, config.ConfigFileName)) != original {
		t.Fatal("dry run wrote")
	}
	if _, err := executeConfigSet(t, "--plugin", "fixture", "--enable", "--issue-provider", "github"); err == nil {
		t.Fatal("mixed plugin and issue flags")
	}
	report, err := executeConfigSet(t, "--plugin", "fixture", "--enable")
	if err != nil || report.Issues.EffectiveProvider() != "linear" || report.Issues.Linear.Workspace != "acme" {
		t.Fatalf("%#v %v", report, err)
	}
	text := readFile(t, filepath.Join(root, config.ConfigFileName))
	if !strings.Contains(text, "# keep comment") || !strings.Contains(text, "workspace: acme") || !strings.Contains(text, "fixture") {
		t.Fatalf("config:\n%s", text)
	}
	report, err = executeConfigSet(t, "--plugin", "fixture", "--enable")
	if err != nil || report.Changed {
		t.Fatalf("repeat %#v %v", report, err)
	}
}

func TestInitPluginWritesBlock(t *testing.T) {
	setupMigrationEnvironment(t)
	root := t.TempDir()
	setWorkingDirectory(t, root)
	writeFixturePlugin(t, root, "Init roster.\n")
	setInitPlugin(t, "fixture")
	withInitFlags(t, func() {
		initOutputOnly = true
		if err := runInitForTest(initCmd, nil); err != nil {
			t.Fatal(err)
		}
	})
	cfg, err := config.Load(root)
	if err != nil || len(cfg.Plugins.Enabled) != 1 || cfg.Plugins.Enabled[0] != "fixture" {
		t.Fatalf("%#v %v", cfg.Plugins, err)
	}
	content := readFile(t, filepath.Join(root, "AGENTS.md"))
	if !strings.Contains(content, "Init roster.") || !strings.Contains(content, "<!-- Plugin config: .kit/plugins/fixture/config.yaml -->") {
		t.Fatalf("init block:\n%s", content)
	}
	if !strings.Contains(content, templates.UniversalContractBlock()) {
		t.Fatal("init dropped the contract")
	}
}

func TestIssueConfigSetJSONOmitsPlugins(t *testing.T) {
	root := t.TempDir()
	setWorkingDirectory(t, root)
	stubReconcileWorktreeInspection(t, worktreeprep.Location{})
	writeFile(t, filepath.Join(root, config.ConfigFileName), "schema_version: 2\n")
	cmd := newConfigSetCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--issue-provider", "github", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "plugins") {
		t.Fatalf("issue JSON included plugins: %s", out.String())
	}
}

func TestCopyMissingPluginFilesSkipsExistingManifest(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	writeFixturePlugin(t, from, "from roster\n")
	writeFixturePlugin(t, to, "keep roster\n")
	if err := copyMissingPluginFiles(from, to); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readFile(t, filepath.Join(to, plugin.RootDir, "fixture", "instructions.md")), "keep roster") {
		t.Fatal("existing plugin was overwritten")
	}
	empty := t.TempDir()
	if err := copyMissingPluginFiles(from, empty); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readFile(t, filepath.Join(empty, plugin.RootDir, "fixture", plugin.ManifestName)), "id: fixture") {
		t.Fatal("missing plugin was not copied")
	}
}

func setInitPlugin(t *testing.T, id string) {
	t.Helper()
	value := initCmd.Flags().Lookup("plugin").Value.(pflag.SliceValue)
	previous := append([]string{}, value.GetSlice()...)
	if err := value.Replace([]string{id}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := value.Replace(previous); err != nil {
			t.Errorf("restore plugin flag: %v", err)
		}
	})
}

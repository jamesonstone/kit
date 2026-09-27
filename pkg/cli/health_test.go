package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestRunHealthExplicitOptOutSkipsNetworkAndWrites(t *testing.T) {
	projectRoot := t.TempDir()
	managed := false
	cfg := config.Default()
	cfg.Health = &config.HealthConfig{Managed: &managed}
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	setWorkingDirectory(t, projectRoot)
	before, err := os.ReadFile(filepath.Join(projectRoot, config.ConfigFileName))
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	registryCalls := 0
	stubRulesetRegistryFunc(t, func(_ context.Context) ([]registryRuleset, error) {
		registryCalls++
		return nil, errors.New("registry should not be called")
	})

	cmd := healthCommandForTest(t, "--json")
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runHealth(cmd, nil); err != nil {
		t.Fatalf("runHealth() error = %v", err)
	}
	if registryCalls != 0 {
		t.Fatalf("registry calls = %d, want 0", registryCalls)
	}
	after, err := os.ReadFile(filepath.Join(projectRoot, config.ConfigFileName))
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("opted-out config changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}

	var report healthReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.State != statusKitManagedStateDisabled || report.Managed || report.ProjectCheck != "not_run" {
		t.Fatalf("report = %#v, want disabled no-op", report)
	}
}

// Health is diagnostic: it reports what reconcile would change and never
// writes, even when changes are pending.
func TestRunHealthReportsPendingChangesWithoutWriting(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	writeInitScaffoldArtifacts(t, projectRoot)
	setWorkingDirectory(t, projectRoot)
	ruleset := registryRulesetForTest("sample-guardrails", []string{"git"})
	stubRulesetRegistry(t, ruleset)
	target := filepath.Join(projectRoot, rulesetTarget(ruleset.Slug))
	configBefore := readFile(t, filepath.Join(projectRoot, config.ConfigFileName))

	for _, flags := range [][]string{{"--json"}, {"--dry-run", "--json"}} {
		cmd := healthCommandForTest(t, flags...)
		out := &strings.Builder{}
		cmd.SetOut(out)
		if err := runHealth(cmd, nil); err != nil {
			t.Fatalf("runHealth(%v) error = %v", flags, err)
		}
		if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("health wrote %s", target)
		}
		if readFile(t, filepath.Join(projectRoot, config.ConfigFileName)) != configBefore {
			t.Fatal("health changed .kit.yaml")
		}
		var report healthReport
		if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if report.State != statusKitManagedStateRefreshAvailable || len(report.Files) == 0 || report.ProjectCheck != "passed" ||
			!strings.Contains(strings.Join(report.NextActions, " "), "kit reconcile") {
			t.Fatalf("report = %#v, want pending changes that point to kit reconcile", report)
		}
	}
}

func TestRunHealthIsCurrentAfterReconcile(t *testing.T) {
	setupMigrationEnvironment(t)
	freshInitProject(t)
	cmd := healthCommandForTest(t, "--json")
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runHealth(cmd, nil); err != nil {
		t.Fatalf("runHealth() error = %v", err)
	}
	var report healthReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatal(err)
	}
	if report.State != statusKitManagedStateCurrent || len(report.Files) != 0 {
		t.Fatalf("fresh project health = %#v", report)
	}
}

func TestRunHealthRegistryFailureIsReadOnlyUnknown(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	setWorkingDirectory(t, projectRoot)
	configPath := filepath.Join(projectRoot, config.ConfigFileName)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	stubRulesetRegistryError(t, errors.New("registry offline"))

	cmd := healthCommandForTest(t, "--json")
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runHealth(cmd, nil); err != nil {
		t.Fatalf("runHealth() error = %v", err)
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("config changed while registry was unavailable")
	}

	var report healthReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.State != statusKitManagedStateUnknown || !strings.Contains(report.CheckError, "registry offline") {
		t.Fatalf("report = %#v, want unknown registry state", report)
	}
}

func TestHealthAndRegistryCommandsSkipAutomaticConfigPreflight(t *testing.T) {
	for _, cmd := range []*cobra.Command{healthCmd, registryStatusCmd} {
		if !skipAutomaticConfigCheck(cmd) {
			t.Fatalf("skipAutomaticConfigCheck(%q) = false, want true", cmd.CommandPath())
		}
	}
}

func TestCheckProjectContractToPropagatesWriterFailure(t *testing.T) {
	want := errors.New("write failed")
	if err := checkProjectContractTo(errorWriter{err: want}, "", nil); !errors.Is(err, want) {
		t.Fatalf("checkProjectContractTo() error = %v, want %v", err, want)
	}
}

func TestRunHealthValidatesFlags(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	setWorkingDirectory(t, projectRoot)

	for _, tt := range []struct {
		name  string
		flags []string
		want  string
	}{
		{name: "diff conflicts with json", flags: []string{"--diff", "--json"}, want: "--diff cannot be combined with --json"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := healthCommandForTest(t, tt.flags...)
			if err := runHealth(cmd, nil); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("runHealth() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func healthCommandForTest(t *testing.T, flags ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("diff", false, "")
	cmd.Flags().Bool("json", false, "")
	cmd.SetContext(context.Background())
	for _, name := range flags {
		if err := cmd.Flags().Set(strings.TrimPrefix(name, "--"), "true"); err != nil {
			t.Fatalf("Flags().Set(%s) error = %v", name, err)
		}
	}
	return cmd
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

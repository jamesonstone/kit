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

	"github.com/jamesonstone/kit/internal/config"
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

func TestRunHealthDryRunPlansWithoutWriting(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	setWorkingDirectory(t, projectRoot)
	ruleset := registryRulesetForTest("safety-guardrails", []string{"git"})
	stubRulesetRegistry(t, ruleset)
	target := filepath.Join(projectRoot, rulesetTarget(ruleset.Slug))

	cmd := healthCommandForTest(t, "--dry-run", "--json")
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runHealth(cmd, nil); err != nil {
		t.Fatalf("runHealth() error = %v", err)
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry-run target stat error = %v, want not exist", err)
	}

	var report healthReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.State != statusKitManagedStateRefreshAvailable || len(report.Files) == 0 {
		t.Fatalf("report = %#v, want planned refresh", report)
	}
}

func TestRunHealthAppliesSafeRegistryUpdateAndChecksProject(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	setWorkingDirectory(t, projectRoot)
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("", ""))
	ruleset := registryRulesetForTest("safety-guardrails", []string{"git"})
	stubRulesetRegistry(t, ruleset)

	cmd := healthCommandForTest(t, "--json")
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runHealth(cmd, nil); err != nil {
		t.Fatalf("runHealth() error = %v\noutput: %s", err, out.String())
	}
	content, err := os.ReadFile(filepath.Join(projectRoot, rulesetTarget(ruleset.Slug)))
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(content) != ruleset.Content {
		t.Fatalf("ruleset content mismatch:\n%s", content)
	}

	var report healthReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, out.String())
	}
	if report.State != healthStateUpdated || report.ProjectCheck != "passed" || report.RegistryState != statusKitManagedStateCurrent {
		t.Fatalf("report = %#v, want updated and healthy", report)
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

func TestRunHealthPreservesConflictedRulesetAndReportsAttention(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	setWorkingDirectory(t, projectRoot)
	writeFile(t, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("", ""))
	base := registryRulesetForTest("github-pr-delivery", []string{"github"})
	local := strings.Replace(base.Content, "## Rules", "- Local rule change.\n\n## Rules", 1)
	remoteContent := strings.Replace(base.Content, "## Rules", "- Remote rule change.\n\n## Rules", 1)
	remote := registryRulesetWithContentForTest(base.Slug, remoteContent, "new-commit")
	recordRulesetRegistryState(cfg, base, registryArtifactStateManaged, base.NormalizedHash, base.Content)
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	target := filepath.Join(projectRoot, rulesetTarget(base.Slug))
	writeFile(t, target, local)
	stubRulesetRegistry(t, remote)
	stubRulesetRegistryContent(t, map[string]string{base.SourceCommit: base.Content})

	cmd := healthCommandForTest(t, "--json")
	out := &strings.Builder{}
	cmd.SetOut(out)
	if err := runHealth(cmd, nil); err != nil {
		t.Fatalf("runHealth() error = %v\noutput: %s", err, out.String())
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(after) != local {
		t.Fatalf("conflicted ruleset was overwritten:\n%s", after)
	}

	var report healthReport
	if err := json.Unmarshal([]byte(out.String()), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.State != statusKitManagedStateAttentionNeeded || report.RegistryState != statusKitManagedStateAttentionNeeded || report.ProjectCheck != "passed" {
		t.Fatalf("report = %#v, want preserved conflict requiring attention", report)
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
		{name: "diff requires dry run", flags: []string{"--diff"}, want: "--diff requires --dry-run"},
		{name: "diff conflicts with json", flags: []string{"--dry-run", "--diff", "--json"}, want: "--diff cannot be combined with --json"},
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

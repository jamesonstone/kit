package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/worktreeprep"
)

func executeConfigSet(t *testing.T, args ...string) (configSetReport, error) {
	t.Helper()
	cmd := newConfigSetCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(append(args, "--json"))
	err := cmd.Execute()
	var report configSetReport
	if out.Len() > 0 {
		if decodeErr := json.Unmarshal(out.Bytes(), &report); decodeErr != nil {
			t.Fatalf("JSON: %s %v", out.String(), decodeErr)
		}
	}
	return report, err
}

func TestConfigSetTransitionsAndDryRun(t *testing.T) {
	root := t.TempDir()
	setWorkingDirectory(t, root)
	stubReconcileWorktreeInspection(t, worktreeprep.Location{})
	original := "# keep comment\nschema_version: 2\ncustom: {preserve: yes}\n"
	writeFile(t, filepath.Join(root, config.ConfigFileName), original)
	args := []string{"--issue-provider", "linear", "--linear-workspace", "acme", "--linear-team-id", "11111111-1111-4111-8111-111111111111"}
	report, err := executeConfigSet(t, append(args, "--dry-run")...)
	if err != nil || !report.Changed || !report.DryRun || len(report.Next) != 2 || !report.ReconciliationRequired {
		t.Fatalf("%#v %v", report, err)
	}
	if readFile(t, filepath.Join(root, config.ConfigFileName)) != original {
		t.Fatal("dry run wrote")
	}
	report, err = executeConfigSet(t, args...)
	if err != nil || !report.Changed {
		t.Fatalf("%#v %v", report, err)
	}
	content := readFile(t, filepath.Join(root, config.ConfigFileName))
	if !strings.Contains(content, "preserve: yes") || !strings.Contains(content, "# keep comment") {
		t.Fatal(content)
	}
	report, err = executeConfigSet(t, args...)
	if err != nil || report.Changed {
		t.Fatalf("not repeatable: %#v %v", report, err)
	}
	report, err = executeConfigSet(t, "--issue-provider", "linear", "--linear-project-id", "22222222-2222-4222-8222-222222222222")
	if err != nil || report.Issues.Linear.ProjectID == "" {
		t.Fatalf("project: %#v %v", report, err)
	}
	report, err = executeConfigSet(t, "--issue-provider", "linear", "--linear-project-id", "")
	if err != nil || report.Issues.Linear.ProjectID != "" {
		t.Fatalf("clear: %#v %v", report, err)
	}
	for _, provider := range []string{"github", "linear"} {
		report, err = executeConfigSet(t, "--issue-provider", provider)
		if err != nil || report.Issues.EffectiveProvider() != provider {
			t.Fatalf("%#v %v", report, err)
		}
	}
}

func TestConfigSetInvalidNoPartialWrites(t *testing.T) {
	root := t.TempDir()
	setWorkingDirectory(t, root)
	original := "schema_version: 2\ncustom: keep\n"
	writeFile(t, filepath.Join(root, config.ConfigFileName), original)
	for _, args := range [][]string{nil, {"--issue-provider", ""}, {"--issue-provider", "unknown"}, {"--issue-provider", "linear"}, {"--issue-provider", "github", "--linear-workspace", "acme"}} {
		if _, err := executeConfigSet(t, args...); err == nil {
			t.Fatalf("accepted %v", args)
		}
		if readFile(t, filepath.Join(root, config.ConfigFileName)) != original {
			t.Fatal("invalid input wrote")
		}
	}
}

func TestConfigSetUsesCanonicalDeliveryTarget(t *testing.T) {
	root, target := t.TempDir(), t.TempDir()
	setWorkingDirectory(t, root)
	before := "schema_version: 2\ncustom: original\nissues:\n  provider: linear\n  linear: {workspace: source, team_id: 11111111-1111-4111-8111-111111111111}\n"
	other := "schema_version: 2\ncustom: target\nissues:\n  provider: linear\n  linear: {workspace: target, team_id: 22222222-2222-4222-8222-222222222222}\n"
	writeFile(t, filepath.Join(root, config.ConfigFileName), before)
	writeFile(t, filepath.Join(target, config.ConfigFileName), other)
	previous := resolveConfigSetTarget
	t.Cleanup(func() { resolveConfigSetTarget = previous })
	resolveConfigSetTarget = func(_ io.Writer, got string, dry, current bool) (reconcileTarget, error) {
		if !sameResolvedPath(got, root) || dry || current {
			t.Fatalf("unexpected target request %s %t %t", got, dry, current)
		}
		return reconcileTarget{projectRoot: target, worktree: true}, nil
	}
	report, err := executeConfigSet(t, "--issue-provider", "linear", "--linear-project-id", "33333333-3333-4333-8333-333333333333")
	if err != nil || report.Target != target || !report.Worktree {
		t.Fatalf("%#v %v", report, err)
	}
	if report.Issues.Linear.Workspace != "target" || report.Issues.Linear.TeamID != "22222222-2222-4222-8222-222222222222" || report.Issues.Linear.ProjectID != "33333333-3333-4333-8333-333333333333" {
		t.Fatalf("did not replan from canonical target: %+v", report.Issues)
	}
	if readFile(t, filepath.Join(root, config.ConfigFileName)) != before {
		t.Fatal("primary changed")
	}
	if !strings.Contains(readFile(t, filepath.Join(target, config.ConfigFileName)), "custom: target") {
		t.Fatal("target content overwritten")
	}
	resolveConfigSetTarget = func(io.Writer, string, bool, bool) (reconcileTarget, error) {
		return reconcileTarget{}, errors.New("target unavailable")
	}
	targetBefore, _ := os.ReadFile(filepath.Join(target, config.ConfigFileName))
	if _, err := executeConfigSet(t, "--issue-provider", "github"); err == nil {
		t.Fatal("unsafe delivery accepted")
	}
	targetAfter, _ := os.ReadFile(filepath.Join(target, config.ConfigFileName))
	if !bytes.Equal(targetBefore, targetAfter) {
		t.Fatal("failed target mutated")
	}
}

func TestConfigSetPrimaryPreviewExplainsWriteTarget(t *testing.T) {
	root := t.TempDir()
	setWorkingDirectory(t, root)
	stubReconcileWorktreeInspection(t, worktreeprep.Location{InsideGit: true, IsPrimary: true, Path: root})
	original := "schema_version: 2\ncustom: keep\n"
	writeFile(t, filepath.Join(root, config.ConfigFileName), original)
	args := []string{"--issue-provider", "linear", "--linear-workspace", "acme", "--linear-team-id", "11111111-1111-4111-8111-111111111111", "--dry-run"}
	report, err := executeConfigSet(t, args...)
	if err != nil || !sameResolvedPath(report.Target, root) || report.Worktree || !report.DryRun {
		t.Fatalf("%#v %v", report, err)
	}
	if !strings.Contains(report.TargetNotice, "Preview of this checkout") || !strings.Contains(report.TargetNotice, "kit-reconcile linked worktree") || !strings.Contains(report.TargetNotice, "are not included") {
		t.Fatalf("missing preview target guidance: %q", report.TargetNotice)
	}
	cmd := newConfigSetCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), report.TargetNotice) {
		t.Fatal(out.String())
	}
	if readFile(t, filepath.Join(root, config.ConfigFileName)) != original {
		t.Fatal("preview wrote configuration")
	}
}

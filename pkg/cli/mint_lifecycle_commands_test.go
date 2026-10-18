package cli

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/minttool"
	"github.com/jamesonstone/kit/v3/internal/worktreeprep"
	"github.com/spf13/cobra"
)

func TestMintMaintenanceCommandSurface(t *testing.T) {
	if _, _, err := rootCmd.Find([]string{"install"}); err == nil {
		t.Fatal("redundant install command remains registered")
	}
	for _, name := range []string{"check", "update"} {
		cmd, _, err := rootCmd.Find([]string{"mint", name})
		if err != nil || cmd.Name() != name {
			t.Fatalf("mint %s unavailable: %v", name, err)
		}
	}
}

func TestReconcileMaintainsMintBeforeWritesAndPreservesPins(t *testing.T) {
	for _, tc := range []struct {
		name          string
		preview, fail bool
		status        string
	}{
		{"invocation", false, false, "installed"}, {"missing", false, false, "missing"},
		{"preview", true, false, "installed"}, {"failure", false, true, "installed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := mintLifecycleProject(t)
			reconcileDryRun = tc.preview
			reconcileOutputOnly = true
			reconcileRefreshFiles = []string{config.ConfigFileName}
			before := readFile(t, filepath.Join(root, config.ConfigFileName))
			calls := 0
			failure := errors.New("fixture maintenance unavailable")
			stubMintLifecycle(t, func(_ context.Context, preview bool) (minttool.Report, error) {
				calls++
				if preview != tc.preview {
					t.Fatalf("wrong reconcile request preview=%t", preview)
				}
				if readFile(t, filepath.Join(root, config.ConfigFileName)) != before {
					t.Fatal("project write preceded maintenance")
				}
				r := minttool.Report{Installed: minttool.Installation{Status: tc.status, Version: "v0.9.0"}, Target: "v0.9.0"}
				if tc.fail {
					return r, failure
				}
				return r, nil
			})
			var out, diagnostics bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetContext(context.Background())
			cmd.SetOut(&out)
			cmd.SetErr(&diagnostics)
			cmd.Flags().Bool("output-only", true, "")
			var err error
			captureStdout(t, func() { err = runReconcile(cmd, nil) })
			if calls != 1 || (err != nil) != tc.fail || (tc.fail && !strings.Contains(err.Error(), failure.Error())) {
				t.Fatalf("calls=%d error=%v", calls, err)
			}
			if !tc.fail {
				captureStdout(t, func() { err = runReconcile(cmd, nil) })
				if err != nil || calls != 2 {
					t.Fatalf("current/repeated reconcile bypassed maintenance: calls=%d err=%v", calls, err)
				}
			}
			if strings.Contains(out.String(), "Mint local") || strings.Contains(out.String(), "v0.9.0") {
				t.Fatalf("output-only diagnostics entered stdout: %s", out.String())
			}
			if !tc.fail && diagnostics.Len() == 0 {
				t.Fatal("output-only lost maintenance diagnostics on stderr")
			}
			if tc.status == "missing" && !strings.Contains(diagnostics.String(), "kit mint update") {
				t.Fatal("missing Mint omitted explicit installation guidance")
			}
			if readFile(t, filepath.Join(root, "go.mod")) != "module fixture\nrequire github.com/jamesonstone/mint v0.5.0\n" || readFile(t, filepath.Join(root, ".github/workflows/ci.yml")) != "# retained pin\njobs: {}\n" {
				t.Fatal("local maintenance changed repository dependency/workflow pins")
			}
			if tc.fail && readFile(t, filepath.Join(root, config.ConfigFileName)) != before {
				t.Fatal("failed maintenance changed config")
			}
		})
	}
}

func TestFeatureReconcileAndInvalidInputsMintOrdering(t *testing.T) {
	root := mintLifecycleProject(t)
	featurePath := filepath.Join(root, "docs/specs/0001-sample")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), withFeatureFrontMatter(validSpecWithRelationships("none\n"), "spec", "0001-sample"))
	writeFile(t, filepath.Join(featurePath, "PLAN.md"), withFeatureFrontMatter(validPlan(), "plan", "0001-sample"))
	writeFile(t, filepath.Join(featurePath, "TASKS.md"), withFeatureFrontMatter(validTasks(), "tasks", "0001-sample"))
	writeFile(t, filepath.Join(root, "docs/PROJECT_PROGRESS_SUMMARY.md"), validProgressSummary("0001", "sample"))
	calls := 0
	stubMintLifecycle(t, func(context.Context, bool) (minttool.Report, error) {
		calls++
		return minttool.Report{Installed: minttool.Installation{Status: "missing"}}, nil
	})
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.Flags().Bool("output-only", true, "")
	if err := runReconcile(cmd, []string{"sample"}); err != nil || calls != 1 {
		t.Fatalf("feature reconcile bypassed maintenance: calls=%d err=%v", calls, err)
	}
	if err := runReconcile(cmd, []string{"missing"}); err == nil || calls != 1 {
		t.Fatalf("unresolved feature reached maintenance: calls=%d err=%v", calls, err)
	}
	reconcileDiff = true
	if err := runReconcile(cmd, nil); err == nil || calls != 1 {
		t.Fatalf("invalid flags reached maintenance: calls=%d err=%v", calls, err)
	}
	reconcileDiff = false
	writeFile(t, filepath.Join(root, config.ConfigFileName), "schema_version: 2\nissues: {provider: invalid}\n")
	if err := runReconcile(cmd, nil); err == nil || calls != 1 {
		t.Fatalf("invalid config reached maintenance: calls=%d err=%v", calls, err)
	}
}

func mintLifecycleProject(t *testing.T) string {
	t.Helper()
	resetReconcileFlags(t)
	setupInitHome(t)
	root := t.TempDir()
	setWorkingDirectory(t, root)
	stubReconcileWorktreeInspection(t, worktreeprep.Location{})
	if err := config.Save(root, defaultInitConfig()); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\nrequire github.com/jamesonstone/mint v0.5.0\n")
	writeFile(t, filepath.Join(root, ".github/workflows/ci.yml"), "# retained pin\njobs: {}\n")
	return root
}

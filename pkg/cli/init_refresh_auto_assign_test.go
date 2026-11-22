package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestRunInitRefresh_UpdatesManagedAutoAssignWorkflowWhenAssigneesChange(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	oldAssignees := []string{"old-owner"}
	newAssignees := []string{"jamesonstone", "octocat"}
	cfg := config.Default()
	cfg.GitHub.DefaultAssignees = &newAssignees
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, autoAssignWorkflowPath), templates.BuildAutoAssignWorkflow(oldAssignees))

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	for _, check := range []string{`"jamesonstone"`, `"octocat"`} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected refreshed workflow to contain %q, got:\n%s", check, content)
		}
	}
	if strings.Contains(content, "old-owner") {
		t.Fatalf("managed workflow kept stale assignee:\n%s", content)
	}
}

func TestRunInitRefresh_DoesNotOverwriteCustomAutoAssignWorkflowWithoutForceTarget(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	assignees := []string{"jamesonstone"}
	cfg := config.Default()
	cfg.GitHub.DefaultAssignees = &assignees
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	custom := "name: Custom auto assign\n"
	writeFile(t, filepath.Join(tempDir, autoAssignWorkflowPath), custom)

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	if content != custom {
		t.Fatalf("custom workflow was overwritten:\n%s", content)
	}
}

func TestRunInitRefresh_FileForceOverwritesCustomAutoAssignWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	assignees := []string{"jamesonstone"}
	cfg := config.Default()
	cfg.GitHub.DefaultAssignees = &assignees
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, autoAssignWorkflowPath), "name: Custom auto assign\n")

	withInitFlags(t, func() {
		initRefresh = true
		initForce = true
		initOutputOnly = true
		initRefreshFiles = []string{autoAssignWorkflowPath}
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	if !strings.Contains(content, `"jamesonstone"`) || !strings.Contains(content, "# Kit-managed auto-assignment workflow.") {
		t.Fatalf("expected force-targeted refresh to write managed workflow, got:\n%s", content)
	}
}

func TestRunInitRefresh_RejectsUnsupportedFileTarget(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initRefresh = true
		initRefreshFiles = []string{"NOT_MANAGED.md"}
		err := runInitForTest(initCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "not a Kit-managed refresh target") {
			t.Fatalf("expected unsupported target error, got %v", err)
		}
	})
}

func TestRunInitRefresh_KeepsNewerPinnedAutoAssignActionMajor(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	assignees := []string{"jamesonstone"}
	cfg := config.Default()
	cfg.GitHub.DefaultAssignees = &assignees
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	template := templates.BuildAutoAssignWorkflow(assignees)
	if !strings.Contains(template, "actions/github-script@v9\n") {
		t.Fatalf("expected template to pin actions/github-script@v9, got:\n%s", template)
	}
	bumped := strings.Replace(template, "actions/github-script@v9", "actions/github-script@v12", 1)
	writeFile(t, filepath.Join(tempDir, autoAssignWorkflowPath), bumped)

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	if content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath)); content != bumped {
		t.Fatalf("reconcile downgraded the newer pinned action major; got:\n%s", content)
	}
}

func TestPreserveNewerWorkflowActionPins(t *testing.T) {
	template := "      - uses: actions/github-script@v9\n      - uses: actions/checkout@v5\n"
	cases := []struct {
		name     string
		existing string
		want     string
	}{
		{"newer major kept", "steps:\n  - uses: actions/github-script@v10.1.0\n", "      - uses: actions/github-script@v10.1.0\n      - uses: actions/checkout@v5\n"},
		{"named step newer major kept", "steps:\n  - name: Task\n    uses: actions/github-script@v12\n", "      - uses: actions/github-script@v12\n      - uses: actions/checkout@v5\n"},
		{"older major upgraded", "steps:\n  - uses: actions/github-script@v7\n  - uses: actions/checkout@v4\n", template},
		{"same major unchanged", "steps:\n  - uses: actions/github-script@v9.2.0\n", template},
		{"other action ignored", "steps:\n  - uses: actions/setup-go@v99\n", template},
		{"commented pin ignored", "steps:\n  # - uses: actions/github-script@v99\n  - uses: actions/github-script@v9\n", template},
		{"run scalar ignored", "steps:\n  - run: |\n      echo uses: actions/github-script@v99\n      uses: actions/github-script@v98\n", template},
		{"sha pin ignored", "steps:\n  - uses: actions/github-script@0123456789abcdef0123456789abcdef01234567 # v10\n", template},
		{"invalid yaml ignored", "steps: [\n  uses: actions/github-script@v99\n", template},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := preserveNewerWorkflowActionPins(tc.existing, template); got != tc.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}

func TestReconcileAutoAssignHonorsNullOverride(t *testing.T) {
	for _, override := range []string{"github: null\n", "github:\n  default_assignees: null\n"} {
		t.Run(strings.TrimSpace(override), func(t *testing.T) {
			home := setupInitHome(t)
			root := t.TempDir()
			writeFile(t, filepath.Join(home, ".config", "kit", config.ConfigFileName), "github:\n  default_assignees: [global-example]\n")
			writeFile(t, filepath.Join(root, config.ConfigFileName), "schema_version: 2\n"+override)
			plan, err := buildInitRefreshPlan(t.Context(), root, initRefreshOptions{files: []string{autoAssignWorkflowPath}})
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.changes) != 1 || strings.Contains(plan.changes[0].after, "global-example") {
				t.Fatalf("workflow retained a global assignee after a project override: %+v", plan.changes)
			}
		})
	}
}

func TestReconcileAutoAssignUsesPlannedConfigOnForceReset(t *testing.T) {
	home := setupInitHome(t)
	root := t.TempDir()
	writeFile(t, filepath.Join(home, ".config", "kit", config.ConfigFileName), "github:\n  default_assignees: [global-example]\n")
	writeFile(t, filepath.Join(root, config.ConfigFileName), "schema_version: 2\ngithub:\n  default_assignees: [project-example]\n")
	opts := initRefreshOptions{force: true, files: []string{config.ConfigFileName, autoAssignWorkflowPath}}
	plan, err := buildInitRefreshPlan(t.Context(), root, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		t.Fatal(err)
	}
	workflow := readFile(t, filepath.Join(root, autoAssignWorkflowPath))
	if !strings.Contains(workflow, "global-example") || strings.Contains(workflow, "project-example") {
		t.Fatalf("workflow does not match the settings after reset: %s", workflow)
	}
}

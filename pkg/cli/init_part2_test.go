package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunInit_UsesProjectAutoAssignAssigneesBeforeGlobalFallback(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	globalAssignees := []string{"jamesonstone"}
	global := config.Default()
	global.GitHub.DefaultAssignees = &globalAssignees
	if _, _, err := config.PopulateGlobalConfig(global); err != nil {
		t.Fatalf("config.PopulateGlobalConfig() error = %v", err)
	}
	projectAssignees := []string{"octocat", "@hubot"}
	project := config.Default()
	project.GitHub.DefaultAssignees = &projectAssignees
	if err := config.Save(tempDir, project); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	for _, check := range []string{`"octocat"`, `"hubot"`} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected project assignee %q in workflow, got:\n%s", check, content)
		}
	}
	if strings.Contains(content, "jamesonstone") {
		t.Fatalf("project assignees should take precedence over global fallback:\n%s", content)
	}
}

func TestRunInit_ExplicitEmptyProjectAutoAssignAssigneesSkipsGlobalFallback(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	globalAssignees := []string{"jamesonstone"}
	global := config.Default()
	global.GitHub.DefaultAssignees = &globalAssignees
	if _, _, err := config.PopulateGlobalConfig(global); err != nil {
		t.Fatalf("config.PopulateGlobalConfig() error = %v", err)
	}
	projectAssignees := []string{}
	project := config.Default()
	project.GitHub.DefaultAssignees = &projectAssignees
	if err := config.Save(tempDir, project); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	if strings.Contains(content, "jamesonstone") {
		t.Fatalf("explicit empty project assignees should not fall back to global config:\n%s", content)
	}
	if !strings.Contains(content, "const assignees = [];") {
		t.Fatalf("expected explicit empty project assignees to render a no-op workflow, got:\n%s", content)
	}
}

func TestRunInit_CreatesNonBlockingAutoAssignWorkflowWithoutAssignees(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, autoAssignWorkflowPath))
	for _, check := range []string{
		"const assignees = [];",
		"No Kit auto-assignees configured; skipping.",
		"continue-on-error: true",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected no-assignee workflow to contain %q, got:\n%s", check, content)
		}
	}
}

func TestRunInit_CreatesLoopReviewAgentConfig(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	created, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	assertDefaultInitLoopAgent(t, created)
}

func TestRunInit_InstallsRegistryRulesetsAndState(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	registry := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	stubRulesetRegistry(t, registry)

	withInitFlags(t, func() {
		initOutputOnly = true

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	rulesetContent, err := os.ReadFile(filepath.Join(tempDir, rulesetTarget(registry.Slug)))
	if err != nil {
		t.Fatalf("expected registry ruleset to be installed by kit init: %v", err)
	}
	if !strings.Contains(string(rulesetContent), "slug: safety-guardrails") {
		t.Fatalf("unexpected ruleset content:\n%s", rulesetContent)
	}

	created, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	artifact, ok := created.RegistryArtifact(rulesetKind, registry.Slug)
	if !ok {
		t.Fatalf("expected registry artifact for %s", registry.Slug)
	}
	if artifact.State != registryArtifactStateManaged || artifact.InstalledHash != registry.NormalizedHash {
		t.Fatalf("artifact = %#v, want managed hash %s", artifact, registry.NormalizedHash)
	}
}

package cli

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
)

// workCleanupRule returns the optional embedded cleanup rule.
func workCleanupRule(t *testing.T) registryRuleset {
	t.Helper()
	registry, err := embeddedRulesetRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range registry {
		if rule.Slug == "work-cleanup" {
			return rule
		}
	}
	t.Fatal("work-cleanup is not shipped")
	return registryRuleset{}
}

// TestWorkCleanupIsOptionalAndInstallsExactly verifies opt-in installation through the existing registry path.
func TestWorkCleanupIsOptionalAndInstallsExactly(t *testing.T) {
	rule := workCleanupRule(t)
	if rule.Metadata.RegistryScope != rulesetRegistryScopeOptional || rule.Metadata.ReadPolicyDefault != document.ReferenceReadPolicyConditional {
		t.Fatalf("unexpected installation/read policy: %#v", rule.Metadata)
	}
	projectRoot := setupRulesProject(t)
	setWorkingDirectory(t, projectRoot)
	resetRulesFlags(t)
	stubRulesetRegistry(t, rule)
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	changes, _, changed, err := planRefreshInitRulesets(t.Context(), projectRoot, initRefreshOptions{}, cfg, nil, []registryRuleset{rule})
	if err != nil || changed || len(changes) != 0 {
		t.Fatalf("default refresh installs optional cleanup: changes=%v changed=%v err=%v", changes, changed, err)
	}
	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	if err := runRulesAdd(cmd, []string{rule.Slug}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(projectRoot, rulesetTarget(rule.Slug))); got != rule.Content {
		t.Fatal("rules add did not install the exact embedded guidance")
	}
	cfg, err = config.Load(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if installed, ok := cfg.RegistryArtifact(rulesetKind, rule.Slug); !ok || installed.InstalledHash != rule.NormalizedHash {
		t.Fatalf("installed artifact lacks embedded identity: %#v", installed)
	}
}

// TestWorkCleanupKeepsSafetyBoundaries checks that aggressive pruning never bypasses hard-delete rules.
func TestWorkCleanupKeepsSafetyBoundaries(t *testing.T) {
	content := workCleanupRule(t).Content
	for _, invariant := range []string{
		"never weakens `deletion-safety`, `delivery`, or `infrastructure-change-approval`",
		"Never delete a branch, stash, or worktree to unblock a new lane",
		"never use `--force` on any Git command",
		"Never touch the primary checkout",
		"Never run `git clean -X`",
		"Close, never delete. Never post to Slack.",
		"Never run `docker volume prune`",
		"Ask once for confirmation of the complete hard-delete list",
		"idle for 7 days with no open pull request",
		"It does not mean the work is completed",
	} {
		if !strings.Contains(content, invariant) {
			t.Errorf("lost cleanup invariant: %s", invariant)
		}
	}
	for _, state := range workCleanupStates {
		if !strings.Contains(content, "`"+state+"`") {
			t.Errorf("rule does not explain classifier state %s", state)
		}
	}
}

// TestWorkCleanupFreshInitKeepsDefaultFootprint verifies no default installation or instruction overhead.
func TestWorkCleanupFreshInitKeepsDefaultFootprint(t *testing.T) {
	rule := workCleanupRule(t)
	root := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, root)
	stubRulesetRegistry(t, rule)
	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
	})
	if document.Exists(filepath.Join(root, rulesetTarget(rule.Slug))) {
		t.Fatal("fresh init installed optional cleanup")
	}
	for _, entry := range []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"} {
		if strings.Contains(readFile(t, filepath.Join(root, entry)), "work-cleanup") {
			t.Fatalf("%s added optional cleanup to default context", entry)
		}
	}
	if words := len(strings.Fields(rule.Content)); words > 1400 {
		t.Fatalf("optional cleanup grew to %d words; keep its one-time context cost bounded", words)
	}
}

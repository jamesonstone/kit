package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestRunInitRefresh_SectionMergesManagedRuleset(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	base := registryRulesetForTest("work-lane-gating", []string{"workflow"})
	localContent := strings.Replace(base.Content, "## Examples", "- Local example.\n\n## Examples", 1)
	remoteContent := strings.Replace(base.Content, "## Verification", "- Remote verification.\n\n## Verification", 1)
	remote := registryRulesetWithContentForTest(base.Slug, remoteContent, "new-commit")
	stubRulesetRegistry(t, remote)
	stubRulesetRegistryContent(t, map[string]string{base.SourceCommit: base.Content})

	cfg := config.Default()
	recordRulesetRegistryState(cfg, base, registryArtifactStateManaged, base.NormalizedHash, base.Content)
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, rulesetTarget(base.Slug)), localContent)

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{rulesetTarget(base.Slug)}
		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content, err := os.ReadFile(filepath.Join(tempDir, rulesetTarget(base.Slug)))
	if err != nil {
		t.Fatalf("failed to read ruleset: %v", err)
	}
	for _, check := range []string{"Local example.", "Remote verification."} {
		if !strings.Contains(string(content), check) {
			t.Fatalf("expected merged content to contain %q, got:\n%s", check, content)
		}
	}
	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	artifact, ok := updated.RegistryArtifact(rulesetKind, base.Slug)
	if !ok || artifact.State != registryArtifactStateManaged {
		t.Fatalf("artifact = %#v, want managed", artifact)
	}
	mergedHash, err := normalizedRulesetContentHash(string(content), remote.Metadata.Status)
	if err != nil {
		t.Fatalf("merged hash error: %v", err)
	}
	if artifact.InstalledHash != mergedHash {
		t.Fatalf("artifact.InstalledHash = %s, want merged hash %s", artifact.InstalledHash, mergedHash)
	}
	if artifact.InstalledHash == remote.NormalizedHash {
		t.Fatalf("section-merged artifact stored remote hash instead of merged content hash")
	}
	if len(artifact.Sections) == 0 {
		t.Fatalf("section-merged artifact should store section hashes for future refreshes")
	}
}

func TestRunInitRefresh_SkipsConflictedManagedRuleset(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	base := registryRulesetForTest("github-pr-delivery", []string{"github"})
	localContent := strings.Replace(base.Content, "## Rules", "- Local rule change.\n\n## Rules", 1)
	remoteContent := strings.Replace(base.Content, "## Rules", "- Remote rule change.\n\n## Rules", 1)
	remote := registryRulesetWithContentForTest(base.Slug, remoteContent, "new-commit")
	stubRulesetRegistry(t, remote)
	stubRulesetRegistryContent(t, map[string]string{base.SourceCommit: base.Content})

	cfg := config.Default()
	recordRulesetRegistryState(cfg, base, registryArtifactStateManaged, base.NormalizedHash, base.Content)
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, rulesetTarget(base.Slug)), localContent)

	var output string
	withInitFlags(t, func() {
		initRefresh = true
		initRefreshFiles = []string{rulesetTarget(base.Slug)}
		output = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content, err := os.ReadFile(filepath.Join(tempDir, rulesetTarget(base.Slug)))
	if err != nil {
		t.Fatalf("failed to read ruleset: %v", err)
	}
	if string(content) != localContent {
		t.Fatalf("conflicted ruleset was overwritten:\n%s", content)
	}
	if !strings.Contains(output, "changed locally and in registry") {
		t.Fatalf("expected conflict note, got:\n%s", output)
	}
	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	artifact, ok := updated.RegistryArtifact(rulesetKind, base.Slug)
	if !ok || artifact.State != registryArtifactStateConflict {
		t.Fatalf("artifact = %#v, want conflict", artifact)
	}
	if artifact.SourceCommit != base.SourceCommit {
		t.Fatalf("artifact.SourceCommit = %q, want retained base %q", artifact.SourceCommit, base.SourceCommit)
	}
}

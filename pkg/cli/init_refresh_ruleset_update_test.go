package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunInitRefresh_ForceAcceptsRegistryContentAndPreservesStatus(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	registry := registryRulesetForTest("github-pr-delivery", []string{"github"})
	stubRulesetRegistry(t, registry)

	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	local := strings.Replace(registry.Content, "status: active", "status: optional", 1)
	local = strings.Replace(local, "## Examples", "Local custom guidance.\n\n## Examples", 1)
	writeFile(t, filepath.Join(tempDir, rulesetTarget(registry.Slug)), local)

	withInitFlags(t, func() {
		initRefresh = true
		initForce = true
		initOutputOnly = true
		initRefreshFiles = []string{rulesetTarget(registry.Slug)}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content, err := os.ReadFile(filepath.Join(tempDir, rulesetTarget(registry.Slug)))
	if err != nil {
		t.Fatalf("failed to read ruleset: %v", err)
	}
	if strings.Contains(string(content), "Local custom guidance.") {
		t.Fatalf("expected force to replace custom content, got:\n%s", content)
	}
	if !strings.Contains(string(content), "status: optional") {
		t.Fatalf("expected force to preserve local status, got:\n%s", content)
	}
	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	artifact, ok := updated.RegistryArtifact(rulesetKind, registry.Slug)
	if !ok || artifact.State != registryArtifactStateManaged || artifact.InstalledHash != registry.NormalizedHash {
		t.Fatalf("artifact = %#v, want managed hash %s", artifact, registry.NormalizedHash)
	}
}

func TestRunInitRefresh_FastForwardsManagedRuleset(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	base := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	remoteContent := strings.Replace(base.Content, "## Verification", "- Registry refresh verification.\n\n## Verification", 1)
	remote := registryRulesetWithContentForTest(base.Slug, remoteContent, "new-commit")
	stubRulesetRegistry(t, remote)

	cfg := config.Default()
	recordRulesetRegistryState(cfg, base, registryArtifactStateManaged, base.NormalizedHash, base.Content)
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, rulesetTarget(base.Slug)), base.Content)

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
	if !strings.Contains(string(content), "Registry refresh verification.") {
		t.Fatalf("expected remote content to be applied, got:\n%s", content)
	}
	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	artifact, ok := updated.RegistryArtifact(rulesetKind, base.Slug)
	if !ok || artifact.InstalledHash != remote.NormalizedHash {
		t.Fatalf("artifact = %#v, want hash %s", artifact, remote.NormalizedHash)
	}
}

func TestRunInitRefresh_PreservesSourceCommitForUnchangedManagedRuleset(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	base := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	remote := registryRulesetWithContentForTest(base.Slug, base.Content, "new-registry-head")
	stubRulesetRegistry(t, remote)

	cfg := defaultInitConfig()
	recordRulesetRegistryState(cfg, base, registryArtifactStateManaged, base.NormalizedHash, base.Content)
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, rulesetTarget(base.Slug)), base.Content)
	configBefore, err := os.ReadFile(filepath.Join(tempDir, config.ConfigFileName))
	if err != nil {
		t.Fatalf("failed to read config before refresh: %v", err)
	}

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

	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	artifact, ok := updated.RegistryArtifact(rulesetKind, base.Slug)
	if !ok {
		t.Fatalf("missing registry artifact %q", base.Slug)
	}
	if artifact.SourceCommit != base.SourceCommit {
		t.Fatalf("artifact.SourceCommit = %q, want unchanged %q", artifact.SourceCommit, base.SourceCommit)
	}
	configAfter, err := os.ReadFile(filepath.Join(tempDir, config.ConfigFileName))
	if err != nil {
		t.Fatalf("failed to read config after refresh: %v", err)
	}
	if string(configAfter) != string(configBefore) {
		t.Fatalf("unchanged ruleset rewrote config:\n%s", configAfter)
	}
}

func TestRunInitRefresh_AdvancesChangedContentOrSourceIdentityCommit(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	unchangedBase := registryRulesetForTest("safety-guardrails", []string{"git", "github"})
	changedBase := registryRulesetForTest("work-lane-gating", []string{"workflow"})
	movedBase := registryRulesetForTest("github-pr-delivery", []string{"github"})
	unchangedRemote := registryRulesetWithContentForTest(unchangedBase.Slug, unchangedBase.Content, "new-registry-head")
	changedContent := strings.Replace(changedBase.Content, "## Verification", "- Registry refresh verification.\n\n## Verification", 1)
	changedRemote := registryRulesetWithContentForTest(changedBase.Slug, changedContent, "new-registry-head")
	movedRemote := registryRulesetWithContentForTest(movedBase.Slug, movedBase.Content, "new-registry-head")
	movedRemote.SourcePath = "docs/references/rules/moved-github-pr-delivery.md"
	stubRulesetRegistry(t, unchangedRemote, changedRemote, movedRemote)

	cfg := defaultInitConfig()
	recordRulesetRegistryState(cfg, unchangedBase, registryArtifactStateManaged, unchangedBase.NormalizedHash, unchangedBase.Content)
	recordRulesetRegistryState(cfg, changedBase, registryArtifactStateManaged, changedBase.NormalizedHash, changedBase.Content)
	recordRulesetRegistryState(cfg, movedBase, registryArtifactStateManaged, movedBase.NormalizedHash, movedBase.Content)
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	writeFile(t, filepath.Join(tempDir, rulesetTarget(unchangedBase.Slug)), unchangedBase.Content)
	writeFile(t, filepath.Join(tempDir, rulesetTarget(changedBase.Slug)), changedBase.Content)
	writeFile(t, filepath.Join(tempDir, rulesetTarget(movedBase.Slug)), movedBase.Content)

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{
			rulesetTarget(unchangedBase.Slug),
			rulesetTarget(changedBase.Slug),
			rulesetTarget(movedBase.Slug),
		}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	updated, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	unchangedArtifact, ok := updated.RegistryArtifact(rulesetKind, unchangedBase.Slug)
	if !ok {
		t.Fatalf("missing registry artifact %q", unchangedBase.Slug)
	}
	if unchangedArtifact.SourceCommit != unchangedBase.SourceCommit {
		t.Fatalf("unchanged artifact.SourceCommit = %q, want %q", unchangedArtifact.SourceCommit, unchangedBase.SourceCommit)
	}
	changedArtifact, ok := updated.RegistryArtifact(rulesetKind, changedBase.Slug)
	if !ok {
		t.Fatalf("missing registry artifact %q", changedBase.Slug)
	}
	if changedArtifact.SourceCommit != changedRemote.SourceCommit {
		t.Fatalf("changed artifact.SourceCommit = %q, want %q", changedArtifact.SourceCommit, changedRemote.SourceCommit)
	}
	if changedArtifact.InstalledHash != changedRemote.NormalizedHash {
		t.Fatalf("changed artifact.InstalledHash = %q, want %q", changedArtifact.InstalledHash, changedRemote.NormalizedHash)
	}
	movedArtifact, ok := updated.RegistryArtifact(rulesetKind, movedBase.Slug)
	if !ok {
		t.Fatalf("missing registry artifact %q", movedBase.Slug)
	}
	if movedArtifact.SourceCommit != movedRemote.SourceCommit {
		t.Fatalf("moved artifact.SourceCommit = %q, want %q", movedArtifact.SourceCommit, movedRemote.SourceCommit)
	}
	if movedArtifact.SourcePath != movedRemote.SourcePath {
		t.Fatalf("moved artifact.SourcePath = %q, want %q", movedArtifact.SourcePath, movedRemote.SourcePath)
	}
}

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

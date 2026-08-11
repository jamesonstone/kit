package releaseprompt

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestResolveUsesExplicitDiscoveredAndDefaultValues(t *testing.T) {
	repository := filepath.Join(t.TempDir(), "service")
	if err := os.Mkdir(repository, 0o755); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{lookPaths: map[string]bool{"gh": true}}
	addFakeRepository(runner, repository, "https://user:secret@github.com/acme/service.git?token=hidden", "main")

	config, err := Resolve(context.Background(), Input{
		Repositories:           []string{repository},
		Project:                "release-x",
		ScopeExpansion:         "strict",
		InfrastructureMode:     "none",
		Environment:            "staging",
		ProductionVerification: "endpoint:https://service.example.test/health",
		IntegrationSuite:       "none",
	}, runner)
	if err != nil {
		t.Fatal(err)
	}
	if config.Project != "release-x" || config.Organization != "acme" || config.ScopeExpansion != "strict" {
		t.Fatalf("unexpected resolved config: %#v", config)
	}
	if got := config.Repositories[0].GitHub; got != "acme/service" {
		t.Fatalf("sanitized GitHub identity = %q", got)
	}
	if strings.Contains(config.Repositories[0].GitHub, "secret") || strings.Contains(config.Repositories[0].GitHub, "token") {
		t.Fatalf("repository identity leaked remote credentials: %q", config.Repositories[0].GitHub)
	}
	wantSources := []ValueSource{SourceExplicit, SourceDiscovered, SourceExplicit}
	gotSources := []ValueSource{config.Resolution[0].Source, config.Resolution[1].Source, config.Resolution[2].Source}
	if !slices.Equal(gotSources, wantSources) {
		t.Fatalf("resolution sources = %#v, want %#v", gotSources, wantSources)
	}
}

func TestRootDiscoveryIsShallowSortedAndDeduplicated(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a-service")
	z := filepath.Join(root, "z-service")
	nested := filepath.Join(a, "nested")
	for _, path := range []string{a, z, nested, filepath.Join(root, "vendor")} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	runner := &fakeRunner{}
	addFakeRepository(runner, a, "git@github.com:acme/a-service.git", "main")
	addFakeRepository(runner, z, "git@github.com:acme/z-service.git", "main")
	config, err := Resolve(context.Background(), Input{Root: root}, runner)
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Repositories) != 2 || config.Repositories[0].Name != "a-service" || config.Repositories[1].Name != "z-service" {
		t.Fatalf("shallow repository order = %#v", config.Repositories)
	}
}

func TestGitHubFallbackRunsOnceForMissingLocalMetadata(t *testing.T) {
	repository := filepath.Join(t.TempDir(), "service")
	if err := os.Mkdir(repository, 0o755); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{lookPaths: map[string]bool{"gh": true}}
	addFakeRepository(runner, repository, "", "")
	runner.ghMetadata.NameWithOwner = "acme/service"
	runner.ghMetadata.DefaultBranchRef.Name = "trunk"
	config, err := Resolve(context.Background(), Input{Repositories: []string{repository, repository}}, runner)
	if err != nil {
		t.Fatal(err)
	}
	if runner.ghCalls != 1 {
		t.Fatalf("gh calls = %d, want 1", runner.ghCalls)
	}
	if len(config.Repositories) != 1 || config.Repositories[0].DefaultBranch != "trunk" {
		t.Fatalf("fallback repository = %#v", config.Repositories)
	}
}

func TestResolveDetectsFilenameLevelClues(t *testing.T) {
	repository := filepath.Join(t.TempDir(), "service")
	if err := os.MkdirAll(filepath.Join(repository, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"cdk.json", "scripts/verify-production.sh", "scripts/integration-test.sh", ".kit.yaml"} {
		full := filepath.Join(repository, filepath.FromSlash(path))
		if err := os.WriteFile(full, []byte("fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runner := &fakeRunner{}
	addFakeRepository(runner, repository, "git@github.com:acme/service.git", "main")
	config, err := Resolve(context.Background(), Input{Repositories: []string{repository}}, runner)
	if err != nil {
		t.Fatal(err)
	}
	if config.Infrastructure.Mode != "iac" || config.Infrastructure.Provider != "aws" || config.Infrastructure.CLI != "cdk" {
		t.Fatalf("infrastructure clues = %#v", config.Infrastructure)
	}
	if config.Production.Verification != "instruction:run script:scripts/verify-production.sh from repository service" {
		t.Fatalf("verification clue = %q", config.Production.Verification)
	}
	if !config.Repositories[0].KitManaged {
		t.Fatal("expected Kit-managed repository detection")
	}
}

func TestResolveRejectsInvalidCombinationsAndSecrets(t *testing.T) {
	repository := filepath.Join(t.TempDir(), "service")
	if err := os.Mkdir(repository, 0o755); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{}
	addFakeRepository(runner, repository, "git@github.com:acme/service.git", "main")
	for _, input := range []Input{
		{Repositories: []string{repository}, Root: repository},
		{Repositories: []string{repository}, ProductionVerification: "shell:make verify"},
		{Repositories: []string{repository}, FeatureContext: "token=super-secret"},
		{Repositories: []string{repository}, InfrastructureMode: "none", InfrastructureProvider: "aws"},
		{Repositories: []string{repository}, InfrastructureMode: "direct"},
	} {
		if _, err := Resolve(context.Background(), input, runner); err == nil {
			t.Fatalf("Resolve(%#v) unexpectedly succeeded", input)
		}
	}
}

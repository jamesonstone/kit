package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/legacy"
	"github.com/jamesonstone/kit/v3/internal/templates"
	"gopkg.in/yaml.v3"
)

var migrationFixtures = []string{"v1", "v2", "v3-precontract", "phase2"}

// migrationFixtureRoot is resolved before any test changes directory.
var migrationFixtureRoot = func() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(wd, "testdata", "migration")
}()

// setupMigrationEnvironment isolates HOME and Git identity while keeping the
// real embedded rule registry.
func setupMigrationEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_AUTHOR_NAME", "Test User")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test User")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
}

// copyMigrationFixture copies a released Kit's output into a committed Git
// repository, the way an existing project looks before reconcile.
func copyMigrationFixture(t *testing.T, name string, mutate func(root string)) string {
	t.Helper()
	source := filepath.Join(migrationFixtureRoot, name)
	root := t.TempDir()
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(source, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		writeFile(t, filepath.Join(root, rel), string(data))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(root)
	}
	commitMigrationFixture(t, root)
	return root
}

func commitMigrationFixture(t *testing.T, root string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		runGitForSourceAuditTest(t, root, "init", "-q", "-b", "main")
	}
	runGitForSourceAuditTest(t, root, "add", "--all")
	runGitForSourceAuditTest(t, root, "commit", "-q", "--allow-empty", "-m", "fixture")
}

// migrateProject runs the reconcile file migration in place and returns the
// plan it applied.
func migrateProject(t *testing.T, root string, force bool) *initRefreshPlan {
	t.Helper()
	plan, err := buildInitRefreshPlan(t.Context(), root, initRefreshOptions{force: force, outputOnly: true})
	if err != nil {
		t.Fatalf("plan migration: %v", err)
	}
	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	return plan
}

func plannedChanges(t *testing.T, root string) []string {
	t.Helper()
	plan, err := buildInitRefreshPlan(t.Context(), root, initRefreshOptions{outputOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	var changed []string
	for _, change := range plan.changes {
		if change.result != instructionFileSkipped {
			changed = append(changed, string(change.result)+" "+change.relativePath)
		}
	}
	return changed
}

// freshInitProject runs `kit init` in an empty committed repository.
func freshInitProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	commitMigrationFixture(t, root)
	setWorkingDirectory(t, root)
	withInitFlags(t, func() {
		initOutputOnly = true
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})
	commitMigrationFixture(t, root)
	return root
}

// kitOwnedState is the structural view of Kit-owned project state that init
// and reconcile must agree on.
type kitOwnedState struct {
	ScaffoldVersion int
	ConfigKeys      []string
	EntryBlocks     map[string]string
	Baseline        string
	Rules           map[string]string // slug -> state:hash, default rules only
	RetiredPresent  []string
	RequiredMissing []string
}

func captureKitOwnedState(t *testing.T, root string) kitOwnedState {
	t.Helper()
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	state := kitOwnedState{ScaffoldVersion: cfg.InstructionScaffoldVersion, EntryBlocks: map[string]string{}, Rules: map[string]string{}}
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(readFile(t, filepath.Join(root, config.ConfigFileName))), &raw); err != nil {
		t.Fatal(err)
	}
	for key := range raw {
		state.ConfigKeys = append(state.ConfigKeys, key)
	}
	sort.Strings(state.ConfigKeys)
	for _, path := range instructionFiles(cfg) {
		content := readFileOrEmpty(filepath.Join(root, path))
		start := strings.Index(content, templates.UniversalContractBeginMarker)
		end := strings.LastIndex(content, templates.UniversalContractEndMarker)
		if start >= 0 && end > start && strings.Count(content, templates.UniversalContractBeginMarker) == 1 {
			state.EntryBlocks[path] = content[start : end+len(templates.UniversalContractEndMarker)]
		}
	}
	constitution := readFileOrEmpty(filepath.Join(root, cfg.ConstitutionPath))
	if i := strings.Index(constitution, "### "+templates.ConstitutionBaselineHeading); i >= 0 {
		end := strings.Index(constitution[i:], "<!-- END KIT-MANAGED BASELINE RULES -->")
		state.Baseline = constitution[i : i+end]
	}
	registry, err := embeddedRulesetRegistry(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range registry {
		if item.Metadata.RegistryScope == rulesetRegistryScopeOptional {
			continue
		}
		artifact, _ := cfg.RegistryArtifact(rulesetKind, item.Slug)
		same := readFileOrEmpty(filepath.Join(root, rulesetTarget(item.Slug))) == item.Content
		state.Rules[item.Slug] = artifact.State + ":" + artifact.InstalledHash + ":" + map[bool]string{true: "embedded", false: "differs"}[same]
	}
	retired := append(legacy.GeneratedPaths(), legacy.ProgressSummaryPath)
	for slug := range retiredRulesets {
		retired = append(retired, rulesetTarget(slug))
	}
	for _, path := range retired {
		if path == legacyConstitutionPath || path == templates.TestingReferencePath || isInstructionEntryPath(cfg, path) {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, path)); err == nil {
			state.RetiredPresent = append(state.RetiredPresent, path)
		}
	}
	sort.Strings(state.RetiredPresent)
	for _, path := range append(instructionArtifactPaths(cfg), cfg.ConstitutionPath, config.ConfigFileName) {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			state.RequiredMissing = append(state.RequiredMissing, path)
		}
	}
	return state
}

func isInstructionEntryPath(cfg *config.Config, path string) bool {
	for _, entry := range instructionFiles(cfg) {
		if filepath.ToSlash(entry) == path {
			return true
		}
	}
	return false
}

func loadMigratedConfig(t *testing.T, root string) *config.Config {
	t.Helper()
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

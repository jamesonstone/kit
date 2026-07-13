package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultIncludesAllRepositoryInstructionFiles(t *testing.T) {
	cfg := Default()
	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("Default().SchemaVersion = %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
	}

	want := []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"}
	if !reflect.DeepEqual(cfg.Agents, want) {
		t.Fatalf("Default().Agents = %v, want %v", cfg.Agents, want)
	}
}

func TestLoadWithInspectionDistinguishesSchemaStates(t *testing.T) {
	tests := []struct {
		name    string
		content string
		state   SchemaState
	}{
		{name: "missing", content: "goal_percentage: 95\n", state: SchemaStateMissing},
		{name: "older", content: "schema_version: 0\n", state: SchemaStateOlder},
		{name: "current", content: "schema_version: 1\n", state: SchemaStateCurrent},
		{name: "newer", content: "schema_version: 2\n", state: SchemaStateNewer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, ConfigFileName), []byte(tt.content), 0644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			cfg, inspection, err := LoadWithInspection(root)
			if err != nil {
				t.Fatalf("LoadWithInspection() error = %v", err)
			}
			if inspection.SchemaState != tt.state {
				t.Fatalf("SchemaState = %q, want %q", inspection.SchemaState, tt.state)
			}
			if cfg.SchemaVersion != CurrentSchemaVersion && tt.state == SchemaStateMissing {
				t.Fatalf("effective SchemaVersion = %d, want default %d", cfg.SchemaVersion, CurrentSchemaVersion)
			}
		})
	}
}

func TestLoadWithInspectionReportsNewerSchemaBeforeTypedDecode(t *testing.T) {
	root := t.TempDir()
	content := "schema_version: 2\ngoal_percentage: automatic\n"
	if err := os.WriteFile(filepath.Join(root, ConfigFileName), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, inspection, err := LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v, want upgrade inspection", err)
	}
	if inspection.SchemaState != SchemaStateNewer || len(inspection.Findings) == 0 || !strings.Contains(inspection.Findings[0].Message, "upgrade Kit") {
		t.Fatalf("inspection = %#v, want newer-schema upgrade guidance", inspection)
	}
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "upgrade Kit") {
		t.Fatalf("Load() error = %v, want upgrade guidance", err)
	}
}

func TestLoadWithInspectionValidatesAWSContext(t *testing.T) {
	root := t.TempDir()
	content := "schema_version: 1\naws:\n  profile: dev\n  account_id: not-an-account\n"
	if err := os.WriteFile(filepath.Join(root, ConfigFileName), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, inspection, err := LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if cfg.AWS == nil || cfg.AWS.Profile != "dev" {
		t.Fatalf("AWS = %#v, want dev profile", cfg.AWS)
	}
	if !inspection.HasErrors() {
		t.Fatal("expected invalid AWS account ID to be an error")
	}
}

func TestLoadWithInspectionRequiresQuotedAWSAccountID(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		wantError bool
	}{
		{name: "double quoted", accountID: `"012345678901"`, wantError: false},
		{name: "single quoted", accountID: `'012345678901'`, wantError: false},
		{name: "unquoted", accountID: `012345678901`, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			content := "schema_version: 1\naws:\n  profile: dev\n  account_id: " + tt.accountID + "\n"
			if err := os.WriteFile(filepath.Join(root, ConfigFileName), []byte(content), 0644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}

			_, inspection, err := LoadWithInspection(root)
			if err != nil {
				t.Fatalf("LoadWithInspection() error = %v", err)
			}
			if inspection.HasErrors() != tt.wantError {
				t.Fatalf("HasErrors() = %v, want %v; findings = %#v", inspection.HasErrors(), tt.wantError, inspection.Findings)
			}
		})
	}
}

func TestDisabledAWSConfigIsSemanticallyValid(t *testing.T) {
	root := t.TempDir()
	cfg := Default()
	cfg.InstructionScaffoldVersion = DefaultInstructionScaffoldVersion
	cfg.AWS = DisabledAWSConfig()
	if err := Save(root, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	_, inspection, err := LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if inspection.HasErrors() {
		t.Fatalf("disabled AWS findings = %#v, want no errors", inspection.Findings)
	}
}

func TestUpdateProjectSchemaAndAWSPreservesUnknownFields(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ConfigFileName)
	content := "goal_percentage: 95\ncustom_root:\n  keep: true\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cfg, _, err := LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	cfg.SchemaVersion = CurrentSchemaVersion
	cfg.AWS = &AWSConfig{Profile: "dev", AccountID: "012345678901"}
	if err := UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	for _, want := range []string{"schema_version: 1", "custom_root:", "keep: true", "profile: dev", `account_id: "012345678901"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("updated config missing %q:\n%s", want, text)
		}
	}
}

func TestDefaultIncludesLoopPolicy(t *testing.T) {
	cfg := Default()

	if cfg.Loop.MinConfidence != 95 {
		t.Fatalf("Loop.MinConfidence = %d, want 95", cfg.Loop.MinConfidence)
	}
	if cfg.Loop.MaxIterations != 20 {
		t.Fatalf("Loop.MaxIterations = %d, want 20", cfg.Loop.MaxIterations)
	}
	if cfg.ProjectRefresh.Constitution.FeatureInterval != 5 {
		t.Fatalf("ProjectRefresh.Constitution.FeatureInterval = %d, want 5", cfg.ProjectRefresh.Constitution.FeatureInterval)
	}
	if cfg.ProjectRefresh.Constitution.MaxAgeDays != 30 {
		t.Fatalf("ProjectRefresh.Constitution.MaxAgeDays = %d, want 30", cfg.ProjectRefresh.Constitution.MaxAgeDays)
	}
}

func TestSaveOmitsDefaultLoopConfigAndKeepsCustomLoopConfig(t *testing.T) {
	projectRoot := t.TempDir()
	defaults := Default()
	if err := Save(projectRoot, defaults); err != nil {
		t.Fatalf("Save(defaults) error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, ConfigFileName))
	if err != nil {
		t.Fatalf("ReadFile(default config) error = %v", err)
	}
	if strings.Contains(string(data), "loop:") {
		t.Fatalf("default loop config should be omitted, got:\n%s", data)
	}
	if strings.Contains(string(data), "github:") {
		t.Fatalf("empty github config should be omitted, got:\n%s", data)
	}

	defaults.Loop.MaxIterations = 7
	defaults.Loop.Agent.Command = "codex"
	if err := Save(projectRoot, defaults); err != nil {
		t.Fatalf("Save(custom loop) error = %v", err)
	}
	data, err = os.ReadFile(filepath.Join(projectRoot, ConfigFileName))
	if err != nil {
		t.Fatalf("ReadFile(custom config) error = %v", err)
	}
	for _, check := range []string{"loop:", "max_iterations: 7", "command: codex"} {
		if !strings.Contains(string(data), check) {
			t.Fatalf("custom loop config missing %q, got:\n%s", check, data)
		}
	}
}

func TestLoadParsesLoopConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`
loop:
  min_confidence: 91
  max_iterations: 7
  agent:
    command: codex
    args:
      - run
      - --stdin
`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Loop.MinConfidence != 91 {
		t.Fatalf("Loop.MinConfidence = %d, want 91", cfg.Loop.MinConfidence)
	}
	if cfg.Loop.MaxIterations != 7 {
		t.Fatalf("Loop.MaxIterations = %d, want 7", cfg.Loop.MaxIterations)
	}
	if cfg.Loop.Agent.Command != "codex" {
		t.Fatalf("Loop.Agent.Command = %q, want codex", cfg.Loop.Agent.Command)
	}
	wantArgs := []string{"run", "--stdin"}
	if !reflect.DeepEqual(cfg.Loop.Agent.Args, wantArgs) {
		t.Fatalf("Loop.Agent.Args = %v, want %v", cfg.Loop.Agent.Args, wantArgs)
	}
}

func TestLoadParsesProjectRefreshConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`
project_refresh:
  constitution:
    feature_interval: 3
    max_age_days: 14
    last_reviewed_at: "2026-06-01T12:00:00Z"
    last_completed_feature_count: 8
`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	refresh := cfg.ProjectRefresh.Constitution
	if refresh.FeatureInterval != 3 || refresh.MaxAgeDays != 14 {
		t.Fatalf("ProjectRefresh.Constitution thresholds = %#v", refresh)
	}
	if refresh.LastReviewedAt != "2026-06-01T12:00:00Z" {
		t.Fatalf("LastReviewedAt = %q", refresh.LastReviewedAt)
	}
	if refresh.LastCompletedFeatureCount != 8 {
		t.Fatalf("LastCompletedFeatureCount = %d, want 8", refresh.LastCompletedFeatureCount)
	}
}

func TestLoadParsesRegistryConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`
registry:
  schema_version: 1
  source:
    repo: jamesonstone/kit
    branch: main
  artifacts:
    - kind: ruleset
      slug: github-pr-delivery
      path: docs/references/rules/github-pr-delivery.md
      source_repo: jamesonstone/kit
      source_branch: main
      source_commit: abc123
      source_path: docs/references/rules/github-pr-delivery.md
      installed_hash: sha256:deadbeef
      state: managed
`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Registry.SchemaVersion != 1 {
		t.Fatalf("Registry.SchemaVersion = %d, want 1", cfg.Registry.SchemaVersion)
	}
	if cfg.Registry.Source.Repo != "jamesonstone/kit" || cfg.Registry.Source.Branch != "main" {
		t.Fatalf("Registry.Source = %#v", cfg.Registry.Source)
	}
	artifact, ok := cfg.RegistryArtifact("ruleset", "github-pr-delivery")
	if !ok {
		t.Fatal("expected registry artifact")
	}
	if artifact.InstalledHash != "sha256:deadbeef" || artifact.State != "managed" {
		t.Fatalf("artifact = %#v", artifact)
	}
	cfg.UpsertRegistryArtifact(RegistryArtifact{
		Kind:          "ruleset",
		Slug:          "github-pr-delivery",
		Path:          "docs/references/rules/github-pr-delivery.md",
		InstalledHash: "sha256:feedface",
		State:         "local-custom",
	})
	artifact, ok = cfg.RegistryArtifact("ruleset", "github-pr-delivery")
	if !ok || artifact.InstalledHash != "sha256:feedface" || artifact.State != "local-custom" {
		t.Fatalf("updated artifact = %#v", artifact)
	}
}

func TestLoadParsesGitHubConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(`
github:
  repository: jamesonstone/kit
  default_branch: main
  default_assignees:
    - jamesonstone
    - octocat
`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GitHub.Repository != "jamesonstone/kit" {
		t.Fatalf("GitHub.Repository = %q, want jamesonstone/kit", cfg.GitHub.Repository)
	}
	if cfg.GitHub.DefaultBranch != "main" {
		t.Fatalf("GitHub.DefaultBranch = %q, want main", cfg.GitHub.DefaultBranch)
	}
	if cfg.GitHub.DefaultAssignees == nil {
		t.Fatalf("GitHub.DefaultAssignees = nil, want configured assignees")
	}
	if !reflect.DeepEqual(*cfg.GitHub.DefaultAssignees, []string{"jamesonstone", "octocat"}) {
		t.Fatalf("GitHub.DefaultAssignees = %v, want jamesonstone and octocat", *cfg.GitHub.DefaultAssignees)
	}

	if err := Save(projectRoot, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	for _, check := range []string{"github:", "repository: jamesonstone/kit", "default_branch: main", "default_assignees:", "- jamesonstone", "- octocat"} {
		if !strings.Contains(string(data), check) {
			t.Fatalf("saved config missing %q, got:\n%s", check, data)
		}
	}
}

func TestLoadAllowsMissingPrompts(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("goal_percentage: 90\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GoalPercentage != 90 {
		t.Fatalf("GoalPercentage = %d, want 90", cfg.GoalPercentage)
	}
	if cfg.Prompts != nil {
		t.Fatalf("Prompts = %v, want nil", cfg.Prompts)
	}
}

func TestLoadParsesPromptSchema(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	configText := []byte(`
prompts:
  coding-agent:
    short:
      content: clarify first
      description: short planning prompt
`)
	if err := os.WriteFile(configPath, configText, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got := cfg.Prompts["coding-agent"]["short"]
	if got.Content != "clarify first" {
		t.Fatalf("Content = %q, want %q", got.Content, "clarify first")
	}
	if got.Description != "short planning prompt" {
		t.Fatalf("Description = %q, want %q", got.Description, "short planning prompt")
	}
}

func TestLoadPromptSchemaIgnoresUnknownMetadata(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	configText := []byte(`
prompts:
  coding-agent:
    short:
      content: clarify first
      description: short planning prompt
      tags:
        - planning
      future:
        owner: agent
`)
	if err := os.WriteFile(configPath, configText, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got := cfg.Prompts["coding-agent"]["short"]
	if got.Content != "clarify first" {
		t.Fatalf("Content = %q, want clarify first", got.Content)
	}
	if got.Description != "short planning prompt" {
		t.Fatalf("Description = %q, want short planning prompt", got.Description)
	}
}

func TestRecordRemovedFeatureClearsPausedStateAndReplacesTombstone(t *testing.T) {
	cfg := Default()
	cfg.SetFeaturePaused("0002-bravo", true)
	cfg.RecordRemovedFeature(RemovedFeature{
		Number:    2,
		Slug:      "bravo",
		DirName:   "0002-bravo",
		CreatedAt: "2026-04-05T00:00:00Z",
		RemovedAt: "2026-05-06T12:00:00Z",
	})

	if cfg.IsFeaturePaused("0002-bravo") {
		t.Fatalf("expected removed feature to clear paused state")
	}
	if len(cfg.RemovedFeatures) != 1 {
		t.Fatalf("RemovedFeatures length = %d, want 1", len(cfg.RemovedFeatures))
	}
	if cfg.RemovedFeatures[0].RemovedAt != "2026-05-06T12:00:00Z" {
		t.Fatalf("RemovedAt = %q, want first timestamp", cfg.RemovedFeatures[0].RemovedAt)
	}

	cfg.RecordRemovedFeature(RemovedFeature{
		Number:    2,
		Slug:      "bravo",
		DirName:   "0002-bravo",
		CreatedAt: "2026-04-05T00:00:00Z",
		RemovedAt: "2026-05-07T12:00:00Z",
	})

	if len(cfg.RemovedFeatures) != 1 {
		t.Fatalf("RemovedFeatures length after replace = %d, want 1", len(cfg.RemovedFeatures))
	}
	if cfg.RemovedFeatures[0].RemovedAt != "2026-05-07T12:00:00Z" {
		t.Fatalf("RemovedAt = %q, want replacement timestamp", cfg.RemovedFeatures[0].RemovedAt)
	}
}

func TestFindProjectRootOptionalReturnsRootWhenConfigExists(t *testing.T) {
	projectRoot := t.TempDir()
	nested := filepath.Join(projectRoot, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, ConfigFileName), []byte("goal_percentage: 95\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Chdir(nested)

	root, found, err := FindProjectRootOptional()
	if err != nil {
		t.Fatalf("FindProjectRootOptional() error = %v", err)
	}
	if !found {
		t.Fatal("FindProjectRootOptional() found = false, want true")
	}
	if root != projectRoot {
		t.Fatalf("FindProjectRootOptional() root = %q, want %q", root, projectRoot)
	}
}

func TestFindProjectRootOptionalReturnsNotFoundWithoutError(t *testing.T) {
	t.Chdir(t.TempDir())

	root, found, err := FindProjectRootOptional()
	if err != nil {
		t.Fatalf("FindProjectRootOptional() error = %v", err)
	}
	if found {
		t.Fatal("FindProjectRootOptional() found = true, want false")
	}
	if root != "" {
		t.Fatalf("FindProjectRootOptional() root = %q, want empty", root)
	}
}

func TestGlobalConfigPathUsesDotConfigKit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := GlobalConfigPath()
	if err != nil {
		t.Fatalf("GlobalConfigPath() error = %v", err)
	}

	want := filepath.Join(home, ".config", "kit", ConfigFileName)
	if got != want {
		t.Fatalf("GlobalConfigPath() = %q, want %q", got, want)
	}
}

func TestLoadGlobalReturnsDefaultWhenAbsent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if found {
		t.Fatal("LoadGlobal() found = true, want false")
	}
	if cfg == nil {
		t.Fatal("LoadGlobal() cfg = nil")
	}
	if cfg.Prompts != nil {
		t.Fatalf("Prompts = %v, want nil", cfg.Prompts)
	}
}

func TestPopulateGlobalConfigCreatesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	defaults := Default()
	defaults.InstructionScaffoldVersion = DefaultInstructionScaffoldVersion

	path, changed, err := PopulateGlobalConfig(defaults)
	if err != nil {
		t.Fatalf("PopulateGlobalConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("PopulateGlobalConfig() changed = false, want true")
	}
	wantPath := filepath.Join(home, ".config", "kit", ConfigFileName)
	if path != wantPath {
		t.Fatalf("path = %q, want %q", path, wantPath)
	}

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if !found {
		t.Fatal("LoadGlobal() found = false, want true")
	}
	if cfg.GoalPercentage != defaults.GoalPercentage {
		t.Fatalf("GoalPercentage = %d, want %d", cfg.GoalPercentage, defaults.GoalPercentage)
	}
	if cfg.InstructionScaffoldVersion != DefaultInstructionScaffoldVersion {
		t.Fatalf("InstructionScaffoldVersion = %d, want %d", cfg.InstructionScaffoldVersion, DefaultInstructionScaffoldVersion)
	}
}

func TestPopulateGlobalConfigPreservesPromptsAndAddsMissingDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".config", "kit", ConfigFileName)
	initial := []byte(`prompts:
  custom:
    review:
      content: keep existing prompt
      description: existing description
`)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	defaults := Default()
	defaults.InstructionScaffoldVersion = DefaultInstructionScaffoldVersion

	_, changed, err := PopulateGlobalConfig(defaults)
	if err != nil {
		t.Fatalf("PopulateGlobalConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("PopulateGlobalConfig() changed = false, want true")
	}

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if !found {
		t.Fatal("LoadGlobal() found = false, want true")
	}
	got := cfg.Prompts["custom"]["review"]
	if got.Content != "keep existing prompt" {
		t.Fatalf("Content = %q, want keep existing prompt", got.Content)
	}
	if got.Description != "existing description" {
		t.Fatalf("Description = %q, want existing description", got.Description)
	}
	if cfg.GoalPercentage != defaults.GoalPercentage {
		t.Fatalf("GoalPercentage = %d, want %d", cfg.GoalPercentage, defaults.GoalPercentage)
	}
	if cfg.InstructionScaffoldVersion != DefaultInstructionScaffoldVersion {
		t.Fatalf("InstructionScaffoldVersion = %d, want %d", cfg.InstructionScaffoldVersion, DefaultInstructionScaffoldVersion)
	}
}

func TestUpsertGlobalPromptCreatesConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := UpsertGlobalPrompt("coding-agent", "short", Prompt{
		Content:     "clarify first",
		Description: "short prompt",
	})
	if err != nil {
		t.Fatalf("UpsertGlobalPrompt() error = %v", err)
	}

	cfg, found, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if !found {
		t.Fatal("LoadGlobal() found = false, want true")
	}
	got := cfg.Prompts["coding-agent"]["short"]
	if got.Content != "clarify first" {
		t.Fatalf("Content = %q, want clarify first", got.Content)
	}
	if got.Description != "short prompt" {
		t.Fatalf("Description = %q, want short prompt", got.Description)
	}
}

func TestUpsertLocalPromptWritesProjectConfig(t *testing.T) {
	projectRoot := t.TempDir()
	configPath := filepath.Join(projectRoot, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("goal_percentage: 90\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := UpsertLocalPrompt(projectRoot, "coding-agent", "short", Prompt{
		Content:     "clarify first",
		Description: "short prompt",
	})
	if err != nil {
		t.Fatalf("UpsertLocalPrompt() error = %v", err)
	}

	cfg, err := Load(projectRoot)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	got := cfg.Prompts["coding-agent"]["short"]
	if got.Content != "clarify first" {
		t.Fatalf("Content = %q, want clarify first", got.Content)
	}
	if got.Description != "short prompt" {
		t.Fatalf("Description = %q, want short prompt", got.Description)
	}
}

func TestUpsertPromptFilePreservesUnrelatedAndUnknownFields(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ConfigFileName)
	initial := []byte(`
goal_percentage: 90
custom_root: keep me
prompts:
  coding-agent:
    short:
      content: old
      description: old description
      tags:
        - keep
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := UpsertPromptFile(configPath, "coding-agent", "short", Prompt{
		Content: "new",
	}, false)
	if err != nil {
		t.Fatalf("UpsertPromptFile() error = %v", err)
	}

	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	for _, want := range []string{
		"custom_root: keep me",
		"content: new",
		"tags:",
		"- keep",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("updated config missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "description: old description") {
		t.Fatalf("description was not removed:\n%s", text)
	}
}

func TestUpsertPromptFileRejectsEmptyContent(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ConfigFileName)
	if err := os.WriteFile(configPath, []byte("{}\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := UpsertPromptFile(configPath, "coding-agent", "short", Prompt{}, false)
	if err == nil {
		t.Fatal("expected empty prompt content to fail")
	}
}

func TestUpsertPromptFileRejectsNonMappingRoot(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ConfigFileName)
	if err := os.WriteFile(configPath, []byte("- not-a-mapping\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := UpsertPromptFile(configPath, "coding-agent", "short", Prompt{Content: "prompt"}, false)
	if err == nil {
		t.Fatal("expected non-mapping config root to fail")
	}
	if !strings.Contains(err.Error(), "config root must be a YAML mapping") {
		t.Fatalf("error = %q, want YAML mapping guidance", err)
	}
}

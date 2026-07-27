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

func TestHealthManagedDefaultsToTrueUnlessExplicitlyDisabled(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "omitted", content: "schema_version: 1\n", want: true},
		{name: "null health", content: "schema_version: 1\nhealth:\n", want: true},
		{name: "empty health", content: "schema_version: 1\nhealth: {}\n", want: true},
		{name: "null managed", content: "schema_version: 1\nhealth:\n  managed:\n", want: true},
		{name: "explicit true", content: "schema_version: 1\nhealth:\n  managed: true\n", want: true},
		{name: "explicit false", content: "schema_version: 1\nhealth:\n  managed: false\n", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, ConfigFileName), []byte(tt.content), 0644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			cfg, err := Load(root)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if got := cfg.IsHealthManaged(); got != tt.want {
				t.Fatalf("IsHealthManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveOmitsDefaultHealthConfig(t *testing.T) {
	root := t.TempDir()
	if err := Save(root, Default()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ConfigFileName))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(data), "health:") {
		t.Fatalf("default config materialized health opt-in:\n%s", data)
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

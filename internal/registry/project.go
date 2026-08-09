package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

const ProjectFile = ".kit.yaml"

func DefaultSource() SourceConfig {
	return SourceConfig{
		Repo:        "jamesonstone/kit",
		Branch:      "main",
		CatalogPath: "registry/catalog.yaml",
	}
}

func NewProjectConfig(source SourceConfig) ProjectConfig {
	return ProjectConfig{
		SchemaVersion: ProjectSchemaVersion,
		Registry: ProjectRegistry{
			SchemaVersion: CatalogSchemaVersion,
			Source:        normalizeSource(source),
		},
	}
}

func LoadProject(root string) (ProjectConfig, bool, error) {
	content, err := os.ReadFile(filepath.Join(root, ProjectFile))
	if err != nil {
		return ProjectConfig{}, false, fmt.Errorf("read %s: %w", ProjectFile, err)
	}
	var cfg ProjectConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return ProjectConfig{}, false, fmt.Errorf("parse %s: %w", ProjectFile, err)
	}
	if cfg.SchemaVersion == 0 {
		cfg.SchemaVersion = 1
	}
	if cfg.SchemaVersion > ProjectSchemaVersion {
		return ProjectConfig{}, false, fmt.Errorf("%s schema_version %d is newer than supported version %d", ProjectFile, cfg.SchemaVersion, ProjectSchemaVersion)
	}
	cfg.Registry.Source = normalizeSource(cfg.Registry.Source)
	migration := cfg.SchemaVersion < ProjectSchemaVersion
	if migration {
		stripLegacyConfig(&cfg)
	}
	if cfg.Registry.SchemaVersion == 0 {
		cfg.Registry.SchemaVersion = CatalogSchemaVersion
	}
	sortRecords(cfg.Registry.Artifacts)
	return cfg, migration, nil
}

func stripLegacyConfig(cfg *ProjectConfig) {
	for _, key := range []string{
		"goal_percentage", "specs_dir", "skills_dir", "constitution_path",
		"allow_out_of_order", "loop", "agents", "instruction_scaffold_version",
		"feature_naming", "removed_features", "health", "github", "aws", "project_refresh",
	} {
		delete(cfg.Extra, key)
	}
}

func MarshalProject(cfg ProjectConfig) ([]byte, error) {
	cfg.SchemaVersion = ProjectSchemaVersion
	cfg.Registry.SchemaVersion = CatalogSchemaVersion
	cfg.Registry.Source = normalizeSource(cfg.Registry.Source)
	sortRecords(cfg.Registry.Artifacts)
	content, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", ProjectFile, err)
	}
	return content, nil
}

func FindProjectRoot(start string) (string, error) {
	root, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, ProjectFile)); err == nil {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", fmt.Errorf("%s not found; run `kit init`", ProjectFile)
		}
		root = parent
	}
}

func RecordByKey(records []ArtifactRecord, kind, slug string) (ArtifactRecord, bool) {
	for _, record := range records {
		if record.Kind == kind && record.Slug == slug {
			return record, true
		}
	}
	return ArtifactRecord{}, false
}

func UpsertRecord(records []ArtifactRecord, record ArtifactRecord) []ArtifactRecord {
	for index := range records {
		if records[index].Kind == record.Kind && records[index].Slug == record.Slug {
			records[index] = record
			sortRecords(records)
			return records
		}
	}
	records = append(records, record)
	sortRecords(records)
	return records
}

func RemoveRecord(records []ArtifactRecord, kind, slug string) []ArtifactRecord {
	result := records[:0]
	for _, record := range records {
		if record.Kind != kind || record.Slug != slug {
			result = append(result, record)
		}
	}
	sortRecords(result)
	return result
}

func normalizeSource(source SourceConfig) SourceConfig {
	defaults := DefaultSource()
	if source.Repo == "" && source.Path == "" {
		source.Repo = defaults.Repo
	}
	if source.Branch == "" {
		source.Branch = defaults.Branch
	}
	if source.CatalogPath == "" {
		source.CatalogPath = defaults.CatalogPath
	}
	return source
}

func sortRecords(records []ArtifactRecord) {
	sort.Slice(records, func(i, j int) bool {
		return ArtifactKey(records[i].Kind, records[i].Slug) < ArtifactKey(records[j].Kind, records[j].Slug)
	})
}

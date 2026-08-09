package registry

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func ValidSlug(slug string) bool {
	return slugPattern.MatchString(slug)
}

func ParseCatalog(content []byte) (Catalog, error) {
	var catalog Catalog
	if err := yaml.Unmarshal(content, &catalog); err != nil {
		return Catalog{}, fmt.Errorf("parse registry catalog: %w", err)
	}
	if err := ValidateCatalog(catalog); err != nil {
		return Catalog{}, err
	}
	sort.Slice(catalog.Artifacts, func(i, j int) bool {
		return ArtifactKey(catalog.Artifacts[i].Kind, catalog.Artifacts[i].Slug) <
			ArtifactKey(catalog.Artifacts[j].Kind, catalog.Artifacts[j].Slug)
	})
	return catalog, nil
}

func ValidateCatalog(catalog Catalog) error {
	if catalog.SchemaVersion != CatalogSchemaVersion {
		return fmt.Errorf("registry catalog schema_version %d is unsupported; expected %d", catalog.SchemaVersion, CatalogSchemaVersion)
	}
	seen := map[string]bool{}
	targets := map[string]string{}
	for index, artifact := range catalog.Artifacts {
		if err := validateCatalogArtifact(artifact); err != nil {
			return fmt.Errorf("registry artifact %d: %w", index+1, err)
		}
		key := ArtifactKey(artifact.Kind, artifact.Slug)
		if seen[key] {
			return fmt.Errorf("registry artifact %q is duplicated", key)
		}
		if previous := targets[artifact.TargetPath]; previous != "" {
			return fmt.Errorf("registry artifacts %q and %q share target path %q", previous, key, artifact.TargetPath)
		}
		seen[key] = true
		targets[artifact.TargetPath] = key
	}
	for _, artifact := range catalog.Artifacts {
		for _, dependency := range artifact.Dependencies {
			if !validArtifactKey(dependency) {
				return fmt.Errorf("registry artifact %q has invalid dependency %q", ArtifactKey(artifact.Kind, artifact.Slug), dependency)
			}
			if !seen[dependency] {
				return fmt.Errorf("registry artifact %q depends on missing %q", ArtifactKey(artifact.Kind, artifact.Slug), dependency)
			}
		}
	}
	return validateDependencyCycles(catalog.Artifacts)
}

func validateCatalogArtifact(artifact CatalogArtifact) error {
	if artifact.Kind != KindRuleset && artifact.Kind != KindWorkflow {
		return fmt.Errorf("kind %q must be ruleset or workflow", artifact.Kind)
	}
	if !slugPattern.MatchString(artifact.Slug) {
		return fmt.Errorf("slug %q is invalid", artifact.Slug)
	}
	if strings.TrimSpace(artifact.Description) == "" {
		return fmt.Errorf("%s description is required", ArtifactKey(artifact.Kind, artifact.Slug))
	}
	if artifact.Visibility != "downstream" && artifact.Visibility != "registry" {
		return fmt.Errorf("%s visibility %q is unsupported", ArtifactKey(artifact.Kind, artifact.Slug), artifact.Visibility)
	}
	if artifact.Version < 1 {
		return fmt.Errorf("%s version must be positive", ArtifactKey(artifact.Kind, artifact.Slug))
	}
	if !digestPattern.MatchString(artifact.Digest) {
		return fmt.Errorf("%s digest must be a sha256 value", ArtifactKey(artifact.Kind, artifact.Slug))
	}
	for _, candidate := range []string{artifact.SourcePath, artifact.TargetPath} {
		if !safeRelativePath(candidate) {
			return fmt.Errorf("%s path %q must stay inside the project", ArtifactKey(artifact.Kind, artifact.Slug), candidate)
		}
	}
	if artifact.ReadPolicy != "must" && artifact.ReadPolicy != "conditional" {
		return fmt.Errorf("%s read_policy must be must or conditional", ArtifactKey(artifact.Kind, artifact.Slug))
	}
	return nil
}

func validArtifactKey(key string) bool {
	parts := strings.Split(key, "/")
	return len(parts) == 2 && (parts[0] == KindRuleset || parts[0] == KindWorkflow) && ValidSlug(parts[1])
}

func safeRelativePath(path string) bool {
	if strings.TrimSpace(path) == "" || filepath.IsAbs(path) {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean != "." && clean != ".." && !strings.HasPrefix(clean, "../")
}

func validateDependencyCycles(artifacts []CatalogArtifact) error {
	graph := map[string][]string{}
	for _, artifact := range artifacts {
		key := ArtifactKey(artifact.Kind, artifact.Slug)
		graph[key] = append(graph[key], artifact.Dependencies...)
	}
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var visit func(string) error
	visit = func(key string) error {
		if visiting[key] {
			return fmt.Errorf("registry dependency cycle includes %q", key)
		}
		if visited[key] {
			return nil
		}
		visiting[key] = true
		for _, dependency := range graph[key] {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		visiting[key] = false
		visited[key] = true
		return nil
	}
	for key := range graph {
		if err := visit(key); err != nil {
			return err
		}
	}
	return nil
}

func VisibleArtifacts(catalog Catalog) []CatalogArtifact {
	result := make([]CatalogArtifact, 0, len(catalog.Artifacts))
	for _, artifact := range catalog.Artifacts {
		if artifact.Visibility == "" || artifact.Visibility == "downstream" {
			result = append(result, artifact)
		}
	}
	return result
}

func FindArtifact(catalog Catalog, kind, slug string) (CatalogArtifact, bool) {
	for _, artifact := range catalog.Artifacts {
		if artifact.Kind == kind && artifact.Slug == slug {
			return artifact, true
		}
	}
	return CatalogArtifact{}, false
}

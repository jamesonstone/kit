package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
)

const (
	frontendProfileDependencyName  = "Frontend profile"
	frontendProfileDependencyType  = "profile"
	frontendProfileReferenceTarget = "--profile=frontend"
	designMaterialsDependencyName  = "Design materials"
	designMaterialsDependencyType  = "design"
)

type profileDependencyRow struct {
	Dependency string
	Type       string
	Location   string
	UsedFor    string
	Status     string
}

func designMaterialsRelPath(featureDirName string) string {
	return filepath.ToSlash(filepath.Join("docs", "notes", featureDirName, "design"))
}

func featureHasActiveFrontendProfileDependency(featurePath string) bool {
	for _, source := range frontendProfileDependencySources(featurePath) {
		if !document.Exists(source.path) {
			continue
		}
		content, err := os.ReadFile(source.path)
		if err != nil {
			continue
		}
		doc := document.Parse(string(content), source.path, source.docType)
		if hasActiveFrontendProfileReference(doc.References()) {
			return true
		}
	}
	return false
}

func ensureFrontendProfileDependencyRows(docPath string, docType document.DocumentType, featureDirName string) (bool, error) {
	content, err := os.ReadFile(docPath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", docPath, err)
	}

	updated, changed, err := appendFrontendProfileDependencyRows(string(content), docType, featureDirName)
	if err != nil {
		return false, fmt.Errorf("failed to update frontend profile references in %s: %w", docPath, err)
	}
	if !changed {
		return false, nil
	}

	if err := document.Write(docPath, updated); err != nil {
		return false, fmt.Errorf("failed to update frontend profile references in %s: %w", docPath, err)
	}

	return true, nil
}

func seedFrontendProfileDependencyRows(content string, docType document.DocumentType, featureDirName string) string {
	updated, _, err := appendFrontendProfileDependencyRows(content, docType, featureDirName)
	if err != nil {
		return content
	}
	return updated
}

func appendFrontendProfileDependencyRows(content string, docType document.DocumentType, featureDirName string) (string, bool, error) {
	updated, changed, err := document.UpsertMetadata(content, docType, document.MetadataUpsert{
		Feature:    document.FeatureMetadataFromDir(featureDirName),
		References: referencesForMetadataUpsert(content, docType, canonicalFrontendProfileReferences(featureDirName)),
	})
	if err != nil {
		return content, false, err
	}
	return updated, changed, nil
}

func referencesForMetadataUpsert(content string, docType document.DocumentType, newReferences []document.MetadataReference) []document.MetadataReference {
	doc := document.Parse(content, "", docType)
	if doc.FrontMatterPresent {
		return newReferences
	}
	return newReferences
}

func hasActiveFrontendProfileReference(references []document.MetadataReference) bool {
	for _, reference := range references {
		if dependencyCellMatches(reference.Name, frontendProfileDependencyName) &&
			dependencyCellMatches(reference.Type, frontendProfileDependencyType) &&
			dependencyCellMatches(reference.Target, frontendProfileReferenceTarget) &&
			strings.EqualFold(normalizeDependencyCell(reference.Status), document.ReferenceStatusActive) {
			return true
		}
	}
	return false
}

func canonicalFrontendProfileReferences(featureDirName string) []document.MetadataReference {
	rows := canonicalFrontendProfileDependencyRows(featureDirName)
	references := make([]document.MetadataReference, 0, len(rows))
	for _, row := range rows {
		references = append(references, document.MetadataReference{
			ID:         frontendProfileReferenceID(row.Dependency),
			Name:       row.Dependency,
			Type:       row.Type,
			Target:     row.Location,
			Relation:   document.ReferenceRelationGuides,
			ReadPolicy: document.ReferenceReadPolicyConditional,
			UsedFor:    row.UsedFor,
			Status:     row.Status,
		})
	}
	return references
}

func frontendProfileReferenceID(name string) string {
	if dependencyCellMatches(name, frontendProfileDependencyName) {
		return "frontend-profile"
	}
	if dependencyCellMatches(name, designMaterialsDependencyName) {
		return "frontend-design-materials"
	}
	return ""
}

func canonicalFrontendProfileDependencyRows(featureDirName string) []profileDependencyRow {
	return []profileDependencyRow{
		{
			Dependency: frontendProfileDependencyName,
			Type:       frontendProfileDependencyType,
			Location:   frontendProfileReferenceTarget,
			UsedFor:    "apply frontend-specific coding-agent instruction set",
			Status:     "active",
		},
		{
			Dependency: designMaterialsDependencyName,
			Type:       designMaterialsDependencyType,
			Location:   designMaterialsRelPath(featureDirName),
			UsedFor:    "optional frontend design input",
			Status:     "optional",
		},
	}
}

func frontendProfileDependencySources(featurePath string) []struct {
	path    string
	docType document.DocumentType
} {
	return []struct {
		path    string
		docType document.DocumentType
	}{
		{path: filepath.Join(featurePath, "BRAINSTORM.md"), docType: document.TypeBrainstorm},
		{path: filepath.Join(featurePath, "SPEC.md"), docType: document.TypeSpec},
		{path: filepath.Join(featurePath, "PLAN.md"), docType: document.TypePlan},
	}
}

func dependencyCellMatches(got, want string) bool {
	return normalizeDependencyCell(got) == normalizeDependencyCell(want)
}

func normalizeDependencyCell(value string) string {
	trimmed := strings.TrimSpace(value)
	for strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`") && len(trimmed) >= 2 {
		trimmed = strings.TrimSpace(strings.Trim(trimmed, "`"))
	}
	return strings.ToLower(trimmed)
}

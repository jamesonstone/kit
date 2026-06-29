package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
)

const featureNotesReferenceName = "Feature notes"

func featureNotesRelPath(featureDirName string) string {
	return filepath.ToSlash(filepath.Join("docs", "notes", featureDirName))
}

func featureNotesPath(projectRoot, featureDirName string) string {
	return filepath.Join(projectRoot, "docs", "notes", featureDirName)
}
func featureNotesPathExists(projectRoot, featureDirName string) bool {
	_, err := os.Stat(featureNotesPath(projectRoot, featureDirName))
	return err == nil
}

func removeFeatureNotesDir(projectRoot, featureDirName string) (string, bool, error) {
	notesPath := featureNotesPath(projectRoot, featureDirName)
	if !featureNotesPathExists(projectRoot, featureDirName) {
		return notesPath, false, nil
	}
	if err := os.RemoveAll(notesPath); err != nil {
		return notesPath, false, fmt.Errorf("failed to remove feature notes directory: %w", err)
	}
	return notesPath, true, nil
}

func featureDesignMaterialsPath(projectRoot, featureDirName string) string {
	return filepath.Join(featureNotesPath(projectRoot, featureDirName), "design")
}

func ensureFeatureNotesDir(projectRoot, featureDirName string) (string, string, error) {
	notesPath := featureNotesPath(projectRoot, featureDirName)
	if err := ensureFeatureNotesScaffold(projectRoot, featureDirName); err != nil {
		return "", "", fmt.Errorf("failed to create feature notes directory: %w", err)
	}

	return notesPath, featureNotesRelPath(featureDirName), nil
}

func ensureFeatureDesignMaterialsDirs(projectRoot, featureDirName string) (string, string, error) {
	designPath := featureDesignMaterialsPath(projectRoot, featureDirName)
	for _, dir := range []string{
		designPath,
		filepath.Join(designPath, "screenshots"),
		filepath.Join(designPath, "references"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create frontend design materials directory: %w", err)
		}
		if err := ensurePlaceholderFile(dir); err != nil {
			return "", "", fmt.Errorf("failed to create frontend design materials placeholder: %w", err)
		}
	}

	return designPath, designMaterialsRelPath(featureDirName), nil
}

func ensurePlaceholderFile(dir string) error {
	gitkeepPath := filepath.Join(dir, ".gitkeep")
	file, err := os.OpenFile(gitkeepPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}

func featureNotesDirName(brainstormPath, fallbackSlug string) string {
	dirName := filepath.Base(filepath.Dir(brainstormPath))
	if dirName == "" || dirName == "." || dirName == string(filepath.Separator) {
		return fallbackSlug
	}
	return dirName
}

func seedBrainstormNotesDependency(content, notesRelPath string) string {
	updated, changed, err := appendBrainstormNotesDependency(content, notesRelPath)
	if err != nil {
		return content
	}
	if !changed {
		return content
	}
	return updated
}

func ensureBrainstormNotesDependency(brainstormPath, notesRelPath string) (bool, error) {
	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", brainstormPath, err)
	}

	updated, changed, err := appendBrainstormNotesDependency(string(content), notesRelPath)
	if err != nil {
		return false, fmt.Errorf("failed to update notes reference in %s: %w", brainstormPath, err)
	}
	if !changed {
		return false, nil
	}

	if err := document.Write(brainstormPath, updated); err != nil {
		return false, fmt.Errorf("failed to update notes reference in %s: %w", brainstormPath, err)
	}

	return true, nil
}

func appendBrainstormNotesDependency(content, notesRelPath string) (string, bool, error) {
	if brainstormNotesDependencyExists(content, notesRelPath) {
		return content, false, nil
	}

	featureMeta := document.FeatureMetadataFromDir(featureDirNameFromNotesRelPath(notesRelPath))
	updated, changed, err := document.UpsertMetadata(content, document.TypeBrainstorm, document.MetadataUpsert{
		Feature:    featureMeta,
		References: referencesForMetadataUpsert(content, document.TypeBrainstorm, []document.MetadataReference{featureNotesReference(notesRelPath)}),
	})
	if err != nil {
		return content, false, err
	}
	return updated, changed, nil
}

func featureNotesReference(notesRelPath string) document.MetadataReference {
	return document.MetadataReference{
		ID:         "feature-notes",
		Name:       featureNotesReferenceName,
		Type:       "notes",
		Target:     notesRelPath,
		Relation:   document.ReferenceRelationInforms,
		ReadPolicy: document.ReferenceReadPolicyConditional,
		UsedFor:    "optional pre-brainstorm research input",
		Status:     document.ReferenceStatusOptional,
	}
}

func featureDirNameFromNotesRelPath(notesRelPath string) string {
	parts := strings.Split(filepath.ToSlash(notesRelPath), "/")
	for i, part := range parts {
		if part == "notes" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func brainstormNotesDependencyExists(content, notesRelPath string) bool {
	doc := document.Parse(content, "", document.TypeBrainstorm)
	if !doc.FrontMatterPresent || doc.Metadata == nil {
		return false
	}
	for _, reference := range doc.Metadata.References {
		if reference.Name == featureNotesReferenceName && reference.Target == notesRelPath {
			return true
		}
	}
	return false
}

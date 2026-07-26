package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func setSpecPhase(specPath, featureDirName, phase string) error {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	doc := document.Parse(string(data), specPath, document.TypeSpec)
	workflowVersion := document.WorkflowVersionV2
	if doc.Metadata != nil && doc.Metadata.WorkflowVersion != 0 {
		workflowVersion = doc.Metadata.WorkflowVersion
	}
	updated, changed, err := document.UpsertMetadata(string(data), document.TypeSpec, document.MetadataUpsert{
		Feature:         document.FeatureMetadataFromDir(featureDirName),
		WorkflowVersion: workflowVersion,
		Phase:           phase,
	})
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return os.WriteFile(specPath, []byte(updated), 0644)
}

func isV2Feature(feat *feature.Feature) bool {
	return workflowVersionForFeature(feat) == document.WorkflowVersionV2
}

func isLivingSpecFeature(feat *feature.Feature) bool {
	version := workflowVersionForFeature(feat)
	return version == document.WorkflowVersionV2 || version == document.WorkflowVersionV3
}

func workflowVersionForFeature(feat *feature.Feature) int {
	if feat == nil {
		return 0
	}
	specPath := filepath.Join(feat.Path, "SPEC.md")
	doc, err := document.ParseFile(specPath, document.TypeSpec)
	if err != nil || doc.Metadata == nil {
		return 0
	}
	return doc.Metadata.WorkflowVersion
}

// appendReflectionMarker appends the REFLECTION_COMPLETE marker to a TASKS.md file.
func appendReflectionMarker(tasksPath string) error {
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return err
	}

	content := string(data)

	// already present
	if strings.Contains(content, feature.ReflectionCompleteMarker) {
		return nil
	}

	// ensure trailing newline before marker
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	content += "\n" + feature.ReflectionCompleteMarker + "\n"

	return os.WriteFile(tasksPath, []byte(content), 0644)
}

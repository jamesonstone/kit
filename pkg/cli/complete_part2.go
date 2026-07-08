package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
)

func markFeaturesComplete(
	out io.Writer,
	errOut io.Writer,
	features []feature.Feature,
	force bool,
	projectRoot string,
	cfg *config.Config,
) error {
	tasksPaths := make([]string, len(features))
	for i := range features {
		tasksPath, err := validateFeatureCanComplete(&features[i], force)
		if err != nil {
			return fmt.Errorf("feature '%s': %w", features[i].Slug, err)
		}
		tasksPaths[i] = tasksPath
	}

	for i := range features {
		if err := markFeatureComplete(&features[i], tasksPaths[i]); err != nil {
			return fmt.Errorf("failed to mark feature '%s' complete: %w", features[i].Slug, err)
		}
		if _, err := fmt.Fprintf(out, "✅ Feature '%s' marked complete\n", features[i].Slug); err != nil {
			return err
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		_, _ = fmt.Fprintf(errOut, "  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
		return printProjectRefreshAdvisory(out, projectRoot, cfg)
	}
	if _, err := fmt.Fprintln(out, "  ✓ Updated PROJECT_PROGRESS_SUMMARY.md"); err != nil {
		return err
	}
	return printProjectRefreshAdvisory(out, projectRoot, cfg)
}

func markFeatureComplete(feat *feature.Feature, path string) error {
	if isV2Feature(feat) {
		return setSpecPhase(path, feat.DirName, string(feature.PhaseComplete))
	}
	return appendReflectionMarker(path)
}

func setSpecPhase(specPath, featureDirName, phase string) error {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	updated, changed, err := document.UpsertMetadata(string(data), document.TypeSpec, document.MetadataUpsert{
		Feature:         document.FeatureMetadataFromDir(featureDirName),
		WorkflowVersion: 2,
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
	if feat == nil {
		return false
	}
	specPath := filepath.Join(feat.Path, "SPEC.md")
	doc, err := document.ParseFile(specPath, document.TypeSpec)
	return err == nil && doc.Metadata != nil && doc.Metadata.WorkflowVersion == 2
}

// appendReflectionMarker appends the REFLECTION_COMPLETE marker to a TASKS.md file.
func appendReflectionMarker(tasksPath string) error {
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return err
	}

	content := string(data)

	if strings.Contains(content, feature.ReflectionCompleteMarker) {
		return nil
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	content += "\n" + feature.ReflectionCompleteMarker + "\n"

	return os.WriteFile(tasksPath, []byte(content), 0644)
}

package cli

import (
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/feature"
)

func loadFeatureWithState(specsDir string, cfg *config.Config, ref string) (*feature.Feature, error) {
	feat, err := feature.Resolve(specsDir, ref)
	if err != nil {
		return nil, err
	}
	feature.ApplyLifecycleState(feat, cfg)
	return feat, nil
}

func isLivingSpecFeature(feat *feature.Feature) bool {
	if feat == nil {
		return false
	}
	doc, err := document.ParseFile(filepath.Join(feat.Path, "SPEC.md"), document.TypeSpec)
	if err != nil || doc.Metadata == nil {
		return false
	}
	return doc.Metadata.WorkflowVersion == document.WorkflowVersionV2 ||
		doc.Metadata.WorkflowVersion == document.WorkflowVersionV3
}

func normalizeSpecAnswer(raw string) string {
	return strings.TrimSpace(strings.ReplaceAll(raw, "\r\n", "\n"))
}

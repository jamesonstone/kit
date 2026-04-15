package feature

import (
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
)

// IsBacklogItem returns true when a feature is explicitly deferred as backlog work
func IsBacklogItem(feat Feature) bool {
	if !feat.Paused || feat.Phase != PhaseBrainstorm {
		return false
	}

	_, err := os.Stat(filepath.Join(feat.Path, "BRAINSTORM.md"))
	return err == nil
}

// ListBacklogFeatures returns paused brainstorm-phase features in numeric order
func ListBacklogFeatures(specsDir string, cfg *config.Config) ([]Feature, error) {
	features, err := ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}

	backlog := make([]Feature, 0, len(features))
	for _, feat := range features {
		if IsBacklogItem(feat) {
			backlog = append(backlog, feat)
		}
	}

	return backlog, nil
}

// FindActiveFeatureWithState returns the newest in-flight non-backlog feature.
func FindActiveFeatureWithState(specsDir string, cfg *config.Config) (*Feature, error) {
	features, err := ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}

	for i := len(features) - 1; i >= 0; i-- {
		if IsBacklogItem(features[i]) {
			continue
		}
		if features[i].Phase == PhaseComplete {
			continue
		}

		active := features[i]
		return &active, nil
	}

	return nil, nil
}

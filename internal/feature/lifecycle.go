package feature

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/config"
)

func ApplyLifecycleState(feat *Feature, cfg *config.Config) {
	if feat == nil {
		return
	}

	feat.Paused = cfg != nil && cfg.IsFeaturePaused(feat.DirName)
}

func ApplyLifecycleStateToFeatures(features []Feature, cfg *config.Config) {
	for i := range features {
		ApplyLifecycleState(&features[i], cfg)
	}
}

func ListFeaturesWithState(specsDir string, cfg *config.Config) ([]Feature, error) {
	features, err := ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	ApplyLifecycleStateToFeatures(features, cfg)
	return features, nil
}

func PersistPaused(projectRoot string, cfg *config.Config, feat *Feature, paused bool) error {
	if feat == nil {
		return fmt.Errorf("feature is required")
	}

	cfg.SetFeaturePaused(feat.DirName, paused)
	if err := config.Save(projectRoot, cfg); err != nil {
		return err
	}

	feat.Paused = paused
	return nil
}

func ClearPersistedState(projectRoot string, cfg *config.Config, feat *Feature) error {
	if feat == nil {
		return fmt.Errorf("feature is required")
	}

	cfg.RemoveFeatureState(feat.DirName)
	if err := config.Save(projectRoot, cfg); err != nil {
		return err
	}

	feat.Paused = false
	return nil
}

package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
)

func clearPausedForExplicitResume(projectRoot string, cfg *config.Config, feat *feature.Feature) error {
	if feat == nil || !feat.Paused {
		return nil
	}

	if err := feature.PersistPaused(projectRoot, cfg, feat, false); err != nil {
		return fmt.Errorf("failed to clear paused state for %s: %w", feat.DirName, err)
	}

	return nil
}

func updateRollupForResume(projectRoot string, cfg *config.Config, dirName string, wasPaused bool) error {
	if !wasPaused {
		return nil
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return fmt.Errorf("failed to refresh PROJECT_PROGRESS_SUMMARY.md after resuming %s: %w", dirName, err)
	}

	return nil
}

func loadFeatureWithState(specsDir string, cfg *config.Config, ref string) (*feature.Feature, error) {
	feat, err := feature.Resolve(specsDir, ref)
	if err != nil {
		return nil, err
	}

	feature.ApplyLifecycleState(feat, cfg)
	return feat, nil
}

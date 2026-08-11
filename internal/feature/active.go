package feature

import "github.com/jamesonstone/kit/internal/config"

// FindActiveFeatureWithState ignores inert legacy lifecycle configuration and
// returns the newest feature whose document phase is not complete.
func FindActiveFeatureWithState(specsDir string, cfg *config.Config) (*Feature, error) {
	features, err := ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	for i := len(features) - 1; i >= 0; i-- {
		if features[i].Phase == PhaseComplete {
			continue
		}
		active := features[i]
		return &active, nil
	}
	return nil, nil
}

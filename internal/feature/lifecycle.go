package feature

import "github.com/jamesonstone/kit/internal/config"

// ApplyLifecycleState remains as a compatibility hook. Legacy lifecycle fields
// are parsed for one major release but no longer alter current feature state.
func ApplyLifecycleState(_ *Feature, _ *config.Config) {}

func ListFeaturesWithState(specsDir string, _ *config.Config) ([]Feature, error) {
	return ListFeatures(specsDir)
}

package cli

import (
	"github.com/jamesonstone/kit/internal/config"
)

type specV2PromptInput struct {
	SpecPath       string
	BrainstormPath string
	FeatureSlug    string
	ProjectRoot    string
	Config         *config.Config
	Answers        *specAnswers
	PromptOnly     bool
	SingleAgent    bool
}

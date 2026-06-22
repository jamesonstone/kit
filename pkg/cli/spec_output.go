package cli

import (
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
)

func outputCompiledPrompt(
	specPath, brainstormPath, featureSlug, projectRoot string,
	cfg *config.Config,
	answers *specAnswers,
	outputOnly bool,
) error {
	prompt := buildSpecV2SupervisorPrompt(specV2PromptInput{
		SpecPath:       specPath,
		BrainstormPath: brainstormPath,
		FeatureSlug:    featureSlug,
		ProjectRoot:    projectRoot,
		Config:         cfg,
		Answers:        answers,
		SingleAgent:    singleAgent,
	})

	preparedPrompt := preparePromptForFeature(prompt, false, filepath.Dir(specPath))
	if err := writePromptWithClipboardDefault(preparedPrompt, outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		printNumberedNextSteps([]string{
			"Paste the copied v2 supervisor prompt into your coding agent",
			"Answer clarification questions until SPEC.md has binary acceptance criteria and mapped validation",
			"Let the supervisor route implementation, reflection, validation/verification, evidence, and delivery gating inside SPEC.md",
		})
	}

	return nil
}

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
	})

	preparedPrompt := preparePromptForFeature(prompt, false, filepath.Dir(specPath))
	if err := writePromptWithClipboardDefault(preparedPrompt, outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		printNumberedNextSteps([]string{
			"Paste the copied prompt into your coding agent",
			"Work with the agent through clarification, implementation, validation, reflection, and delivery gating inside SPEC.md",
		})
	}

	return nil
}

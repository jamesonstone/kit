package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/feature"
)

func outputCatchupPromptForFeature(
	feat *feature.Feature,
	projectRoot string,
	outputOnly bool,
	copy bool,
	currentStep string,
) error {
	status, err := feature.GetFeatureStatus(feat)
	if err != nil {
		return fmt.Errorf("failed to get feature status: %w", err)
	}

	prompt := buildCatchupPrompt(feat, status, projectRoot)
	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, copy); err != nil {
		return err
	}

	if !outputOnly {
		printWorkflowInstructions(currentStep, []string{
			"answer the agent's clarification questions to restore context",
			"approve a move to implementation only when you want coding to begin",
		})
	}

	return nil
}

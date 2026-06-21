package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
)

func runSpecTemplate(
	specPath, brainstormPath, featureSlug, projectRoot string,
	cfg *config.Config,
	outputOnly bool,
	promptOnly bool,
) error {
	prompt := buildSpecTemplatePrompt(specPath, brainstormPath, featureSlug, projectRoot, cfg, promptOnly)

	if !outputOnly {
		fmt.Println()
		fmt.Println(dim + "⚠️ IMPORTANT: Before submitting this prompt, fill in the context section" + reset)
		fmt.Println(dim + "   with your idea, known constraints, acceptance criteria, and delivery intent." + reset)
		fmt.Println(dim + "   Supporting artifacts can be placed in the feature notes/design directories" + reset)
		fmt.Println(dim + "   referenced by the generated prompt." + reset)
		fmt.Println()
		fmt.Println(dim + "   Tip: Run 'kit spec <feature> --interactive' for a guided" + reset)
		fmt.Println(dim + "   editor-first experience, or add '--inline' for terminal multiline entry." + reset)
		fmt.Println()
	}

	preparedPrompt := preparePromptForFeature(prompt, false, filepath.Dir(specPath))
	if err := writePromptWithClipboardDefault(preparedPrompt, outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		printNumberedNextSteps([]string{
			fmt.Sprintf("Review or edit %s as the single durable workflow artifact", specPath),
			"Paste the copied v2 supervisor prompt into your coding agent",
		})
	}

	return nil
}

func buildSpecTemplatePrompt(
	specPath, brainstormPath, featureSlug, projectRoot string,
	cfg *config.Config,
	promptOnly bool,
) string {
	return buildSpecV2SupervisorPrompt(specV2PromptInput{
		SpecPath:       specPath,
		BrainstormPath: brainstormPath,
		FeatureSlug:    featureSlug,
		ProjectRoot:    projectRoot,
		Config:         cfg,
		PromptOnly:     promptOnly,
	})
}

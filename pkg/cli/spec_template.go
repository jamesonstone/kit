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
		fmt.Println(dim + "ℹ️  This prompt uses the current SPEC.md as the durable workflow state." + reset)
		fmt.Println(dim + "   New SPEC.md files are seeded by one thesis/goal editor entry plus delivery intent." + reset)
		fmt.Println(dim + "   Existing SPEC.md files are preserved unless you explicitly pass --revise-thesis." + reset)
		fmt.Println(dim + "   Supporting artifacts can be placed in the feature notes/design directories referenced by the prompt." + reset)
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

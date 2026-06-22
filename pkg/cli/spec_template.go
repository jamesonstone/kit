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
		style := styleForStdout()
		printSectionBanner("🧠", "V2 Supervisor Prompt")
		fmt.Println(style.bullet(style.label("Source of truth:") + " SPEC.md is the durable workflow state."))
		fmt.Println(style.bullet(style.label("New specs:") + " one thesis/goal entry plus delivery intent starts the flow."))
		fmt.Println(style.bullet(style.label("Existing specs:") + " content is preserved unless you pass --revise-thesis."))
		fmt.Println(style.bullet(style.label("Supporting inputs:") + " add files under the referenced notes/design directories."))
		fmt.Println()
	}

	preparedPrompt := preparePromptForFeature(prompt, false, filepath.Dir(specPath))
	if err := writePromptWithClipboardDefault(preparedPrompt, outputOnly, specCopy); err != nil {
		return err
	}
	if !outputOnly {
		printNumberedNextSteps([]string{
			fmt.Sprintf("Review %s as the single durable workflow artifact", specPath),
			"Paste the copied v2 supervisor prompt into your coding agent",
			"Use the clarification loop to fill every unresolved SPEC.md section before implementation",
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
		SingleAgent:    singleAgent,
	})
}

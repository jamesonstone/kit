package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/feature"
)

func promptBrainstormFeatureRef(args []string) (string, error) {

	featureRef := ""
	if len(args) == 1 {
		featureRef = normalizeSpecAnswer(args[0])
	}
	if featureRef == "" {
		rl, err := newMultilineReadline()
		if err != nil {
			return "", fmt.Errorf("failed to initialize readline: %w", err)
		}
		defer closeMultilineReadline(rl)
		style := styleForStdout()
		printSectionBanner("🧠", "Brainstorm Builder")
		fmt.Println(style.muted("Step 1 of 2: Enter a feature/project name."))
		fmt.Println(style.muted("It will be normalized to lowercase kebab-case and must be 5 words or fewer."))
		featureRef = readLineRL(rl)
	}

	if featureRef == "" {
		return "", fmt.Errorf("feature name cannot be empty")
	}

	normalized := feature.NormalizeSlug(featureRef)
	if err := feature.ValidateSlug(normalized); err != nil {
		return "", err
	}

	if normalized != featureRef {
		fmt.Printf(dim+"Using normalized feature slug: %s"+reset+"\n\n", normalized)
	}
	return normalized, nil
}

func promptBrainstormThesis(inputCfg freeTextInputConfig) (string, error) {
	style := styleForStdout()

	fmt.Println()
	fmt.Println(style.muted("Step 2 of 2: Describe the issue or feature in a few sentences."))
	if inputCfg.usesEditor() {
		fmt.Printf("%s\n", style.muted(fmt.Sprintf("A %s will open for this response.", inputCfg.editorLabel())))
		return readEditorText(inputCfg, "brainstorm thesis", false)
	}

	rl, err := newMultilineReadline()
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	fmt.Println(style.muted("Press Enter to submit. Use Shift+Enter or Ctrl+J to insert newlines."))
	fmt.Println(style.muted("Consecutive blank lines are preserved."))
	thesis := readLineRL(rl)
	if thesis == "" {
		return "", fmt.Errorf("brainstorm thesis cannot be empty")
	}

	return thesis, nil
}

package cli

import "fmt"

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

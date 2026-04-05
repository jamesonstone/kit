package cli

import (
	"fmt"
	"io"

	"github.com/chzyer/readline"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func readLineRL(rl *readline.Instance) string {
	line, err := rl.Readline()
	if err != nil {
		if err == readline.ErrInterrupt || err == io.EOF {
			return ""
		}
		return ""
	}
	return normalizeSpecAnswer(line)
}

func runSpecInteractive(
	specPath, brainstormPath string,
	feat *feature.Feature,
	projectRoot string,
	cfg *config.Config,
	inputCfg freeTextInputConfig,
	outputOnly bool,
) error {
	if inputCfg.usesEditor() {
		return runSpecInteractiveWithEditor(
			specPath,
			brainstormPath,
			feat,
			projectRoot,
			cfg,
			inputCfg,
			outputOnly,
		)
	}

	return runSpecInteractiveWithReadline(specPath, brainstormPath, feat, projectRoot, cfg, outputOnly)
}

func runSpecInteractiveWithReadline(
	specPath, brainstormPath string,
	feat *feature.Feature,
	projectRoot string,
	cfg *config.Config,
	outputOnly bool,
) error {
	rl, err := newMultilineReadline()
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	printSectionBanner("📝", "Interactive Spec Builder")
	style := styleForStdout()

	fmt.Println(style.muted("Answer the following questions to generate a complete prompt for your coding agent."))
	fmt.Println(style.muted("Use ←/→ arrow keys to move through your text and correct mistakes."))
	fmt.Println(style.muted("Press Enter to continue; use Shift+Enter or Ctrl+J to add newlines."))
	fmt.Println(style.muted("Consecutive blank lines are preserved."))
	fmt.Println(style.muted("Press Enter on an empty response to skip a question."))
	if document.Exists(brainstormPath) {
		fmt.Println(style.muted("Existing brainstorm research will also be referenced in the generated prompt."))
	}
	fmt.Println()

	rl.SetPrompt(whiteBold + "   > " + reset)

	answers := specAnswers{}

	fmt.Println(spec + "1. PROBLEM" + reset + " - What problem does this feature solve?")
	fmt.Println(dim + "   Example: Users cannot export their data in CSV format" + reset)
	answers.Problem = readLineRL(rl)

	fmt.Println()
	fmt.Println(spec + "2. GOALS" + reset + " - What are the measurable outcomes? (comma-separated)")
	fmt.Println(dim + "   Example: Export completes in <5s, supports 100k+ rows, CSV is RFC-compliant" + reset)
	answers.Goals = readLineRL(rl)

	fmt.Println()
	fmt.Println(spec + "3. NON-GOALS" + reset + " - What is explicitly out of scope?")
	fmt.Println(dim + "   Example: Excel format, scheduled exports, email delivery" + reset)
	answers.NonGoals = readLineRL(rl)

	fmt.Println()
	fmt.Println(spec + "4. USERS" + reset + " - Who will use this feature?")
	fmt.Println(dim + "   Example: Admin users, API consumers, data analysts" + reset)
	answers.Users = readLineRL(rl)

	fmt.Println()
	fmt.Println(spec + "5. REQUIREMENTS" + reset + " - What must be true for this feature to be complete?")
	fmt.Println(dim + "   Example: Must handle Unicode, must include headers, must stream large files" + reset)
	answers.Requirements = readLineRL(rl)

	fmt.Println()
	fmt.Println(spec + "6. ACCEPTANCE" + reset + " - How do we verify the feature works?")
	fmt.Println(dim + "   Example: Unit tests pass, integration tests cover edge cases, manual QA sign-off" + reset)
	answers.Acceptance = readLineRL(rl)

	fmt.Println()
	fmt.Println(spec + "7. EDGE-CASES" + reset + " - What unusual scenarios must be handled?")
	fmt.Println(dim + "   Example: Empty dataset, special characters in data, network timeout during export" + reset)
	answers.EdgeCases = readLineRL(rl)

	fmt.Println()

	return outputCompiledPrompt(specPath, brainstormPath, feat.Slug, projectRoot, cfg, &answers, outputOnly)
}

func runSpecInteractiveWithEditor(
	specPath, brainstormPath string,
	feat *feature.Feature,
	projectRoot string,
	cfg *config.Config,
	inputCfg freeTextInputConfig,
	outputOnly bool,
) error {
	printSectionBanner("📝", "Interactive Spec Builder")
	style := styleForStdout()

	fmt.Println(style.muted("Answer the following questions to generate a complete prompt for your coding agent."))
	fmt.Printf("%s\n", style.muted(fmt.Sprintf("A %s will open for each free-text response.", inputCfg.editorLabel())))
	fmt.Println(style.muted("Save and quit to submit. Quit without save to skip that question."))
	if document.Exists(brainstormPath) {
		fmt.Println(style.muted("Existing brainstorm research will also be referenced in the generated prompt."))
	}
	fmt.Println()

	answers := specAnswers{}
	var err error

	fmt.Println(spec + "1. PROBLEM" + reset + " - What problem does this feature solve?")
	fmt.Println(dim + "   Example: Users cannot export their data in CSV format" + reset)
	answers.Problem, err = readEditorText(inputCfg, "problem", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "2. GOALS" + reset + " - What are the measurable outcomes? (comma-separated)")
	fmt.Println(dim + "   Example: Export completes in <5s, supports 100k+ rows, CSV is RFC-compliant" + reset)
	answers.Goals, err = readEditorText(inputCfg, "goals", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "3. NON-GOALS" + reset + " - What is explicitly out of scope?")
	fmt.Println(dim + "   Example: Excel format, scheduled exports, email delivery" + reset)
	answers.NonGoals, err = readEditorText(inputCfg, "non-goals", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "4. USERS" + reset + " - Who will use this feature?")
	fmt.Println(dim + "   Example: Admin users, API consumers, data analysts" + reset)
	answers.Users, err = readEditorText(inputCfg, "users", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "5. REQUIREMENTS" + reset + " - What must be true for this feature to be complete?")
	fmt.Println(dim + "   Example: Must handle Unicode, must include headers, must stream large files" + reset)
	answers.Requirements, err = readEditorText(inputCfg, "requirements", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "6. ACCEPTANCE" + reset + " - How do we verify the feature works?")
	fmt.Println(dim + "   Example: Unit tests pass, integration tests cover edge cases, manual QA sign-off" + reset)
	answers.Acceptance, err = readEditorText(inputCfg, "acceptance", true)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(spec + "7. EDGE-CASES" + reset + " - What unusual scenarios must be handled?")
	fmt.Println(dim + "   Example: Empty dataset, special characters in data, network timeout during export" + reset)
	answers.EdgeCases, err = readEditorText(inputCfg, "edge-cases", true)
	if err != nil {
		return err
	}

	fmt.Println()

	return outputCompiledPrompt(specPath, brainstormPath, feat.Slug, projectRoot, cfg, &answers, outputOnly)
}

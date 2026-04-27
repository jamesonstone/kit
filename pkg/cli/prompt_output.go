package cli

import (
	"fmt"
	"strings"
)

var clipboardCopyFunc = copyToClipboard

func formatAgentInstructionBlock(prompt string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(prompt)
	if !strings.HasSuffix(prompt, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("---\n")
	return sb.String()
}

func outputPrompt(prompt string, outputOnly, copy bool) error {
	return writePrompt(prepareAgentPrompt(prompt), outputOnly, copy)
}

func outputPromptWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	return writePromptWithClipboardDefault(prepareAgentPrompt(prompt), outputOnly, copy)
}

func outputPromptForFeatureWithClipboardDefault(prompt, featurePath string, outputOnly, copy bool) error {
	return writePromptWithClipboardDefault(prepareAgentPromptForFeature(prompt, featurePath), outputOnly, copy)
}

func outputPromptWithoutSubagentsWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	return writePromptWithClipboardDefault(preparePromptWithoutSubagents(prompt), outputOnly, copy)
}

func writePrompt(prompt string, outputOnly, copy bool) error {
	if copy {
		if err := clipboardCopyFunc(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		if outputOnly {
			fmt.Print(prompt)
			return nil
		}
		fmt.Println("Copied agent instructions to clipboard.")
		return nil
	}
	if outputOnly {
		fmt.Print(prompt)
		return nil
	}

	fmt.Println("Copy this section to the Agent:")
	fmt.Print(formatAgentInstructionBlock(prompt))
	return nil
}

func writePromptWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	shouldCopy := !outputOnly || copy
	if shouldCopy {
		if err := clipboardCopyFunc(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
	}

	if outputOnly {
		fmt.Print(prompt)
		return nil
	}

	fmt.Println(styleForStdout().clipboardAcknowledgement())
	return nil
}

func printWorkflowInstructions(currentStep string, nextSteps []string) {
	style := styleForStdout()

	fmt.Println(style.title("🧭", "Workflow"))
	if divider := style.sectionDivider(); divider != "" {
		fmt.Println(divider)
	}
	fmt.Println(style.muted("Pipeline: [optional brainstorm] -> spec -> plan -> tasks -> implement -> reflect"))
	fmt.Println()
	fmt.Println(style.currentStepLine(currentStep))
	if len(nextSteps) > 0 {
		fmt.Println()
		fmt.Println(style.nextStepsTitle())
		for _, step := range nextSteps {
			fmt.Printf("  %s\n", style.bullet(step))
		}
	}
	fmt.Println()
}

package cli

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/promptlib"
)

const codingAgentPromptPrefix = "---\n"

func outputPromptLibraryPrompt(prompt promptlib.EffectivePrompt, outputOnly, copy bool) error {
	body, err := renderPromptLibraryBody(prompt)
	if err != nil {
		return err
	}

	shouldCopy := !outputOnly || copy
	if shouldCopy {
		if err := clipboardCopyFunc(body); err != nil {
			return fmt.Errorf("failed to copy prompt to clipboard: %w", err)
		}
	}

	if outputOnly {
		fmt.Print(body)
		return nil
	}

	style := styleForStdout()
	fmt.Println(style.clipboardAcknowledgement())
	fmt.Println(style.title("📚", "Prompt Library"))
	fmt.Printf("Command: %s\n", prompt.CommandName())
	fmt.Printf("Origin: %s\n", promptLibraryOrigin(prompt))
	if shadow := prompt.ShadowSummary(); shadow != "" {
		fmt.Printf("Overrides: %s\n", shadow)
	} else {
		fmt.Println("Overrides: none")
	}
	fmt.Println()
	fmt.Println("Prompt:")
	fmt.Print(formatPromptLibraryBodyForDisplay(prompt, body))
	return nil
}

func renderPromptLibraryBody(prompt promptlib.EffectivePrompt) (string, error) {
	var (
		body string
		err  error
	)

	if prompt.Prompt.Render != nil {
		body, err = prompt.Prompt.Render()
		if err != nil {
			return "", err
		}
	} else {
		body = prompt.Prompt.Content
	}

	if strings.TrimSpace(body) == "" {
		return "", fmt.Errorf("prompt %q has empty content", prompt.CommandName())
	}
	return applyPromptLibraryOutputConventions(prompt, body), nil
}

func applyPromptLibraryOutputConventions(prompt promptlib.EffectivePrompt, body string) string {
	if !isCodingAgentPrompt(prompt) {
		return body
	}
	if strings.HasPrefix(body, codingAgentPromptPrefix) {
		return body
	}
	return codingAgentPromptPrefix + body
}

func formatPromptLibraryBodyForDisplay(prompt promptlib.EffectivePrompt, body string) string {
	if isCodingAgentPrompt(prompt) && strings.HasPrefix(body, codingAgentPromptPrefix) {
		return formatAgentInstructionBlock(strings.TrimPrefix(body, codingAgentPromptPrefix))
	}
	return formatAgentInstructionBlock(body)
}

func isCodingAgentPrompt(prompt promptlib.EffectivePrompt) bool {
	return prompt.Prompt.Identity.Noun == "coding-agent"
}

func promptLibraryOrigin(prompt promptlib.EffectivePrompt) string {
	if prompt.Location == "" {
		return string(prompt.Kind)
	}
	return fmt.Sprintf("%s (%s)", prompt.Kind, prompt.Location)
}

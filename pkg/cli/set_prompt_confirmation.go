package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/promptlib"
)

func confirmGlobalPromptSave(reader *bufio.Reader) (bool, error) {
	return readPromptConfirmation(reader, "No Kit project found. Save this prompt globally? [y/N]: ")
}

func confirmSetPromptOverwrites(
	identity promptlib.Identity,
	targets []setPromptTarget,
	reader *bufio.Reader,
) ([]setPromptTarget, error) {
	confirmed := make([]setPromptTarget, 0, len(targets))
	for _, target := range targets {
		if !target.Exists {
			confirmed = append(confirmed, target)
			continue
		}

		ok, err := readPromptConfirmation(
			reader,
			fmt.Sprintf("Overwrite existing %s prompt %s? [y/N]: ", target.Scope, identity.CommandName()),
		)
		if err != nil {
			return nil, err
		}
		if ok {
			confirmed = append(confirmed, target)
		}
	}
	return confirmed, nil
}

func readPromptConfirmation(reader *bufio.Reader, prompt string) (bool, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "y" || answer == "yes", nil
}

func saveSetPromptTarget(target setPromptTarget, identity promptlib.Identity, prompt config.Prompt) error {
	switch target.Scope {
	case promptlib.SourceLocal:
		return config.UpsertLocalPrompt(target.ProjectRoot, identity.Noun, identity.Verb, prompt)
	case promptlib.SourceGlobal:
		return config.UpsertGlobalPrompt(identity.Noun, identity.Verb, prompt)
	default:
		return fmt.Errorf("unsupported prompt scope %q", target.Scope)
	}
}

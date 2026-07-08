package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/promptlib"
)

func resolveSetPromptTargets(
	identity promptlib.Identity,
	localScope bool,
	globalScope bool,
	reader *bufio.Reader,
) ([]setPromptTarget, error) {
	projectRoot, hasProject, err := config.FindProjectRootOptional()
	if err != nil {
		return nil, err
	}

	if localScope && !hasProject {
		return nil, fmt.Errorf("--local requires a Kit project .kit.yaml; run from a Kit project or use --global")
	}

	if !localScope && !globalScope {
		if hasProject {
			localScope = true
		} else {
			saveGlobal, err := confirmGlobalPromptSave(reader)
			if err != nil {
				return nil, err
			}
			if !saveGlobal {
				fmt.Println("No prompt saved.")
				return nil, nil
			}
			globalScope = true
		}
	}

	var targets []setPromptTarget
	if localScope {
		target, err := localSetPromptTarget(projectRoot, identity)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	if globalScope {
		target, err := globalSetPromptTarget(identity)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	return targets, nil
}

func localSetPromptTarget(projectRoot string, identity promptlib.Identity) (setPromptTarget, error) {
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return setPromptTarget{}, err
	}

	exists, description := promptExistsInConfig(cfg, identity)
	return setPromptTarget{
		Scope:       promptlib.SourceLocal,
		Location:    filepath.Join(projectRoot, config.ConfigFileName),
		ProjectRoot: projectRoot,
		Exists:      exists,
		Description: description,
	}, nil
}

func globalSetPromptTarget(identity promptlib.Identity) (setPromptTarget, error) {
	location, err := config.GlobalConfigPath()
	if err != nil {
		return setPromptTarget{}, err
	}

	cfg, _, err := config.LoadGlobal()
	if err != nil {
		return setPromptTarget{}, err
	}

	exists, description := promptExistsInConfig(cfg, identity)
	return setPromptTarget{
		Scope:       promptlib.SourceGlobal,
		Location:    location,
		Exists:      exists,
		Description: description,
	}, nil
}

func promptExistsInConfig(cfg *config.Config, identity promptlib.Identity) (bool, string) {
	if cfg == nil || cfg.Prompts == nil {
		return false, ""
	}
	verbs, ok := cfg.Prompts[identity.Noun]
	if !ok {
		return false, ""
	}
	prompt, ok := verbs[identity.Verb]
	if !ok {
		return false, ""
	}
	return true, prompt.Description
}

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

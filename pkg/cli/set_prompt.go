package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/promptlib"
)

var setPromptLocal bool
var setPromptGlobal bool

var setPromptCmd = &cobra.Command{
	Use:   "prompt [noun] [verb]",
	Short: "Create or update a local or global prompt",
	Long: `Create or update a reusable prompt through the editor.

With no arguments, prompts for noun, verb, optional description, scope, and
content. With noun and verb, opens the editor for prompt content directly.`,
	Args: cobra.MaximumNArgs(2),
	RunE: runSetPrompt,
}

type setPromptInput struct {
	Identity       promptlib.Identity
	Content        string
	Description    string
	DescriptionSet bool
}

type setPromptTarget struct {
	Scope       promptlib.SourceKind
	Location    string
	ProjectRoot string
	Exists      bool
	Description string
}

func init() {
	setPromptCmd.Flags().BoolVar(&setPromptLocal, "local", false, "save prompt to the project .kit.yaml")
	setPromptCmd.Flags().BoolVar(&setPromptGlobal, "global", false, "save prompt to the user global .kit.yaml")
	setCmd.AddCommand(setPromptCmd)
}

func runSetPrompt(cmd *cobra.Command, args []string) error {
	localOnly, _ := cmd.Flags().GetBool("local")
	globalOnly, _ := cmd.Flags().GetBool("global")
	return runSetPromptWithOptions(args, localOnly, globalOnly)
}

func runSetPromptWithOptions(args []string, localScope, globalScope bool) error {
	reader := bufio.NewReader(os.Stdin)

	input, err := collectSetPromptInput(args, reader)
	if err != nil {
		return err
	}

	targets, err := resolveSetPromptTargets(input.Identity, localScope, globalScope, reader)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}

	targets, err = confirmSetPromptOverwrites(input.Identity, targets, reader)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		fmt.Println("No prompt scopes selected. Nothing was changed.")
		return nil
	}

	content, err := readEditorText(promptDefaultEditorConfig(), "prompt content", false)
	if err != nil {
		return err
	}
	input.Content = content

	for _, target := range targets {
		prompt := config.Prompt{
			Content:     input.Content,
			Description: input.Description,
		}
		if !input.DescriptionSet {
			prompt.Description = target.Description
		}

		if err := saveSetPromptTarget(target, input.Identity, prompt); err != nil {
			return err
		}
		fmt.Printf("✓ Saved prompt %s to %s (%s)\n", input.Identity.CommandName(), target.Scope, target.Location)
	}

	return nil
}

func collectSetPromptInput(args []string, reader *bufio.Reader) (setPromptInput, error) {
	switch len(args) {
	case 0:
		return collectSetPromptWizardInput(reader)
	case 2:
		identity, err := normalizeSetPromptIdentity(args[0], args[1])
		if err != nil {
			return setPromptInput{}, err
		}
		return setPromptInput{Identity: identity}, nil
	default:
		return setPromptInput{}, fmt.Errorf("kit set prompt requires both noun and verb when arguments are provided")
	}
}

func collectSetPromptWizardInput(reader *bufio.Reader) (setPromptInput, error) {
	noun, err := readPromptLine(reader, "Prompt noun")
	if err != nil {
		return setPromptInput{}, err
	}
	verb, err := readPromptLine(reader, "Prompt verb")
	if err != nil {
		return setPromptInput{}, err
	}
	description, err := readOptionalPromptLine(reader, "Description (optional)")
	if err != nil {
		return setPromptInput{}, err
	}

	identity, err := normalizeSetPromptIdentity(noun, verb)
	if err != nil {
		return setPromptInput{}, err
	}

	return setPromptInput{
		Identity:       identity,
		Description:    description,
		DescriptionSet: true,
	}, nil
}

func normalizeSetPromptIdentity(noun, verb string) (promptlib.Identity, error) {
	identity, err := promptlib.NormalizeIdentity(noun, verb)
	if err != nil {
		return promptlib.Identity{}, err
	}
	if identity.Noun == "list" {
		return promptlib.Identity{}, fmt.Errorf("prompt noun %q is reserved for `kit prompt list`; choose a different noun", identity.Noun)
	}
	return identity, nil
}

func readPromptLine(reader *bufio.Reader, label string) (string, error) {
	value, err := readOptionalPromptLine(reader, label)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("%s cannot be empty", strings.ToLower(label))
	}
	return value, nil
}

func readOptionalPromptLine(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read %s: %w", strings.ToLower(label), err)
	}
	return strings.TrimSpace(input), nil
}

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

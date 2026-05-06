package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/promptlib"
)

var promptCopy bool
var promptOutputOnly bool

var promptCmd = &cobra.Command{
	Use:   "prompt [noun] [verb]",
	Short: "Resolve reusable prompts from built-in, global, and local prompt libraries",
	Long: `Resolve reusable prompts by noun and verb.

With no arguments, prompts for a noun and then a verb.
With one argument, prompts for a verb under that noun.
With two arguments, resolves the prompt directly.

Prompts are resolved by precedence:
  local project .kit.yaml > global ~/.config/kit/.kit.yaml > built-in Kit prompts`,
	Args: cobra.MaximumNArgs(2),
	RunE: runPrompt,
}

func init() {
	promptCmd.Flags().BoolVar(&promptCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	promptCmd.Flags().BoolVar(&promptOutputOnly, "output-only", false, "output raw prompt text and skip default clipboard copy")
	rootCmd.AddCommand(promptCmd)
}

func runPrompt(cmd *cobra.Command, args []string) error {
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	copyPrompt, _ := cmd.Flags().GetBool("copy")
	return runPromptWithOptions(args, outputOnly, copyPrompt)
}

func runPromptWithOptions(args []string, outputOnly, copyPrompt bool) error {
	prompts, err := loadPromptLibrary()
	if err != nil {
		return err
	}
	if len(prompts) == 0 {
		return fmt.Errorf("no prompts are available")
	}

	selected, err := selectPrompt(args, prompts)
	if err != nil {
		return err
	}

	return outputPromptLibraryPrompt(selected, outputOnly, copyPrompt)
}

func loadPromptLibrary() ([]promptlib.EffectivePrompt, error) {
	sources := builtInPromptSources()

	globalConfigPath, err := config.GlobalConfigPath()
	if err != nil {
		return nil, err
	}
	globalCfg, _, err := config.LoadGlobal()
	if err != nil {
		return nil, err
	}
	sources = append(sources, promptlib.SourceFromConfig(promptlib.SourceGlobal, globalConfigPath, globalCfg))

	projectRoot, found, err := config.FindProjectRootOptional()
	if err != nil {
		return nil, err
	}
	if found {
		localCfg, err := config.Load(projectRoot)
		if err != nil {
			return nil, err
		}
		sources = append(
			sources,
			promptlib.SourceFromConfig(promptlib.SourceLocal, filepath.Join(projectRoot, config.ConfigFileName), localCfg),
		)
	}

	return promptlib.Merge(sources...)
}

func selectPrompt(args []string, prompts []promptlib.EffectivePrompt) (promptlib.EffectivePrompt, error) {
	reader := bufio.NewReader(os.Stdin)
	switch len(args) {
	case 0:
		noun, err := selectPromptNoun(reader, prompts)
		if err != nil {
			return promptlib.EffectivePrompt{}, err
		}
		verb, err := selectPromptVerb(reader, prompts, noun)
		if err != nil {
			return promptlib.EffectivePrompt{}, err
		}
		return promptlib.Resolve(prompts, noun, verb)
	case 1:
		if strings.EqualFold(args[0], "list") {
			return promptlib.EffectivePrompt{}, fmt.Errorf("use `kit prompt list` to list available prompts")
		}
		noun, err := promptlib.NormalizePart(args[0], "noun")
		if err != nil {
			return promptlib.EffectivePrompt{}, err
		}
		verb, err := selectPromptVerb(reader, prompts, noun)
		if err != nil {
			return promptlib.EffectivePrompt{}, err
		}
		return promptlib.Resolve(prompts, noun, verb)
	default:
		return promptlib.Resolve(prompts, args[0], args[1])
	}
}

func selectPromptNoun(reader *bufio.Reader, prompts []promptlib.EffectivePrompt) (string, error) {
	nouns := promptlib.Nouns(prompts)
	if len(nouns) == 0 {
		return "", fmt.Errorf("no prompt nouns are available")
	}

	printSelectionHeader("Select a prompt noun:")
	for i, noun := range nouns {
		fmt.Printf("  [%d] %s\n", i+1, noun)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

	selection, err := readPromptSelectionFrom(reader, len(nouns))
	if err != nil {
		return "", err
	}
	return nouns[selection-1], nil
}

func selectPromptVerb(reader *bufio.Reader, prompts []promptlib.EffectivePrompt, noun string) (string, error) {
	verbs := promptlib.VerbsForNoun(prompts, noun)
	if len(verbs) == 0 {
		return "", promptlib.NoMatchError{
			Identity: promptlib.Identity{Noun: noun},
			Nouns:    promptlib.Nouns(prompts),
		}
	}

	printSelectionHeader(fmt.Sprintf("Select a prompt verb for %s:", noun))
	for i, verb := range verbs {
		description := promptDescription(prompts, noun, verb)
		if description == "" {
			fmt.Printf("  [%d] %s\n", i+1, verb)
			continue
		}
		fmt.Printf("  [%d] %s — %s\n", i+1, verb, description)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

	selection, err := readPromptSelectionFrom(reader, len(verbs))
	if err != nil {
		return "", err
	}
	return verbs[selection-1], nil
}

func promptDescription(prompts []promptlib.EffectivePrompt, noun, verb string) string {
	for _, prompt := range prompts {
		identity := prompt.Prompt.Identity
		if identity.Noun == noun && identity.Verb == verb {
			return prompt.Prompt.Description
		}
	}
	return ""
}

func readPromptSelection(max int) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	return readPromptSelectionFrom(reader, max)
}

func readPromptSelectionFrom(reader *bufio.Reader, max int) (int, error) {
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("failed to read selection: %w", err)
	}
	input = strings.TrimSpace(input)

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > max {
		return 0, fmt.Errorf("invalid selection: %s", input)
	}
	return selection, nil
}

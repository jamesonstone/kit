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
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

var skillCopy bool
var skillOutputOnly bool

func init() {
	rootCmd.AddCommand(newSkillRootCommand("skill", []string{"skills"}))
	rootCmd.AddCommand(newSkillRootCommand("skills", []string{"skill"}))
}

func newSkillRootCommand(use string, aliases []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   "Output skill extraction prompts for completed features",
		Long: `Output a structured prompt that tells the active coding agent how to
mine a reusable skill from a feature's spec pipeline and implemented delta.

The command never writes a skill directly. It only outputs the prompt.

Commands:
  kit skill mine [feature]
  kit skills mine [feature]`,
	}

	mineCmd := &cobra.Command{
		Use:   "mine [feature]",
		Short: "Output skill extraction prompt for the active coding agent",
		Long: `Output a structured markdown prompt that instructs the active coding
agent to analyze a feature's documents, compare planned work to implemented
work, and draft a reusable SKILL.md only when a real cross-feature pattern
exists.

If no feature is specified, shows an interactive selection of features that
have reached at least TASKS.md.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runSkillMine,
	}

	mineCmd.Flags().BoolVar(&skillCopy, "copy", false, "copy agent prompt to clipboard")
	mineCmd.Flags().BoolVar(
		&skillOutputOnly,
		"output-only",
		false,
		"output prompt only, suppressing status messages",
	)

	cmd.AddCommand(mineCmd)

	return cmd
}

func runSkillMine(cmd *cobra.Command, args []string) error {
	outputOnly, _ := cmd.Flags().GetBool("output-only")

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := cfg.SpecsPath(projectRoot)
	skillsDir := cfg.SkillsPath(projectRoot)

	var feat *feature.Feature

	if len(args) == 1 {
		feat, err = feature.Resolve(specsDir, args[0])
		if err != nil {
			return fmt.Errorf("failed to resolve feature: %w", err)
		}
	} else {
		feat, err = selectFeatureForSkillMine(specsDir)
		if err != nil {
			return err
		}
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")

	if !document.Exists(tasksPath) {
		return fmt.Errorf(
			"TASKS.md not found. Run 'kit tasks %s' first",
			feat.Slug,
		)
	}
	if !document.Exists(specPath) {
		return fmt.Errorf(
			"SPEC.md not found for %s. Run 'kit spec %s' first",
			feat.Slug,
			feat.Slug,
		)
	}
	if !document.Exists(planPath) {
		return fmt.Errorf(
			"PLAN.md not found for %s. Run 'kit plan %s' first",
			feat.Slug,
			feat.Slug,
		)
	}

	prompt := buildSkillMinePrompt(
		feat,
		brainstormPath,
		specPath,
		planPath,
		tasksPath,
		skillsDir,
		projectRoot,
	)

	if err := outputPrompt(prompt, outputOnly, skillCopy); err != nil {
		return err
	}

	if !outputOnly {
		printWorkflowInstructions("skill mine (post-reflect step)", []string{
			"review the generated SKILL.md draft before committing",
			"run kit skill mine again on future features to grow the skills library",
		})
	}

	return nil
}

// selectFeatureForSkillMine shows an interactive numbered list of features
// that have TASKS.md and includes the current phase label.
func selectFeatureForSkillMine(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "TASKS.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf(
			"no features ready for skill mining (need TASKS.md)\n\nRun 'kit tasks <feature>' to create tasks first",
		)
	}

	fmt.Println()
	fmt.Println(whiteBold + "Select a feature to mine a skill from:" + reset)
	fmt.Println()
	for i, f := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
	}
	fmt.Println()
	fmt.Print(whiteBold + "Enter number: " + reset)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(candidates) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := candidates[num-1]
	return &selected, nil
}

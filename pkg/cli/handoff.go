// package cli implements the Kit command-line interface.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/spf13/cobra"
)

var handoffCopy bool
var handoffOutputOnly bool

var handoffCmd = &cobra.Command{
	Use:   "handoff [feature]",
	Short: "Output a documentation-sync handoff prompt for the current coding agent session",
	Long: `Output instructions for the current coding agent session to reconcile
feature documentation with implementation reality before transfer.

Use this when switching between agents or sessions and you want the current
session to leave behind accurate docs plus a concise handoff summary.

Without a feature argument, shows an interactive selector with:
  - numbered features from docs/specs/
  - [0] no specific feature for a project-wide handoff
With a feature argument, outputs a feature-scoped handoff-preparation prompt.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHandoff,
}

func init() {
	handoffCmd.Flags().BoolVarP(&handoffCopy, "copy", "c", false, "copy output to clipboard even with --output-only")
	handoffCmd.Flags().BoolVar(&handoffOutputOnly, "output-only", false, "output text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(handoffCmd)
	rootCmd.AddCommand(handoffCmd)
}

func runHandoff(cmd *cobra.Command, args []string) error {
	var output string
	var featurePath string
	var err error

	if len(args) == 1 {
		output, featurePath, err = featureHandoffWithPath(args[0])
	} else {
		output, featurePath, err = interactiveHandoffWithPath()
	}

	if err != nil {
		return err
	}
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	if !outputOnly {
		printWorkflowInstructions("handoff (supporting step)", []string{
			"run the generated prompt in the current coding agent session",
			"let it reconcile docs before transferring work",
		})
	}

	if featurePath != "" {
		return outputPromptForFeatureWithClipboardDefault(output, featurePath, outputOnly, handoffCopy)
	}
	return outputPromptWithClipboardDefault(output, outputOnly, handoffCopy)
}

func interactiveHandoff() (string, error) {
	output, _, err := interactiveHandoffWithPath()
	return output, err
}

func interactiveHandoffWithPath() (string, string, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return genericHandoffInstructions(), "", nil
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", "", fmt.Errorf("failed to load config: %w", err)
	}

	feat, projectWide, err := selectFeatureForHandoff(cfg.SpecsPath(projectRoot))
	if err != nil {
		return "", "", err
	}

	if projectWide || feat == nil {
		output, err := projectHandoffWithConfig(projectRoot, cfg)
		return output, "", err
	}

	return featureHandoffWithPath(feat.Slug)
}

func selectFeatureForHandoff(specsDir string) (*feature.Feature, bool, error) {
	return selectFeatureForHandoffWithIO(specsDir, os.Stdin, os.Stdout)
}

func selectFeatureForHandoffWithIO(
	specsDir string,
	input io.Reader,
	output io.Writer,
) (*feature.Feature, bool, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, false, err
	}

	if len(features) == 0 {
		return nil, true, nil
	}

	printSelectionHeaderTo(output, "Select a feature to hand off:")
	for i, f := range features {
		fmt.Fprintf(output, "  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
	}
	fmt.Fprintln(output, "  [0] no specific feature (project-wide handoff)")
	fmt.Fprintln(output)
	fmt.Fprint(output, selectionPrompt(output))

	reader := bufio.NewReader(input)
	selection, err := reader.ReadString('\n')
	if err != nil {
		return nil, false, fmt.Errorf("failed to read selection: %w", err)
	}
	selection = strings.TrimSpace(selection)

	choice, err := strconv.Atoi(selection)
	if err != nil || choice < 0 || choice > len(features) {
		return nil, false, fmt.Errorf("invalid selection: %s", selection)
	}

	if choice == 0 {
		return nil, true, nil
	}

	selected := features[choice-1]
	return &selected, false, nil
}

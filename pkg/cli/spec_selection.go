package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func selectFeatureForSpec(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		brainstormPath := filepath.Join(f.Path, "BRAINSTORM.md")
		specPath := filepath.Join(f.Path, "SPEC.md")
		if document.Exists(brainstormPath) || document.Exists(specPath) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no brainstorms or specifications found\n\nRun 'kit brainstorm' or 'kit spec <feature>' to start a feature")
	}

	printSelectionHeader("Select a feature to continue into spec:")
	for i, f := range candidates {
		label := f.DirName
		if document.Exists(filepath.Join(f.Path, "BRAINSTORM.md")) &&
			!document.Exists(filepath.Join(f.Path, "SPEC.md")) {
			label += " (brainstorm)"
		}
		fmt.Printf("  [%d] %s\n", i+1, label)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

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

func selectFeatureForSpecPromptOnly(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "SPEC.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no specifications available to regenerate prompts for\n\nRun 'kit spec <feature>' first")
	}

	printSelectionHeader("Select a feature to regenerate the spec prompt for:")
	for i, f := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

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

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

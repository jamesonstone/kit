package cli

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

func resolveNotesFeature(projectRoot string, cfg *config.Config, specsDir string, args []string, reader *bufio.Reader, out io.Writer) (*feature.Feature, bool, error) {
	if len(args) == 1 {
		return feature.EnsureExists(cfg, projectRoot, specsDir, args[0])
	}
	return selectOrCreateFeatureForNotes(projectRoot, cfg, specsDir, reader, out)
}

func selectOrCreateFeatureForNotes(projectRoot string, cfg *config.Config, specsDir string, reader *bufio.Reader, out io.Writer) (*feature.Feature, bool, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, false, err
	}
	if len(features) == 0 {
		_, _ = fmt.Fprintln(out, "No feature directories found.")
		return promptNewNotesFeature(projectRoot, cfg, specsDir, reader, out)
	}

	printSelectionHeaderTo(out, "Select a feature notes directory:")
	for i, feat := range features {
		pausedSuffix := ""
		if feat.Paused {
			pausedSuffix = ", paused"
		}
		_, _ = fmt.Fprintf(out, "  [%d] %s (%s%s)\n", i+1, feat.DirName, feat.Phase, pausedSuffix)
	}
	_, _ = fmt.Fprintln(out, "  [n] Create a new feature notes directory")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprint(out, selectionPrompt(out))

	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)
	if selection == "n" || selection == "N" {
		return promptNewNotesFeature(projectRoot, cfg, specsDir, reader, out)
	}
	num, err := strconv.Atoi(selection)
	if err != nil || num < 1 || num > len(features) {
		return nil, false, fmt.Errorf("invalid selection: %s", selection)
	}
	selected := features[num-1]
	return &selected, false, nil
}

func promptNewNotesFeature(projectRoot string, cfg *config.Config, specsDir string, reader *bufio.Reader, out io.Writer) (*feature.Feature, bool, error) {
	_, _ = fmt.Fprint(out, "Feature name: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, false, fmt.Errorf("feature name cannot be empty")
	}
	return feature.EnsureExists(cfg, projectRoot, specsDir, input)
}

func promptNotesAction(reader *bufio.Reader, out io.Writer, options *notesOptions) error {
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Choose notes action:")
	_, _ = fmt.Fprintln(out, "  [1] Ensure/show notes scaffold")
	_, _ = fmt.Fprintln(out, "  [2] Add inbox note")
	_, _ = fmt.Fprintln(out, "  [3] Add reference note")
	_, _ = fmt.Fprintln(out, "  [4] Add response note")
	_, _ = fmt.Fprintln(out, "  [5] Add private note")
	_, _ = fmt.Fprintln(out, "  [6] Copy notes path")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprint(out, selectionPrompt(out))

	selection, _ := reader.ReadString('\n')
	switch strings.TrimSpace(selection) {
	case "", "1":
		return nil
	case "2":
		options.add = true
		options.section = "inbox"
	case "3":
		options.add = true
		options.section = "references"
	case "4":
		options.add = true
		options.section = "responses"
	case "5":
		options.add = true
		options.private = true
	case "6":
		options.copyPath = true
	default:
		return fmt.Errorf("invalid selection: %s", strings.TrimSpace(selection))
	}
	return nil
}

func promptNoteMetadata(cmd *cobra.Command, reader *bufio.Reader, out io.Writer, options *notesOptions) error {
	if !cmd.Flags().Changed("title") {
		options.title = promptStringDefault(reader, out, "Title", defaultNoteTitle(*options))
	}
	if !cmd.Flags().Changed("source") {
		options.source = promptStringDefault(reader, out, "Source", options.source)
	}
	if !cmd.Flags().Changed("status") {
		options.status = promptStringDefault(reader, out, "Status", options.status)
	}
	if !cmd.Flags().Changed("sensitivity") && !options.private {
		options.sensitivity = promptStringDefault(reader, out, "Sensitivity", options.sensitivity)
	}
	return nil
}

func promptStringDefault(reader *bufio.Reader, out io.Writer, label, fallback string) string {
	_, _ = fmt.Fprintf(out, "%s [%s]: ", label, fallback)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return fallback
	}
	return input
}

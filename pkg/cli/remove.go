package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
)

var removeYes bool
var removeNotes bool

var removeCmd = &cobra.Command{
	Use:     "rm [feature]",
	Aliases: []string{"remove"},
	Short:   "Remove a feature and all its docs",
	Long: `Remove a feature directory, all files under it, and its persisted
lifecycle state.

If no feature is specified, shows an interactive selection of existing
features.

Feature notes under docs/notes/<feature> are retained by default for follow-up
work. Use --notes or answer the interactive notes prompt to remove them too.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&removeYes, "yes", "y", false, "skip the deletion confirmation prompt")
	removeCmd.Flags().BoolVar(&removeNotes, "notes", false, "also remove retained notes under docs/notes/<feature>")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	var feat *feature.Feature
	if len(args) == 0 {
		feat, err = selectFeatureForRemove(projectRoot, specsDir, cfg)
		if err != nil {
			return err
		}
	} else {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found: %w", args[0], err)
		}
	}

	var reader *bufio.Reader
	if !removeYes {
		reader = bufio.NewReader(os.Stdin)
		confirmed, err := confirmFeatureRemovalWithReader(feat, reader)
		if err != nil {
			return err
		}
		if !confirmed {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Removal canceled.")
			return err
		}
	}
	removeFeatureNotes := removeNotes
	if !removeNotes && !removeYes && featureNotesPathExists(projectRoot, feat.DirName) {
		removeFeatureNotes, err = confirmFeatureNotesRemovalWithReader(feat, featureNotesRelPath(feat.DirName), reader)
		if err != nil {
			return err
		}
	}

	if err := os.RemoveAll(feat.Path); err != nil {
		return fmt.Errorf("failed to remove feature directory %s: %w", feat.Path, err)
	}
	notesRelPath := featureNotesRelPath(feat.DirName)
	notesRemoved := false
	var notesErr error
	if removeFeatureNotes {
		_, notesRemoved, notesErr = removeFeatureNotesDir(projectRoot, feat.DirName)
	}

	if err := feature.PersistRemoved(projectRoot, cfg, feat, time.Now()); err != nil {
		return fmt.Errorf("feature '%s' was removed but failed to record removed lifecycle state: %w", feat.Slug, err)
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return fmt.Errorf("feature '%s' was removed but failed to refresh PROJECT_PROGRESS_SUMMARY.md: %w", feat.Slug, err)
	}
	if notesErr != nil {
		return fmt.Errorf("feature '%s' was removed and marked removed but failed to remove notes at %s: %w", feat.Slug, notesRelPath, notesErr)
	}

	return printRemoveResult(cmd, feat, projectRoot, removeFeatureNotes, notesRemoved, notesRelPath)
}

func selectFeatureForRemove(projectRoot string, specsDir string, cfg *config.Config) (*feature.Feature, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	liveFeatureDirs := make(map[string]struct{}, len(features))
	for _, feat := range features {
		liveFeatureDirs[feat.DirName] = struct{}{}
	}
	if len(features) == 0 {
		printRemovedFeatureHistoryForRemove(projectRoot, cfg, liveFeatureDirs)
		return nil, fmt.Errorf("no feature docs available to remove")
	}

	printSelectionHeader("Select a feature to remove:")
	for i, feat := range features {
		label := feat.DirName
		if feat.Paused {
			label += " (paused)"
		}
		fmt.Printf("  [%d] %s (%s)\n", i+1, label, feat.Phase)
	}
	printRemovedFeatureHistoryForRemove(projectRoot, cfg, liveFeatureDirs)
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(features) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := features[num-1]
	return &selected, nil
}

func printRemoveResult(
	cmd *cobra.Command,
	feat *feature.Feature,
	projectRoot string,
	removeFeatureNotes bool,
	notesRemoved bool,
	notesRelPath string,
) error {
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "🗑️ Removed feature '%s' (status: removed)\n", feat.Slug); err != nil {
		return err
	}
	if removeFeatureNotes {
		if notesRemoved {
			_, err := fmt.Fprintf(out, "🗑️ Removed notes at %s\n", notesRelPath)
			return err
		}
		_, err := fmt.Fprintf(out, "ℹ️ No notes found at %s\n", notesRelPath)
		return err
	}
	if featureNotesPathExists(projectRoot, feat.DirName) {
		_, err := fmt.Fprintf(out, "📝 Retained notes at %s\n", notesRelPath)
		return err
	}
	_, err := fmt.Fprintf(out, "ℹ️ No notes found at %s\n", notesRelPath)
	return err
}

func printRemovedFeatureHistoryForRemove(projectRoot string, cfg *config.Config, liveFeatureDirs map[string]struct{}) {
	if cfg == nil || len(cfg.RemovedFeatures) == 0 {
		return
	}

	printedHeader := false
	for _, removed := range cfg.RemovedFeatures {
		if removed.DirName == "" {
			continue
		}
		if _, exists := liveFeatureDirs[removed.DirName]; exists {
			continue
		}
		if !printedHeader {
			fmt.Println()
			fmt.Println("Already removed:")
			printedHeader = true
		}
		fmt.Printf("  - %s (removed, %s)\n", removed.DirName, removedNotesLabel(projectRoot, removed.DirName))
	}
}

func removedNotesLabel(projectRoot string, dirName string) string {
	if featureNotesPathExists(projectRoot, dirName) {
		return fmt.Sprintf("notes retained: %s", featureNotesRelPath(dirName))
	}
	return "notes removed"
}

func confirmFeatureRemoval(feat *feature.Feature) (bool, error) {
	return confirmFeatureRemovalWithReader(feat, bufio.NewReader(os.Stdin))
}

func confirmFeatureRemovalWithReader(feat *feature.Feature, reader *bufio.Reader) (bool, error) {
	fmt.Printf("Remove feature '%s' at %s? [y/N]: ", feat.DirName, feat.Path)

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "y" || answer == "yes", nil
}

func confirmFeatureNotesRemoval(feat *feature.Feature, notesRelPath string) (bool, error) {
	return confirmFeatureNotesRemovalWithReader(feat, notesRelPath, bufio.NewReader(os.Stdin))
}

func confirmFeatureNotesRemovalWithReader(
	feat *feature.Feature,
	notesRelPath string,
	reader *bufio.Reader,
) (bool, error) {
	fmt.Printf("Remove notes for feature '%s' at %s too? [y/N]: ", feat.DirName, notesRelPath)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "y" || answer == "yes", nil
}

package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

type backlogEntry struct {
	Feature     feature.Feature
	Description string
}

func loadBacklogEntries(specsDir string, cfg *config.Config) ([]backlogEntry, error) {
	features, err := feature.ListBacklogFeatures(specsDir, cfg)
	if err != nil {
		return nil, err
	}

	entries := make([]backlogEntry, 0, len(features))
	for _, feat := range features {
		entries = append(entries, backlogEntry{
			Feature:     feat,
			Description: backlogDescription(&feat),
		})
	}

	return entries, nil
}

func backlogDescription(feat *feature.Feature) string {
	if feat == nil {
		return "(no description)"
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	description, err := feature.ExtractBrainstormSummary(brainstormPath)
	if err != nil || strings.TrimSpace(description) == "" {
		return "(no description)"
	}

	return strings.Join(strings.Fields(description), " ")
}

func resolveBacklogFeature(specsDir string, cfg *config.Config, ref string) (*feature.Feature, error) {
	feat, err := loadFeatureWithState(specsDir, cfg, ref)
	if err != nil {
		return nil, err
	}
	if !feature.IsBacklogItem(*feat) {
		return nil, fmt.Errorf(
			"feature '%s' is not a backlog item. Backlog items must be paused brainstorm-phase features",
			feat.Slug,
		)
	}

	return feat, nil
}

func selectBacklogFeature(specsDir string, cfg *config.Config, title string) (*feature.Feature, error) {
	entries, err := loadBacklogEntries(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no backlog items available")
	}

	printSelectionHeader(title)
	for i, entry := range entries {
		fmt.Printf("  [%d] %s\n", i+1, entry.Feature.DirName)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(entries) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := entries[num-1].Feature
	return &selected, nil
}

func resumeBacklogFeature(
	projectRoot string,
	cfg *config.Config,
	feat *feature.Feature,
	outputOnly bool,
	copy bool,
	outputPath string,
	currentStep string,
) error {
	if feat == nil {
		return fmt.Errorf("backlog feature is required")
	}
	if !feature.IsBacklogItem(*feat) {
		return fmt.Errorf(
			"feature '%s' is not a backlog item. Backlog items must be paused brainstorm-phase features",
			feat.Slug,
		)
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	_, notesRelPath, err := ensureFeatureNotesDir(projectRoot, feat.DirName)
	if err != nil {
		return err
	}
	if _, err := ensureBrainstormNotesDependency(brainstormPath, notesRelPath); err != nil {
		return err
	}

	thesis := existingBrainstormThesis(brainstormPath)
	wasPaused := feat.Paused

	if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
		return err
	}
	if err := updateRollupForResume(projectRoot, cfg, feat.DirName, wasPaused); err != nil {
		return err
	}

	prompt := buildBrainstormPrompt(brainstormPath, feat.Slug, projectRoot, thesis, cfg.GoalPercentage)
	preparedPrompt := prepareAgentPrompt(prompt)
	if outputPath != "" {
		if err := document.Write(outputPath, preparedPrompt); err != nil {
			return fmt.Errorf("failed to write prompt file: %w", err)
		}
		if !outputOnly {
			fmt.Printf("✓ Written prompt to %s\n", outputPath)
		}
	}
	if err := writePromptWithClipboardDefault(preparedPrompt, outputOnly, copy); err != nil {
		return err
	}

	if !outputOnly {
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
		printWorkflowInstructions(currentStep, []string{
			fmt.Sprintf("review and refine %s", brainstormPath),
			fmt.Sprintf("run kit spec %s when the brainstorm is complete", feat.Slug),
			"then continue spec -> plan -> tasks -> implement -> reflect",
		})
	}

	return nil
}

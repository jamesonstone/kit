package cli

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current feature status for coding agents",
	Long: `Display the active feature's status, including:
  - Feature name and ID
  - Business summary from SPEC.md
  - File existence (SPEC, PLAN, TASKS)
  - Task completion progress (from markdown checkboxes)
  - Suggested next action

Output is optimized for coding agents to quickly understand
which files to investigate for the current feature.`,
	Args: cobra.NoArgs,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().Bool("json", false, "output status as JSON")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	// find active feature
	feat, err := feature.FindActiveFeature(specsDir)
	if err != nil {
		return fmt.Errorf("failed to find active feature: %w", err)
	}

	if feat == nil {
		return outputNoActiveFeature(jsonOutput)
	}

	// get full status
	status, err := feature.GetFeatureStatus(feat)
	if err != nil {
		return fmt.Errorf("failed to get feature status: %w", err)
	}

	if jsonOutput {
		return outputStatusJSON(status)
	}

	return outputStatusText(status, specsDir)
}

func outputNoActiveFeature(asJSON bool) error {
	if asJSON {
		output := map[string]interface{}{
			"active_feature": nil,
			"message":        "No active feature in progress",
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println()
	fmt.Println("🤷 No active feature in progress 📭")
	fmt.Println()
	fmt.Println("Run `kit spec <feature-name>` to start a new feature.")
	fmt.Println()
	return nil
}

func outputStatusJSON(status *feature.FeatureStatus) error {
	output := map[string]interface{}{
		"active_feature": status,
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputStatusText(status *feature.FeatureStatus, specsDir string) error {
	fmt.Println()
	fmt.Printf("📊 "+whiteBold+"Active Feature: "+reset+"%s-%s\n", status.ID, status.Name)
	fmt.Println()

	// summary
	if status.Summary != "" {
		fmt.Printf("📝 "+whiteBold+"Summary: "+reset+"%s\n", status.Summary)
		fmt.Println()
	}

	// files
	fmt.Println("📁 " + whiteBold + "Files:" + reset)
	printFileStatus("SPEC.md", status.Files["spec"])
	printFileStatus("PLAN.md", status.Files["plan"])
	printFileStatus("TASKS.md", status.Files["tasks"])
	fmt.Println()

	// progress
	fmt.Print("📈 " + whiteBold + "Progress: " + reset)
	printProgressLine(status)
	fmt.Println()
	fmt.Println()

	// next action
	nextAction := determineNextAction(status)
	fmt.Printf("🎯 "+whiteBold+"Next: "+reset+"%s\n", nextAction)
	fmt.Println()

	// all features progress chart
	printAllFeaturesProgress(specsDir)

	return nil
}

func printFileStatus(name string, fs feature.FileStatus) {
	if fs.Exists {
		fmt.Printf("   %s   ✓  %s\n", padRight(name, 10), fs.Path)
	} else {
		fmt.Printf("   %s   ✗  "+dim+"(not created)"+reset+"\n", padRight(name, 10))
	}
}

func printProgressLine(status *feature.FeatureStatus) {
	specMark := "✗"
	if status.Files["spec"].Exists {
		specMark = "✓"
	}
	planMark := "✗"
	if status.Files["plan"].Exists {
		planMark = "✓"
	}
	tasksMark := "✗"
	if status.Files["tasks"].Exists {
		tasksMark = "✓"
	}

	fmt.Printf("SPEC %s → PLAN %s → TASKS %s", specMark, planMark, tasksMark)

	if status.Progress != nil && status.Progress.HasTasks() {
		fmt.Printf(" (%d/%d complete)", status.Progress.Complete, status.Progress.Total)
	}
}

func determineNextAction(status *feature.FeatureStatus) string {
	// check files in order
	if !status.Files["spec"].Exists {
		return fmt.Sprintf("Create specification: run `kit spec %s`", status.Name)
	}

	if !status.Files["plan"].Exists {
		return fmt.Sprintf("Create implementation plan: run `kit plan %s`", status.Name)
	}

	if !status.Files["tasks"].Exists {
		return fmt.Sprintf("Create task list: run `kit tasks %s`", status.Name)
	}

	// tasks exist, check progress
	if status.Progress != nil && status.Progress.HasTasks() {
		incomplete := status.Progress.Incomplete()
		if incomplete > 0 {
			return fmt.Sprintf("Complete %d remaining task(s) in %s", incomplete, status.Files["tasks"].Path)
		}
		return "All tasks complete! Review and verify implementation."
	}

	// tasks file exists but no checkboxes found
	return fmt.Sprintf("Define tasks with markdown checkboxes in %s", status.Files["tasks"].Path)
}

func padRight(s string, width int) string {
	runeCount := utf8.RuneCountInString(s)
	if runeCount >= width {
		return s
	}
	return s + spaces(width-runeCount)
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

// printAllFeaturesProgress displays a markdown-style progress chart for all features.
func printAllFeaturesProgress(specsDir string) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil || len(features) == 0 {
		return
	}

	fmt.Println("🗺️  " + whiteBold + "All Features:" + reset)
	fmt.Println()

	// table header (6 stages: spec → plan → task → impl → refl → done)
	fmt.Println(dim + "| Feature              | SPEC | PLAN | TASK | IMPL | REFL | DONE |" + reset)
	fmt.Println(dim + "|----------------------|------|------|------|------|------|------|" + reset)

	for _, feat := range features {
		printFeatureProgressRow(&feat)
	}
	fmt.Println()
}

// printFeatureProgressRow prints a single row in the progress table.
func printFeatureProgressRow(feat *feature.Feature) {
	name := truncateString(feat.DirName, 20)
	name = padRight(name, 20)

	// phase markers (6 stages)
	specM := phaseMarker(feat.Phase, feature.PhaseSpec)
	planM := phaseMarker(feat.Phase, feature.PhasePlan)
	taskM := phaseMarker(feat.Phase, feature.PhaseTasks)
	implM := phaseMarker(feat.Phase, feature.PhaseImplement)
	reflM := phaseMarker(feat.Phase, feature.PhaseReflect)
	doneM := phaseMarker(feat.Phase, feature.PhaseComplete)

	fmt.Printf("| %s | %s | %s | %s | %s | %s | %s |\n",
		name, specM, planM, taskM, implM, reflM, doneM)
}

// phaseMarker returns a visual marker for the phase state.
// returns: ●  (complete), ◐  (current), ○  (pending)
func phaseMarker(current feature.Phase, target feature.Phase) string {
	order := map[feature.Phase]int{
		feature.PhaseSpec:      1,
		feature.PhasePlan:      2,
		feature.PhaseTasks:     3,
		feature.PhaseImplement: 4,
		feature.PhaseReflect:   5,
		feature.PhaseComplete:  6,
	}

	currentIdx := order[current]
	targetIdx := order[target]

	if targetIdx < currentIdx {
		return plan + " ●  " + reset // complete (green)
	}
	if targetIdx == currentIdx {
		if current == feature.PhaseComplete {
			return plan + " ●  " + reset // complete (green) — terminal phase
		}
		return implement + " ◐  " + reset // current (orange)
	}
	return dim + " ○  " + reset // pending (dim)
}

// truncateString truncates a string to maxLen, adding ellipsis if needed.
func truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen-1]) + "…"
}

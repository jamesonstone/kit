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
	"github.com/jamesonstone/kit/internal/promptdoc"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var planCopy bool
var planOutputOnly bool

var planCmd = &cobra.Command{
	Use:   "plan [feature]",
	Short: "Deprecated v1 staged workflow: create PLAN.md",
	Long: `Deprecated v1 staged workflow: create a new implementation plan for a feature.

The default workflow uses the host agent's native planning capability and
captures an accepted plan in SPEC.md only when durable rationale is required.
Use this command only for the legacy staged artifact flow.

Creates:
  - PLAN.md with required sections and placeholder comments

Prerequisites:
  - SPEC.md must exist (unless --force)

If no feature is specified, shows an interactive selection of features
that have SPEC.md but no PLAN.md yet.

Updates PROJECT_PROGRESS_SUMMARY.md after creation.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPlan,
}

func init() {
	planCmd.Flags().Bool("force", false, "create missing SPEC.md with headers if it doesn't exist")
	planCmd.Flags().Bool("warp", false, "output prompt for Warp coding agent to fill PLAN.md from Warp plan")
	planCmd.Flags().BoolVar(&planCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	planCmd.Flags().BoolVar(&planOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(planCmd)
	legacyCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	planForce, _ := cmd.Flags().GetBool("force")
	warpMode, _ := cmd.Flags().GetBool("warp")
	outputOnly, _ := cmd.Flags().GetBool("output-only")
	promptOnly := promptOnlyEnabled(cmd)

	// find project root
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	specsDir := cfg.SpecsPath(projectRoot)

	if promptOnly {
		if planForce {
			return fmt.Errorf("--prompt-only cannot be used with --force")
		}
		return runPlanPromptOnly(args, projectRoot, cfg, warpMode, outputOnly)
	}

	var feat *feature.Feature

	if len(args) == 0 {
		// interactive mode: select from features with SPEC but no PLAN
		feat, err = selectFeatureForPlan(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	} else {
		// direct mode: resolve feature by name
		featureRef := args[0]
		feat, err = loadFeatureWithState(specsDir, cfg, featureRef)
		if err != nil {
			return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", featureRef, featureRef)
		}
	}

	if !outputOnly {
		fmt.Printf("📋 Creating plan for feature: %s\n", feat.DirName)
	}

	// check prerequisites
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		if planForce || cfg.AllowOutOfOrder {
			// create empty SPEC.md
			content := templates.BuildSpecV2ArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(specPath, content); err != nil {
				return fmt.Errorf("failed to create SPEC.md: %w", err)
			}
			if !outputOnly {
				fmt.Println("  ✓ Created SPEC.md (--force)")
			}
		} else {
			return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first or use --force", feat.Slug)
		}
	}

	// create PLAN.md if it doesn't exist
	planPath := filepath.Join(feat.Path, "PLAN.md")
	if !document.Exists(planPath) {
		content := templates.BuildPlanArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(planPath, content); err != nil {
			return fmt.Errorf("failed to create PLAN.md: %w", err)
		}
		if !outputOnly {
			fmt.Println("  ✓ Created PLAN.md")
		}
	} else if !outputOnly {
		fmt.Println("  ✓ PLAN.md already exists")
	}
	if effectivePromptProfile(feat.Path) == promptProfileFrontend {
		if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
			return err
		}
		if _, err := ensureFrontendProfileDependencyRows(planPath, document.TypePlan, feat.DirName); err != nil {
			return err
		}
	}

	wasPaused := feat.Paused
	if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
		return err
	}

	// update PROJECT_PROGRESS_SUMMARY.md
	if err := rollup.Update(projectRoot, cfg); err != nil {
		if !outputOnly {
			fmt.Printf("  ⚠ Could not update PROJECT_PROGRESS_SUMMARY.md: %v\n", err)
		}
	} else if !outputOnly {
		fmt.Println("  ✓ Updated PROJECT_PROGRESS_SUMMARY.md")
	}

	if !outputOnly {
		fmt.Printf("\n✅ Plan for '%s' ready!\n", feat.Slug)
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
		printNumberedNextSteps([]string{
			fmt.Sprintf("Edit %s to define the implementation approach", planPath),
			fmt.Sprintf("Run 'kit legacy tasks %s' to create executable tasks", feat.Slug),
		})
	}
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")

	if warpMode {
		return outputWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
	}

	return outputStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
}

func runPlanPromptOnly(args []string, projectRoot string, cfg *config.Config, warpMode, outputOnly bool) error {
	specsDir := cfg.SpecsPath(projectRoot)

	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 0 {
		feat, err = selectFeatureForPlanPromptOnly(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	} else {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", args[0], args[0])
		}
	}

	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")

	if !document.Exists(specPath) {
		return fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}
	if !document.Exists(planPath) {
		return fmt.Errorf("PLAN.md not found. Run 'kit legacy plan %s' first", feat.Slug)
	}

	if warpMode {
		return outputWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
	}

	return outputStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, outputOnly)
}

// selectFeatureForPlan shows an interactive numbered list of features
// that have SPEC.md but no PLAN.md yet.
func selectFeatureForPlan(specsDir string) (*feature.Feature, error) {
	candidates, err := workflowStageCandidates(specsDir, workflowSelectionStagePlan)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features ready for planning (need SPEC.md without PLAN.md)\n\nRun 'kit spec <feature>' to create a new feature first")
	}

	printSelectionHeader("Select a feature to plan:")
	for i, f := range candidates {
		fmt.Printf("  [%d] %s\n", i+1, f.DirName)
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

func selectFeatureForPlanPromptOnly(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "SPEC.md")) &&
			document.Exists(filepath.Join(f.Path, "PLAN.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no plans available to regenerate prompts for\n\nRun 'kit legacy plan <feature>' first")
	}

	printSelectionHeader("Select a feature to regenerate the plan prompt for:")
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

// outputStandardPlanPrompt outputs the standard coding agent prompt.
func outputStandardPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	prompt := buildStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, projectRoot)

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, planCopy); err != nil {
		return fmt.Errorf("failed to output prompt: %w", err)
	}

	return nil
}
func buildStandardPlanPrompt(
	planPath string,
	specPath string,
	brainstormPath string,
	feat *feature.Feature,
	cfg *config.Config,
	projectRoot string,
) string {
	constitutionPath := filepath.Join(projectRoot, cfg.ConstitutionPath)
	hasBrainstorm := document.Exists(brainstormPath)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Complete the legacy implementation plan for feature `%s`. This is documentation-only; do not implement product code.", feat.Slug))
		doc.Heading(2, "Context")
		rows := [][]string{
			{"SPEC.md", specPath, "Binding requirements and acceptance"},
			{"PLAN.md", planPath, "Artifact to update"},
			{"Constitution", constitutionPath, "Project invariants"},
			{"Project root", projectRoot, "Discover existing implementation patterns"},
		}
		if hasBrainstorm {
			rows = append(rows, []string{"BRAINSTORM.md", brainstormPath, "Non-binding research context"})
		}
		doc.Table([]string{"Input", "Path", "Use"}, rows)

		doc.Heading(2, "Planning Contract")
		doc.OrderedList(1,
			"Read SPEC.md as fixed scope. Inspect the smallest relevant code, tests, docs, and prior-feature context needed to ground design decisions; do not invent files or APIs.",
			"Resolve repository-discoverable gaps yourself. Ask concise numbered questions only for a material non-discoverable design choice, with a recommended default and impact; stop before writing a final plan while such a choice remains.",
			fmt.Sprintf("Update `%s` directly with the simplest viable approach, explicit tradeoffs, components/responsibilities, data and interfaces, exact dependencies/references, touched surfaces, sequencing, risks/rollback, and validation strategy.", planPath),
			"Map every binding acceptance criterion to implementation responsibility and evidence. Keep exact external and repo references in front matter.",
			"Make the resulting task breakdown deterministic without writing TASKS.md or implementation code.",
		)

		doc.Heading(2, "Success Criteria")
		doc.BulletList(
			fmt.Sprintf("Confidence is at least %d and no material design question remains.", cfg.GoalPercentage),
			"PLAN.md adds implementation strategy rather than restating requirements, introduces no new scope, and follows repository invariants.",
			"Each planned surface, risk, test, and documentation obligation traces to SPEC.md acceptance and has a concrete evidence method.",
			"Empty optional sections state `not applicable`; placeholder comments are removed and the project progress summary remains accurate.",
		)

		doc.Heading(2, "Output")
		doc.BulletList(
			"Update PLAN.md and supporting documentation only.",
			"Report key decisions, validation strategy, remaining risk, and the next legacy task-generation step.",
		)
		addFinalResponseContract(doc, planFinalResponseContract(feat.Slug)...)
	})
}

func outputWarpPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	prompt := buildWarpPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, projectRoot)

	if !outputOnly {
		fmt.Println()
		fmt.Println(whiteBold + "Warp Plan Integration" + reset)
		fmt.Println(dim + "The following files have been created:" + reset)
		fmt.Printf("  • PLAN.md: %s\n", planPath)
		fmt.Printf("  • SPEC.md: %s\n\n", specPath)
	}

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, planCopy); err != nil {
		return fmt.Errorf("failed to output prompt: %w", err)
	}

	return nil
}

func buildWarpPlanPrompt(
	planPath string,
	specPath string,
	brainstormPath string,
	feat *feature.Feature,
	cfg *config.Config,
	projectRoot string,
) string {
	constitutionPath := filepath.Join(projectRoot, cfg.ConstitutionPath)
	hasBrainstorm := document.Exists(brainstormPath)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Convert the Warp plan in the current conversation into the legacy PLAN.md for feature `%s`. This is documentation-only; do not implement product code.", feat.Slug))
		doc.Heading(2, "Context")
		rows := [][]string{
			{"Warp plan", "Current conversation", "Non-binding design input"},
			{"SPEC.md", specPath, "Binding requirements and acceptance"},
			{"PLAN.md", planPath, "Artifact to update"},
			{"Constitution", constitutionPath, "Project invariants"},
			{"Project root", projectRoot, "Discover implementation patterns"},
		}
		if hasBrainstorm {
			rows = append(rows, []string{"BRAINSTORM.md", brainstormPath, "Non-binding research context"})
		}
		doc.Table([]string{"Input", "Path", "Use"}, rows)

		doc.Heading(2, "Planning Contract")
		doc.OrderedList(1,
			"Extract the Warp plan's concrete design decisions, then verify them against SPEC.md, the constitution, and the smallest relevant code and test surfaces. SPEC.md wins on conflict.",
			"Resolve repository-discoverable gaps yourself. Ask concise numbered questions only for a material non-discoverable design choice, with a recommended default and impact; stop before finalizing while such a choice remains.",
			fmt.Sprintf("Update `%s` directly with the simplest viable approach, explicit tradeoffs, components/responsibilities, data and interfaces, exact dependencies/references, touched surfaces, sequencing, risks/rollback, and validation strategy.", planPath),
			"Map every binding acceptance criterion to implementation responsibility and evidence. Keep exact external and repository references in front matter.",
			"Add implementation detail where the Warp plan is abstract, but introduce no scope beyond SPEC.md and do not write TASKS.md or product code.",
		)

		doc.Heading(2, "Success Criteria")
		doc.BulletList(
			fmt.Sprintf("Confidence is at least %d and no material design question remains.", cfg.GoalPercentage),
			"PLAN.md adds implementation strategy beyond the Warp plan without restating requirements or changing binding scope.",
			"Each planned surface, risk, test, and documentation obligation traces to SPEC.md acceptance and has concrete evidence.",
			"Empty optional sections state `not applicable`; placeholder comments are removed and the project progress summary remains accurate.",
		)

		doc.Heading(2, "Output")
		doc.BulletList(
			"Update PLAN.md and supporting documentation only.",
			"Report key decisions carried forward or changed, validation strategy, remaining risk, and the next legacy task-generation step.",
		)
		addFinalResponseContract(doc, planFinalResponseContract(feat.Slug)...)
	})
}

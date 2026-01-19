package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

var checkAll bool

var checkCmd = &cobra.Command{
	Use:   "check [feature]",
	Short: "Validate feature documents",
	Long: `Validate a feature's documents for completeness and correctness.

Validates:
  - Required documents exist (SPEC.md, PLAN.md, TASKS.md)
  - Required sections are present in each document
  - Traceability between spec → plan → tasks
  - No unresolved placeholders

Use --all to validate all features in the project.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().BoolVar(&checkAll, "all", false, "validate all features in docs/specs/")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
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

	if checkAll {
		return checkAllFeatures(specsDir)
	}

	if len(args) == 0 {
		return fmt.Errorf("feature name required. Use --all to check all features")
	}

	return checkFeature(specsDir, args[0])
}

func checkFeature(specsDir string, featureRef string) error {
	feat, err := feature.Resolve(specsDir, featureRef)
	if err != nil {
		return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", featureRef, featureRef)
	}

	fmt.Printf("🔎 Checking feature: %s\n", feat.DirName)

	var errors []string
	var warnings []string

	// check SPEC.md
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		errors = append(errors, fmt.Sprintf("SPEC.md not found. Run 'kit spec %s' to create it", feat.Slug))
	} else {
		doc, err := document.ParseFile(specPath, document.TypeSpec)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse SPEC.md: %v", err))
		} else {
			for _, e := range doc.Validate() {
				errors = append(errors, e.Error())
			}
			if doc.HasUnresolvedPlaceholders() {
				warnings = append(warnings, "SPEC.md has unresolved TODO placeholders")
			}
		}
	}

	// check PLAN.md
	planPath := filepath.Join(feat.Path, "PLAN.md")
	if !document.Exists(planPath) {
		warnings = append(warnings, fmt.Sprintf("PLAN.md not found. Run 'kit plan %s' to create it", feat.Slug))
	} else {
		doc, err := document.ParseFile(planPath, document.TypePlan)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse PLAN.md: %v", err))
		} else {
			for _, e := range doc.Validate() {
				errors = append(errors, e.Error())
			}
			if doc.HasUnresolvedPlaceholders() {
				warnings = append(warnings, "PLAN.md has unresolved TODO placeholders")
			}
		}
	}

	// check TASKS.md
	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	if !document.Exists(tasksPath) {
		warnings = append(warnings, fmt.Sprintf("TASKS.md not found. Run 'kit tasks %s' to create it", feat.Slug))
	} else {
		doc, err := document.ParseFile(tasksPath, document.TypeTasks)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse TASKS.md: %v", err))
		} else {
			for _, e := range doc.Validate() {
				errors = append(errors, e.Error())
			}
			if doc.HasUnresolvedPlaceholders() {
				warnings = append(warnings, "TASKS.md has unresolved TODO placeholders")
			}
		}
	}

	// print results
	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Printf("  ✅ All checks passed!\n")
		return nil
	}

	if len(warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings (%d):\n", len(warnings))
		for _, w := range warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\n❌ Errors (%d):\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("validation failed with %d error(s)", len(errors))
	}

	return nil
}

func checkAllFeatures(specsDir string) error {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return fmt.Errorf("failed to list features: %w", err)
	}

	if len(features) == 0 {
		fmt.Println("No features found. Run 'kit spec <feature>' to create one.")
		return nil
	}

	fmt.Printf("🔎 Checking %d feature(s)...\n\n", len(features))

	var totalErrors int
	for _, feat := range features {
		err := checkFeature(specsDir, feat.Slug)
		if err != nil {
			totalErrors++
		}
		fmt.Println()
	}

	if totalErrors > 0 {
		return fmt.Errorf("%d feature(s) have validation errors", totalErrors)
	}

	fmt.Printf("✅ All %d feature(s) passed validation!\n", len(features))
	return nil
}

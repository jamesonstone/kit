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
var checkProject bool

var checkCmd = &cobra.Command{
	Use:   "check [feature]",
	Short: "Validate feature or project documents",
	Long: `Validate Kit-managed documents for completeness and correctness.

Validates:
  - Optional BRAINSTORM.md when present
  - Required documents exist (SPEC.md, PLAN.md, TASKS.md)
  - Required sections are present and populated in each document
  - Traceability between spec → plan → tasks
  - No unresolved placeholders

Use --all to validate all features in the project.
Use --project to validate the repo-level document contract.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().BoolVar(&checkAll, "all", false, "validate all features in docs/specs/")
	checkCmd.Flags().BoolVar(&checkProject, "project", false, "validate the repo-level document and instruction contract")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	if checkProject && len(args) > 0 {
		return fmt.Errorf("--project cannot be used with a feature argument")
	}

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

	if checkProject {
		return checkProjectContract(projectRoot, cfg)
	}

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

	// check BRAINSTORM.md when present
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if document.Exists(brainstormPath) {
		doc, err := document.ParseFile(brainstormPath, document.TypeBrainstorm)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse BRAINSTORM.md: %v", err))
		} else {
			for _, e := range doc.Validate() {
				errors = append(errors, e.Error())
			}
			if doc.HasUnresolvedPlaceholders() {
				warnings = append(warnings, "BRAINSTORM.md has unresolved TODO placeholders")
			}
		}
	}

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
		fmt.Printf("\n⚠️ Warnings (%d):\n", len(warnings))
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

func checkProjectContract(projectRoot string, cfg *config.Config) error {
	fmt.Printf("🔎 Checking project contract...\n")

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		return err
	}

	if len(report.Findings) == 0 {
		fmt.Printf("  ✅ Project contract is coherent!\n")
		return nil
	}

	var errors []reconcileFinding
	var warnings []reconcileFinding
	for _, finding := range report.Findings {
		if finding.Severity == reconcileSeverityError {
			errors = append(errors, finding)
			continue
		}
		warnings = append(warnings, finding)
	}

	if len(warnings) > 0 {
		fmt.Printf("\n⚠️ Warnings (%d):\n", len(warnings))
		for _, finding := range warnings {
			fmt.Printf("  - [%s] %s\n", relativeCheckPath(projectRoot, finding.FilePath), finding.Issue)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\n❌ Errors (%d):\n", len(errors))
		for _, finding := range errors {
			fmt.Printf("  - [%s] %s\n", relativeCheckPath(projectRoot, finding.FilePath), finding.Issue)
		}
	}

	return fmt.Errorf("project validation failed with %d finding(s)", len(report.Findings))
}

func relativeCheckPath(projectRoot, path string) string {
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return path
	}

	return rel
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

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
  - Required v2 SPEC.md exists and is valid
  - Optional legacy BRAINSTORM.md, PLAN.md, and TASKS.md when present
  - Required sections are present and populated in each parsed document
  - Legacy traceability when staged artifacts exist
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
		return checkAllFeatures(projectRoot, specsDir)
	}

	if len(args) == 0 {
		return fmt.Errorf("feature name required. Use --all to check all features")
	}

	return checkFeature(projectRoot, specsDir, args[0])
}

func checkFeature(projectRoot string, specsDir string, featureRef string) error {
	feat, err := feature.Resolve(specsDir, featureRef)
	if err != nil {
		return fmt.Errorf("feature '%s' not found. Run 'kit spec %s' first to create it", featureRef, featureRef)
	}

	fmt.Printf("🔎 Checking feature: %s\n", feat.DirName)

	var errors []string
	var warnings []string
	v2Feature := isV2Feature(feat)

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if document.Exists(brainstormPath) {
		doc, err := document.ParseFile(brainstormPath, document.TypeBrainstorm)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse BRAINSTORM.md: %v", err))
		} else {
			for _, e := range doc.Validate() {
				errors = append(errors, e.Error())
			}
			errors = append(errors, featureMetadataIdentityErrors(doc, feat.DirName)...)
			errors = append(errors, featureRulesetReferenceErrors(projectRoot, doc)...)
			warnings = append(warnings, metadataDiagnosticWarnings(doc)...)
			warnings = append(warnings, metadataConflictWarnings(doc)...)
			if doc.HasUnresolvedPlaceholders() {
				warnings = append(warnings, "BRAINSTORM.md has unresolved TODO placeholders")
			}
		}
	}

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
			errors = append(errors, featureMetadataIdentityErrors(doc, feat.DirName)...)
			errors = append(errors, featureRulesetReferenceErrors(projectRoot, doc)...)
			warnings = append(warnings, metadataDiagnosticWarnings(doc)...)
			warnings = append(warnings, metadataConflictWarnings(doc)...)
			if doc.HasUnresolvedPlaceholders() {
				warnings = append(warnings, "SPEC.md has unresolved TODO placeholders")
			}
		}
	}

	planPath := filepath.Join(feat.Path, "PLAN.md")
	if !document.Exists(planPath) {
		if !v2Feature {
			warnings = append(warnings, fmt.Sprintf("legacy PLAN.md not found. Run 'kit legacy plan %s' only when continuing staged v1 work", feat.Slug))
		}
	} else {
		doc, err := document.ParseFile(planPath, document.TypePlan)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse PLAN.md: %v", err))
		} else {
			for _, e := range doc.Validate() {
				errors = append(errors, e.Error())
			}
			errors = append(errors, featureMetadataIdentityErrors(doc, feat.DirName)...)
			errors = append(errors, featureRulesetReferenceErrors(projectRoot, doc)...)
			warnings = append(warnings, metadataDiagnosticWarnings(doc)...)
			warnings = append(warnings, metadataConflictWarnings(doc)...)
			if doc.HasUnresolvedPlaceholders() {
				warnings = append(warnings, "PLAN.md has unresolved TODO placeholders")
			}
		}
	}

	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	if !document.Exists(tasksPath) {
		if !v2Feature {
			warnings = append(warnings, fmt.Sprintf("legacy TASKS.md not found. Run 'kit legacy tasks %s' only when continuing staged v1 work", feat.Slug))
		}
	} else {
		doc, err := document.ParseFile(tasksPath, document.TypeTasks)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to parse TASKS.md: %v", err))
		} else {
			for _, e := range doc.Validate() {
				errors = append(errors, e.Error())
			}
			errors = append(errors, featureMetadataIdentityErrors(doc, feat.DirName)...)
			errors = append(errors, featureRulesetReferenceErrors(projectRoot, doc)...)
			warnings = append(warnings, metadataDiagnosticWarnings(doc)...)
			warnings = append(warnings, metadataConflictWarnings(doc)...)
			if doc.HasUnresolvedPlaceholders() {
				warnings = append(warnings, "TASKS.md has unresolved TODO placeholders")
			}
		}
	}

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

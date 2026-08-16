package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/feature"
)

func checkProjectContractTo(out io.Writer, projectRoot string, cfg *config.Config) error {
	if _, err := fmt.Fprintln(out, "🔎 Checking project contract..."); err != nil {
		return err
	}

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		return err
	}
	if len(report.Findings) == 0 {
		if _, err := fmt.Fprintln(out, "  ✅ Project contract is coherent!"); err != nil {
			return err
		}
		return nil
	}

	var errors []reconcileFinding
	var warnings []reconcileFinding
	var blockingWarnings []reconcileFinding
	for _, finding := range report.Findings {
		if finding.Severity == reconcileSeverityError {
			errors = append(errors, finding)
			continue
		}
		warnings = append(warnings, finding)
		if !finding.NonBlocking {
			blockingWarnings = append(blockingWarnings, finding)
		}
	}

	if len(warnings) > 0 {
		if _, err := fmt.Fprintf(out, "\n⚠️ Warnings (%d):\n", len(warnings)); err != nil {
			return err
		}
		for _, finding := range warnings {
			if _, err := fmt.Fprintf(out, "  - [%s] %s\n", relativeCheckPath(projectRoot, finding.FilePath), finding.Issue); err != nil {
				return err
			}
		}
	}
	if len(errors) > 0 {
		if _, err := fmt.Fprintf(out, "\n❌ Errors (%d):\n", len(errors)); err != nil {
			return err
		}
		for _, finding := range errors {
			if _, err := fmt.Fprintf(out, "  - [%s] %s\n", relativeCheckPath(projectRoot, finding.FilePath), finding.Issue); err != nil {
				return err
			}
		}
	}

	if len(errors) == 0 && len(blockingWarnings) == 0 {
		if _, err := fmt.Fprintln(out, "  ℹ Compatibility advisories do not block project validation."); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("project validation failed with %d blocking finding(s)", len(errors)+len(blockingWarnings))
}

func relativeCheckPath(projectRoot, path string) string {
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return path
	}

	return rel
}

func checkAllFeatures(projectRoot string, specsDir string) error {
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
		err := checkFeature(projectRoot, specsDir, feat.Slug)
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

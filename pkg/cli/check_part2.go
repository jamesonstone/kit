package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func featureMetadataIdentityErrors(doc *document.Document, featureDirName string) []string {
	if doc.Metadata == nil {
		return nil
	}

	expected := document.FeatureMetadataFromDir(featureDirName)
	var errors []string
	if doc.Metadata.Feature.ID != "" && doc.Metadata.Feature.ID != expected.ID {
		errors = append(errors, fmt.Sprintf(
			"%s: front matter feature.id %q does not match containing feature directory id %q",
			doc.Path,
			doc.Metadata.Feature.ID,
			expected.ID,
		))
	}
	if doc.Metadata.Feature.Slug != "" && doc.Metadata.Feature.Slug != expected.Slug {
		errors = append(errors, fmt.Sprintf(
			"%s: front matter feature.slug %q does not match containing feature directory slug %q",
			doc.Path,
			doc.Metadata.Feature.Slug,
			expected.Slug,
		))
	}
	if doc.Metadata.Feature.Dir != "" && doc.Metadata.Feature.Dir != expected.Dir {
		errors = append(errors, fmt.Sprintf(
			"%s: front matter feature.dir %q does not match containing feature directory %q",
			doc.Path,
			doc.Metadata.Feature.Dir,
			expected.Dir,
		))
	}
	return errors
}

func metadataConflictWarnings(doc *document.Document) []string {
	warnings := make([]string, 0, len(doc.MetadataConflictWarnings))
	for _, conflict := range doc.MetadataConflictWarnings {
		warnings = append(warnings, fmt.Sprintf("%s: %s", doc.Path, conflict.Message))
	}
	return warnings
}

func metadataDiagnosticWarnings(doc *document.Document) []string {
	var warnings []string
	for _, diagnostic := range doc.MetadataDiagnostics {
		if diagnostic.Severity != document.MetadataDiagnosticWarning {
			continue
		}
		warnings = append(warnings, fmt.Sprintf("%s: %s. %s", doc.Path, diagnostic.Message, diagnostic.Fix))
	}
	return warnings
}

func checkProjectContract(projectRoot string, cfg *config.Config) error {
	fmt.Printf("🔎 Checking project contract...\n")

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		return err
	}
	refreshStatus, refreshErr := calculateProjectRefreshStatus(projectRoot, cfg, time.Now().UTC())

	if len(report.Findings) == 0 {
		fmt.Printf("  ✅ Project contract is coherent!\n")
		if refreshErr == nil {
			if refreshStatus.Due {
				fmt.Printf("  ⚠ Project refresh due: %s. Run `kit project refresh`.\n", strings.Join(refreshStatus.Reasons, "; "))
			} else {
				fmt.Printf("  ℹ Project refresh %s.\n", formatProjectRefreshDueSummary(refreshStatus))
			}
			return nil
		}
		fmt.Printf("  ⚠ Project refresh due status unavailable: %v\n", refreshErr)
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
	if refreshErr != nil {
		fmt.Printf("\n⚠️ Project refresh due status unavailable: %v\n", refreshErr)
	} else if refreshStatus.Due {
		fmt.Printf("\n⚠️ Project refresh due: %s\n", formatProjectRefreshDueSummary(refreshStatus))
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

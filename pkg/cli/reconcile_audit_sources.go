package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func newFinding(
	severity reconcileSeverity,
	path, issue, contractSource, update string,
	searchHints []string,
) reconcileFinding {
	return reconcileFinding{
		Severity:          severity,
		FilePath:          path,
		Issue:             issue,
		ContractSource:    contractSource,
		UpdateInstruction: update,
		SearchHints:       searchHints,
	}
}

func templateSource(projectRoot string) string {
	return filepath.Join(projectRoot, "internal", "templates", "templates.go")
}

func constitutionSource(projectRoot string) string {
	return filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
}

func initProjectSource(projectRoot string) string {
	return filepath.Join(projectRoot, "docs", "specs", "0000_INIT_PROJECT.md")
}

func contractSourceForSection(projectRoot string, docType document.DocumentType, section string) string {
	if section == "RELATIONSHIPS" {
		return constitutionSource(projectRoot)
	}
	if docType == document.TypeTasks &&
		(section == "PROGRESS TABLE" || section == "TASK LIST" || section == "TASK DETAILS") {
		return initProjectSource(projectRoot)
	}
	return templateSource(projectRoot)
}

func genericFeatureSearchHints(projectRoot string, feat *feature.Feature, path, documentName string) []string {
	return []string{
		fmt.Sprintf("sed -n '1,240p' %s", path),
		fmt.Sprintf(
			"rg -n \"%s|%s\" %s %s %s",
			feat.Slug,
			strings.ReplaceAll(feat.Slug, "-", " "),
			filepath.Join(projectRoot, "docs"),
			filepath.Join(projectRoot, "pkg"),
			filepath.Join(projectRoot, "internal"),
		),
		fmt.Sprintf("rg -n \"^## %s|%s\" %s", documentName, feat.Slug, feat.Path),
	}
}

func searchHintsForSection(projectRoot, path, section string) []string {
	switch section {
	case "RELATIONSHIPS":
		return []string{
			fmt.Sprintf("sed -n '1,240p' %s", path),
			fmt.Sprintf(
				"rg -n \"^## RELATIONSHIPS|builds on:|depends on:|related to:\" %s",
				filepath.Join(projectRoot, "docs", "specs"),
			),
		}
	case "DEPENDENCIES":
		return []string{
			fmt.Sprintf("sed -n '1,240p' %s", path),
			fmt.Sprintf(
				"rg -n \"## DEPENDENCIES|Dependency |Used For|Status\" %s %s %s",
				filepath.Join(projectRoot, "docs"),
				filepath.Join(projectRoot, "pkg"),
				filepath.Join(projectRoot, "internal"),
			),
		}
	default:
		return []string{
			fmt.Sprintf("sed -n '1,240p' %s", path),
			fmt.Sprintf(
				"rg -n \"^## %s|SUMMARY|PROBLEM|APPROACH|QUESTIONS\" %s %s %s",
				section,
				filepath.Join(projectRoot, "docs"),
				filepath.Join(projectRoot, "pkg"),
				filepath.Join(projectRoot, "internal"),
			),
		}
	}
}

func searchHintsForTable(projectRoot, path, section string) []string {
	return []string{
		fmt.Sprintf("sed -n '1,240p' %s", path),
		fmt.Sprintf("rg -n \"## %s|\\|\" %s", section, templateSource(projectRoot)),
	}
}

func searchHintsForTaskAlignment(path string) []string {
	return []string{
		fmt.Sprintf("rg -n \"^\\| T[0-9]{3} \\||^- \\[[ xX]\\] T[0-9]{3}:|^### T[0-9]{3}$\" %s", path),
		fmt.Sprintf("sed -n '1,260p' %s", path),
	}
}

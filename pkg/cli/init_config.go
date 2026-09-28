package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/promptdoc"
)

func populateGlobalConfig(outputOnly bool) error {
	configPath, changed, err := config.PopulateGlobalConfig(defaultInitConfig())
	if err != nil {
		return fmt.Errorf("failed to populate global config: %w", err)
	}

	if outputOnly {
		return nil
	}
	if changed {
		fmt.Printf("  ✓ Populated %s\n", configPath)
		return nil
	}
	fmt.Printf("  ✓ %s exists\n", configPath)
	return nil
}

func buildProjectInitPrompt(
	projectRoot,
	constitutionFullPath string,
	snapshots ...[]managedFileDeliverySnapshot,
) string {
	makefileFullPath := filepath.Join(projectRoot, makefilePath)
	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Initialize project memory and verified command entrypoints for the repository at %s.", projectRoot))
		doc.Paragraph("Constitution:")
		doc.BulletList(
			fmt.Sprintf("The generated starter at %s is a valid bootstrap Constitution", constitutionFullPath),
			"Record a principle, constraint, non-goal, definition, or workflow boundary only when implemented behavior, validated outcomes, or recurring conventions already demonstrate it; otherwise leave the starter sections unchanged",
			"Product ideas and feature intent belong in the relevant SPEC.md, not the Constitution",
		)
		doc.Paragraph(fmt.Sprintf("Makefile (%s): add targets only when backed by this repository's real commands.", makefileFullPath))
		doc.BulletList(
			"Leave the safe starter unchanged when the repository has no verified development, build, test, lint, formatting, or validation commands",
			"Add only applicable canonical targets (`dev`, `build`, `test`, `check`, `lint`, `fmt`, `clean`) as thin wrappers around repository-native commands, declared `.PHONY`, with no placeholder or guessed recipes",
			"Run `make help` and each added target that is safe to execute",
		)
		changed, removed := initDeliveredFiles(snapshots...)
		if len(changed) > 0 {
			doc.Paragraph("Files `kit init` created or changed (deliver them with this work): " + strings.Join(changed, ", "))
		}
		if len(removed) > 0 {
			doc.Paragraph("Retired Kit files `kit init` removed (deliver the deletions with this work): " + strings.Join(removed, ", "))
		}
	})
}

// initDeliveredFiles splits the repository files kit init touched into those
// present afterwards and those it removed. The snapshots already exclude
// local-only, secret-like, and ignored paths.
func initDeliveredFiles(snapshots ...[]managedFileDeliverySnapshot) ([]string, []string) {
	if len(snapshots) == 0 {
		return nil, nil
	}
	var changed, removed []string
	for _, change := range snapshots[0] {
		path := "`" + normalizeManagedFileDeliveryPath(change.Path) + "`"
		if change.ResultState == managedFileAbsentState {
			removed = append(removed, path)
		} else {
			changed = append(changed, path)
		}
	}
	sort.Strings(changed)
	sort.Strings(removed)
	return changed, removed
}

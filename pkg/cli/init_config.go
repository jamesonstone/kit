package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/promptdoc"
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

func buildProjectInitPrompt(projectRoot, constitutionFullPath string) string {
	makefileFullPath := filepath.Join(projectRoot, makefilePath)
	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Initialize project memory and verified command entrypoints for the repository at %s.", projectRoot))
		doc.Paragraph("Constitution guidance:")
		doc.BulletList(
			"Read docs/agents/README.md before inspecting repository evidence or modifying project memory",
			fmt.Sprintf("Treat the exact generated starter at %s as a valid bootstrap Constitution", constitutionFullPath),
			"Inspect implemented behavior, validated outcomes, current canonical documentation, and recurring repository conventions as evidence",
			"Do not ask the user to explain the entire project, infer permanent rules from initial aspiration, or derive project truth from Kit-generated scaffolding",
			"Update the Constitution only when repository evidence already demonstrates a durable project-wide principle, constraint, non-goal, definition, vocabulary term, or workflow boundary",
			"When evidence is insufficient, leave the project-specific starter sections unchanged; normal post-validation curation will evolve them as the project matures",
			"Follow docs/references/rules/constitution-curation.md when that registry ruleset is present",
		)
		doc.Paragraph(fmt.Sprintf(
			"Populate %s with a canonical project command interface only when it can be backed by this repository's real commands.",
			makefileFullPath,
		))
		doc.BulletList(
			"Inspect package scripts, toolchain configuration, development documentation, and existing automation before choosing recipe commands",
			"Leave the safe starter Makefile unchanged when the repository has no verified development, build, test, lint, formatting, or validation commands",
			"Expose `make dev` when the repository has a verified local development or run workflow",
			"Add only applicable canonical targets such as `build`, `test`, `check`, `lint`, `fmt`, and `clean`, plus useful project-specific targets",
			"Keep recipes as thin wrappers around repository-native commands; let composite targets reuse atomic targets instead of duplicating their commands",
			"Declare non-file targets with `.PHONY` and use overridable tool variables when they improve portability",
			"Do not leave TODO recipes, echo-only placeholders, guessed commands, or duplicated build logic",
			"Run `make help` and each added target that is safe to execute, and report any target that could not be validated",
		)
		doc.Paragraph("Rules:")
		doc.BulletList(
			"Initial product ideas and feature intent belong in the accepted native plan and relevant SPEC.md, not in the Constitution until implementation demonstrates project-wide truth",
			"PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times",
		)
		doc.Paragraph("Delivery of command-created files:")
		doc.BulletList(managedFileDeliveryInstructions(projectRoot)...)
	})
}

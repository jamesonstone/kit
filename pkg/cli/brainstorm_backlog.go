package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

func runBrainstormBacklog(
	projectRoot string,
	cfg *config.Config,
	specsDir string,
	args []string,
) error {
	currentActive, err := feature.FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to resolve active feature context: %w", err)
	}

	featureRef, err := promptBrainstormFeatureRef(args)
	if err != nil {
		return err
	}

	feat, created, err := feature.EnsureExists(cfg, specsDir, featureRef)
	if err != nil {
		return err
	}
	feature.ApplyLifecycleState(feat, cfg)

	if !created && feat.Phase != feature.PhaseBrainstorm {
		return fmt.Errorf(
			"feature '%s' is in %s phase. Backlog capture only supports brainstorm-phase features",
			feat.Slug,
			feat.Phase,
		)
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	createdBrainstorm := false
	if !document.Exists(brainstormPath) {
		thesis, err := promptBrainstormThesis(
			newFreeTextInputConfig(brainstormUseVim, brainstormEditor, brainstormInline, true),
		)
		if err != nil {
			return err
		}
		if err := document.Write(brainstormPath, templates.BuildBrainstormArtifact(thesis)); err != nil {
			return fmt.Errorf("failed to create BRAINSTORM.md: %w", err)
		}
		createdBrainstorm = true
	}

	linked, err := addBacklogRelationship(brainstormPath, currentActive, feat)
	if err != nil {
		return err
	}

	alreadyPaused := feat.Paused
	if !alreadyPaused {
		if err := feature.PersistPaused(projectRoot, cfg, feat, true); err != nil {
			return fmt.Errorf("failed to persist paused state: %w", err)
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return fmt.Errorf(
			"backlog item '%s' state updated but failed to refresh PROJECT_PROGRESS_SUMMARY.md: %w",
			feat.Slug,
			err,
		)
	}

	if created {
		fmt.Printf("📁 Created feature directory: %s\n", feat.DirName)
	} else {
		fmt.Printf("📁 Using existing feature: %s\n", feat.DirName)
	}
	if createdBrainstorm {
		fmt.Println("  ✓ Created BRAINSTORM.md")
	} else {
		fmt.Println("  ✓ BRAINSTORM.md already exists")
	}
	if linked && currentActive != nil {
		fmt.Printf("  ✓ Linked backlog item to %s\n", currentActive.DirName)
	}
	if alreadyPaused {
		fmt.Printf("⏸️ Backlog item '%s' is already deferred\n", feat.Slug)
	} else {
		fmt.Printf("⏸️ Captured backlog item '%s'\n", feat.Slug)
	}

	printWorkflowInstructions("backlog capture", []string{
		"run kit backlog to review deferred items",
		fmt.Sprintf("run kit resume %s when you are ready to resume this item", feat.Slug),
		fmt.Sprintf("or run kit backlog --pickup %s for the backlog-specific shortcut", feat.Slug),
	})

	return nil
}

func runBrainstormPickup(
	projectRoot string,
	cfg *config.Config,
	specsDir string,
	args []string,
	outputOnly bool,
) error {
	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 1 {
		feat, err = resolveBacklogFeature(specsDir, cfg, args[0])
		if err != nil {
			return err
		}
	} else {
		feat, err = selectBacklogFeature(specsDir, cfg, "Select a backlog item to pick up:")
		if err != nil {
			return err
		}
	}

	return resumeBacklogFeature(
		projectRoot,
		cfg,
		feat,
		outputOnly,
		brainstormCopy,
		brainstormOutput,
		"brainstorm pickup",
	)
}

func addBacklogRelationship(
	brainstormPath string,
	currentActive *feature.Feature,
	backlogFeat *feature.Feature,
) (bool, error) {
	if currentActive == nil || backlogFeat == nil || currentActive.DirName == backlogFeat.DirName {
		return false, nil
	}

	content, err := os.ReadFile(brainstormPath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", brainstormPath, err)
	}

	relation := fmt.Sprintf("- related to: %s", currentActive.DirName)
	doc := document.Parse(string(content), brainstormPath, document.TypeBrainstorm)
	section := doc.GetSection("RELATIONSHIPS")
	if section == nil {
		return false, nil
	}

	current := strings.TrimSpace(section.Content)
	if strings.Contains(current, relation) {
		return false, nil
	}

	updatedSection := relation
	if current != "" && !strings.EqualFold(current, "none") {
		updatedSection = current + "\n" + relation
	}

	updated, ok := replaceMarkdownSection(string(content), "RELATIONSHIPS", updatedSection)
	if !ok || updated == string(content) {
		return false, nil
	}

	if err := document.Write(brainstormPath, updated); err != nil {
		return false, fmt.Errorf("failed to update relationships in %s: %w", brainstormPath, err)
	}

	return true, nil
}

func replaceMarkdownSection(content, sectionName, sectionBody string) (string, bool) {
	lines := strings.Split(content, "\n")
	header := "## " + sectionName
	start := -1
	end := len(lines)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if start == -1 {
			if trimmed == header {
				start = i
			}
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			end = i
			break
		}
	}

	if start == -1 {
		return content, false
	}

	replacementLines := []string{header, ""}
	replacementLines = append(replacementLines, strings.Split(sectionBody, "\n")...)
	replacementLines = append(replacementLines, "")

	updatedLines := append([]string{}, lines[:start]...)
	updatedLines = append(updatedLines, replacementLines...)
	updatedLines = append(updatedLines, lines[end:]...)

	return strings.Join(updatedLines, "\n"), true
}

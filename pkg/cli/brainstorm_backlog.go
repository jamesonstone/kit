package cli

import (
	"fmt"
	"os"
	"path/filepath"

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

	feat, created, err := feature.EnsureExists(cfg, projectRoot, specsDir, featureRef)
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

	_, notesRelPath, err := ensureFeatureNotesDir(projectRoot, feat.DirName)
	if err != nil {
		return err
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
		content := templates.BuildBrainstormArtifactForFeature(
			thesis,
			document.FeatureMetadataFromDir(feat.DirName),
			[]document.MetadataDependency{featureNotesDependency(notesRelPath)},
		)
		if err := document.Write(brainstormPath, content); err != nil {
			return fmt.Errorf("failed to create BRAINSTORM.md: %w", err)
		}
		createdBrainstorm = true
	} else {
		if _, err := ensureBrainstormNotesDependency(brainstormPath, notesRelPath); err != nil {
			return err
		}
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

	relationship := document.MetadataRelationship{
		Type:   document.RelationshipRelatedTo,
		Target: currentActive.DirName,
	}
	doc := document.Parse(string(content), brainstormPath, document.TypeBrainstorm)
	existingRelationships, _ := doc.Relationships()
	for _, existing := range existingRelationships {
		if existing.Type == "related to" && existing.Target == currentActive.DirName {
			return false, nil
		}
	}
	if containsMetadataRelationship(doc.Metadata, relationship) {
		return false, nil
	}

	relationships := []document.MetadataRelationship{relationship}
	for _, existing := range existingRelationships {
		machine, ok := document.RelationshipHumanToMachine(existing.Type)
		if !ok || existing.Target == "" {
			continue
		}
		relationships = append(relationships, document.MetadataRelationship{
			Type:   machine,
			Target: existing.Target,
		})
	}

	updated, changed, err := document.UpsertMetadata(string(content), document.TypeBrainstorm, document.MetadataUpsert{
		Feature:       document.FeatureMetadataFromDir(backlogFeat.DirName),
		Relationships: relationships,
	})
	if err != nil {
		return false, fmt.Errorf("failed to update relationships in %s: %w", brainstormPath, err)
	}
	if !changed {
		return false, nil
	}

	if err := document.Write(brainstormPath, updated); err != nil {
		return false, fmt.Errorf("failed to update relationships in %s: %w", brainstormPath, err)
	}

	return true, nil
}

func containsMetadataRelationship(metadata *document.Metadata, relationship document.MetadataRelationship) bool {
	if metadata == nil {
		return false
	}
	for _, existing := range metadata.Relationships {
		if existing.Type == relationship.Type && existing.Target == relationship.Target {
			return true
		}
	}
	return false
}

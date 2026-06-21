package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/runstore"
	"github.com/jamesonstone/kit/internal/verify"
)

func validateLoopProgress(before, after loopStageState, _ feature.TaskProgress) error {
	if after.Stage == loopStageBlocked {
		return nil
	}
	if loopStageRank(after.Stage) > loopStageRank(before.Stage) {
		return nil
	}
	if len(after.Diagnostics) > 0 {
		return fmt.Errorf("strict validation failed for %s: %s", after.Stage, strings.Join(after.Diagnostics, "; "))
	}
	return fmt.Errorf("stage %s did not advance after agent reported done", before.Stage)
}

func stopOnFailedVerification(projectRoot string, feat *feature.Feature, since time.Time) error {
	run, ok, err := runstore.LatestForFeature(projectRoot, feat.DirName)
	if err != nil {
		return err
	}
	if ok && !since.IsZero() && run.StartedAt.Before(since) {
		return nil
	}
	if !ok || run.Status != verify.RunStatusFail {
		return nil
	}
	return fmt.Errorf("latest verification run failed: %s", run.RunID)
}

func resolveStrictLoopStage(projectRoot string, feat *feature.Feature) loopStageState {
	specPath := filepath.Join(feat.Path, "SPEC.md")

	if diagnostics := validateLoopDocument(projectRoot, feat, specPath, document.TypeSpec); len(diagnostics) > 0 {
		return loopStageState{Stage: loopStageClarify, Diagnostics: diagnostics}
	}

	doc, err := document.ParseFile(specPath, document.TypeSpec)
	if err != nil {
		return loopStageState{Stage: loopStageClarify, Diagnostics: []string{err.Error()}}
	}
	if doc.Metadata == nil || doc.Metadata.WorkflowVersion != 2 {
		return loopStageState{Stage: loopStageClarify, Diagnostics: []string{"SPEC.md is not marked workflow_version: 2"}}
	}
	phase, ok := feature.V2PhaseFromString(doc.Metadata.Phase)
	if !ok {
		return loopStageState{Stage: loopStageClarify, Diagnostics: []string{"SPEC.md has missing or invalid v2 phase"}}
	}
	return loopStageState{Stage: loopStage(phase)}
}

func validateLoopDocument(projectRoot string, feat *feature.Feature, path string, docType document.DocumentType) []string {
	if !document.Exists(path) {
		return []string{fmt.Sprintf("%s not found", filepath.Base(path))}
	}
	doc, err := document.ParseFile(path, docType)
	if err != nil {
		return []string{err.Error()}
	}
	var diagnostics []string
	for _, validationErr := range doc.Validate() {
		diagnostics = append(diagnostics, validationErr.Error())
	}
	diagnostics = append(diagnostics, featureMetadataIdentityErrors(doc, feat.DirName)...)
	diagnostics = append(diagnostics, featureRulesetReferenceErrors(projectRoot, doc)...)
	if doc.HasUnresolvedPlaceholders() {
		diagnostics = append(diagnostics, fmt.Sprintf("%s has unresolved TODO placeholders", filepath.Base(path)))
	}
	return diagnostics
}

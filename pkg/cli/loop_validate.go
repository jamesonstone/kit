package cli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/runstore"
	"github.com/jamesonstone/kit/internal/verify"
)

var stableAcceptanceIDPattern = regexp.MustCompile(`\bAC-\d{3}\b`)

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
	return resolveStrictLoopStageWithMinConfidence(projectRoot, feat, 95)
}

func resolveStrictLoopStageWithMinConfidence(projectRoot string, feat *feature.Feature, minConfidence int) loopStageState {
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
	stage := loopStage(phase)
	if loopStageRank(stage) > loopStageRank(loopStageClarify) {
		if diagnostics := clarifyReadinessDiagnostics(doc, minConfidence); len(diagnostics) > 0 {
			return loopStageState{Stage: loopStageClarify, Diagnostics: diagnostics}
		}
	}
	return loopStageState{Stage: stage}
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

func clarifyReadinessDiagnostics(doc *document.Document, minConfidence int) []string {
	if minConfidence <= 0 {
		minConfidence = 95
	}
	clarification, ok := doc.ClarificationState()
	if !ok {
		return []string{"SPEC.md cannot leave clarify without front matter clarification state"}
	}

	var diagnostics []string
	if clarification.Status != document.ClarificationStatusReady {
		diagnostics = append(diagnostics, fmt.Sprintf("clarification.status is %q, want %q before leaving clarify", clarification.Status, document.ClarificationStatusReady))
	}
	confidence, hasConfidence := clarification.ConfidenceValue()
	if !hasConfidence {
		diagnostics = append(diagnostics, "clarification.confidence is missing before leaving clarify")
	} else if confidence < minConfidence {
		diagnostics = append(diagnostics, fmt.Sprintf("clarification.confidence %d is below required %d before leaving clarify", confidence, minConfidence))
	}
	unresolved, hasUnresolved := clarification.UnresolvedQuestionsValue()
	if !hasUnresolved {
		diagnostics = append(diagnostics, "clarification.unresolved_questions is missing before leaving clarify")
	} else if unresolved != 0 {
		diagnostics = append(diagnostics, fmt.Sprintf("clarification.unresolved_questions is %d, want 0 before leaving clarify", unresolved))
	}
	acceptanceSection := doc.GetSection("ACCEPTANCE CRITERIA")
	acceptanceIDs := stableAcceptanceIDs(acceptanceSection)
	if len(acceptanceIDs) == 0 {
		diagnostics = append(diagnostics, "Acceptance Criteria must include stable AC-### IDs before leaving clarify")
	}
	validationSection := doc.GetSection("VALIDATION MAP")
	validationIDs := stableAcceptanceIDs(validationSection)
	if len(validationIDs) == 0 {
		diagnostics = append(diagnostics, "Validation Map must reference stable AC-### IDs before leaving clarify")
	} else {
		for id := range acceptanceIDs {
			if !validationIDs[id] {
				diagnostics = append(diagnostics, fmt.Sprintf("Validation Map must reference acceptance criterion %s before leaving clarify", id))
			}
		}
	}
	return diagnostics
}

func stableAcceptanceIDs(section *document.Section) map[string]bool {
	ids := make(map[string]bool)
	if section == nil {
		return ids
	}
	for _, id := range stableAcceptanceIDPattern.FindAllString(section.Content, -1) {
		ids[id] = true
	}
	return ids
}

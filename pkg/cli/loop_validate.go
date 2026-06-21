package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/runstore"
	"github.com/jamesonstone/kit/internal/verify"
)

func validateLoopProgress(before, after loopStageState, previousImplement feature.TaskProgress) error {
	if loopStageRank(after.Stage) > loopStageRank(before.Stage) {
		return nil
	}
	if len(after.Diagnostics) > 0 {
		return fmt.Errorf("strict validation failed for %s: %s", after.Stage, strings.Join(after.Diagnostics, "; "))
	}
	if before.Stage == loopStageImplement && after.Stage == loopStageImplement {
		if after.TasksDone > previousImplement.Complete || after.TasksTotal != previousImplement.Total {
			return nil
		}
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
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")

	if diagnostics := validateLoopDocument(projectRoot, feat, specPath, document.TypeSpec); len(diagnostics) > 0 {
		return loopStageState{Stage: loopStageSpec, Diagnostics: diagnostics}
	}
	if diagnostics := validateLoopDocument(projectRoot, feat, planPath, document.TypePlan); len(diagnostics) > 0 {
		return loopStageState{Stage: loopStagePlan, Diagnostics: diagnostics}
	}
	if diagnostics := validateLoopDocument(projectRoot, feat, tasksPath, document.TypeTasks); len(diagnostics) > 0 {
		progress, _ := feature.ParseTaskProgress(tasksPath)
		return loopStageState{Stage: loopStageTasks, Diagnostics: diagnostics, TasksTotal: progress.Total, TasksDone: progress.Complete}
	}
	progress, err := feature.ParseTaskProgress(tasksPath)
	if err != nil {
		return loopStageState{Stage: loopStageTasks, Diagnostics: []string{err.Error()}}
	}
	if progress.Total == 0 {
		return loopStageState{Stage: loopStageTasks, Diagnostics: []string{"TASKS.md has no markdown checkbox tasks"}}
	}
	if progress.Complete < progress.Total {
		return loopStageState{Stage: loopStageImplement, TasksTotal: progress.Total, TasksDone: progress.Complete}
	}
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		return loopStageState{Stage: loopStageReflect, Diagnostics: []string{err.Error()}, TasksTotal: progress.Total, TasksDone: progress.Complete}
	}
	if strings.Contains(string(data), feature.ReflectionCompleteMarker) {
		return loopStageState{Stage: loopStageComplete, TasksTotal: progress.Total, TasksDone: progress.Complete}
	}
	return loopStageState{Stage: loopStageReflect, TasksTotal: progress.Total, TasksDone: progress.Complete}
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

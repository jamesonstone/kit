package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

func buildLoopPromptForStage(projectRoot string, cfg *config.Config, feat *feature.Feature, stage loopStage, minConfidence int) (string, error) {
	if err := ensureLoopStageArtifact(projectRoot, cfg, feat, stage); err != nil {
		return "", err
	}
	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")

	var base string
	switch stage {
	case loopStageSpec:
		base = buildSpecTemplatePrompt(specPath, brainstormPath, feat.Slug, projectRoot, cfg)
	case loopStagePlan:
		base = buildStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, projectRoot)
	case loopStageTasks:
		base = buildTasksPrompt(feat, projectRoot, cfg)
	case loopStageImplement:
		summary, _ := feature.ExtractSpecSummary(specPath)
		base = buildImplementationPrompt(feat, brainstormPath, specPath, planPath, tasksPath, summary, projectRoot)
	case loopStageReflect:
		base = buildReflectPrompt(
			projectRoot,
			filepath.Join(projectRoot, "docs", "CONSTITUTION.md"),
			cfg.ProgressSummaryPath(projectRoot),
			brainstormPath,
			specPath,
			planPath,
			tasksPath,
			feat.Slug,
		)
	default:
		return "", fmt.Errorf("stage %s does not produce a loop prompt", stage)
	}

	return appendLoopContract(prepareAgentPromptForFeature(base, feat.Path), stage, minConfidence), nil
}

func ensureLoopStageArtifact(projectRoot string, cfg *config.Config, feat *feature.Feature, stage loopStage) error {
	switch stage {
	case loopStageSpec:
		specPath := filepath.Join(feat.Path, "SPEC.md")
		if !document.Exists(specPath) {
			content := templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(specPath, content); err != nil {
				return fmt.Errorf("failed to create SPEC.md: %w", err)
			}
		}
		if effectivePromptProfile(feat.Path) == promptProfileFrontend {
			if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
				return err
			}
		}
	case loopStagePlan:
		planPath := filepath.Join(feat.Path, "PLAN.md")
		if !document.Exists(planPath) {
			content := templates.BuildPlanArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(planPath, content); err != nil {
				return fmt.Errorf("failed to create PLAN.md: %w", err)
			}
		}
		if effectivePromptProfile(feat.Path) == promptProfileFrontend {
			specPath := filepath.Join(feat.Path, "SPEC.md")
			if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
				return err
			}
			if _, err := ensureFrontendProfileDependencyRows(planPath, document.TypePlan, feat.DirName); err != nil {
				return err
			}
		}
	case loopStageTasks:
		tasksPath := filepath.Join(feat.Path, "TASKS.md")
		if !document.Exists(tasksPath) {
			content := templates.BuildTasksArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
			if err := document.Write(tasksPath, content); err != nil {
				return fmt.Errorf("failed to create TASKS.md: %w", err)
			}
		}
	}
	return rollup.Update(projectRoot, cfg)
}

func appendLoopContract(prompt string, stage loopStage, minConfidence int) string {
	var builder strings.Builder
	builder.WriteString(strings.TrimRight(prompt, "\n"))
	builder.WriteString("\n\n## Kit Loop Contract\n\n")
	builder.WriteString("This run is controlled by `kit loop`. Complete the current stage, write all durable changes to repository files, and end your final output with exactly one machine-readable result line.\n\n")
	builder.WriteString("- Do not report `status: \"done\"` unless the stage artifact or implementation state is actually complete.\n")
	builder.WriteString(fmt.Sprintf("- Do not proceed with unresolved assumptions or confidence below %d.\n", minConfidence))
	builder.WriteString("- If blocked, set `status` to `blocked`, include concrete blockers, and do not guess.\n")
	builder.WriteString("- The result line must be a single line of JSON prefixed with `KIT_LOOP_RESULT:`.\n\n")
	builder.WriteString("Required result line:\n\n")
	builder.WriteString("```text\n")
	builder.WriteString(fmt.Sprintf("KIT_LOOP_RESULT: {\"stage\":\"%s\",\"status\":\"done\",\"confidence\":%d,\"blockers\":[]}\n", stage, minConfidence))
	builder.WriteString("```\n")
	return builder.String()
}

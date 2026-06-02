package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/feature"
)

func renderSupportHandoffPrompt() (string, error) {
	prompt, err := projectHandoff()
	if err != nil {
		return "", err
	}
	return prepareAgentPrompt(prompt), nil
}

func renderSupportSummarizePrompt() (string, error) {
	return prepareAgentPrompt(genericSummarizeInstructions()), nil
}

func renderSupportCodeReviewPrompt() (string, error) {
	return prepareAgentPrompt(codeReviewInstructions()), nil
}

func renderWorkflowBrainstormPrompt() (string, error) {
	ctx, err := activePromptFeatureContext("workflow brainstorm", "BRAINSTORM.md")
	if err != nil {
		return "", err
	}

	brainstormPath := filepath.Join(ctx.Feature.Path, "BRAINSTORM.md")
	prompt := buildBrainstormPrompt(
		brainstormPath,
		ctx.Feature.Slug,
		ctx.ProjectRoot,
		existingBrainstormThesis(brainstormPath),
		ctx.Config.GoalPercentage,
	)
	return prepareAgentPromptForFeature(prompt, ctx.Feature.Path), nil
}

func renderWorkflowSpecPrompt() (string, error) {
	ctx, err := activePromptFeatureContext("workflow spec", "SPEC.md")
	if err != nil {
		return "", err
	}

	specPath := filepath.Join(ctx.Feature.Path, "SPEC.md")
	brainstormPath := filepath.Join(ctx.Feature.Path, "BRAINSTORM.md")
	prompt := buildSpecTemplatePrompt(specPath, brainstormPath, ctx.Feature.Slug, ctx.ProjectRoot, ctx.Config)
	return prepareAgentPromptForFeature(prompt, ctx.Feature.Path), nil
}

func renderWorkflowPlanPrompt() (string, error) {
	ctx, err := activePromptFeatureContext("workflow plan", "SPEC.md", "PLAN.md")
	if err != nil {
		return "", err
	}

	specPath := filepath.Join(ctx.Feature.Path, "SPEC.md")
	planPath := filepath.Join(ctx.Feature.Path, "PLAN.md")
	brainstormPath := filepath.Join(ctx.Feature.Path, "BRAINSTORM.md")
	prompt := buildStandardPlanPrompt(planPath, specPath, brainstormPath, ctx.Feature, ctx.Config, ctx.ProjectRoot)
	return prepareAgentPromptForFeature(prompt, ctx.Feature.Path), nil
}

func renderWorkflowTasksPrompt() (string, error) {
	ctx, err := activePromptFeatureContext("workflow tasks", "SPEC.md", "PLAN.md", "TASKS.md")
	if err != nil {
		return "", err
	}

	prompt := buildTasksPrompt(ctx.Feature, ctx.ProjectRoot, ctx.Config)
	return prepareAgentPromptForFeature(prompt, ctx.Feature.Path), nil
}

func renderWorkflowImplementPrompt() (string, error) {
	ctx, err := activePromptFeatureContext("workflow implement", "SPEC.md", "PLAN.md", "TASKS.md")
	if err != nil {
		return "", err
	}

	specPath := filepath.Join(ctx.Feature.Path, "SPEC.md")
	planPath := filepath.Join(ctx.Feature.Path, "PLAN.md")
	tasksPath := filepath.Join(ctx.Feature.Path, "TASKS.md")
	brainstormPath := filepath.Join(ctx.Feature.Path, "BRAINSTORM.md")
	summary, _ := feature.ExtractSpecSummary(specPath)

	prompt := buildImplementationPrompt(
		ctx.Feature,
		brainstormPath,
		specPath,
		planPath,
		tasksPath,
		summary,
		ctx.ProjectRoot,
	)
	return prepareAgentPromptForFeature(prompt, ctx.Feature.Path), nil
}

func renderWorkflowReflectPrompt() (string, error) {
	ctx, err := optionalPromptFeatureContext()
	if err != nil {
		return "", err
	}

	constitutionPath := filepath.Join(ctx.ProjectRoot, "docs", "CONSTITUTION.md")
	summaryPath := ctx.Config.ProgressSummaryPath(ctx.ProjectRoot)
	if ctx.Feature == nil {
		return prepareAgentPrompt(buildReflectPrompt(ctx.ProjectRoot, constitutionPath, summaryPath, "", "", "", "", "")), nil
	}

	brainstormPath := filepath.Join(ctx.Feature.Path, "BRAINSTORM.md")
	specPath := filepath.Join(ctx.Feature.Path, "SPEC.md")
	planPath := filepath.Join(ctx.Feature.Path, "PLAN.md")
	tasksPath := filepath.Join(ctx.Feature.Path, "TASKS.md")
	prompt := buildReflectPrompt(
		ctx.ProjectRoot,
		constitutionPath,
		summaryPath,
		brainstormPath,
		specPath,
		planPath,
		tasksPath,
		ctx.Feature.Slug,
	)
	return prepareAgentPromptForFeature(prompt, ctx.Feature.Path), nil
}

func renderSupportResumePrompt() (string, error) {
	ctx, err := activePromptFeatureContext("support resume")
	if err != nil {
		return "", err
	}

	var prompt string
	if feature.IsBacklogItem(*ctx.Feature) {
		if err := requirePromptFeatureDocs("support resume", ctx.Feature, "BRAINSTORM.md"); err != nil {
			return "", err
		}
		brainstormPath := filepath.Join(ctx.Feature.Path, "BRAINSTORM.md")
		prompt = buildBrainstormPrompt(
			brainstormPath,
			ctx.Feature.Slug,
			ctx.ProjectRoot,
			existingBrainstormThesis(brainstormPath),
			ctx.Config.GoalPercentage,
		)
		return prepareAgentPromptForFeature(prompt, ctx.Feature.Path), nil
	}

	status, err := feature.GetFeatureStatus(ctx.Feature)
	if err != nil {
		return "", fmt.Errorf("failed to get feature status: %w", err)
	}
	prompt = buildCatchupPrompt(ctx.Feature, status, ctx.ProjectRoot)
	return prepareAgentPromptForFeature(prompt, ctx.Feature.Path), nil
}

func renderSupportReconcilePrompt() (string, error) {
	projectRoot, cfg, err := promptProjectContext()
	if err != nil {
		return "", err
	}

	report, err := buildReconcileReport(projectRoot, cfg, nil)
	if err != nil {
		return "", err
	}
	if len(report.Findings) == 0 {
		return "", fmt.Errorf("%s", report.cleanResult())
	}
	return prepareAgentPrompt(buildReconcilePrompt(report)), nil
}

func renderSupportDispatchPrompt() (string, error) {
	context, err := collectMissingPromptContext(
		"support dispatch",
		"a task list to dispatch",
		"dispatch tasks",
		promptDefaultEditorConfig(),
	)
	if err != nil {
		return "", err
	}

	tasks, err := normalizeDispatchTasks(context)
	if err != nil {
		return "", err
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	prompt := buildDispatchPrompt(tasks, 10, workingDirectory, dispatchInputSourceEditor, dispatchPromptOptions{})
	return preparePromptWithoutSubagents(prompt), nil
}

func renderSkillMinePrompt() (string, error) {
	ctx, err := activePromptFeatureContext("skill mine", "SPEC.md", "PLAN.md", "TASKS.md")
	if err != nil {
		return "", err
	}

	brainstormPath := filepath.Join(ctx.Feature.Path, "BRAINSTORM.md")
	specPath := filepath.Join(ctx.Feature.Path, "SPEC.md")
	planPath := filepath.Join(ctx.Feature.Path, "PLAN.md")
	tasksPath := filepath.Join(ctx.Feature.Path, "TASKS.md")

	prompt := buildSkillMinePrompt(
		ctx.Feature,
		brainstormPath,
		specPath,
		planPath,
		tasksPath,
		ctx.Config.SkillsPath(ctx.ProjectRoot),
		ctx.ProjectRoot,
	)
	return prepareAgentPromptForFeature(prompt, ctx.Feature.Path), nil
}

func renderProjectInitPrompt() (string, error) {
	projectRoot, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	return prepareAgentPrompt(buildProjectInitPrompt(projectRoot, constitutionPath)), nil
}

func renderProjectRefreshPrompt() (string, error) {
	projectRoot, cfg, err := promptProjectContext()
	if err != nil {
		return "", err
	}

	return prepareAgentPrompt(buildProjectRefreshPrompt(projectRoot, cfg)), nil
}

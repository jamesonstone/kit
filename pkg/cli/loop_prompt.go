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

	if stage == loopStageComplete || stage == loopStageBlocked {
		return "", fmt.Errorf("stage %s does not produce a loop prompt", stage)
	}

	base := buildSpecTemplatePrompt(specPath, brainstormPath, feat.Slug, projectRoot, cfg, false)
	return appendLoopContract(prepareAgentPromptForFeature(base, feat.Path), stage, minConfidence), nil
}

func ensureLoopStageArtifact(projectRoot string, cfg *config.Config, feat *feature.Feature, stage loopStage) error {
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		content := templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(specPath, content); err != nil {
			return fmt.Errorf("failed to create SPEC.md: %w", err)
		}
	}
	if _, err := ensureSpecV2Adoption(specPath, projectRoot, feat.DirName, cfg.GoalPercentage); err != nil {
		return err
	}
	if effectivePromptProfile(feat.Path) == promptProfileFrontend {
		if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
			return err
		}
	}
	return rollup.Update(projectRoot, cfg)
}

func appendLoopContract(prompt string, stage loopStage, minConfidence int) string {
	var builder strings.Builder
	builder.WriteString(strings.TrimRight(prompt, "\n"))
	builder.WriteString("\n\n## Kit Loop Contract\n\n")
	builder.WriteString("This run is controlled by `kit loop workflow`. Advance the current v2 SPEC.md phase, write all durable changes to repository files, and end your final output with exactly one machine-readable result line.\n\n")
	builder.WriteString("- Do not report `status: \"done\"` unless SPEC.md is updated and the current phase has advanced, completed, or become blocked with explicit blockers.\n")
	builder.WriteString(fmt.Sprintf("- Do not proceed with unresolved assumptions or confidence below %d.\n", minConfidence))
	builder.WriteString("- During the clarify stage, research only discoverable facts. If user intent is still ambiguous, set `status` to `blocked`, include the exact clarification questions, and do not guess.\n")
	builder.WriteString("- If blocked, set `status` to `blocked`, include concrete blockers, and do not guess.\n")
	builder.WriteString("- The result line must be a single line of JSON prefixed with `KIT_LOOP_RESULT:`.\n\n")
	builder.WriteString("Required result line:\n\n")
	builder.WriteString("```text\n")
	builder.WriteString(fmt.Sprintf("KIT_LOOP_RESULT: {\"stage\":\"%s\",\"status\":\"done\",\"confidence\":%d,\"blockers\":[]}\n", stage, minConfidence))
	builder.WriteString("```\n")
	return builder.String()
}

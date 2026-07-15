package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

func buildLoopPromptForStage(projectRoot string, cfg *config.Config, feat *feature.Feature, stage loopStage, minConfidence int) (string, error) {
	if err := ensureLoopStageArtifact(projectRoot, cfg, feat, stage); err != nil {
		return "", err
	}
	specPath := filepath.Join(feat.Path, "SPEC.md")

	if stage == loopStageComplete || stage == loopStageBlocked {
		return "", fmt.Errorf("stage %s does not produce a loop prompt", stage)
	}

	base := buildLoopStagePrompt(projectRoot, feat, specPath, stage, minConfidence)
	return appendLoopContract(prepareAgentPromptForFeature(base, feat.Path), stage, minConfidence), nil
}

func buildLoopStagePrompt(projectRoot string, feat *feature.Feature, specPath string, stage loopStage, minConfidence int) string {
	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Advance feature `%s` through the `%s` phase only.", feat.Slug, stage))
		doc.Heading(2, "Goal")
		doc.BulletList(
			fmt.Sprintf("Treat `%s` as the source of truth and update it with every durable decision, result, and evidence item from this phase.", specPath),
			"Complete this phase's success criteria and advance the front matter phase, or record a concrete blocker.",
		)
		doc.Heading(2, "Context")
		doc.BulletList(
			fmt.Sprintf("Project root: `%s`", projectRoot),
			fmt.Sprintf("Feature directory: `%s`", feat.Path),
			"Load repository guidance and referenced sources only when they affect this phase's decision.",
		)
		doc.Heading(2, "Phase Contract")
		doc.BulletList(loopStageContract(stage, minConfidence)...)
		doc.Heading(2, "Constraints")
		doc.BulletList(
			"Preserve unrelated or user-owned work and obey code-enforced phase, validation, evidence, and delivery gates.",
			"Proceed with safe in-scope discovery and edits. Ask only when a material decision is non-discoverable or an external/irreversible action needs authority.",
			"Do not perform work belonging to a later phase merely to make this phase appear complete.",
		)
	})
}

func loopStageContract(stage loopStage, minConfidence int) []string {
	switch stage {
	case loopStageClarify:
		return []string{
			"Research repository-discoverable facts and make requirements, non-goals, assumptions, source map, acceptance criteria, validation map, rollback, and delivery intent explicit.",
			fmt.Sprintf("Move to `ready` only when confidence is at least %d, unresolved questions are 0, and criteria are binary-verifiable.", minConfidence),
			"If a material user choice remains, update `SPEC.md`, set the phase/status to blocked or open as appropriate, and return exact numbered questions with recommended defaults.",
		}
	case loopStageReady:
		return []string{
			"Audit readiness: stable requirements and acceptance IDs, 1:1 validation, implementation plan, task checklist, predicted files, dirty-work ownership, rollback, delivery intent, and agent routing.",
			"Resolve discoverable gaps in `SPEC.md`; block only on a material non-discoverable decision.",
			"When every readiness gate passes, set phase to `implement` without starting unrelated implementation work in this run.",
		}
	case loopStageImplement:
		return []string{
			"Execute the in-scope task checklist using existing repository patterns and the smallest coherent diff.",
			"Update focused tests and affected documentation with the behavior; record files, task state, risks, and validation still required.",
			"Run focused checks needed to keep the implementation safe. Move to `validate` only when implementation tasks are complete and no known implementation defect remains.",
		}
	case loopStageValidate:
		return []string{
			"Run the validation map and relevant regressions; record exact commands/results and documentation or runtime evidence for every acceptance criterion.",
			"Use read-only verification when required. Fix valid findings and rerun affected checks; route implementation gaps back to `implement` rather than waiving them.",
			"Move to `reflect` only when every criterion is proven or a genuine blocker is recorded with risk and owner.",
		}
	case loopStageReflect:
		return []string{
			"Review the integrated diff against the thesis, requirements, acceptance criteria, tests, runtime behavior, docs, evidence, and scope.",
			"Find regressions, hidden scope creep, dead code, unnecessary surfaces, error-handling gaps, and stale documentation; route any gap back to implementation/validation.",
			"Record reflection and remaining risk. Move to `deliver` only when the result is coherent and no known correctness gap remains.",
		}
	case loopStageDeliver:
		return []string{
			"Read the repo-local delivery rules and current Delivery Decision; establish the exact issue, branch, identity, staging, commit, push, PR, assignee, and check contract before mutation.",
			"Perform only delivery actions already authorized and only after implementation, validation, reflection, docs, and evidence are complete. Stop on an unknown contract field.",
			"Record exact delivery results and literal check state. Move to `complete` only when the requested delivery outcome or documented no-delivery decision is satisfied.",
		}
	default:
		return []string{"Record that this stage cannot be executed and return a concrete blocker."}
	}
}

func ensureLoopStageArtifact(projectRoot string, cfg *config.Config, feat *feature.Feature, stage loopStage) error {
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		content := templates.BuildSpecV2ArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
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
	builder.WriteString("`kit loop workflow` controls this run. Update SPEC.md, advance only the current phase, and end with exactly one result line.\n\n")
	builder.WriteString("- Use `status: \"done\"` only after the phase success criteria pass and SPEC.md advances. Otherwise use `blocked` with concrete blockers.\n")
	builder.WriteString(fmt.Sprintf("- Report confidence at least %d only when supported by the completed phase evidence.\n", minConfidence))
	if stage == loopStageClarify {
		builder.WriteString("- If a material user decision remains, include `Open Questions` with numbered questions and recommended defaults. If none remain, state `Open Questions: none`.\n")
	}
	builder.WriteString("- The final line is single-line JSON prefixed with `KIT_LOOP_RESULT:`.\n\n")
	builder.WriteString("Required result line:\n\n")
	builder.WriteString("```text\n")
	builder.WriteString(fmt.Sprintf("KIT_LOOP_RESULT: {\"stage\":\"%s\",\"status\":\"done\",\"confidence\":%d,\"blockers\":[]}\n", stage, minConfidence))
	builder.WriteString("```\n")
	return builder.String()
}

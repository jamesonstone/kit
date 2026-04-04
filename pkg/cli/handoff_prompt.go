package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

type handoffDocument struct {
	File     string
	FullPath string
	HowToUse string
}

func projectHandoff() (string, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return genericHandoffInstructions(), nil
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	return projectHandoffWithConfig(projectRoot, cfg)
}

func projectHandoffWithConfig(projectRoot string, cfg *config.Config) (string, error) {
	specsDir := cfg.SpecsPath(projectRoot)
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return "", fmt.Errorf("projectHandoffWithConfig: failed to list features in %s: %w", specsDir, err)
	}

	activeFeatures := activeHandoffFeatures(features)
	docs := projectHandoffDocuments(projectRoot, cfg, activeFeatures)

	var sb strings.Builder
	sb.WriteString("# Handoff Preparation\n\n")
	sb.WriteString("You are the current coding agent session preparing this project for handoff.\n\n")
	sb.WriteString("Your job is to reconcile project and active-feature documentation with repository reality before transfer.\n\n")

	if len(activeFeatures) > 0 {
		sb.WriteString("## Active Features\n\n")
		sb.WriteString("| Feature | Phase | Full Path |\n")
		sb.WriteString("| ------- | ----- | --------- |\n")
		for _, feat := range activeFeatures {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", feat.Slug, feat.Phase, feat.Path))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Documentation Inventory\n\n")
	sb.WriteString(renderHandoffDocumentTable(docs))
	sb.WriteString("\n")

	sb.WriteString("## Work Instructions\n\n")
	sb.WriteString("1. Review the most recent conversation context already available in this session before changing anything.\n")
	sb.WriteString("2. Summarize that recent context into high-signal facts covering decisions made, blockers, validation results, open questions, and next steps.\n")
	sb.WriteString("3. Read the documentation inventory in order, starting with `CONSTITUTION.md` and `PROJECT_PROGRESS_SUMMARY.md`.\n")
	if len(activeFeatures) > 0 {
		sb.WriteString("4. For each active feature, compare current implementation reality, task state, repository findings, and phase dependency inventories against the listed feature docs.\n")
		sb.WriteString("5. Update any stale feature docs first. If implementation reality diverges from the docs, fix the docs before handoff.\n")
		sb.WriteString("6. For every touched `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md`, make sure the `## DEPENDENCIES` table lists current `active`, `optional`, and `stale` dependencies with exact locations.\n")
		sb.WriteString("7. Update `PROJECT_PROGRESS_SUMMARY.md` so it reflects the reconciled state of every active feature.\n")
		sb.WriteString("8. Keep changes limited to documentation and handoff accuracy. Do not begin unrelated implementation work.\n")
		sb.WriteString("9. If a listed doc is stale, update it before producing your final handoff response.\n")
		sb.WriteString("10. Prefer repository files and current code over memory when they disagree.\n\n")
	} else {
		sb.WriteString("4. Compare the project summary and repository findings to confirm there is no undocumented active work.\n")
		sb.WriteString("5. If you touch any feature docs during reconciliation, make sure each touched `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md` keeps its `## DEPENDENCIES` table current with exact locations.\n")
		sb.WriteString("6. Update any stale project docs first so the handoff is accurate.\n")
		sb.WriteString("7. Keep changes limited to documentation and handoff accuracy. Do not begin unrelated implementation work.\n")
		sb.WriteString("8. If a listed doc is stale, update it before producing your final handoff response.\n")
		sb.WriteString("9. Prefer repository files and current code over memory when they disagree.\n\n")
	}

	sb.WriteString("## Final Response Contract\n\n")
	sb.WriteString("After the documentation is reconciled, reply in stdout/chat with exactly these sections:\n\n")
	sb.WriteString("1. `Documentation Sync`\n")
	sb.WriteString("   - one concise paragraph confirming all relevant documentation files and dependency inventories have been updated and are up to date\n")
	sb.WriteString("   - if you updated docs or dependency tables, name the files you changed in that paragraph\n")
	sb.WriteString("2. `Documentation Files`\n")
	sb.WriteString("   - a markdown table with columns `File`, `Full Path`, and `How To Use`\n")
	sb.WriteString("   - include the reconciled project docs and every relevant active-feature doc\n")
	sb.WriteString("3. `Recent Context`\n")
	sb.WriteString("   - flat bullets for decisions made, blockers, validation results, open questions, and next steps\n")
	sb.WriteString("   - keep this concise and factual\n")

	return sb.String(), nil
}

func featureHandoff(featureRef string) (string, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return "", err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := cfg.SpecsPath(projectRoot)
	feat, err := loadFeatureWithState(specsDir, cfg, featureRef)
	if err != nil {
		return "", fmt.Errorf("feature '%s' not found: %w", featureRef, err)
	}

	docs := featureHandoffDocuments(projectRoot, cfg, feat)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Handoff Preparation — Feature: %s\n\n", feat.Slug))
	sb.WriteString("You are the current coding agent session preparing this feature for handoff.\n\n")
	sb.WriteString("Before transfer, make sure the feature documentation reflects the realities of the implementation and current task state.\n\n")

	sb.WriteString("## Feature State\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("| ----- | ----- |\n")
	sb.WriteString(fmt.Sprintf("| Feature | %s |\n", feat.Slug))
	sb.WriteString(fmt.Sprintf("| Phase | %s |\n", feat.Phase))
	sb.WriteString(fmt.Sprintf("| Paused | %t |\n", feat.Paused))
	sb.WriteString(fmt.Sprintf("| Directory | %s |\n", feat.DirName))
	sb.WriteString(fmt.Sprintf("| Full Path | %s |\n\n", feat.Path))

	sb.WriteString("## Documentation Inventory\n\n")
	sb.WriteString(renderHandoffDocumentTable(docs))
	sb.WriteString("\n")

	sb.WriteString("## Work Instructions\n\n")
	sb.WriteString("1. Review the most recent conversation context already available in this session before changing anything.\n")
	sb.WriteString("2. Summarize that recent context into high-signal facts covering decisions made, blockers, validation results, open questions, and next steps.\n")
	sb.WriteString("3. Read the listed docs in order, starting with `CONSTITUTION.md`, then the feature docs, then `PROJECT_PROGRESS_SUMMARY.md`.\n")
	sb.WriteString("4. Compare current implementation reality, task status, repository findings, and phase dependency inventories against each feature document.\n")
	sb.WriteString("5. If any feature specification document is stale, update it first so it matches reality. Do this before preparing the handoff summary.\n")
	sb.WriteString("6. Verify that `BRAINSTORM.md`, `SPEC.md`, and `PLAN.md` keep their `## DEPENDENCIES` tables current with exact locations and `active`, `optional`, or `stale` status values when those docs exist.\n")
	sb.WriteString("7. Keep `PROJECT_PROGRESS_SUMMARY.md` aligned with the reconciled feature state.\n")
	sb.WriteString("8. Limit your work to documentation reconciliation and handoff preparation. Do not start unrelated implementation work.\n")
	if feat.Phase == feature.PhaseSpec || feat.Phase == feature.PhasePlan {
		sb.WriteString("9. Preserve the planning approval semantics when they still apply: " + approvalSyntaxSummary + ".\n")
	} else {
		sb.WriteString("9. Be explicit about the actual implementation and task state, especially when code has outpaced `PLAN.md` or `TASKS.md`.\n")
	}
	sb.WriteString("10. Prefer repository files and current code over memory when they disagree.\n\n")

	sb.WriteString("## Final Response Contract\n\n")
	sb.WriteString("After the documentation is reconciled, reply in stdout/chat with exactly these sections:\n\n")
	sb.WriteString("1. `Documentation Sync`\n")
	sb.WriteString("   - one concise paragraph confirming all relevant documentation files and dependency inventories have been updated and are up to date\n")
	sb.WriteString("   - if you updated docs or dependency tables, name the files you changed in that paragraph\n")
	sb.WriteString("2. `Documentation Files`\n")
	sb.WriteString("   - a markdown table with columns `File`, `Full Path`, and `How To Use`\n")
	sb.WriteString("   - include every reconciled feature document and relevant project-level doc\n")
	sb.WriteString("3. `Recent Context`\n")
	sb.WriteString("   - flat bullets for decisions made, blockers, validation results, open questions, and next steps\n")
	sb.WriteString("   - keep this concise and factual\n")

	return sb.String(), nil
}

func genericHandoffInstructions() string {
	var sb strings.Builder
	sb.WriteString("# Handoff Preparation\n\n")
	sb.WriteString("You are the current coding agent session preparing this project for handoff.\n\n")
	sb.WriteString("## Work Instructions\n\n")
	sb.WriteString("1. Review the most recent conversation context already available in this session\n")
	sb.WriteString("2. Summarize that recent context into high-signal facts covering decisions made, blockers, validation results, open questions, and next steps\n")
	sb.WriteString("3. Identify the authoritative project documents and make sure they reflect current implementation reality before handoff\n")
	sb.WriteString("4. If this is a Kit project, use `kit handoff` from the project root to generate a feature-aware documentation inventory\n")
	sb.WriteString("5. If the relevant docs include `## DEPENDENCIES` tables, make sure they reflect current `active`, `optional`, and `stale` dependencies with exact locations\n")
	sb.WriteString("6. Prefer repository files and current code over memory when they disagree\n\n")
	sb.WriteString("## Final Response Contract\n\n")
	sb.WriteString("After the docs are reconciled, reply in stdout/chat with:\n\n")
	sb.WriteString("1. `Documentation Sync`\n")
	sb.WriteString("   - one concise paragraph confirming all relevant documentation files and dependency inventories have been updated and are up to date\n")
	sb.WriteString("2. `Documentation Files`\n")
	sb.WriteString("   - a markdown table with columns `File`, `Full Path`, and `How To Use`\n")
	sb.WriteString("3. `Recent Context`\n")
	sb.WriteString("   - flat bullets for decisions made, blockers, validation results, open questions, and next steps\n")
	return sb.String()
}

func activeHandoffFeatures(features []feature.Feature) []feature.Feature {
	var active []feature.Feature
	for _, feat := range features {
		if feat.Phase == feature.PhaseComplete || feat.Paused {
			continue
		}
		active = append(active, feat)
	}
	return active
}

func projectHandoffDocuments(projectRoot string, cfg *config.Config, features []feature.Feature) []handoffDocument {
	docs := []handoffDocument{}

	constitutionPath := cfg.ConstitutionAbsPath(projectRoot)
	if document.Exists(constitutionPath) {
		docs = append(docs, handoffDocument{
			File:     filepath.Base(constitutionPath),
			FullPath: constitutionPath,
			HowToUse: "Project-wide constraints and invariants; read first before reconciling any feature docs",
		})
	}

	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	if document.Exists(summaryPath) {
		docs = append(docs, handoffDocument{
			File:     filepath.Base(summaryPath),
			FullPath: summaryPath,
			HowToUse: "Cross-feature rollup; update after reconciling active feature state",
		})
	}

	for _, feat := range features {
		docs = append(docs, featureScopedDocuments(&feat)...)
	}

	return docs
}

func featureHandoffDocuments(projectRoot string, cfg *config.Config, feat *feature.Feature) []handoffDocument {
	docs := []handoffDocument{}

	constitutionPath := cfg.ConstitutionAbsPath(projectRoot)
	if document.Exists(constitutionPath) {
		docs = append(docs, handoffDocument{
			File:     filepath.Base(constitutionPath),
			FullPath: constitutionPath,
			HowToUse: "Project-wide constraints and invariants; read first before reconciling feature docs",
		})
	}

	docs = append(docs, featureScopedDocuments(feat)...)

	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	if document.Exists(summaryPath) {
		docs = append(docs, handoffDocument{
			File:     filepath.Base(summaryPath),
			FullPath: summaryPath,
			HowToUse: "Cross-feature rollup; update after reconciling this feature's current state",
		})
	}

	return docs
}

func featureScopedDocuments(feat *feature.Feature) []handoffDocument {
	entries := []struct {
		name string
		use  string
	}{
		{"BRAINSTORM.md", fmt.Sprintf("Upstream research for %s; preserve validated findings, affected-file context, and the phase dependency inventory", feat.Slug)},
		{"SPEC.md", fmt.Sprintf("Feature requirements for %s; keep scope, acceptance, edge cases, and the dependency inventory aligned with reality", feat.Slug)},
		{"PLAN.md", fmt.Sprintf("Implementation approach for %s; update the written design and planning dependency inventory when execution diverges", feat.Slug)},
		{"TASKS.md", fmt.Sprintf("Execution state for %s; keep task status and evidence aligned with actual progress", feat.Slug)},
		{"ANALYSIS.md", fmt.Sprintf("Optional scratchpad for %s; keep open questions and assumptions current if present", feat.Slug)},
	}

	docs := make([]handoffDocument, 0, len(entries))
	for _, entry := range entries {
		fullPath := filepath.Join(feat.Path, entry.name)
		if !document.Exists(fullPath) {
			continue
		}
		docs = append(docs, handoffDocument{
			File:     entry.name,
			FullPath: fullPath,
			HowToUse: entry.use,
		})
	}

	return docs
}

func renderHandoffDocumentTable(docs []handoffDocument) string {
	var sb strings.Builder
	sb.WriteString("| File | Full Path | How To Use |\n")
	sb.WriteString("| ---- | --------- | ---------- |\n")
	for _, doc := range docs {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", doc.File, doc.FullPath, doc.HowToUse))
	}
	return sb.String()
}

func ensureHandoffTestWorkingDirectory(projectRoot string) (func(), error) {
	originalWD, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		return nil, fmt.Errorf("failed to change directory: %w", err)
	}

	return func() {
		_ = os.Chdir(originalWD)
	}, nil
}

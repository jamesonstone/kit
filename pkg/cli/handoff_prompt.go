package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
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
	version := detectInstructionScaffoldVersion(projectRoot, cfg)

	workSteps := []string{
		"Review the most recent conversation context already available in this session before changing anything.",
		"Summarize that recent context into high-signal facts covering decisions made, blockers, validation results, open questions, and next steps.",
	}
	if config.UsesInstructionSupportDocs(version) {
		workSteps = append(workSteps, "Read the documentation inventory in order, starting with `CONSTITUTION.md`, then the thin instruction entrypoints and `docs/agents/*`, then `PROJECT_PROGRESS_SUMMARY.md`.")
	} else {
		workSteps = append(workSteps, "Read the documentation inventory in order, starting with `CONSTITUTION.md` and `PROJECT_PROGRESS_SUMMARY.md`.")
	}
	if len(activeFeatures) > 0 {
		workSteps = append(workSteps,
			"For each active feature, compare current implementation reality, SPEC.md phase/task/evidence state, repository findings, and reference inventories against the listed feature docs.",
			"Update any stale feature docs first. If implementation reality diverges from the docs, fix the docs before handoff.",
			"For every touched `SPEC.md`, make sure canonical front matter phase, references, relationships, skills, and evidence remain current. If you touch legacy `BRAINSTORM.md`, `PLAN.md`, or `TASKS.md`, keep their front matter references current too.",
			"Update `PROJECT_PROGRESS_SUMMARY.md` so it reflects the reconciled state of every active feature.",
			"Keep changes limited to documentation and handoff accuracy. Do not begin unrelated implementation work.",
			"If a listed doc is stale, update it before producing your final handoff response.",
			"Prefer repository files and current code over memory when they disagree.",
		)
	} else {
		workSteps = append(workSteps,
			"Compare the project summary and repository findings to confirm there is no undocumented active work.",
			"If you touch any feature docs during reconciliation, make sure each touched `SPEC.md` keeps canonical front matter phase, references, relationships, skills, and evidence current; legacy staged docs only need reference updates when they are touched.",
			"Update any stale project docs first so the handoff is accurate.",
			"Keep changes limited to documentation and handoff accuracy. Do not begin unrelated implementation work.",
			"If a listed doc is stale, update it before producing your final handoff response.",
			"Prefer repository files and current code over memory when they disagree.",
		)
	}

	output := renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Heading(1, "Handoff Preparation")
		doc.Paragraph("You are the current coding agent session preparing this project for handoff.")
		doc.Paragraph("Your job is to reconcile project and active-feature documentation with repository reality before transfer.")
		if len(activeFeatures) > 0 {
			doc.Heading(2, "Active Features")
			rows := make([][]string, 0, len(activeFeatures))
			for _, feat := range activeFeatures {
				rows = append(rows, []string{feat.Slug, string(feat.Phase), feat.Path})
			}
			doc.Table([]string{"Feature", "Phase", "Full Path"}, rows)
		}
		doc.Heading(2, "Documentation Inventory")
		doc.Table([]string{"File", "Full Path", "How To Use"}, handoffDocumentRows(docs))
		doc.Heading(2, "Work Instructions")
		doc.OrderedList(1, workSteps...)
		doc.Heading(2, "Final Response Contract")
		doc.Paragraph("After the documentation is reconciled, reply in stdout/chat with exactly these sections:")
		doc.OrderedList(1,
			"`Documentation Sync`\n- one concise paragraph confirming all relevant documentation files and reference inventories have been updated and are up to date\n- if you updated docs or reference metadata, name the files you changed in that paragraph",
			"`Documentation Files`\n- a markdown table with columns `File`, `Full Path`, and `How To Use`\n- include the reconciled project docs and every relevant active-feature doc",
			"`Recent Context`\n- flat bullets for decisions made, blockers, validation results, open questions, and next steps\n- keep this concise and factual",
		)
	})

	return output, nil
}

func featureHandoff(featureRef string) (string, error) {
	output, _, err := featureHandoffWithPath(featureRef)
	return output, err
}

func featureHandoffWithPath(featureRef string) (string, string, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return "", "", err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", "", fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := cfg.SpecsPath(projectRoot)
	feat, err := loadFeatureWithState(specsDir, cfg, featureRef)
	if err != nil {
		return "", "", fmt.Errorf("feature '%s' not found: %w", featureRef, err)
	}

	docs := featureHandoffDocuments(projectRoot, cfg, feat)
	version := detectInstructionScaffoldVersion(projectRoot, cfg)

	workSteps := []string{
		"Review the most recent conversation context already available in this session before changing anything.",
		"Summarize that recent context into high-signal facts covering decisions made, blockers, validation results, open questions, and next steps.",
	}
	if config.UsesInstructionSupportDocs(version) {
		workSteps = append(workSteps, "Read the listed docs in order, starting with `CONSTITUTION.md`, then the thin instruction entrypoints and `docs/agents/*`, then the feature docs, then `PROJECT_PROGRESS_SUMMARY.md`.")
	} else {
		workSteps = append(workSteps, "Read the listed docs in order, starting with `CONSTITUTION.md`, then the feature docs, then `PROJECT_PROGRESS_SUMMARY.md`.")
	}
	workSteps = append(workSteps,
		"Compare current implementation reality, SPEC.md phase/task/evidence state, repository findings, and reference inventories against each feature document.",
		"If any feature specification document is stale, update it first so it matches reality. Do this before preparing the handoff summary.",
		"Verify that `SPEC.md` keeps canonical front matter phase, references, relationships, skills, validation evidence, and delivery state current. If legacy `BRAINSTORM.md`, `PLAN.md`, or `TASKS.md` exist and are touched, keep their front matter references current too.",
		"Keep `PROJECT_PROGRESS_SUMMARY.md` aligned with the reconciled feature state.",
		"Limit your work to documentation reconciliation and handoff preparation. Do not start unrelated implementation work.",
	)
	if feat.Phase == feature.PhaseClarify || feat.Phase == feature.PhaseReady || feat.Phase == feature.PhaseSpec || feat.Phase == feature.PhasePlan {
		workSteps = append(workSteps, "Preserve clarification/readiness approval semantics when they still apply: "+approvalSyntaxSummary+".")
	} else {
		workSteps = append(workSteps, "Be explicit about the actual implementation, validation, reflection, and delivery state, especially when code has outpaced `SPEC.md`.")
	}
	workSteps = append(workSteps, "Prefer repository files and current code over memory when they disagree.")

	output := renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Heading(1, fmt.Sprintf("Handoff Preparation — Feature: %s", feat.Slug))
		doc.Paragraph("You are the current coding agent session preparing this feature for handoff.")
		doc.Paragraph("Before transfer, make sure the feature documentation reflects the realities of the implementation and current task state.")
		doc.Heading(2, "Feature State")
		doc.Table([]string{"Field", "Value"}, [][]string{
			{"Feature", feat.Slug},
			{"Phase", string(feat.Phase)},
			{"Paused", fmt.Sprintf("%t", feat.Paused)},
			{"Directory", feat.DirName},
			{"Full Path", feat.Path},
		})
		doc.Heading(2, "Documentation Inventory")
		doc.Table([]string{"File", "Full Path", "How To Use"}, handoffDocumentRows(docs))
		doc.Heading(2, "Work Instructions")
		doc.OrderedList(1, workSteps...)
		doc.Heading(2, "Final Response Contract")
		doc.Paragraph("After the documentation is reconciled, reply in stdout/chat with exactly these sections:")
		doc.OrderedList(1,
			"`Documentation Sync`\n- one concise paragraph confirming all relevant documentation files and reference inventories have been updated and are up to date\n- if you updated docs or reference metadata, name the files you changed in that paragraph",
			"`Documentation Files`\n- a markdown table with columns `File`, `Full Path`, and `How To Use`\n- include every reconciled feature document and relevant project-level doc",
			"`Recent Context`\n- flat bullets for decisions made, blockers, validation results, open questions, and next steps\n- keep this concise and factual",
		)
	})

	return output, feat.Path, nil
}

func genericHandoffInstructions() string {
	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Heading(1, "Handoff Preparation")
		doc.Paragraph("You are the current coding agent session preparing this project for handoff.")
		doc.Heading(2, "Work Instructions")
		doc.OrderedList(1,
			"Review the most recent conversation context already available in this session",
			"Summarize that recent context into high-signal facts covering decisions made, blockers, validation results, open questions, and next steps",
			"Identify the authoritative project documents and make sure they reflect current implementation reality before handoff",
			"If this is a Kit project, use `kit handoff` from the project root to generate a feature-aware documentation inventory",
			"If the relevant docs include canonical front matter references, make sure they reflect current `active`, `optional`, and `stale` references with exact targets, stable ids when needed, selector types, stable selectors, relations, and read policies",
			"Prefer repository files and current code over memory when they disagree",
		)
		doc.Heading(2, "Final Response Contract")
		doc.Paragraph("After the docs are reconciled, reply in stdout/chat with:")
		doc.OrderedList(1,
			"`Documentation Sync`\n- one concise paragraph confirming all relevant documentation files and reference inventories have been updated and are up to date",
			"`Documentation Files`\n- a markdown table with columns `File`, `Full Path`, and `How To Use`",
			"`Recent Context`\n- flat bullets for decisions made, blockers, validation results, open questions, and next steps",
		)
	})
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

	docs = appendHandoffRepoContextDocuments(docs, existingRepoInstructionDocs(projectRoot, cfg))
	docs = appendHandoffRepoContextDocuments(docs, existingRepoKnowledgeDocs(projectRoot, cfg))

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

	docs = appendHandoffRepoContextDocuments(docs, existingRepoInstructionDocs(projectRoot, cfg))
	docs = appendHandoffRepoContextDocuments(docs, existingRepoKnowledgeDocs(projectRoot, cfg))

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

func appendHandoffRepoContextDocuments(dst []handoffDocument, docs []repoContextDoc) []handoffDocument {
	for _, doc := range docs {
		dst = append(dst, handoffDocument{
			File:     doc.Label,
			FullPath: doc.Path,
			HowToUse: doc.Use,
		})
	}

	return dst
}

func featureScopedDocuments(feat *feature.Feature) []handoffDocument {
	entries := []struct {
		name string
		use  string
	}{
		{"SPEC.md", fmt.Sprintf("V2 durable workflow artifact for %s; keep phase, requirements, plan, task checklist, validation, reflection, delivery, evidence, and references aligned with reality", feat.Slug)},
		{"BRAINSTORM.md", fmt.Sprintf("Optional legacy research for %s; historical context unless a legacy staged command is in use", feat.Slug)},
		{"PLAN.md", fmt.Sprintf("Optional legacy implementation approach for %s; historical context unless a legacy staged command is in use", feat.Slug)},
		{"TASKS.md", fmt.Sprintf("Optional legacy execution state for %s; historical context unless a legacy staged command is in use", feat.Slug)},
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

func handoffDocumentRows(docs []handoffDocument) [][]string {
	rows := make([][]string, 0, len(docs))
	for _, doc := range docs {
		rows = append(rows, []string{doc.File, doc.FullPath, doc.HowToUse})
	}
	return rows
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

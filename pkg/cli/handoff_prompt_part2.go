package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

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

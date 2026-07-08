package cli

import (
	"fmt"
	"os"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/templates"
)

func ensureSpecV2Adoption(specPath, projectRoot, featureDirName string, goalPercentage int) (bool, error) {
	before, err := os.ReadFile(specPath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", specPath, err)
	}

	if err := document.MergeDocument(specPath, templates.Spec, document.TypeSpec); err != nil {
		return false, fmt.Errorf("failed to add v2 SPEC.md sections to %s: %w", specPath, err)
	}

	afterMerge, err := os.ReadFile(specPath)
	if err != nil {
		return false, fmt.Errorf("failed to read %s after v2 section adoption: %w", specPath, err)
	}

	_, notesRelPath, err := ensureFeatureNotesDir(projectRoot, featureDirName)
	if err != nil {
		return false, err
	}

	phase := "clarify"
	doc := document.Parse(string(afterMerge), specPath, document.TypeSpec)
	if doc.Metadata != nil && doc.Metadata.WorkflowVersion == 2 && doc.Metadata.Phase != "" {
		phase = doc.Metadata.Phase
	}
	clarification := clarificationMetadataForAdoption(doc, phase, goalPercentage)

	updated, _, err := document.UpsertMetadata(string(afterMerge), document.TypeSpec, document.MetadataUpsert{
		Feature:         document.FeatureMetadataFromDir(featureDirName),
		WorkflowVersion: 2,
		Phase:           phase,
		Clarification:   clarification,
		References:      referencesForMetadataUpsert(string(afterMerge), document.TypeSpec, []document.MetadataReference{featureNotesReference(notesRelPath)}),
	})
	if err != nil {
		return false, fmt.Errorf("failed to update v2 SPEC.md metadata in %s: %w", specPath, err)
	}
	if updated != string(afterMerge) {
		if err := document.Write(specPath, updated); err != nil {
			return false, fmt.Errorf("failed to write v2 SPEC.md metadata in %s: %w", specPath, err)
		}
	}

	return string(before) != updated, nil
}

func clarificationMetadataForAdoption(doc *document.Document, phase string, goalPercentage int) *document.MetadataClarification {
	status := document.ClarificationStatusOpen
	confidence := 0
	unresolvedQuestions := 1
	switch feature.Phase(phase) {
	case feature.PhaseReady, feature.PhaseImplement, feature.PhaseValidate, feature.PhaseReflect, feature.PhaseDeliver, feature.PhaseComplete:
		status = document.ClarificationStatusReady
		confidence = clampPercentage(goalPercentage)
		if confidence == 0 {
			confidence = 95
		}
		unresolvedQuestions = 0
	case feature.PhaseBlocked:
		status = document.ClarificationStatusBlocked
	}

	if doc != nil {
		if existing, ok := doc.ClarificationState(); ok {
			if existing.Status != "" {
				status = existing.Status
			}
			if value, ok := existing.ConfidenceValue(); ok {
				confidence = value
			}
			if value, ok := existing.UnresolvedQuestionsValue(); ok {
				unresolvedQuestions = value
			}
		}
	}
	clarification := document.NewMetadataClarification(status, confidence, unresolvedQuestions)
	return &clarification
}

func readSpecFeatureRef() (string, error) {
	rl, err := newMultilineReadline()
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer closeMultilineReadline(rl)

	style := styleForStdout()
	printSectionBanner("🏷️", "Feature Name")
	fmt.Println(style.muted("No eligible v2 feature candidates were found."))
	fmt.Println(style.muted("Enter a short feature or project name; Kit will normalize it to lowercase kebab-case."))
	fmt.Println(style.muted("Keep it 5 words or fewer."))
	featureRef := readLineRL(rl)
	if featureRef == "" {
		return "", fmt.Errorf("feature name cannot be empty")
	}

	normalized := feature.NormalizeSlug(featureRef)
	if err := feature.ValidateSlug(normalized); err != nil {
		return "", err
	}

	if normalized != featureRef {
		fmt.Printf(dim+"Using normalized feature slug: %s"+reset+"\n\n", normalized)
	}
	return normalized, nil
}

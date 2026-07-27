package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

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

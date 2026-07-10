package cli

import (
	"strings"

	"github.com/jamesonstone/kit/internal/document"
)

type specAnswers struct {
	Problem        string
	Goals          string
	NonGoals       string
	Users          string
	Requirements   string
	Acceptance     string
	EdgeCases      string
	DeliveryIntent string
}

const (
	specDeliveryIntentIdeaOnly           = "idea_only"
	specDeliveryIntentIssueBranchPRLater = "issue_branch_pr_later"
	specDeliveryIntentContinueCurrent    = "continue_current"
)

func clarificationState(status string, confidence int, unresolvedQuestions int) *document.MetadataClarification {
	clarification := document.NewMetadataClarification(status, confidence, unresolvedQuestions)
	return &clarification
}

func normalizeSpecAnswer(raw string) string {
	return strings.TrimSpace(raw)
}

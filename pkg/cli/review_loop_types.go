package cli

import (
	"time"
)

const (
	reviewLoopInitialWait = 90 * time.Second
	reviewLoopPollEvery   = 15 * time.Second
	reviewLoopTimeout     = 15 * time.Minute
	reviewLoopQuietWindow = 60 * time.Second
)

type reviewLoopOptions struct {
	PRRef          string
	CodeRabbitOnly bool
	Watch          bool
	Copy           bool
	OutputOnly     bool
	UseVim         bool
	Editor         string
	MaxSubagents   int
	InputConfig    freeTextInputConfig
}

type reviewLoopPRContext struct {
	Target       dispatchPRTarget
	URL          string
	Title        string
	Body         string
	HeadRefOID   string
	IssueHints   []string
	RepoFullName string
	LocalRoot    string
}

type reviewLoopFinding struct {
	Task dispatchReviewTask
}

type reviewLoopClassification string

const (
	reviewLoopFix             reviewLoopClassification = "FIX"
	reviewLoopValidOutOfScope reviewLoopClassification = "VALID_OUT_OF_SCOPE"
	reviewLoopFalsePositive   reviewLoopClassification = "FALSE_POSITIVE"
	reviewLoopStale           reviewLoopClassification = "STALE"
	reviewLoopNeedsHuman      reviewLoopClassification = "NEEDS_HUMAN"
)

type reviewLoopClassifiedFinding struct {
	Finding reviewLoopFinding
	Kind    reviewLoopClassification
	Reason  string
}

type reviewLoopCheck struct {
	Name        string `json:"name"`
	Workflow    string `json:"workflow"`
	State       string `json:"state"`
	Bucket      string `json:"bucket"`
	CompletedAt string `json:"completedAt"`
	Link        string `json:"link"`
	Description string `json:"description"`
}

// Package ledger describes optional work-ledger adapters. Implementations
// perform no I/O and store no credentials. Native agents execute the plan
// through their own connectors.
package ledger

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
)

// Phase is a ticket lifecycle moment that may update the work ledger.
type Phase string

const (
	PhaseStart  Phase = "start"
	PhaseReview Phase = "review"
	PhaseDone   Phase = "done"

	// OpNoop means the ledger is off. The other operations are connector steps.
	OpNoop         = "noop"
	OpFind         = "find"
	OpFindOrCreate = "find_or_create"
	OpLink         = "link"
	OpSetState     = "set_state"
	OpLabel        = "label"

	LinkIssue       = "issue"
	LinkPullRequest = "pull_request"
)

// Item is the ticket the ledger item mirrors. The adapter does not fetch it.
type Item struct {
	IssueID    string
	IssueURL   string
	Title      string
	PullURL    string
	OwnerLabel string
}

// Action is one connector step. Operation "noop" means the ledger is off.
type Action struct {
	Operation string `json:"operation"`
	Provider  string `json:"provider,omitempty"`
	Workspace string `json:"workspace,omitempty"`
	Project   string `json:"project,omitempty"`
	State     string `json:"state,omitempty"`
	LinkKind  string `json:"link_kind,omitempty"`
	LinkURL   string `json:"link_url,omitempty"`
	Label     string `json:"label,omitempty"`
	Summary   string `json:"summary"`
}

// Adapter describes ledger work for a native agent.
type Adapter interface {
	Provider() string
	Enabled() bool
	Description() string
	Plan(phase Phase, item Item) ([]Action, error)
}

// New selects the adapter for cfg. An omitted provider returns the no-op.
// linear, jira, and notion are reserved and rejected until an adapter exists.
func New(cfg config.WorkLedgerConfig) (Adapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if !cfg.Enabled() {
		return noop{}, nil
	}
	switch strings.TrimSpace(cfg.Provider) {
	case "plane":
		return plane{cfg: cfg.Plane, ownerLabel: strings.TrimSpace(cfg.OwnerLabel)}, nil
	default:
		return nil, fmt.Errorf("work_ledger.provider %q has no adapter", cfg.Provider)
	}
}

type noop struct{}

func (noop) Provider() string { return "" }
func (noop) Enabled() bool    { return false }
func (noop) Description() string {
	return "The work ledger is off. Do nothing."
}

func (noop) Plan(phase Phase, _ Item) ([]Action, error) {
	if !knownPhase(phase) {
		return nil, fmt.Errorf("unknown work ledger phase %q", phase)
	}
	return []Action{{Operation: OpNoop, Summary: "work ledger is not configured"}}, nil
}

func knownPhase(phase Phase) bool {
	switch phase {
	case PhaseStart, PhaseReview, PhaseDone:
		return true
	default:
		return false
	}
}

var suppliedLabelPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,63}$`)

func validateOwnerLabel(label string) error {
	if label == "" || suppliedLabelPattern.MatchString(label) {
		return nil
	}
	return fmt.Errorf("owner label must be a short label with no URL or spaces")
}

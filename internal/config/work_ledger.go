package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkLedgerConfig selects an optional work-ledger adapter. Credentials never
// belong here. An omitted provider is off and selects the no-op adapter.
type WorkLedgerConfig struct {
	Provider   string            `yaml:"provider,omitempty" json:"provider,omitempty"`
	OwnerLabel string            `yaml:"owner_label,omitempty" json:"owner_label,omitempty"`
	Plane      PlaneLedgerConfig `yaml:"plane,omitempty" json:"plane,omitempty"`
}

// PlaneLedgerConfig is the non-secret Plane target. Workspace is a URL slug.
// Project is the project identifier. Name is the optional title. Kit does not
// supply either value.
type PlaneLedgerConfig struct {
	Workspace string            `yaml:"workspace,omitempty" json:"workspace,omitempty"`
	Project   string            `yaml:"project,omitempty" json:"project,omitempty"`
	Name      string            `yaml:"name,omitempty" json:"name,omitempty"`
	States    PlaneLedgerStates `yaml:"states,omitempty" json:"states,omitempty"`
}

// PlaneLedgerStates names the lifecycle states. Empty fields resolve to
// Active, In review, and Done.
type PlaneLedgerStates struct {
	Active   string `yaml:"active,omitempty" json:"active,omitempty"`
	InReview string `yaml:"in_review,omitempty" json:"in_review,omitempty"`
	Done     string `yaml:"done,omitempty" json:"done,omitempty"`
}

// Enabled reports a configured adapter. An omitted provider and "none" are off.
func (c WorkLedgerConfig) Enabled() bool {
	switch strings.TrimSpace(c.Provider) {
	case "", "none":
		return false
	default:
		return true
	}
}

// Resolved fills omitted Plane state names with the lifecycle defaults.
func (s PlaneLedgerStates) Resolved() PlaneLedgerStates {
	if strings.TrimSpace(s.Active) == "" {
		s.Active = "Active"
	}
	if strings.TrimSpace(s.InReview) == "" {
		s.InReview = "In review"
	}
	if strings.TrimSpace(s.Done) == "" {
		s.Done = "Done"
	}
	return s
}

var (
	ledgerSlugPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)
	ledgerStatePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 .'_-]{0,63}$`)
	ownerLabelPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,63}$`)
)

// Validate checks local configuration only. Agents verify the target with
// their own connector before any ledger write. Kit does not call the API.
func (c WorkLedgerConfig) Validate() error {
	switch strings.TrimSpace(c.Provider) {
	case "", "none":
		return nil
	case "plane":
		if err := c.Plane.validate(); err != nil {
			return err
		}
		return validateOwnerLabel(c.OwnerLabel)
	case "linear", "jira", "notion":
		return fmt.Errorf("work_ledger.provider %q is reserved and not implemented", c.Provider)
	default:
		return fmt.Errorf("work_ledger.provider must be plane or omitted; got %q", c.Provider)
	}
}

func appendLegacyWorkLedgerFinding(raw map[string]yaml.Node, findings []Finding) []Finding {
	for _, key := range []string{"tool", "workspace", "project", "label", "rules"} {
		if _, ok := raw[key]; ok {
			return append(findings, Finding{
				Field:    "work_ledger." + key,
				Severity: FindingError,
				Message:  "obsolete flat ledger keys: use work_ledger.provider, owner_label, and plane.workspace/plane.project in both config files; keep rules links in project guidance",
			})
		}
	}
	return findings
}

func (i Inspection) workLedgerFormatError() error {
	for _, finding := range i.Findings {
		if strings.HasPrefix(finding.Field, "work_ledger.") && finding.Severity == FindingError {
			return errors.New(finding.Message)
		}
	}
	return nil
}

func appendIssueAndLedgerFindings(cfg *Config, findings []Finding) []Finding {
	if err := cfg.Issues.Validate(); err != nil {
		findings = append(findings, Finding{Field: "issues", Severity: FindingError, Message: err.Error()})
	}
	if err := cfg.WorkLedger.Validate(); err != nil {
		findings = append(findings, Finding{Field: "work_ledger", Severity: FindingError, Message: err.Error()})
	}
	return findings
}

func validateOwnerLabel(label string) error {
	if label == "" || ownerLabelPattern.MatchString(label) {
		return nil
	}
	return fmt.Errorf("work_ledger.owner_label must be a short label with no URL or spaces")
}

func (c PlaneLedgerConfig) validate() error {
	if !ledgerSlugPattern.MatchString(c.Workspace) {
		return fmt.Errorf("work_ledger.plane.workspace must be a workspace URL slug")
	}
	if !ledgerSlugPattern.MatchString(c.Project) {
		return fmt.Errorf("work_ledger.plane.project must be a project identifier")
	}
	if strings.TrimSpace(c.Name) != "" && !ledgerStatePattern.MatchString(strings.TrimSpace(c.Name)) {
		return fmt.Errorf("work_ledger.plane.name must be a short project title")
	}
	resolved := c.States.Resolved()
	for _, state := range []string{resolved.Active, resolved.InReview, resolved.Done} {
		if !ledgerStatePattern.MatchString(state) {
			return fmt.Errorf("work_ledger.plane.states must be short state names")
		}
	}
	return nil
}

package config

import (
	"fmt"
	"regexp"
	"strings"
)

// IssueConfig selects the repository's development tracker. Credentials never
// belong here. An omitted provider means GitHub Issues for existing projects.
type IssueConfig struct {
	Provider      string       `yaml:"provider,omitempty" json:"provider"`
	IssueTemplate *bool        `yaml:"issue_template,omitempty" json:"issue_template,omitempty"`
	Linear        LinearConfig `yaml:"linear,omitempty" json:"linear,omitempty"`
}

// IssueTemplateEnabled reports an explicit opt-in. An omitted value stays off
// so init does not install a GitHub issue template over an existing setup.
func (c IssueConfig) IssueTemplateEnabled() bool {
	return c.IssueTemplate != nil && *c.IssueTemplate
}

// LinearConfig stores verified, non-secret target identities. IDs avoid
// ambiguous team/project display names; an empty project means team-only work.
type LinearConfig struct {
	Workspace string `yaml:"workspace,omitempty" json:"workspace,omitempty"`
	TeamID    string `yaml:"team_id,omitempty" json:"team_id,omitempty"`
	TeamKey   string `yaml:"team_key,omitempty" json:"team_key,omitempty"`
	ProjectID string `yaml:"project_id,omitempty" json:"project_id,omitempty"`
}

// SameTeam reports whether both targets identify the same workspace and team.
func (c LinearConfig) SameTeam(other LinearConfig) bool {
	return c.Workspace == other.Workspace && strings.EqualFold(c.TeamID, other.TeamID)
}

func (c IssueConfig) EffectiveProvider() string {
	if c.Provider == "" {
		return "github"
	}
	return c.Provider
}

// BranchPrefix reports the configured prefix. An empty value means the agent
// must resolve the Linear team key. Linear's returned issue ID names the branch.
func (c IssueConfig) BranchPrefix() string {
	switch c.EffectiveProvider() {
	case "github":
		return "GH-"
	case "linear":
		if c.Linear.TeamKey != "" {
			return c.Linear.TeamKey + "-"
		}
	}
	return ""
}

var linearWorkspacePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
var linearIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var linearTeamKeyPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// Validate checks local configuration only. Agents must verify these identities
// and the active human's access with Linear before any ticket mutation.
func (c IssueConfig) Validate() error {
	switch c.EffectiveProvider() {
	case "github":
		return nil
	case "linear":
		if !linearWorkspacePattern.MatchString(c.Linear.Workspace) {
			return fmt.Errorf("issues.linear.workspace must be a Linear workspace URL slug")
		}
		if !linearIDPattern.MatchString(c.Linear.TeamID) {
			return fmt.Errorf("issues.linear.team_id must be a verified Linear team UUID")
		}
		if c.Linear.TeamKey != "" && !linearTeamKeyPattern.MatchString(c.Linear.TeamKey) {
			return fmt.Errorf("issues.linear.team_key must be the Linear team identifier without a trailing hyphen")
		}
		if c.Linear.ProjectID != "" && !linearIDPattern.MatchString(c.Linear.ProjectID) {
			return fmt.Errorf("issues.linear.project_id must be a verified Linear project UUID or omitted")
		}
		return nil
	default:
		return fmt.Errorf("issues.provider must be github or linear; got %q", c.Provider)
	}
}

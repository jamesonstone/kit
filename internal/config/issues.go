package config

import (
	"fmt"
	"regexp"
)

// IssueConfig selects the repository's development tracker. Credentials never
// belong here. An omitted provider means GitHub Issues for existing projects.
type IssueConfig struct {
	Provider string       `yaml:"provider,omitempty"`
	Linear   LinearConfig `yaml:"linear,omitempty"`
}

// LinearConfig stores verified, non-secret target identities. IDs avoid
// ambiguous team/project display names; an empty project means team-only work.
type LinearConfig struct {
	Workspace string `yaml:"workspace,omitempty"`
	TeamID    string `yaml:"team_id,omitempty"`
	ProjectID string `yaml:"project_id,omitempty"`
}

func (c IssueConfig) EffectiveProvider() string {
	if c.Provider == "" {
		return "github"
	}
	return c.Provider
}

var linearWorkspacePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
var linearIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

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
		if c.Linear.ProjectID != "" && !linearIDPattern.MatchString(c.Linear.ProjectID) {
			return fmt.Errorf("issues.linear.project_id must be a verified Linear project UUID or omitted")
		}
		return nil
	default:
		return fmt.Errorf("issues.provider must be github or linear; got %q", c.Provider)
	}
}

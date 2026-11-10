package guidance

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

type Metadata struct {
	RegistryScope string   `yaml:"registry_scope" json:"registry_scope,omitempty"`
	Kind          string   `yaml:"kind" json:"kind"`
	Slug          string   `yaml:"slug" json:"slug"`
	Description   string   `yaml:"description" json:"description"`
	Status        string   `yaml:"status" json:"status"`
	Owner         string   `yaml:"owner" json:"owner,omitempty"`
	AppliesTo     []string `yaml:"applies_to" json:"applies_to,omitempty"`
	ReadPolicy    string   `yaml:"read_policy_default" json:"read_policy_default,omitempty"`
}

func parseMetadata(data []byte, kind string) (*Metadata, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	if strings.TrimSpace(lines[0]) != "---" {
		return nil, fmt.Errorf("missing YAML frontmatter")
	}
	end := 0
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == 0 {
		return nil, fmt.Errorf("unterminated YAML frontmatter")
	}
	var meta Metadata
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:end], "\n")), &meta); err != nil {
		return nil, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	if meta.Kind != kind || strings.TrimSpace(meta.Slug) == "" {
		return &meta, fmt.Errorf("metadata requires kind %q and slug", kind)
	}
	if !slugPattern.MatchString(meta.Slug) {
		return &meta, fmt.Errorf("invalid slug %q", meta.Slug)
	}
	validStatus := meta.Status == "active"
	if kind == "ruleset" {
		validStatus = validStatus || meta.Status == "optional" || meta.Status == "stale"
	}
	if kind == "sop" {
		validStatus = validStatus || meta.Status == "draft" || meta.Status == "retired"
	}
	if !validStatus {
		return &meta, fmt.Errorf("unknown or missing status %q", meta.Status)
	}
	if kind == "ruleset" && (len(meta.AppliesTo) == 0 || meta.ReadPolicy == "") {
		return &meta, fmt.Errorf("ruleset metadata requires applies_to and read_policy_default")
	}
	if kind == "ruleset" {
		scope := strings.TrimSpace(meta.RegistryScope)
		if scope != "" && scope != "downstream" && scope != "optional" {
			return &meta, fmt.Errorf("unknown registry_scope %q", meta.RegistryScope)
		}
		for _, applies := range meta.AppliesTo {
			if !slugPattern.MatchString(applies) {
				return &meta, fmt.Errorf("invalid applies_to entry %q", applies)
			}
		}
	}
	if kind == "sop" && (strings.TrimSpace(meta.Owner) == "" || strings.TrimSpace(meta.Description) == "") {
		return &meta, fmt.Errorf("SOP metadata requires owner and description")
	}
	if meta.ReadPolicy != "" {
		switch meta.ReadPolicy {
		case "must", "conditional", "evidence", "skip":
		default:
			return &meta, fmt.Errorf("unknown read_policy_default %q", meta.ReadPolicy)
		}
	}
	return &meta, nil
}

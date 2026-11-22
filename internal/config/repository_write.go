package config

import (
	"errors"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// MarshalRepositoryConfig updates managed settings without replacing absent or
// explicit zero-valued project overrides with struct defaults. An explicit
// force reset uses the reset config and still preserves the project ledger.
// A nil before means a new file. An empty existing file is an owned input.
func MarshalRepositoryConfig(cfg *Config, before []byte, reset bool) ([]byte, error) {
	if before == nil {
		return yaml.Marshal(cfg)
	}
	var owned map[string]any
	if err := yaml.Unmarshal(before, &owned); err != nil {
		return nil, err
	}
	if owned == nil {
		owned = make(map[string]any)
	}
	doc, err := configDocument(cfg)
	if err != nil {
		return nil, err
	}
	root := documentRoot(doc)
	shape := reflect.TypeFor[Config]()
	for index := 0; index < shape.NumField(); index++ {
		key, _, _ := strings.Cut(shape.Field(index).Tag.Get("yaml"), ",")
		if reset && key != "work_ledger" {
			continue
		}
		switch key {
		case "schema_version", "instruction_scaffold_version", "registry":
			continue // Reconcile writes these repository-managed fields.
		case "issues":
			if cfg.Issues.IssueTemplate != nil {
				issues, _ := owned[key].(map[string]any)
				if issues == nil {
					issues = make(map[string]any)
				}
				issues["issue_template"] = *cfg.Issues.IssueTemplate
				owned[key] = issues
			}
		}
		value, present := owned[key]
		if !present {
			removeKey(root, key)
			continue
		}
		node := findOrCreateMapping(root, key)
		if err := node.Encode(value); err != nil {
			return nil, err
		}
	}
	return yaml.Marshal(doc)
}

// ParseForRepositoryWrite starts from the repository file only. A project can
// override just part of a global setting, so merged-config validation belongs
// to the effective read. Reject formats a rewrite would lose.
func ParseForRepositoryWrite(data []byte) (*Config, Inspection, error) {
	cfg, inspection, err := ParseWithInspection(data)
	if err != nil {
		return nil, inspection, err
	}
	if inspection.SchemaState == SchemaStateNewer {
		return nil, inspection, errors.New(inspection.Findings[0].Message)
	}
	return cfg, inspection, inspection.workLedgerFormatError()
}

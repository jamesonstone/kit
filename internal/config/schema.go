package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type SchemaState string

const (
	SchemaStateMissing SchemaState = "missing"
	SchemaStateOlder   SchemaState = "older"
	SchemaStateCurrent SchemaState = "current"
	SchemaStateNewer   SchemaState = "newer"
)

type FindingSeverity string

const (
	FindingWarning FindingSeverity = "warning"
	FindingError   FindingSeverity = "error"
)

type Finding struct {
	Field      string          `json:"field"`
	Severity   FindingSeverity `json:"severity"`
	Message    string          `json:"message"`
	Repairable bool            `json:"repairable"`
}

type Inspection struct {
	SchemaVersion        int         `json:"schema_version"`
	CurrentSchemaVersion int         `json:"current_schema_version"`
	SchemaState          SchemaState `json:"schema_state"`
	Findings             []Finding   `json:"findings"`
}

type rawConfigInspection struct {
	SchemaVersion *int `yaml:"schema_version"`
	AWS           *struct {
		AccountID yaml.Node `yaml:"account_id"`
	} `yaml:"aws"`
}

func (i Inspection) HasErrors() bool {
	for _, finding := range i.Findings {
		if finding.Severity == FindingError {
			return true
		}
	}
	return false
}

func (i Inspection) NeedsSchemaMigration() bool {
	return i.SchemaState == SchemaStateMissing || i.SchemaState == SchemaStateOlder
}

func LoadWithInspection(projectRoot string) (*Config, Inspection, error) {
	configPath := filepath.Join(projectRoot, ConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, Inspection{}, fmt.Errorf("failed to read %s: %w", ConfigFileName, err)
	}

	raw, inspection, err := inspectRawData(data)
	if err != nil {
		return nil, Inspection{}, fmt.Errorf("failed to inspect %s: %w", ConfigFileName, err)
	}
	if inspection.SchemaState == SchemaStateNewer {
		return Default(), inspection, nil
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, Inspection{}, fmt.Errorf("failed to parse %s: %w", ConfigFileName, err)
	}
	inspection.Findings = append(inspection.Findings, semanticFindings(cfg)...)
	if cfg.AWS != nil && cfg.AWS.IsEnabled() && awsAccountIDPattern.MatchString(strings.TrimSpace(cfg.AWS.AccountID)) {
		if raw.AWS != nil && !isQuotedYAMLString(&raw.AWS.AccountID) {
			inspection.Findings = append(inspection.Findings, Finding{
				Field:      "aws.account_id",
				Severity:   FindingError,
				Message:    "aws.account_id must be a quoted 12-digit string",
				Repairable: true,
			})
		}
	}
	return cfg, inspection, nil
}

func inspectRawData(data []byte) (rawConfigInspection, Inspection, error) {
	var raw rawConfigInspection
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return rawConfigInspection{}, Inspection{}, err
	}

	inspection := Inspection{CurrentSchemaVersion: CurrentSchemaVersion}
	switch {
	case raw.SchemaVersion == nil:
		inspection.SchemaState = SchemaStateMissing
		inspection.Findings = append(inspection.Findings, Finding{
			Field:      "schema_version",
			Severity:   FindingWarning,
			Message:    fmt.Sprintf("schema_version is missing; current version is %d", CurrentSchemaVersion),
			Repairable: true,
		})
	case *raw.SchemaVersion < CurrentSchemaVersion:
		inspection.SchemaVersion = *raw.SchemaVersion
		inspection.SchemaState = SchemaStateOlder
		inspection.Findings = append(inspection.Findings, Finding{
			Field:      "schema_version",
			Severity:   FindingWarning,
			Message:    fmt.Sprintf("schema_version %d is older than current version %d", *raw.SchemaVersion, CurrentSchemaVersion),
			Repairable: true,
		})
	case *raw.SchemaVersion > CurrentSchemaVersion:
		inspection.SchemaVersion = *raw.SchemaVersion
		inspection.SchemaState = SchemaStateNewer
		inspection.Findings = append(inspection.Findings, Finding{
			Field:      "schema_version",
			Severity:   FindingError,
			Message:    fmt.Sprintf("schema_version %d is newer than this Kit version supports (%d); upgrade Kit", *raw.SchemaVersion, CurrentSchemaVersion),
			Repairable: false,
		})
	default:
		inspection.SchemaVersion = *raw.SchemaVersion
		inspection.SchemaState = SchemaStateCurrent
	}

	return raw, inspection, nil
}

func isQuotedYAMLString(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return false
	}
	return node.Style == yaml.SingleQuotedStyle || node.Style == yaml.DoubleQuotedStyle
}

var awsAccountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)

func semanticFindings(cfg *Config) []Finding {
	var findings []Finding
	if cfg.GoalPercentage < 1 || cfg.GoalPercentage > 100 {
		findings = append(findings, Finding{Field: "goal_percentage", Severity: FindingError, Message: "goal_percentage must be between 1 and 100"})
	}
	for _, item := range []struct {
		field string
		value string
	}{
		{field: "specs_dir", value: cfg.SpecsDir},
		{field: "skills_dir", value: cfg.SkillsDir},
		{field: "constitution_path", value: cfg.ConstitutionPath},
	} {
		if strings.TrimSpace(item.value) == "" {
			findings = append(findings, Finding{Field: item.field, Severity: FindingError, Message: item.field + " must not be empty"})
		}
	}
	if cfg.InstructionScaffoldVersion != 0 && !IsInstructionScaffoldVersionSupported(cfg.InstructionScaffoldVersion) {
		findings = append(findings, Finding{Field: "instruction_scaffold_version", Severity: FindingError, Message: "instruction_scaffold_version is unsupported"})
	}
	if cfg.FeatureNaming.NumericWidth <= 0 {
		findings = append(findings, Finding{Field: "feature_naming.numeric_width", Severity: FindingError, Message: "feature_naming.numeric_width must be greater than zero"})
	}
	if strings.TrimSpace(cfg.FeatureNaming.Separator) == "" {
		findings = append(findings, Finding{Field: "feature_naming.separator", Severity: FindingError, Message: "feature_naming.separator must not be empty"})
	}

	if cfg.AWS == nil || !cfg.AWS.IsEnabled() {
		return findings
	}
	if strings.TrimSpace(cfg.AWS.Profile) == "" {
		findings = append(findings, Finding{Field: "aws.profile", Severity: FindingError, Message: "aws.profile is required when AWS context is enabled", Repairable: true})
	}
	if !awsAccountIDPattern.MatchString(strings.TrimSpace(cfg.AWS.AccountID)) {
		findings = append(findings, Finding{Field: "aws.account_id", Severity: FindingError, Message: "aws.account_id must be a quoted 12-digit string", Repairable: true})
	}
	return findings
}

// UpdateProjectSchemaAndAWS updates only Kit's schema and AWS keys while
// preserving unrelated and unknown project configuration fields.
func UpdateProjectSchemaAndAWS(projectRoot string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	configPath := filepath.Join(projectRoot, ConfigFileName)
	doc, err := readYAMLDocument(configPath, false)
	if err != nil {
		return err
	}
	root, err := documentMapping(doc, configPath)
	if err != nil {
		return err
	}

	setTypedScalar(root, "schema_version", fmt.Sprintf("%d", cfg.SchemaVersion), "!!int", 0)
	if cfg.AWS != nil {
		aws := findOrCreateMapping(root, "aws")
		if !cfg.AWS.IsEnabled() {
			setTypedScalar(aws, "enabled", "false", "!!bool", 0)
			removeKey(aws, "profile")
			removeKey(aws, "account_id")
		} else {
			removeKey(aws, "enabled")
			setTypedScalar(aws, "profile", strings.TrimSpace(cfg.AWS.Profile), "!!str", 0)
			setTypedScalar(aws, "account_id", strings.TrimSpace(cfg.AWS.AccountID), "!!str", yaml.DoubleQuotedStyle)
		}
	}

	return writeProjectYAMLDocument(configPath, doc)
}

func setTypedScalar(parent *yaml.Node, key, value, tag string, style yaml.Style) {
	node := &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value, Style: style}
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			parent.Content[i+1] = node
			return
		}
	}
	parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, node)
}

func writeProjectYAMLDocument(configPath string, doc *yaml.Node) (resultErr error) {
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(4)
	if err := encoder.Encode(doc); err != nil {
		encodeErr := fmt.Errorf("failed to encode %s: %w", configPath, err)
		if closeErr := encoder.Close(); closeErr != nil {
			return errors.Join(encodeErr, fmt.Errorf("failed to close YAML encoder for %s: %w", configPath, closeErr))
		}
		return encodeErr
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close YAML encoder for %s: %w", configPath, err)
	}

	temp, err := os.CreateTemp(filepath.Dir(configPath), ".kit-config-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary config file: %w", err)
	}
	tempPath := temp.Name()
	tempClosed := false
	defer func() {
		if !tempClosed {
			if err := temp.Close(); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("failed to close temporary config: %w", err))
			}
		}
		if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			resultErr = errors.Join(resultErr, fmt.Errorf("failed to remove temporary config %s: %w", tempPath, err))
		}
	}()
	if err := temp.Chmod(0644); err != nil {
		return fmt.Errorf("failed to set temporary config permissions: %w", err)
	}
	if _, err := temp.Write(output.Bytes()); err != nil {
		return fmt.Errorf("failed to write temporary config: %w", err)
	}
	if err := temp.Close(); err != nil {
		tempClosed = true
		return fmt.Errorf("failed to close temporary config: %w", err)
	}
	tempClosed = true
	if err := os.Rename(tempPath, configPath); err != nil {
		return fmt.Errorf("failed to replace %s: %w", configPath, err)
	}
	return nil
}

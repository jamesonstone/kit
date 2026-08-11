package releaseprompt

import (
	"fmt"
	"strings"
)

var validScopes = map[string]bool{"strict": true, "related": true, "organization": true}
var validInfrastructureModes = map[string]bool{
	"auto": true, "none": true, "direct": true, "iac": true, "mixed": true, "custom": true,
}

func Validate(config Config) error {
	if strings.TrimSpace(config.Project) == "" {
		return fmt.Errorf("project is required")
	}
	if len(config.Repositories) == 0 {
		return fmt.Errorf("at least one repository is required")
	}
	if !validScopes[config.ScopeExpansion] {
		return fmt.Errorf("scope must be strict, related, or organization")
	}
	if !validInfrastructureModes[config.Infrastructure.Mode] {
		return fmt.Errorf("infrastructure mode must be auto, none, direct, iac, mixed, or custom")
	}
	if config.Infrastructure.Mode == "none" {
		if config.Infrastructure.Provider != "not applicable" || config.Infrastructure.CLI != "not applicable" {
			return fmt.Errorf("infrastructure mode none cannot configure a provider or CLI")
		}
	} else if strings.TrimSpace(config.Infrastructure.Provider) == "" || strings.TrimSpace(config.Infrastructure.CLI) == "" {
		return fmt.Errorf("infrastructure provider and CLI resolution must be explicit")
	}
	if strings.TrimSpace(config.Production.Environment) == "" {
		return fmt.Errorf("target environment is required")
	}
	if err := validateTaggedValue(config.Production.Verification, false); err != nil {
		return fmt.Errorf("production verification: %w", err)
	}
	if err := validateTaggedValue(config.IntegrationSuite, true); err != nil {
		return fmt.Errorf("integration suite: %w", err)
	}
	for _, value := range []string{
		config.Project, config.Organization, config.FeatureContext,
		config.Infrastructure.Provider, config.Infrastructure.CLI,
		config.Production.Environment, config.Production.Verification, config.IntegrationSuite,
		config.AdditionalHardRules, config.FinalReportRequirements,
	} {
		if strings.ContainsRune(value, '\x00') {
			return fmt.Errorf("configuration contains a NUL byte")
		}
		if looksLikeSecret(value) {
			return fmt.Errorf("configuration appears to contain a secret; provide a secret name or lookup instruction instead of its value")
		}
	}
	return nil
}

func requiresConcreteInfrastructure(mode string) bool {
	return mode == "direct" || mode == "iac" || mode == "mixed" || mode == "custom"
}

func validateTaggedValue(value string, allowNone bool) error {
	value = strings.TrimSpace(value)
	if value == "auto" || (allowNone && value == "none") {
		return nil
	}
	for _, prefix := range []string{"command:", "script:", "endpoint:", "instruction:"} {
		if strings.HasPrefix(value, prefix) && strings.TrimSpace(strings.TrimPrefix(value, prefix)) != "" {
			return nil
		}
	}
	allowed := "auto, command:, script:, endpoint:, or instruction:"
	if allowNone {
		allowed = "auto, command:, script:, endpoint:, instruction:, or none"
	}
	return fmt.Errorf("value must be %s", allowed)
}

func looksLikeSecret(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{"-----begin private key", "ghp_", "github_pat_", "akia", "password=", "token=", "secret="} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

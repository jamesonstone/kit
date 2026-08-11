package releaseprompt

import (
	"strings"
	"testing"
)

func TestValidateScansOptionalFreeText(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "additional hard rules secret", mutate: func(config *Config) { config.AdditionalHardRules = "token=secret" }},
		{name: "additional hard rules NUL", mutate: func(config *Config) { config.AdditionalHardRules = "rule\x00value" }},
		{name: "final report secret", mutate: func(config *Config) { config.FinalReportRequirements = "password=secret" }},
		{name: "final report NUL", mutate: func(config *Config) { config.FinalReportRequirements = "report\x00value" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := goldenConfig()
			test.mutate(&config)
			if err := Validate(config); err == nil {
				t.Fatal("Validate() unexpectedly succeeded")
			}
		})
	}
}

func TestValidateTaggedValueErrorListsAllowedValues(t *testing.T) {
	withoutNone := validateTaggedValue("invalid", false)
	if got, want := withoutNone.Error(), "value must be auto, command:, script:, endpoint:, or instruction:"; got != want {
		t.Fatalf("allowNone=false error = %q, want %q", got, want)
	}
	withNone := validateTaggedValue("invalid", true)
	if got, want := withNone.Error(), "value must be auto, command:, script:, endpoint:, instruction:, or none"; got != want {
		t.Fatalf("allowNone=true error = %q, want %q", got, want)
	}
	if strings.Count(withNone.Error(), "or") != 1 {
		t.Fatalf("allowNone=true error must contain exactly one or: %q", withNone)
	}
}

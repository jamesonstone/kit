package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidAWSRegion(t *testing.T) {
	tests := map[string]bool{
		"us-east-1":     true,
		"us-gov-west-1": true,
		"cn-north-1":    true,
		" eu-west-3 ":   true,
		"":              false,
		"us-east":       false,
		"US-EAST-1":     false,
		"us_east_1":     false,
	}
	for value, want := range tests {
		if got := ValidAWSRegion(value); got != want {
			t.Errorf("ValidAWSRegion(%q) = %v, want %v", value, got, want)
		}
	}
}

func TestEnabledAWSContextRequiresRegion(t *testing.T) {
	root := t.TempDir()
	content := "schema_version: 2\naws:\n  profile: dev\n  account_id: \"012345678901\"\n"
	if err := os.WriteFile(filepath.Join(root, ConfigFileName), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, inspection, err := LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	found := false
	for _, finding := range inspection.Findings {
		if finding.Field == "aws.region" && finding.Severity == FindingError && finding.Repairable {
			found = true
		}
	}
	if !found {
		t.Fatalf("findings = %#v, want repairable aws.region error", inspection.Findings)
	}
}

func TestUpdateProjectSchemaAndAWSRemovesDisabledRegion(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ConfigFileName)
	content := "schema_version: 1\naws:\n  profile: dev\n  account_id: \"012345678901\"\n  region: us-west-1\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cfg := Default()
	cfg.AWS = DisabledAWSConfig()
	if err := UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(updated)
	for _, unwanted := range []string{"profile:", "account_id:", "region:"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("disabled AWS config retained %q:\n%s", unwanted, text)
		}
	}
	if !strings.Contains(text, "enabled: false") || !strings.Contains(text, "schema_version: 2") {
		t.Fatalf("disabled AWS config =\n%s", text)
	}
}

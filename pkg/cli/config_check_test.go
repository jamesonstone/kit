package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestAWSConfigRemediationSingleProfileDefaultsYes(t *testing.T) {
	root, cfg, inspection := setupConfigCheckProject(t)
	stubAWSContext(t, "dev\n", `{"Account":"012345678901","Arn":"arn:aws:sts::012345678901:assumed-role/Developer/test","UserId":"user"}`)
	var out bytes.Buffer

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("\n"),
		Output:      &out,
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}
	if !strings.Contains(out.String(), `Use the only AWS profile "dev" for this project? [Y/n]:`) {
		t.Fatalf("output missing default-yes prompt:\n%s", out.String())
	}

	updated, _, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if updated.AWS == nil || updated.AWS.Profile != "dev" || updated.AWS.AccountID != "012345678901" {
		t.Fatalf("AWS = %#v, want verified dev context", updated.AWS)
	}
}

func TestAWSConfigRemediationSingleProfileNoDisablesAWS(t *testing.T) {
	root, cfg, inspection := setupConfigCheckProject(t)
	stubAWSContext(t, "dev\n", "")

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("n\n"),
		Output:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}
	updated, _, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if updated.AWS == nil || updated.AWS.IsEnabled() {
		t.Fatalf("AWS = %#v, want explicit disabled config", updated.AWS)
	}
}

func TestAWSConfigRemediationMultipleProfilesRequiresSelection(t *testing.T) {
	root, cfg, inspection := setupConfigCheckProject(t)
	stubAWSContext(t, "prod\ndev\n", `{"Account":"111122223333","Arn":"arn:aws:sts::111122223333:assumed-role/Developer/test"}`)
	var out bytes.Buffer

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("2\n"),
		Output:      &out,
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}
	updated, _, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if updated.AWS == nil || updated.AWS.Profile != "prod" {
		t.Fatalf("AWS = %#v, want explicitly selected prod profile", updated.AWS)
	}
	if !strings.Contains(out.String(), "0. Do not use AWS") {
		t.Fatalf("output missing explicit disable selection:\n%s", out.String())
	}
}

func TestAWSConfigRemediationMissingCLINoOps(t *testing.T) {
	root, cfg, inspection := setupConfigCheckProject(t)
	previousLookPath := awsLookPath
	awsLookPath = func(string) (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { awsLookPath = previousLookPath })

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader(""),
		Output:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if changed {
		t.Fatal("changed = true, want clean no-op")
	}
}

func TestAWSConfigRemediationNoProfilesNoOps(t *testing.T) {
	root, cfg, inspection := setupConfigCheckProject(t)
	stubAWSContext(t, "", "")

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader(""),
		Output:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if changed {
		t.Fatal("changed = true, want clean no-op")
	}
}

func TestAWSConfigRemediationMismatchDoesNotWrite(t *testing.T) {
	root, cfg, _ := setupConfigCheckProject(t)
	cfg.AWS = &config.AWSConfig{AccountID: "999900001111"}
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}
	cfg, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	stubAWSContext(t, "dev\n", `{"Account":"012345678901","Arn":"arn:aws:sts::012345678901:assumed-role/Developer/test"}`)
	before, _ := os.ReadFile(filepath.Join(root, config.ConfigFileName))

	_, err = remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("\n"),
		Output:      &bytes.Buffer{},
	})
	if err == nil || !strings.Contains(err.Error(), "expects 999900001111") {
		t.Fatalf("error = %v, want account mismatch", err)
	}
	after, _ := os.ReadFile(filepath.Join(root, config.ConfigFileName))
	if !bytes.Equal(before, after) {
		t.Fatalf("config changed after mismatch:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestAWSConfigRemediationRepairsInvalidAccountID(t *testing.T) {
	root, cfg, _ := setupConfigCheckProject(t)
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "invalid"}
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}
	cfg, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	stubAWSContext(t, "", `{"Account":"012345678901","Arn":"arn:aws:sts::012345678901:assumed-role/Developer/test"}`)

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("\n"),
		Output:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want repaired AWS account")
	}

	updated, updatedInspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if updatedInspection.HasErrors() {
		t.Fatalf("updated findings = %#v, want valid config", updatedInspection.Findings)
	}
	if updated.AWS == nil || updated.AWS.Profile != "dev" || updated.AWS.AccountID != "012345678901" {
		t.Fatalf("AWS = %#v, want repaired verified context", updated.AWS)
	}
}

func TestAWSConfigRemediationQuotesUnquotedAccountID(t *testing.T) {
	root, _, _ := setupConfigCheckProject(t)
	path := filepath.Join(root, config.ConfigFileName)
	content := "schema_version: 1\naws:\n  profile: dev\n  account_id: 012345678901\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cfg, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if !inspection.HasErrors() {
		t.Fatal("expected unquoted account ID finding")
	}

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("\n"),
		Output:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want quoted account ID repair")
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(updated), `account_id: "012345678901"`) {
		t.Fatalf("updated config does not quote account ID:\n%s", updated)
	}
	_, updatedInspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if updatedInspection.HasErrors() {
		t.Fatalf("updated findings = %#v, want valid config", updatedInspection.Findings)
	}
}

func TestAWSConfigRemediationQuoteDeclineDoesNotWrite(t *testing.T) {
	root, _, _ := setupConfigCheckProject(t)
	path := filepath.Join(root, config.ConfigFileName)
	content := []byte("schema_version: 1\naws:\n  account_id: 012345678901\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cfg, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("n\n"),
		Output:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if changed {
		t.Fatal("changed = true, want declined repair to be read-only")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(after, content) {
		t.Fatalf("config changed after declining quote repair:\n%s", after)
	}
}

package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

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
	root, cfg, inspection := setupConfigCheckProject(t)
	cfg.AWS = &config.AWSConfig{AccountID: "999900001111"}
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}
	cfg, inspection, _ = config.LoadWithInspection(root)
	stubAWSContext(t, "dev\n", `{"Account":"012345678901","Arn":"arn:aws:sts::012345678901:assumed-role/Developer/test"}`)
	before, _ := os.ReadFile(filepath.Join(root, config.ConfigFileName))

	_, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
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

func TestAWSConfigRemediationAuthenticationFailureDoesNotWrite(t *testing.T) {
	root, cfg, inspection := setupConfigCheckProject(t)
	stubAWSContext(t, "dev\n", "")
	before, err := os.ReadFile(filepath.Join(root, config.ConfigFileName))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	_, err = remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("\n"),
		Output:      &bytes.Buffer{},
	})
	if err == nil || !strings.Contains(err.Error(), "verify AWS profile") {
		t.Fatalf("error = %v, want authentication failure", err)
	}
	after, readErr := os.ReadFile(filepath.Join(root, config.ConfigFileName))
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("config changed after authentication failure")
	}
}

func TestRunConfigCheckJSONIsReadOnly(t *testing.T) {
	root, _, _ := setupConfigCheckProject(t)
	t.Chdir(root)
	cfg, _, _ := config.LoadWithInspection(root)
	cfg.AWS = config.DisabledAWSConfig()
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}
	before, _ := os.ReadFile(filepath.Join(root, config.ConfigFileName))
	previousJSON := configCheckJSON
	configCheckJSON = true
	t.Cleanup(func() { configCheckJSON = previousJSON })
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)

	if err := runConfigCheck(cmd, nil); err != nil {
		t.Fatalf("runConfigCheck() error = %v", err)
	}
	if !strings.Contains(out.String(), `"schema_state": "current"`) || !strings.Contains(out.String(), `"valid": true`) {
		t.Fatalf("JSON output = %s", out.String())
	}
	after, _ := os.ReadFile(filepath.Join(root, config.ConfigFileName))
	if !bytes.Equal(before, after) {
		t.Fatal("--json modified .kit.yaml")
	}
}

func TestAutomaticConfigCheckFastPathRunsNoAWSSubprocess(t *testing.T) {
	root, cfg, _ := setupConfigCheckProject(t)
	t.Chdir(root)
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "012345678901"}
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}
	before, err := os.ReadFile(filepath.Join(root, config.ConfigFileName))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	calls := 0
	previousOutput := awsCombinedOutput
	awsCombinedOutput = func(context.Context, string, ...string) ([]byte, error) {
		calls++
		return nil, errors.New("unexpected AWS call")
	}
	t.Cleanup(func() { awsCombinedOutput = previousOutput })
	parent := &cobra.Command{Use: "kit"}
	cmd := &cobra.Command{Use: "status"}
	parent.AddCommand(cmd)
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	if err := runAutomaticConfigCheck(cmd, nil); err != nil {
		t.Fatalf("runAutomaticConfigCheck() error = %v", err)
	}
	if calls != 0 {
		t.Fatalf("AWS subprocess calls = %d, want 0", calls)
	}
	after, err := os.ReadFile(filepath.Join(root, config.ConfigFileName))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("automatic fast path modified .kit.yaml")
	}
}

func TestAutomaticConfigCheckRejectsNewerSchema(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.WriteFile(filepath.Join(root, config.ConfigFileName), []byte("schema_version: 2\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	parent := &cobra.Command{Use: "kit"}
	cmd := &cobra.Command{Use: "status"}
	parent.AddCommand(cmd)
	err := runAutomaticConfigCheck(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "upgrade Kit") {
		t.Fatalf("error = %v, want upgrade guidance", err)
	}
}

func TestRunAWSVerifyMatchesConfiguredAccount(t *testing.T) {
	root, cfg, _ := setupConfigCheckProject(t)
	t.Chdir(root)
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "012345678901"}
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}
	stubAWSContext(t, "", `{"Account":"012345678901","Arn":"arn:aws:sts::012345678901:assumed-role/Developer/test","UserId":"user"}`)
	t.Setenv("AWS_PROFILE", "")
	previousJSON := awsVerifyJSON
	awsVerifyJSON = false
	t.Cleanup(func() { awsVerifyJSON = previousJSON })
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)

	if err := runAWSVerify(cmd, nil); err != nil {
		t.Fatalf("runAWSVerify() error = %v", err)
	}
	for _, want := range []string{"AWS context verified", "Profile: dev", "Account: 012345678901", "assumed-role/Developer"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunAWSVerifyRejectsConflictingEnvironmentProfile(t *testing.T) {
	root, cfg, _ := setupConfigCheckProject(t)
	t.Chdir(root)
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "012345678901"}
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}
	t.Setenv("AWS_PROFILE", "prod")

	err := runAWSVerify(&cobra.Command{}, nil)
	if err == nil || !strings.Contains(err.Error(), `AWS_PROFILE "prod" does not match .kit.yaml profile "dev"`) {
		t.Fatalf("error = %v, want conflicting profile rejection", err)
	}
}

func TestRunAWSVerifyRejectsUnquotedAccountID(t *testing.T) {
	root, _, _ := setupConfigCheckProject(t)
	t.Chdir(root)
	content := "schema_version: 1\naws:\n  profile: dev\n  account_id: 012345678901\n"
	if err := os.WriteFile(filepath.Join(root, config.ConfigFileName), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("AWS_PROFILE", "")

	err := runAWSVerify(&cobra.Command{}, nil)
	if err == nil || !strings.Contains(err.Error(), "quoted 12-digit string") || !strings.Contains(err.Error(), "kit config check") {
		t.Fatalf("error = %v, want actionable quote validation", err)
	}
}

func setupConfigCheckProject(t *testing.T) (string, *config.Config, config.Inspection) {
	t.Helper()
	root := t.TempDir()
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.DefaultInstructionScaffoldVersion
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	return root, loaded, inspection
}

func stubAWSContext(t *testing.T, profiles, identity string) {
	t.Helper()
	previousLookPath := awsLookPath
	previousOutput := awsCombinedOutput
	awsLookPath = func(string) (string, error) { return "/usr/local/bin/aws", nil }
	awsCombinedOutput = func(_ context.Context, _ string, args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "configure list-profiles"):
			return []byte(profiles), nil
		case strings.Contains(joined, "sts get-caller-identity"):
			if identity == "" {
				return nil, errors.New("unexpected STS call")
			}
			return []byte(identity), nil
		default:
			return nil, errors.New("unexpected AWS command: " + joined)
		}
	}
	t.Cleanup(func() {
		awsLookPath = previousLookPath
		awsCombinedOutput = previousOutput
	})
}

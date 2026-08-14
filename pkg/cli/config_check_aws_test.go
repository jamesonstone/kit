package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

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

func TestAWSConfigRemediationPersistsSchemaWhenQuoteDeclined(t *testing.T) {
	root, _, _ := setupConfigCheckProject(t)
	path := filepath.Join(root, config.ConfigFileName)
	content := "aws:\n  profile: dev\n  account_id: 012345678901\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cfg, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("\nn\n"),
		Output:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want accepted schema migration persisted")
	}
	updated, _, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if updated.SchemaVersion != config.CurrentSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", updated.SchemaVersion, config.CurrentSchemaVersion)
	}
	updatedContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(updatedContent), "account_id: 012345678901") {
		t.Fatalf("declined account ID quote was applied:\n%s", updatedContent)
	}
}

func TestAWSConfigRemediationPersistsSchemaBeforeAWSError(t *testing.T) {
	root, _, _ := setupConfigCheckProject(t)
	path := filepath.Join(root, config.ConfigFileName)
	if err := os.WriteFile(path, []byte("goal_percentage: 80\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	cfg, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	stubAWSContext(t, "dev\n", "")

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("\n\n"),
		Output:      &bytes.Buffer{},
	})
	if err == nil || !strings.Contains(err.Error(), "verify AWS profile") {
		t.Fatalf("error = %v, want authentication failure", err)
	}
	if !changed {
		t.Fatal("changed = false, want accepted schema migration persisted")
	}
	updated, _, loadErr := config.LoadWithInspection(root)
	if loadErr != nil {
		t.Fatalf("LoadWithInspection() error = %v", loadErr)
	}
	if updated.SchemaVersion != config.CurrentSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", updated.SchemaVersion, config.CurrentSchemaVersion)
	}
}

func TestReadDefaultYesRejectsEmptyEOF(t *testing.T) {
	_, err := readDefaultYes(bufio.NewReader(strings.NewReader("")), &bytes.Buffer{}, "Continue? [Y/n]: ")
	if !errors.Is(err, io.EOF) {
		t.Fatalf("error = %v, want EOF", err)
	}
}

func TestReadDefaultYesAcceptsFinalResponseAtEOF(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  bool
	}{
		{input: "yes", want: true},
		{input: "no", want: false},
	} {
		t.Run(tt.input, func(t *testing.T) {
			got, err := readDefaultYes(bufio.NewReader(strings.NewReader(tt.input)), &bytes.Buffer{}, "Continue? [Y/n]: ")
			if err != nil {
				t.Fatalf("readDefaultYes() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("readDefaultYes() = %v, want %v", got, tt.want)
			}
		})
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
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "012345678901", Region: "us-east-1"}
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
	if err := os.WriteFile(filepath.Join(root, config.ConfigFileName), []byte("schema_version: 3\n"), 0644); err != nil {
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
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "012345678901", Region: "us-east-1"}
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
	for _, want := range []string{"AWS context verified", "Profile: dev", "Account: 012345678901", "Region: us-east-1", "assumed-role/Developer"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunAWSVerifyRejectsConflictingEnvironmentProfile(t *testing.T) {
	root, cfg, _ := setupConfigCheckProject(t)
	t.Chdir(root)
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "012345678901", Region: "us-east-1"}
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
	content := "schema_version: 2\naws:\n  profile: dev\n  account_id: 012345678901\n  region: us-east-1\n"
	if err := os.WriteFile(filepath.Join(root, config.ConfigFileName), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("AWS_PROFILE", "")

	err := runAWSVerify(&cobra.Command{}, nil)
	if err == nil || !strings.Contains(err.Error(), "quoted 12-digit string") || !strings.Contains(err.Error(), "kit config check") {
		t.Fatalf("error = %v, want actionable quote validation", err)
	}
}

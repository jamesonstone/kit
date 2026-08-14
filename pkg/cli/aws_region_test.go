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

func TestAWSConfigRemediationSelectsEnabledRegionForExistingContext(t *testing.T) {
	root, cfg, _ := setupConfigCheckProject(t)
	cfg.AWS = &config.AWSConfig{Profile: "appliedsymbolics", AccountID: "012345678901"}
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}
	cfg, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}

	var calls []string
	previousLookPath := awsLookPath
	previousOutput := awsCombinedOutput
	awsLookPath = func(string) (string, error) { return "/usr/local/bin/aws", nil }
	awsCombinedOutput = func(_ context.Context, _ string, args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		calls = append(calls, joined)
		switch {
		case strings.Contains(joined, "sts get-caller-identity"):
			return []byte(`{"Account":"012345678901","Arn":"arn:aws:sts::012345678901:assumed-role/Developer/test"}`), nil
		case strings.Contains(joined, "configure get region"):
			return []byte("us-east-2\n"), nil
		case strings.Contains(joined, "ec2 describe-regions"):
			return []byte("us-west-1\tus-east-2\tus-east-1\n"), nil
		default:
			return nil, errors.New("unexpected AWS command: " + joined)
		}
	}
	t.Cleanup(func() {
		awsLookPath = previousLookPath
		awsCombinedOutput = previousOutput
	})

	var out bytes.Buffer
	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader("3\n"),
		Output:      &out,
	})
	if err != nil {
		t.Fatalf("remediateProjectConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want selected Region persisted")
	}
	if strings.Contains(out.String(), "Verify and complete AWS profile") {
		t.Fatalf("completed profile/account was reselected:\n%s", out.String())
	}
	for _, region := range []string{"us-east-1", "us-east-2", "us-west-1"} {
		if !strings.Contains(out.String(), region) {
			t.Fatalf("selector missing Region %q:\n%s", region, out.String())
		}
	}
	updated, _, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	if updated.AWS == nil || updated.AWS.Region != "us-west-1" {
		t.Fatalf("AWS = %#v, want selected us-west-1", updated.AWS)
	}
	if !containsCommand(calls, "ec2 describe-regions --profile appliedsymbolics --region us-east-2") {
		t.Fatalf("AWS calls = %#v, want live enabled-Region discovery", calls)
	}
}

func TestSelectAWSRegionRejectsEmptyEOF(t *testing.T) {
	_, err := selectAWSRegion(
		bufio.NewReader(strings.NewReader("")),
		io.Discard,
		[]string{"us-east-1"},
		"us-east-1",
	)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("selectAWSRegion() error = %v, want EOF", err)
	}
}

func TestAWSConfigRemediationRegionEOFDoesNotWrite(t *testing.T) {
	root, cfg, _ := setupConfigCheckProject(t)
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "012345678901"}
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}
	cfg, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		t.Fatalf("LoadWithInspection() error = %v", err)
	}
	stubAWSContext(t, "", `{"Account":"012345678901","Arn":"arn:aws:sts::012345678901:assumed-role/Developer/test"}`)
	path := filepath.Join(root, config.ConfigFileName)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	changed, err := remediateProjectConfig(root, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       strings.NewReader(""),
		Output:      io.Discard,
	})
	if !errors.Is(err, io.EOF) || changed {
		t.Fatalf("remediateProjectConfig() = (%v, %v), want (false, EOF)", changed, err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("Region selection EOF modified .kit.yaml")
	}
}

func TestRunAWSVerifyUsesConfiguredRegion(t *testing.T) {
	root, cfg, _ := setupConfigCheckProject(t)
	t.Chdir(root)
	cfg.AWS = &config.AWSConfig{Profile: "dev", AccountID: "012345678901", Region: "us-west-1"}
	if err := config.UpdateProjectSchemaAndAWS(root, cfg); err != nil {
		t.Fatalf("UpdateProjectSchemaAndAWS() error = %v", err)
	}

	var call string
	previousLookPath := awsLookPath
	previousOutput := awsCombinedOutput
	awsLookPath = func(string) (string, error) { return "/usr/local/bin/aws", nil }
	awsCombinedOutput = func(_ context.Context, _ string, args ...string) ([]byte, error) {
		call = strings.Join(args, " ")
		return []byte(`{"Account":"012345678901","Arn":"arn:aws:sts::012345678901:assumed-role/Developer/test"}`), nil
	}
	t.Cleanup(func() {
		awsLookPath = previousLookPath
		awsCombinedOutput = previousOutput
	})
	t.Setenv("AWS_PROFILE", "")

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	if err := runAWSVerify(cmd, nil); err != nil {
		t.Fatalf("runAWSVerify() error = %v", err)
	}
	if !strings.Contains(call, "--profile dev --region us-west-1") {
		t.Fatalf("STS command = %q, want configured profile and Region", call)
	}
}

func TestReconcileRunsAutomaticConfigRemediation(t *testing.T) {
	parent := &cobra.Command{Use: "kit"}
	command := &cobra.Command{Use: "reconcile"}
	parent.AddCommand(command)
	if skipAutomaticConfigCheck(command) {
		t.Fatal("reconcile skipped automatic config remediation")
	}
}

func containsCommand(calls []string, fragment string) bool {
	for _, call := range calls {
		if strings.Contains(call, fragment) {
			return true
		}
	}
	return false
}

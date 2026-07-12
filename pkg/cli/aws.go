package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
)

var awsVerifyJSON bool

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Verify the project AWS context",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var awsVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify the configured AWS profile and account against .kit.yaml",
	Args:  cobra.NoArgs,
	RunE:  runAWSVerify,
}

type awsVerifyReport struct {
	SchemaVersion int    `json:"schema_version"`
	Profile       string `json:"profile"`
	AccountID     string `json:"account_id"`
	ARN           string `json:"arn"`
	UserID        string `json:"user_id,omitempty"`
}

func init() {
	awsVerifyCmd.Flags().BoolVar(&awsVerifyJSON, "json", false, "emit the verified AWS identity as JSON")
	awsCmd.AddCommand(awsVerifyCmd)
	rootCmd.AddCommand(awsCmd)
}

func runAWSVerify(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, inspection, err := config.LoadWithInspection(projectRoot)
	if err != nil {
		return err
	}
	if inspection.SchemaState == config.SchemaStateNewer {
		return fmt.Errorf("%s", inspection.Findings[0].Message)
	}
	for _, finding := range inspection.Findings {
		if finding.Severity == config.FindingError {
			return fmt.Errorf("invalid .kit.yaml field %s: %s; run `kit config check`", finding.Field, finding.Message)
		}
	}
	if cfg.AWS == nil {
		return fmt.Errorf("AWS context is not configured; run `kit config check`")
	}
	if !cfg.AWS.IsEnabled() {
		return fmt.Errorf("AWS context is disabled in .kit.yaml")
	}
	if !validAWSAccountID(cfg.AWS.AccountID) || strings.TrimSpace(cfg.AWS.Profile) == "" {
		return fmt.Errorf("AWS context is incomplete; run `kit config check`")
	}

	profile := strings.TrimSpace(cfg.AWS.Profile)
	if environmentProfile := strings.TrimSpace(os.Getenv("AWS_PROFILE")); environmentProfile != "" && environmentProfile != profile {
		return fmt.Errorf(
			"AWS_PROFILE %q does not match .kit.yaml profile %q; unset it or select the configured project profile",
			environmentProfile,
			profile,
		)
	}
	identity, err := resolveAWSIdentity(profile)
	if err != nil {
		return err
	}
	if identity.Account != cfg.AWS.AccountID {
		return fmt.Errorf(
			"AWS account mismatch: profile %q resolves to %s, but .kit.yaml expects %s",
			profile,
			identity.Account,
			cfg.AWS.AccountID,
		)
	}
	report := awsVerifyReport{
		SchemaVersion: config.CurrentSchemaVersion,
		Profile:       profile,
		AccountID:     identity.Account,
		ARN:           identity.ARN,
		UserID:        identity.UserID,
	}
	if awsVerifyJSON {
		return outputJSON(cmd.OutOrStdout(), report)
	}
	_, err = fmt.Fprintf(
		cmd.OutOrStdout(),
		"✅ AWS context verified\n  Profile: %s\n  Account: %s\n  ARN: %s\n",
		report.Profile,
		report.AccountID,
		report.ARN,
	)
	return err
}

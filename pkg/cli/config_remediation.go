package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

func remediateProjectConfig(
	projectRoot string,
	cfg *config.Config,
	inspection config.Inspection,
	opts configRemediationOptions,
) (bool, error) {
	if !opts.Interactive {
		return false, nil
	}
	reader := bufio.NewReader(opts.Input)
	changed := false
	persistAWS := false
	continueRemediation := true
	var remediationErr error
	if inspection.NeedsSchemaMigration() {
		ok, err := readDefaultYes(reader, opts.Output, fmt.Sprintf(
			"Update .kit.yaml to schema_version %d? [Y/n]: ",
			config.CurrentSchemaVersion,
		))
		if err != nil {
			remediationErr = err
		} else if ok {
			cfg.SchemaVersion = config.CurrentSchemaVersion
			changed = true
		} else {
			return false, nil
		}
	}
	if remediationErr == nil && awsAccountIDNeedsQuoting(cfg, inspection) {
		ok, err := readDefaultYes(reader, opts.Output, "Quote aws.account_id as a YAML string? [Y/n]: ")
		if err != nil {
			remediationErr = err
		} else if ok {
			changed = true
			persistAWS = true
		} else if !changed {
			return false, nil
		} else {
			continueRemediation = false
		}
	}

	if remediationErr == nil && continueRemediation {
		awsChanged, err := remediateAWSConfig(reader, opts.Output, cfg)
		changed = awsChanged || changed
		persistAWS = awsChanged || persistAWS
		remediationErr = err
	}
	if !changed {
		return false, remediationErr
	}
	persistedConfig := *cfg
	if !persistAWS {
		persistedConfig.AWS = nil
	}
	if err := config.UpdateProjectSchemaAndAWS(projectRoot, &persistedConfig); err != nil {
		if remediationErr != nil {
			return false, errors.Join(remediationErr, err)
		}
		return false, err
	}
	_, _ = fmt.Fprintln(opts.Output, "  ✓ Updated .kit.yaml")
	return true, remediationErr
}

func awsAccountIDNeedsQuoting(cfg *config.Config, inspection config.Inspection) bool {
	if cfg == nil || cfg.AWS == nil || !cfg.AWS.IsEnabled() || !validAWSAccountID(strings.TrimSpace(cfg.AWS.AccountID)) {
		return false
	}
	for _, finding := range inspection.Findings {
		if finding.Field == "aws.account_id" && finding.Repairable {
			return true
		}
	}
	return false
}

func remediateAWSConfig(reader *bufio.Reader, out io.Writer, cfg *config.Config) (bool, error) {
	if cfg.AWS != nil && !cfg.AWS.IsEnabled() {
		return false, nil
	}
	if cfg.AWS != nil && strings.TrimSpace(cfg.AWS.Profile) != "" && validAWSAccountID(cfg.AWS.AccountID) {
		return false, nil
	}

	profile := ""
	if cfg.AWS != nil {
		profile = strings.TrimSpace(cfg.AWS.Profile)
	}
	if profile != "" {
		ok, err := readDefaultYes(reader, out, fmt.Sprintf("Verify and complete AWS profile %q? [Y/n]: ", profile))
		if err != nil || !ok {
			return false, err
		}
	} else {
		profiles, err := discoverAWSProfiles()
		if errors.Is(err, errAWSCLINotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if len(profiles) == 0 {
			return false, nil
		}
		if len(profiles) == 1 {
			ok, err := readDefaultYes(reader, out, fmt.Sprintf("Use the only AWS profile %q for this project? [Y/n]: ", profiles[0]))
			if err != nil {
				return false, err
			}
			if !ok {
				cfg.AWS = config.DisabledAWSConfig()
				return true, nil
			}
			profile = profiles[0]
		} else {
			selected, err := selectAWSProfile(reader, out, profiles)
			if err != nil {
				return false, err
			}
			if selected == "" {
				cfg.AWS = config.DisabledAWSConfig()
				return true, nil
			}
			profile = selected
		}
	}

	identity, err := resolveAWSIdentity(profile)
	if err != nil {
		return false, err
	}
	if !validAWSAccountID(identity.Account) {
		return false, fmt.Errorf("AWS profile %q returned invalid account ID %q", profile, identity.Account)
	}
	configuredAccountID := ""
	if cfg.AWS != nil {
		configuredAccountID = strings.TrimSpace(cfg.AWS.AccountID)
	}
	if validAWSAccountID(configuredAccountID) && configuredAccountID != identity.Account {
		return false, fmt.Errorf(
			"AWS profile %q resolves to account %s, but .kit.yaml expects %s",
			profile,
			identity.Account,
			configuredAccountID,
		)
	}
	cfg.AWS = &config.AWSConfig{Profile: profile, AccountID: identity.Account}
	return true, nil
}

func readDefaultYes(reader *bufio.Reader, out io.Writer, prompt string) (bool, error) {
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return false, err
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) || line == "" {
			return false, err
		}
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("answer must be yes or no")
	}
}

func selectAWSProfile(reader *bufio.Reader, out io.Writer, profiles []string) (string, error) {
	_, _ = fmt.Fprintln(out, "Select an AWS profile for this project:")
	_, _ = fmt.Fprintln(out, "  0. Do not use AWS for this project")
	for i, profile := range profiles {
		_, _ = fmt.Fprintf(out, "  %d. %s\n", i+1, profile)
	}
	_, _ = fmt.Fprintf(out, "Enter number [0-%d]: ", len(profiles))
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	selection, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || selection < 0 || selection > len(profiles) {
		return "", fmt.Errorf("selection must be between 0 and %d", len(profiles))
	}
	if selection == 0 {
		return "", nil
	}
	return profiles[selection-1], nil
}

func validAWSAccountID(value string) bool {
	if len(value) != 12 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

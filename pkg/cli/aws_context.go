package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const (
	awsProfileDiscoveryTimeout = 3 * time.Second
	awsIdentityTimeout         = 15 * time.Second
)

var (
	errAWSCLINotFound = errors.New("AWS CLI not found")
	awsLookPath       = exec.LookPath
	awsCombinedOutput = runAWSCombinedOutput
)

type awsIdentity struct {
	Account string `json:"Account"`
	ARN     string `json:"Arn"`
	UserID  string `json:"UserId"`
}

func discoverAWSProfiles() ([]string, error) {
	path, err := awsLookPath("aws")
	if err != nil {
		return nil, errAWSCLINotFound
	}
	ctx, cancel := context.WithTimeout(context.Background(), awsProfileDiscoveryTimeout)
	defer cancel()
	output, err := awsCombinedOutput(ctx, path, "configure", "list-profiles", "--no-cli-pager")
	if err != nil {
		return nil, fmt.Errorf("list AWS profiles: %w", err)
	}

	seen := make(map[string]bool)
	var profiles []string
	for _, line := range strings.Split(string(output), "\n") {
		profile := strings.TrimSpace(line)
		if profile == "" || seen[profile] {
			continue
		}
		seen[profile] = true
		profiles = append(profiles, profile)
	}
	sort.Strings(profiles)
	return profiles, nil
}

func resolveAWSIdentity(profile string) (awsIdentity, error) {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return awsIdentity{}, fmt.Errorf("AWS profile is required")
	}
	path, err := awsLookPath("aws")
	if err != nil {
		return awsIdentity{}, errAWSCLINotFound
	}
	ctx, cancel := context.WithTimeout(context.Background(), awsIdentityTimeout)
	defer cancel()
	output, err := awsCombinedOutput(
		ctx,
		path,
		"sts", "get-caller-identity",
		"--profile", profile,
		"--output", "json",
		"--no-cli-pager",
	)
	if err != nil {
		return awsIdentity{}, fmt.Errorf("verify AWS profile %q: %w", profile, err)
	}

	var identity awsIdentity
	if err := json.Unmarshal(output, &identity); err != nil {
		return awsIdentity{}, fmt.Errorf("parse caller identity for profile %q: %w", profile, err)
	}
	if identity.Account == "" || identity.ARN == "" {
		return awsIdentity{}, fmt.Errorf("caller identity for profile %q is missing account or ARN", profile)
	}
	return identity, nil
}

func runAWSCombinedOutput(ctx context.Context, path string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = append(os.Environ(), "AWS_PAGER=")
	output, err := cmd.CombinedOutput()
	if err == nil {
		return output, nil
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return nil, err
	}
	return nil, fmt.Errorf("%w: %s", err, detail)
}

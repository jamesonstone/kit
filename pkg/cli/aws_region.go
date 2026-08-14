package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
)

const awsRegionDiscoveryTimeout = 15 * time.Second

func discoverEnabledAWSRegions(profile string) ([]string, string, error) {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return nil, "", fmt.Errorf("AWS profile is required")
	}
	path, err := awsLookPath("aws")
	if err != nil {
		return nil, "", errAWSCLINotFound
	}

	configuredRegion := discoverConfiguredAWSRegion(path, profile)
	bootstrapRegion := configuredRegion
	if !config.ValidAWSRegion(bootstrapRegion) {
		bootstrapRegion = "us-east-1"
	}
	ctx, cancel := context.WithTimeout(context.Background(), awsRegionDiscoveryTimeout)
	defer cancel()
	output, err := awsCombinedOutput(
		ctx,
		path,
		"ec2", "describe-regions",
		"--profile", profile,
		"--region", bootstrapRegion,
		"--query", "Regions[].RegionName",
		"--output", "text",
		"--no-cli-pager",
	)
	if err != nil {
		return nil, "", fmt.Errorf("list enabled AWS Regions for profile %q: %w", profile, err)
	}

	seen := make(map[string]bool)
	regions := make([]string, 0)
	for _, region := range strings.Fields(string(output)) {
		if !config.ValidAWSRegion(region) || seen[region] {
			continue
		}
		seen[region] = true
		regions = append(regions, region)
	}
	if len(regions) == 0 {
		return nil, "", fmt.Errorf("profile %q returned no enabled AWS Regions", profile)
	}
	sort.Strings(regions)
	return regions, preferredAWSRegion(regions, configuredRegion), nil
}

func discoverConfiguredAWSRegion(path, profile string) string {
	ctx, cancel := context.WithTimeout(context.Background(), awsProfileDiscoveryTimeout)
	defer cancel()
	output, err := awsCombinedOutput(
		ctx,
		path,
		"configure", "get", "region",
		"--profile", profile,
		"--no-cli-pager",
	)
	if err != nil {
		return ""
	}
	region := strings.TrimSpace(string(output))
	if !config.ValidAWSRegion(region) {
		return ""
	}
	return region
}

func preferredAWSRegion(regions []string, configured string) string {
	for _, candidate := range []string{configured, "us-east-1"} {
		for _, region := range regions {
			if candidate != "" && region == candidate {
				return candidate
			}
		}
	}
	return regions[0]
}

func selectAWSRegion(reader *bufio.Reader, out io.Writer, regions []string, defaultRegion string) (string, error) {
	if len(regions) == 0 {
		return "", fmt.Errorf("at least one AWS Region is required")
	}
	defaultIndex := 0
	_, _ = fmt.Fprintln(out, "Select the default AWS Region for this project:")
	for index, region := range regions {
		if region == defaultRegion {
			defaultIndex = index
		}
		marker := ""
		if region == defaultRegion {
			marker = " (default)"
		}
		_, _ = fmt.Fprintf(out, "  %d. %s%s\n", index+1, region, marker)
	}
	_, _ = fmt.Fprintf(out, "Enter number [1-%d] (default %d): ", len(regions), defaultIndex+1)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if errors.Is(err, io.EOF) && line == "" {
		return "", io.EOF
	}
	answer := strings.TrimSpace(line)
	if answer == "" {
		return regions[defaultIndex], nil
	}
	selection, err := strconv.Atoi(answer)
	if err != nil || selection < 1 || selection > len(regions) {
		return "", fmt.Errorf("selection must be between 1 and %d", len(regions))
	}
	return regions[selection-1], nil
}

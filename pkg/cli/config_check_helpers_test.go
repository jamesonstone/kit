package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

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

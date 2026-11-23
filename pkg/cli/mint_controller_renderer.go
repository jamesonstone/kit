package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var renderMintController = runMintControllerRenderer

func runMintControllerRenderer(root string, release mintActionRelease, environment string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	workspace, err := os.MkdirTemp("", "kit-mint-renderer-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(workspace) }()
	// A standalone module lets Go select Mint's toolchain even if the project
	// has no go.mod. The build cannot use the project's module or replace rules.
	if err := os.WriteFile(filepath.Join(workspace, "go.mod"), []byte("module kit.invalid/mint-renderer\n"), 0o600); err != nil {
		return "", err
	}
	install := exec.CommandContext(ctx, "go", "install", "github.com/jamesonstone/mint/cmd/mint@"+release.SHA)
	install.Dir = workspace
	install.Env = append(install.Environ(), "GOWORK=off", "GOFLAGS=", "GOBIN="+workspace, "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH)
	if _, err := mintRendererOutput(install); err != nil {
		return "", fmt.Errorf("build Mint %s renderer: %w", release.Version, err)
	}
	binary := filepath.Join(workspace, "mint")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	cmd := exec.CommandContext(ctx, binary, "release", "production", "workflow", "--config", mintPolicyPath,
		"--mint-ref", release.SHA, "--environment", environment)
	cmd.Dir = root
	output, err := mintRendererOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("render controller with Mint %s: %w", release.Version, err)
	}
	if len(output) == 0 {
		return "", fmt.Errorf("empty controller workflow from Mint %s", release.Version)
	}
	return output, nil
}

func mintRendererOutput(cmd *exec.Cmd) (string, error) {
	output, err := cmd.Output()
	if exit, ok := err.(*exec.ExitError); ok && len(exit.Stderr) != 0 {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exit.Stderr)))
	}
	return string(output), err
}

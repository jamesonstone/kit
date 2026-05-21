package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/runstore"
	verifyengine "github.com/jamesonstone/kit/internal/verify"
)

var verifyTaskID string
var verifyJSON bool
var verifyDryRun bool
var verifyNoWrite bool
var verifyAllowShell bool
var verifyTimeout string

var verifyCmd = &cobra.Command{
	Use:          "verify [feature]",
	Short:        "Run declared verification checks",
	SilenceUsage: true,
	Long: `Run verification commands declared in TASKS.md.

By default commands are parsed as argv and executed without shell evaluation.
Use --dry-run to inspect selected commands without running them.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runVerify,
}

func init() {
	verifyCmd.Flags().StringVar(&verifyTaskID, "task", "", "run verification for a single task ID")
	verifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "output verification result as JSON")
	verifyCmd.Flags().BoolVar(&verifyDryRun, "dry-run", false, "show selected commands without executing them")
	verifyCmd.Flags().BoolVar(&verifyNoWrite, "no-write", false, "do not write .kit/runs artifacts")
	verifyCmd.Flags().BoolVar(&verifyAllowShell, "allow-shell", false, "allow shell command declarations")
	verifyCmd.Flags().StringVar(&verifyTimeout, "timeout", "", "per-command timeout such as 30s or 2m")
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(cmd *cobra.Command, args []string) error {
	projectRoot, _, feat, err := resolveOptionalFeature(args)
	if err != nil {
		return err
	}

	timeout, err := parseOptionalDuration(verifyTimeout)
	if err != nil {
		return err
	}
	bundles, err := verifyengine.LoadTaskBundles(
		featureTasksPath(feat.Path),
		verifyengine.FeatureRefFromDir(feat.Path),
		verifyAllowShell,
	)
	if err != nil {
		return err
	}
	if verifyTaskID != "" {
		if _, ok := verifyengine.FindTaskBundle(bundles, verifyTaskID); !ok {
			return fmt.Errorf("task %s not found in %s", verifyTaskID, featureTasksPath(feat.Path))
		}
	}
	commands, taskIDs := verifyengine.SelectCommands(bundles, verifyTaskID)
	expectedFiles := verifyengine.SelectExpectedFiles(bundles, verifyTaskID)
	run := verifyengine.ExecuteRun(context.Background(), verifyengine.RunOptions{
		ProjectRoot:   projectRoot,
		Feature:       verifyengine.FeatureRefFromDir(feat.Path),
		TaskIDs:       taskIDs,
		ExpectedFiles: expectedFiles,
		Commands:      commands,
		DryRun:        verifyDryRun,
		Timeout:       timeout,
	})
	if !verifyDryRun && !verifyNoWrite {
		if err := runstore.Write(projectRoot, &run); err != nil {
			return err
		}
	}

	var outputErr error
	if verifyJSON {
		outputErr = outputJSON(cmd.OutOrStdout(), run)
	} else {
		outputErr = outputVerifyText(cmd, run)
	}
	if outputErr != nil {
		return outputErr
	}
	if run.Status == verifyengine.RunStatusFail {
		return fmt.Errorf("verification failed")
	}
	return nil
}

func resolveOptionalFeature(args []string) (string, *config.Config, *feature.Feature, error) {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return "", nil, nil, err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", nil, nil, err
	}
	specsDir := cfg.SpecsPath(projectRoot)
	if len(args) == 1 {
		feat, err := loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return "", nil, nil, err
		}
		return projectRoot, cfg, feat, nil
	}
	feat, err := feature.FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		return "", nil, nil, err
	}
	if feat == nil {
		return "", nil, nil, fmt.Errorf("feature name required when there is no active feature")
	}
	return projectRoot, cfg, feat, nil
}

func featureTasksPath(featurePath string) string {
	return filepath.Join(featurePath, "TASKS.md")
}

func outputVerifyText(cmd *cobra.Command, run verifyengine.Run) error {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Verification run: %s\n", run.RunID)
	fmt.Fprintf(out, "Feature: %s\n", run.Feature.DirName)
	fmt.Fprintf(out, "Status: %s\n", run.Status)
	if run.ArtifactDir != "" {
		fmt.Fprintf(out, "Artifacts: %s\n", run.ArtifactDir)
	}
	if len(run.Commands) == 0 {
		fmt.Fprintln(out, "Commands: none")
		return nil
	}
	fmt.Fprintln(out, "Commands:")
	if run.Status == verifyengine.RunStatusDryRun {
		for _, command := range run.Commands {
			fmt.Fprintf(out, "  - [%s] %s\n", command.TaskID, command.Raw)
		}
		return nil
	}
	for _, result := range run.Results {
		fmt.Fprintf(out, "  - [%s] %s: %s", result.TaskID, result.Raw, result.Status)
		if result.ExitCode != 0 {
			fmt.Fprintf(out, " (exit %d)", result.ExitCode)
		}
		fmt.Fprintln(out)
	}
	return nil
}

func parseOptionalDuration(value string) (time.Duration, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout %q: %w", value, err)
	}
	return duration, nil
}

func outputJSON(out interface{ Write([]byte) (int, error) }, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

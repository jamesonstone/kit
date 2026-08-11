package cli

import (
	"bytes"
	"context"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/commandset"
	"github.com/jamesonstone/kit/internal/releaseprompt"
)

func TestPROrchestrateIsProtectedTelemeteredAndCapabilityDescribed(t *testing.T) {
	if !slices.Contains(commandset.ProtectedPaths(), "pr orchestrate") {
		t.Fatal("pr orchestrate is not in the protected v2 command surface")
	}
	if !commandset.IsTelemetryPath("pr orchestrate") {
		t.Fatal("pr orchestrate is not in bounded usage telemetry")
	}
	record, ok := capabilityByCommandPath("pr orchestrate")
	if !ok {
		t.Fatal("pr orchestrate capability is missing")
	}
	if record.MutationLevel != mutationNetwork || record.GitMutation.Summary == "" || !strings.Contains(record.WhenNotToUse[0], "does not enumerate PRs") {
		t.Fatalf("pr orchestrate capability = %#v", record)
	}
}

func TestPROrchestrateCommandExposesAcceptedFlags(t *testing.T) {
	registered, _, err := rootCmd.Find([]string{"pr", "orchestrate"})
	if err != nil || registered == nil || registered.CommandPath() != "kit pr orchestrate" {
		t.Fatalf("registered pr orchestrate command = %#v, err=%v", registered, err)
	}
	cmd := newPROrchestrateCommand()
	for _, name := range []string{
		"repos", "root", "project", "organization", "context", "scope",
		"infra", "infra-provider", "infra-cli", "environment", "verify",
		"integration-suite", "dry-run", "output-only", "copy",
	} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("pr orchestrate missing --%s", name)
		}
	}
	if !strings.Contains(cmd.Long, "does not enumerate the release set") {
		t.Fatalf("command help omits prompt-only boundary:\n%s", cmd.Long)
	}
}

func TestPROrchestrateNoninteractiveRequiresScope(t *testing.T) {
	restorePROrchestrateGlobals(t)
	prOrchestrateInteractiveCheck = func(io.Reader, io.Writer) bool { return false }
	cmd := newPROrchestrateCommand()
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "use --repos or --root") {
		t.Fatalf("missing-scope error = %v", err)
	}
}

func TestPROrchestrateNoninteractiveWritesOnlyRawPrompt(t *testing.T) {
	var captured releaseprompt.Input
	stubPROrchestratePipeline(t, false, func(input releaseprompt.Input) {
		captured = input
	})
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	cmd := newPROrchestrateCommand()
	cmd.SetOut(output)
	cmd.SetErr(errors)
	cmd.SetArgs([]string{"--repos", "/repo/a", "--repos", "/repo/b", "--context", "release context"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if output.String() != "rendered release prompt\n" || errors.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q", output.String(), errors.String())
	}
	if !slices.Equal(captured.Repositories, []string{"/repo/a", "/repo/b"}) || captured.FeatureContext != "release context" {
		t.Fatalf("captured input = %#v", captured)
	}
}

func TestPROrchestrateInteractiveScopeCopiesAndAcknowledges(t *testing.T) {
	var captured releaseprompt.Input
	copied := stubPROrchestratePipeline(t, true, func(input releaseprompt.Input) {
		captured = input
	})
	output := &bytes.Buffer{}
	errors := &bytes.Buffer{}
	cmd := newPROrchestrateCommand()
	cmd.SetIn(strings.NewReader("repos:/repo/a,/repo/b\n"))
	cmd.SetOut(output)
	cmd.SetErr(errors)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(captured.Repositories, []string{"/repo/a", "/repo/b"}) || captured.Root != "" {
		t.Fatalf("interactive scope = %#v", captured)
	}
	if *copied != "rendered release prompt\n" {
		t.Fatalf("copied prompt = %q", *copied)
	}
	if output.String() != "Copied the prepared text to the clipboard.\n" {
		t.Fatalf("interactive output = %q", output.String())
	}
	if !strings.Contains(errors.String(), "Repository scope") {
		t.Fatalf("interactive prompt was not written to stderr: %q", errors.String())
	}
}

func TestPROrchestrateDryRunAlwaysWritesBundleWithoutCopy(t *testing.T) {
	copied := stubPROrchestratePipeline(t, true, nil)
	output := &bytes.Buffer{}
	cmd := newPROrchestrateCommand()
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(output)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--root", "/repo", "--dry-run", "--copy"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if output.String() != "dry-run bundle\n" || *copied != "" {
		t.Fatalf("dry-run stdout=%q copied=%q", output.String(), *copied)
	}
}

func TestPROrchestrateOutputOnlyAndCopyDoesBoth(t *testing.T) {
	copied := stubPROrchestratePipeline(t, false, nil)
	output := &bytes.Buffer{}
	cmd := newPROrchestrateCommand()
	cmd.SetOut(output)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--root", "/repo", "--output-only", "--copy"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if output.String() != "rendered release prompt\n" || *copied != output.String() {
		t.Fatalf("output=%q copied=%q", output.String(), *copied)
	}
}

func stubPROrchestratePipeline(t *testing.T, interactive bool, capture func(releaseprompt.Input)) *string {
	t.Helper()
	restorePROrchestrateGlobals(t)
	prOrchestrateInteractiveCheck = func(io.Reader, io.Writer) bool { return interactive }
	prOrchestrateResolve = func(_ context.Context, input releaseprompt.Input, _ releaseprompt.Runner) (releaseprompt.Config, error) {
		if capture != nil {
			capture(input)
		}
		return releaseprompt.Config{Project: "test"}, nil
	}
	prOrchestrateRender = func(releaseprompt.Config) (string, error) { return "rendered release prompt\n", nil }
	prOrchestrateRenderDryRun = func(releaseprompt.Config, string) (string, error) { return "dry-run bundle\n", nil }
	copied := ""
	clipboardCopyFunc = func(value string) error {
		copied = value
		return nil
	}
	return &copied
}

func restorePROrchestrateGlobals(t *testing.T) {
	t.Helper()
	previousRunner := prOrchestrateRunner
	previousResolve := prOrchestrateResolve
	previousRender := prOrchestrateRender
	previousDryRun := prOrchestrateRenderDryRun
	previousInteractive := prOrchestrateInteractiveCheck
	previousClipboard := clipboardCopyFunc
	t.Cleanup(func() {
		prOrchestrateRunner = previousRunner
		prOrchestrateResolve = previousResolve
		prOrchestrateRender = previousRender
		prOrchestrateRenderDryRun = previousDryRun
		prOrchestrateInteractiveCheck = previousInteractive
		clipboardCopyFunc = previousClipboard
	})
}

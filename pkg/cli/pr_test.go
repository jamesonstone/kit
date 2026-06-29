package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPRFixCommandIsRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"pr", "fix"})
	if err != nil {
		t.Fatalf("rootCmd.Find(pr fix) error = %v", err)
	}
	if cmd == nil || cmd.CommandPath() != "kit pr fix" {
		t.Fatalf("expected kit pr fix command, got %#v", cmd)
	}
	if flag := cmd.Flags().Lookup("pr"); flag == nil {
		t.Fatal("expected pr fix to expose --pr")
	}
	if flag := cmd.Flags().Lookup("coderabbit"); flag == nil {
		t.Fatal("expected pr fix to expose --coderabbit")
	}
	if flag := cmd.Flags().Lookup("output-only"); flag == nil {
		t.Fatal("expected pr fix to expose --output-only")
	}
	if flag := cmd.Flags().Lookup("max-subagents"); flag == nil {
		t.Fatal("expected pr fix to expose --max-subagents")
	} else if flag.DefValue != "3" || !strings.Contains(flag.Usage, "hard ceiling 4") {
		t.Fatalf("unexpected --max-subagents flag metadata: def=%q usage=%q", flag.DefValue, flag.Usage)
	}
	if !strings.Contains(cmd.Long, "only active (unresolved, non-outdated) review threads") {
		t.Fatalf("expected pr fix help to document active review-thread filtering, got:\n%s", cmd.Long)
	}
	if !strings.Contains(cmd.Long, "post-push reflection") {
		t.Fatalf("expected pr fix help to document post-push reflection, got:\n%s", cmd.Long)
	}
}

func TestRunPRFixCommandRoutesExplicitPRToDispatchPrompt(t *testing.T) {
	var gotOpts prFixDispatchOptions
	restore := installPRFixFakes(t,
		func(_ *cobra.Command, opts prFixDispatchOptions) error {
			gotOpts = opts
			return nil
		},
		nil,
	)
	defer restore()

	cmd := newPRFixCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"--pr", "Patient-Driven-Care/cortex#67",
		"--coderabbit",
		"--output-only",
		"--copy",
		"--max-subagents", "3",
		"--editor", "vim",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("kit pr fix --pr error = %v", err)
	}
	if gotOpts.PRRef != "Patient-Driven-Care/cortex#67" {
		t.Fatalf("PRRef = %q", gotOpts.PRRef)
	}
	if !gotOpts.CodeRabbitOnly || !gotOpts.OutputOnly || !gotOpts.Copy {
		t.Fatalf("expected dispatch prompt flags to be forwarded: %#v", gotOpts)
	}
	if gotOpts.MaxSubagents != 3 || gotOpts.Editor != "vim" {
		t.Fatalf("dispatch prompt options not forwarded: %#v", gotOpts)
	}
}

func TestRunPRFixCommandSelectsOpenPRWhenOmitted(t *testing.T) {
	var gotPRRef string
	restore := installPRFixFakes(t,
		func(_ *cobra.Command, opts prFixDispatchOptions) error {
			gotPRRef = opts.PRRef
			return nil
		},
		func() ([]prFixOpenPullRequest, error) {
			return []prFixOpenPullRequest{
				{Number: 12, Title: "first", URL: "https://github.com/acme/app/pull/12", HeadRefName: "GH-12", BaseRefName: "main", ReviewDecision: "REVIEW_REQUIRED"},
				{Number: 67, Title: "fix review", URL: "https://github.com/acme/app/pull/67", HeadRefName: "GH-67", BaseRefName: "main"},
			}, nil
		},
	)
	defer restore()

	cmd := newPRFixCommand()
	input := strings.NewReader("2\n")
	output := &bytes.Buffer{}
	cmd.SetIn(input)
	cmd.SetOut(output)
	cmd.SetErr(output)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("kit pr fix selector error = %v", err)
	}
	if gotPRRef != "https://github.com/acme/app/pull/67" {
		t.Fatalf("selected PR ref = %q", gotPRRef)
	}
	rendered := output.String()
	if !strings.Contains(rendered, "Open pull requests:") || !strings.Contains(rendered, "#67 fix review") {
		t.Fatalf("selector output missing PR list:\n%s", rendered)
	}
}

func TestRunPRFixCommandSelectorAcceptsPullRequestNumber(t *testing.T) {
	var gotPRRef string
	restore := installPRFixFakes(t,
		func(_ *cobra.Command, opts prFixDispatchOptions) error {
			gotPRRef = opts.PRRef
			return nil
		},
		func() ([]prFixOpenPullRequest, error) {
			return []prFixOpenPullRequest{
				{Number: 12, Title: "first", URL: "https://github.com/acme/app/pull/12"},
				{Number: 67, Title: "fix review", URL: "https://github.com/acme/app/pull/67"},
			}, nil
		},
	)
	defer restore()

	cmd := newPRFixCommand()
	output := &bytes.Buffer{}
	cmd.SetIn(strings.NewReader("67\n"))
	cmd.SetOut(output)
	cmd.SetErr(output)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("kit pr fix selector error = %v", err)
	}
	if gotPRRef != "https://github.com/acme/app/pull/67" {
		t.Fatalf("selected PR ref = %q", gotPRRef)
	}
}

func TestRunPRFixCommandRejectsLegacyFeatureArgument(t *testing.T) {
	cmd := newPRFixCommand()
	cmd.SetArgs([]string{"legacy-feature"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "legacy-feature") {
		t.Fatalf("expected no-args guard, got %v", err)
	}
}

func TestRunPRFixCommandReportsNoOpenPullRequests(t *testing.T) {
	restore := installPRFixFakes(t,
		func(_ *cobra.Command, _ prFixDispatchOptions) error {
			t.Fatal("dispatch prompt runner should not be called")
			return nil
		},
		func() ([]prFixOpenPullRequest, error) {
			return nil, nil
		},
	)
	defer restore()

	cmd := newPRFixCommand()
	cmd.SetIn(strings.NewReader("\n"))
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "no open pull requests") {
		t.Fatalf("expected no-open-PR error, got %v", err)
	}
}

func TestRunPRFixCommandPropagatesPullRequestListError(t *testing.T) {
	wantErr := errors.New("gh unavailable")
	restore := installPRFixFakes(t, nil, func() ([]prFixOpenPullRequest, error) {
		return nil, wantErr
	})
	defer restore()

	cmd := newPRFixCommand()
	cmd.SetIn(strings.NewReader("\n"))
	if err := cmd.Execute(); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func installPRFixFakes(
	t *testing.T,
	runner func(*cobra.Command, prFixDispatchOptions) error,
	lister func() ([]prFixOpenPullRequest, error),
) func() {
	t.Helper()
	previousRunner := prFixDispatchRunner
	previousLister := prFixOpenPRLister
	if runner != nil {
		prFixDispatchRunner = runner
	}
	if lister != nil {
		prFixOpenPRLister = lister
	}
	return func() {
		prFixDispatchRunner = previousRunner
		prFixOpenPRLister = previousLister
	}
}

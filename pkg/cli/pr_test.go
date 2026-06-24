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
	if flag := cmd.Flags().Lookup("subagents"); flag == nil {
		t.Fatal("expected pr fix to expose --subagents")
	}
	if flag := cmd.Flags().Lookup("wait-for-coderabbit"); flag == nil {
		t.Fatal("expected pr fix to expose --wait-for-coderabbit")
	}
}

func TestRunPRFixCommandRoutesExplicitPRToLoopReview(t *testing.T) {
	var gotArgs []string
	var gotOpts loopReviewOptions
	restore := installPRFixFakes(t,
		func(cmd *cobra.Command, args []string, opts loopReviewOptions) error {
			gotArgs = append([]string(nil), args...)
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
		"--watch",
		"--dry-run",
		"--subagents",
		"--min-confidence", "98",
		"--max-iterations", "3",
		"review-loop",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("kit pr fix --pr error = %v", err)
	}
	if strings.Join(gotArgs, " ") != "review-loop" {
		t.Fatalf("args = %#v, want feature arg", gotArgs)
	}
	if gotOpts.PRRef != "Patient-Driven-Care/cortex#67" {
		t.Fatalf("PRRef = %q", gotOpts.PRRef)
	}
	if !gotOpts.WaitForCodeRabbit || !gotOpts.DryRun || !gotOpts.UseSubagents {
		t.Fatalf("expected watch, dry-run, and subagents to be forwarded: %#v", gotOpts)
	}
	if !gotOpts.ResolvePRFeedback {
		t.Fatalf("expected pr fix to enable review-thread resolution guidance: %#v", gotOpts)
	}
	if gotOpts.MinConfidence != 98 || gotOpts.MaxIterations != 3 {
		t.Fatalf("loop bounds not forwarded: %#v", gotOpts)
	}
}

func TestRunPRFixCommandSelectsOpenPRWhenOmitted(t *testing.T) {
	var gotPRRef string
	restore := installPRFixFakes(t,
		func(_ *cobra.Command, _ []string, opts loopReviewOptions) error {
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
		func(_ *cobra.Command, _ []string, opts loopReviewOptions) error {
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

func TestRunPRFixCommandRequiresPRForJSON(t *testing.T) {
	cmd := newPRFixCommand()
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--json requires --pr") {
		t.Fatalf("expected JSON selection guard, got %v", err)
	}
}

func TestRunPRFixCommandReportsNoOpenPullRequests(t *testing.T) {
	restore := installPRFixFakes(t,
		func(_ *cobra.Command, _ []string, _ loopReviewOptions) error {
			t.Fatal("loop review runner should not be called")
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
	runner func(*cobra.Command, []string, loopReviewOptions) error,
	lister func() ([]prFixOpenPullRequest, error),
) func() {
	t.Helper()
	previousRunner := prFixLoopReviewRunner
	previousLister := prFixOpenPRLister
	if runner != nil {
		prFixLoopReviewRunner = runner
	}
	if lister != nil {
		prFixOpenPRLister = lister
	}
	return func() {
		prFixLoopReviewRunner = previousRunner
		prFixOpenPRLister = previousLister
	}
}

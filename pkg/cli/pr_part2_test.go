package cli

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

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

func installPRFixInputFakes(
	t *testing.T,
	taskLoader func(string, bool) (dispatchPRInput, bool, error),
	editorLoader func(string, bool, freeTextInputConfig) (dispatchPRInput, bool, error),
) func() {
	t.Helper()
	previousTaskLoader := prFixDispatchTasksLoader
	previousEditorLoader := prFixDispatchInputLoader
	previousRepairResolver := resolvePRRepairContext
	prFixDispatchTasksLoader = taskLoader
	prFixDispatchInputLoader = editorLoader
	resolvePRRepairContext = func(
		_ context.Context,
		_ io.Reader,
		_ io.Writer,
		_ string,
		_ string,
	) (*repairContext, error) {
		return &repairContext{
			Repository:      "jamesonstone/kit",
			PRNumber:        67,
			PRURL:           "https://github.com/jamesonstone/kit/pull/67",
			HeadBranch:      "GH-67",
			ExpectedHeadOID: "remote-head",
			LocalHeadOID:    "local-head",
			WorktreePath:    "/tmp/kit/GH-67",
			PushTarget:      "origin/GH-67",
			ExistingChanges: repairChangesNone,
		}, nil
	}
	return func() {
		prFixDispatchTasksLoader = previousTaskLoader
		prFixDispatchInputLoader = previousEditorLoader
		resolvePRRepairContext = previousRepairResolver
	}
}

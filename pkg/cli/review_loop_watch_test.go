package cli

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestReviewLoopWatchTimingAndQuietWindow(t *testing.T) {
	fakeClock := &fakeReviewLoopClock{now: time.Unix(0, 0)}
	restore := installReviewLoopFakes(t, fakeClock, nil)
	defer restore()

	checkCalls := 0
	reviewLoopRunner = fakeReviewLoopRunner{
		output: func(_ string, _ string, _ ...string) ([]byte, error) {
			return reviewLoopPRPayload("abc123"), nil
		},
		outputAllowError: func(_ string, _ string, _ ...string) ([]byte, error) {
			checkCalls++
			if checkCalls == 1 {
				return []byte(`[{"name":"CodeRabbit","state":"PENDING","bucket":"pending"}]`), nil
			}
			return []byte(`[{"name":"CodeRabbit","state":"SUCCESS","bucket":"pass","completedAt":"2026-06-14T00:00:00Z"}]`), nil
		},
	}

	ctx := reviewLoopPRContext{
		Target:       dispatchPRTarget{Owner: "Patient-Driven-Care", Repo: "cortex", Number: 67},
		RepoFullName: "Patient-Driven-Care/cortex",
		HeadRefOID:   "abc123",
	}
	if err := waitForReviewLoopCodeRabbit(ctx); err != nil {
		t.Fatalf("waitForReviewLoopCodeRabbit() error = %v", err)
	}

	wantSleeps := []time.Duration{
		reviewLoopInitialWait,
		reviewLoopPollEvery,
		reviewLoopPollEvery,
		reviewLoopPollEvery,
		reviewLoopPollEvery,
		reviewLoopPollEvery,
	}
	if fmt.Sprint(fakeClock.sleeps) != fmt.Sprint(wantSleeps) {
		t.Fatalf("sleeps = %v, want %v", fakeClock.sleeps, wantSleeps)
	}
}

func TestReviewLoopWatchRejectsHeadChange(t *testing.T) {
	fakeClock := &fakeReviewLoopClock{now: time.Unix(0, 0)}
	restore := installReviewLoopFakes(t, fakeClock, nil)
	defer restore()

	reviewLoopRunner = fakeReviewLoopRunner{
		output: func(_ string, _ string, _ ...string) ([]byte, error) {
			return reviewLoopPRPayload("newsha"), nil
		},
		outputAllowError: func(_ string, _ string, _ ...string) ([]byte, error) {
			return []byte(`[{"name":"CodeRabbit","state":"SUCCESS","bucket":"pass"}]`), nil
		},
	}

	ctx := reviewLoopPRContext{
		Target:       dispatchPRTarget{Owner: "Patient-Driven-Care", Repo: "cortex", Number: 67},
		RepoFullName: "Patient-Driven-Care/cortex",
		HeadRefOID:   "oldsha",
	}
	err := waitForReviewLoopCodeRabbit(ctx)
	if err == nil || !strings.Contains(err.Error(), "PR head changed") {
		t.Fatalf("expected head-change error, got %v", err)
	}
}

func TestRunReviewLoopDoesNotCallMutatingGitHubCommands(t *testing.T) {
	fakeRunner := fakeReviewLoopRunner{
		output: func(_ string, name string, args ...string) ([]byte, error) {
			if name == "git" || containsReviewLoopMutation(args) {
				return nil, fmt.Errorf("unexpected mutation command: %s %s", name, strings.Join(args, " "))
			}
			return reviewLoopPRPayload("abc123"), nil
		},
		outputAllowError: func(_ string, name string, args ...string) ([]byte, error) {
			if name == "git" || containsReviewLoopMutation(args) {
				return nil, fmt.Errorf("unexpected mutation command: %s %s", name, strings.Join(args, " "))
			}
			return []byte(`[]`), nil
		},
	}
	restore := installReviewLoopFakes(t, realReviewLoopClock{}, fakeRunner)
	defer restore()

	reviewLoopLoadReviewTasks = func(_ string, _ bool) ([]dispatchReviewTask, string, bool, error) {
		return nil, "", false, nil
	}

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	err := runReviewLoop(cmd, reviewLoopOptions{PRRef: "Patient-Driven-Care/cortex#67", MaxSubagents: 1})
	if err != nil {
		t.Fatalf("runReviewLoop() error = %v", err)
	}
}

func TestFetchReviewLoopPRContextParsesMetadata(t *testing.T) {
	restore := installReviewLoopFakes(t, realReviewLoopClock{}, fakeReviewLoopRunner{
		output: func(_ string, name string, args ...string) ([]byte, error) {
			if name != "gh" || !strings.Contains(strings.Join(args, " "), "--repo Patient-Driven-Care/cortex") {
				return nil, fmt.Errorf("unexpected command: %s %v", name, args)
			}
			return reviewLoopPRPayload("abc123"), nil
		},
	})
	defer restore()

	ctx, err := fetchReviewLoopPRContext("Patient-Driven-Care/cortex#67")
	if err != nil {
		t.Fatalf("fetchReviewLoopPRContext() error = %v", err)
	}
	if ctx.HeadRefOID != "abc123" || ctx.RepoFullName != "Patient-Driven-Care/cortex" || len(ctx.IssueHints) != 1 || ctx.IssueHints[0] != "#67" {
		t.Fatalf("unexpected context: %#v", ctx)
	}
}

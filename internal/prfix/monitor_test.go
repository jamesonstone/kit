package prfix

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/registry"
)

type statusSequence struct {
	items []registry.PRFeedbackObservation
	calls int
}

func (sequence *statusSequence) Status(_ context.Context, _ Target, _ string) (registry.PRFeedbackObservation, error) {
	index := sequence.calls
	sequence.calls++
	if index >= len(sequence.items) {
		index = len(sequence.items) - 1
	}
	return sequence.items[index], nil
}

func TestMonitorDistinguishesCompletedAndSkippedSuccess(t *testing.T) {
	contract := monitorContract()
	completed := registry.PRFeedbackObservation{
		ExpectedHead: "abc", ObservedHead: "abc", PullRequestState: "OPEN",
		ContextPresent: true, ContextState: "SUCCESS", Description: "Review completed",
		RateCost: 1, RateLimit: 5000, RateRemaining: 4500, ResetAt: "later",
	}
	for _, test := range []struct {
		name  string
		items []registry.PRFeedbackObservation
		state string
		calls int
	}{
		{"completed after quiet window", []registry.PRFeedbackObservation{completed, completed}, registry.PRFeedbackCompleted, 2},
		{"skipped terminal", []registry.PRFeedbackObservation{{
			ExpectedHead: "abc", ObservedHead: "abc", PullRequestState: "OPEN",
			ContextPresent: true, ContextState: "SUCCESS", Description: "Review skipped: quota exceeded",
			RateCost: 1, RateLimit: 5000, RateRemaining: 4500, ResetAt: "later",
		}}, registry.PRFeedbackSkipped, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			clock := time.Unix(0, 0)
			monitor := Monitor{Now: func() time.Time { return clock }, Sleep: func(_ context.Context, delay time.Duration) error {
				clock = clock.Add(delay)
				return nil
			}}
			client := &statusSequence{items: test.items}
			result := monitor.Await(context.Background(), client, Target{Repository{"acme", "app"}, 1}, "abc", contract, 2*time.Second, &Budget{Limit: 9})
			if result.State != test.state || client.calls != test.calls ||
				result.RateCost != 1 || result.RateRemaining != 4500 || result.RateResetAt != "later" {
				t.Fatalf("result=%#v calls=%d", result, client.calls)
			}
		})
	}
}

func TestMonitorFailsClosedForHeadTimeoutAndBudget(t *testing.T) {
	pending := registry.PRFeedbackObservation{
		ExpectedHead: "abc", ObservedHead: "abc", PullRequestState: "OPEN",
		ContextPresent: true, ContextState: "PENDING", RateLimit: 5000, RateRemaining: 4500,
	}
	contract := monitorContract()
	clock := time.Unix(0, 0)
	monitor := Monitor{Now: func() time.Time { return clock }, Sleep: func(_ context.Context, delay time.Duration) error {
		clock = clock.Add(delay)
		return nil
	}}
	client := &statusSequence{items: []registry.PRFeedbackObservation{pending}}
	result := monitor.Await(context.Background(), client, Target{Repository{"acme", "app"}, 1}, "abc", contract, 2*time.Second, &Budget{Limit: 9})
	if result.State != registry.PRFeedbackTimedOut || client.calls != 3 {
		t.Fatalf("timeout result=%#v calls=%d", result, client.calls)
	}
	client = &statusSequence{items: []registry.PRFeedbackObservation{pending}}
	result = monitor.Await(context.Background(), client, Target{Repository{"acme", "app"}, 1}, "abc", contract, 2*time.Second, &Budget{Limit: 1})
	if result.State != registry.PRFeedbackRateLimited || !strings.Contains(result.Reason, "budget") {
		t.Fatalf("budget result=%#v", result)
	}
	headChanged := pending
	headChanged.ObservedHead = "def"
	client = &statusSequence{items: []registry.PRFeedbackObservation{headChanged}}
	result = monitor.Await(context.Background(), client, Target{Repository{"acme", "app"}, 1}, "abc", contract, 2*time.Second, &Budget{Limit: 9})
	if result.State != registry.PRFeedbackHeadChanged {
		t.Fatalf("head result=%#v", result)
	}
}

func TestStatusParsesRateLimitedResponse(t *testing.T) {
	for _, status := range []int{403, 429} {
		runner := scriptedRunner{run: func(_ string, _ string, _ []string) ([]byte, error) {
			response := fmt.Sprintf("HTTP/2.0 %d Limited\r\nRetry-After: 120\r\n\r\n{\"data\":{\"rateLimit\":{\"resetAt\":\"later\"}}}", status)
			return []byte(response), errors.New("exit status 1")
		}}
		observation, err := NewGitHubClient(runner).Status(context.Background(), Target{Repository{"acme", "app"}, 1}, "abc")
		if err != nil {
			t.Fatal(err)
		}
		if observation.HTTPStatus != status || observation.RetryAfter != "120" || observation.ResetAt != "later" {
			t.Fatalf("observation = %#v", observation)
		}
	}
}

func monitorContract() registry.PRFeedbackContract {
	return registry.PRFeedbackContract{
		CompletedDescription: "Review completed", SkippedDescriptionPrefix: "Review skipped:",
		StatusScheduleSeconds: []int{0, 1, 2}, QuietWindowSeconds: 1,
		MaxTimeoutSeconds: 60, JitterPercent: 0, RateReservePoints: 500, RateReservePercent: 10,
	}
}

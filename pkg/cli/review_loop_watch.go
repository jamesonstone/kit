package cli

import (
	"fmt"
	"strings"
	"time"
)

var reviewLoopClock reviewLoopClockSource = realReviewLoopClock{}

type reviewLoopClockSource interface {
	Now() time.Time
	Sleep(time.Duration)
}

type realReviewLoopClock struct{}

func (realReviewLoopClock) Now() time.Time {
	return time.Now()
}

func (realReviewLoopClock) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

type reviewLoopCheckStatus string

const (
	reviewLoopCheckUnavailable reviewLoopCheckStatus = "unavailable"
	reviewLoopCheckPending     reviewLoopCheckStatus = "pending"
	reviewLoopCheckComplete    reviewLoopCheckStatus = "complete"
)

func waitForReviewLoopCodeRabbit(ctx reviewLoopPRContext) error {
	start := reviewLoopClock.Now()
	deadline := start.Add(reviewLoopTimeout)
	reviewLoopClock.Sleep(reviewLoopInitialWait)

	var quietSince time.Time
	for {
		now := reviewLoopClock.Now()
		if !now.Before(deadline) {
			return fmt.Errorf(
				"timed out waiting for CodeRabbit review completion after %s",
				reviewLoopTimeout,
			)
		}

		current, err := fetchReviewLoopPRContext(reviewLoopTargetRef(ctx.Target))
		if err != nil {
			return err
		}
		if current.HeadRefOID != ctx.HeadRefOID {
			return fmt.Errorf(
				"PR head changed during review-loop watch: started at %s, now %s; rerun review-loop for the current head",
				ctx.HeadRefOID,
				current.HeadRefOID,
			)
		}

		checks, err := fetchReviewLoopChecks(ctx)
		if err != nil {
			return err
		}
		status := summarizeReviewLoopCodeRabbitChecks(checks)
		switch status {
		case reviewLoopCheckUnavailable:
			return fmt.Errorf("no CodeRabbit check status was available for PR #%d", ctx.Target.Number)
		case reviewLoopCheckPending:
			quietSince = time.Time{}
		case reviewLoopCheckComplete:
			if quietSince.IsZero() {
				quietSince = now
			}
			if now.Sub(quietSince) >= reviewLoopQuietWindow {
				return nil
			}
		}

		reviewLoopClock.Sleep(reviewLoopPollEvery)
	}
}

func summarizeReviewLoopCodeRabbitChecks(checks []reviewLoopCheck) reviewLoopCheckStatus {
	found := false
	for _, check := range checks {
		if !isReviewLoopCodeRabbitCheck(check) {
			continue
		}
		found = true
		if isReviewLoopCheckPending(check) {
			return reviewLoopCheckPending
		}
	}
	if !found {
		return reviewLoopCheckUnavailable
	}
	return reviewLoopCheckComplete
}

func isReviewLoopCodeRabbitCheck(check reviewLoopCheck) bool {
	haystack := strings.ToLower(strings.Join([]string{
		check.Name,
		check.Workflow,
		check.Description,
	}, " "))
	return strings.Contains(haystack, "coderabbit") ||
		strings.Contains(haystack, "code rabbit")
}

func isReviewLoopCheckPending(check reviewLoopCheck) bool {
	state := strings.ToLower(strings.TrimSpace(check.State))
	bucket := strings.ToLower(strings.TrimSpace(check.Bucket))
	combined := state + " " + bucket
	for _, token := range []string{
		"pending",
		"queued",
		"in_progress",
		"in progress",
		"running",
		"waiting",
		"requested",
		"expected",
	} {
		if strings.Contains(combined, token) {
			return true
		}
	}
	if state == "" && bucket == "" {
		return true
	}
	return false
}

func reviewLoopTargetRef(target dispatchPRTarget) string {
	return fmt.Sprintf("%s/%s#%d", target.Owner, target.Repo, target.Number)
}

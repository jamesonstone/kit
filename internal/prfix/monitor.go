package prfix

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/jamesonstone/kit/internal/registry"
)

type StatusClient interface {
	Status(context.Context, Target, string) (registry.PRFeedbackObservation, error)
}

type AwaitResult struct {
	SchemaVersion       int    `json:"schema_version"`
	Mode                string `json:"mode"`
	State               string `json:"state"`
	Repository          string `json:"repository"`
	PullRequest         int    `json:"pull_request"`
	ExpectedHead        string `json:"expected_head"`
	ObservedHead        string `json:"observed_head,omitempty"`
	ProviderState       string `json:"provider_state,omitempty"`
	ProviderDescription string `json:"provider_description,omitempty"`
	Reason              string `json:"reason"`
	RetryAt             string `json:"retry_at,omitempty"`
	RateCost            int    `json:"rate_cost,omitempty"`
	RateLimit           int    `json:"rate_limit,omitempty"`
	RateRemaining       int    `json:"rate_remaining,omitempty"`
	RateResetAt         string `json:"rate_reset_at,omitempty"`
	RequestCount        int    `json:"request_count"`
}

type Monitor struct {
	Now   func() time.Time
	Sleep Sleeper
}

func NewMonitor() Monitor {
	return Monitor{Now: time.Now, Sleep: sleepContext}
}

func (monitor Monitor) Await(
	ctx context.Context,
	client StatusClient,
	target Target,
	expectedHead string,
	contract registry.PRFeedbackContract,
	timeout time.Duration,
	budget *Budget,
) AwaitResult {
	result := AwaitResult{
		SchemaVersion: 1, Mode: "await", State: registry.PRFeedbackPending,
		Repository: target.Slug(), PullRequest: target.Number, ExpectedHead: expectedHead,
	}
	if timeout <= 0 || timeout > time.Duration(contract.MaxTimeoutSeconds)*time.Second {
		result.State, result.Reason = registry.PRFeedbackUnavailable, "timeout exceeds workflow bounds"
		return result
	}
	started := monitor.Now()
	for index, seconds := range contract.StatusScheduleSeconds {
		delay := scheduledDelay(started, monitor.Now(), time.Duration(seconds)*time.Second,
			index, len(contract.StatusScheduleSeconds), contract.JitterPercent, expectedHead)
		if delay > 0 {
			if monitor.Now().Add(delay).Sub(started) > timeout {
				break
			}
			if err := monitor.Sleep(ctx, delay); err != nil {
				result.State, result.Reason = registry.PRFeedbackUnavailable, err.Error()
				result.RequestCount = budgetUsed(budget)
				return result
			}
		}
		if err := budget.Take(); err != nil {
			result.State, result.Reason = registry.PRFeedbackRateLimited, err.Error()
			result.RequestCount = budgetUsed(budget)
			return result
		}
		observation, err := client.Status(ctx, target, expectedHead)
		if err != nil {
			result.State, result.Reason = registry.PRFeedbackUnavailable, err.Error()
			result.RequestCount = budgetUsed(budget)
			return result
		}
		applyObservation(&result, observation, contract)
		result.RequestCount = budgetUsed(budget)
		if result.State == registry.PRFeedbackCompleted {
			return monitor.confirmCompletion(ctx, client, target, expectedHead, contract, budget, result)
		}
		if result.State != registry.PRFeedbackPending {
			return result
		}
	}
	result.State = registry.PRFeedbackTimedOut
	result.Reason = "bounded await expired while feedback remained pending"
	result.RequestCount = budgetUsed(budget)
	return result
}

func (monitor Monitor) confirmCompletion(
	ctx context.Context,
	client StatusClient,
	target Target,
	expectedHead string,
	contract registry.PRFeedbackContract,
	budget *Budget,
	result AwaitResult,
) AwaitResult {
	if err := monitor.Sleep(ctx, time.Duration(contract.QuietWindowSeconds)*time.Second); err != nil {
		result.State, result.Reason = registry.PRFeedbackUnavailable, err.Error()
		return result
	}
	if err := budget.Take(); err != nil {
		result.State, result.Reason = registry.PRFeedbackRateLimited, err.Error()
		result.RequestCount = budgetUsed(budget)
		return result
	}
	observation, err := client.Status(ctx, target, expectedHead)
	if err != nil {
		result.State, result.Reason = registry.PRFeedbackUnavailable, err.Error()
		result.RequestCount = budgetUsed(budget)
		return result
	}
	applyObservation(&result, observation, contract)
	result.RequestCount = budgetUsed(budget)
	if result.State != registry.PRFeedbackCompleted {
		result.State = registry.PRFeedbackUnavailable
		result.Reason = "provider completion did not remain stable through the quiet window"
	}
	return result
}

func applyObservation(
	result *AwaitResult,
	observation registry.PRFeedbackObservation,
	contract registry.PRFeedbackContract,
) {
	classified := contract.Classify(observation)
	result.State = classified.State
	result.Reason = classified.Reason
	result.RetryAt = classified.RetryAt
	result.ObservedHead = observation.ObservedHead
	result.ProviderState = observation.ContextState
	result.ProviderDescription = observation.Description
	result.RateCost = observation.RateCost
	result.RateLimit = observation.RateLimit
	result.RateRemaining = observation.RateRemaining
	result.RateResetAt = observation.ResetAt
}

func scheduledDelay(
	started time.Time,
	now time.Time,
	offset time.Duration,
	index int,
	count int,
	jitterPercent int,
	key string,
) time.Duration {
	if index > 0 && index < count-1 && jitterPercent > 0 {
		digest := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", key, index)))
		span := 2*jitterPercent + 1
		percent := int(digest[0])%span - jitterPercent
		offset += offset * time.Duration(percent) / 100
	}
	delay := started.Add(offset).Sub(now)
	if delay < 0 {
		return 0
	}
	return delay
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

package registry

import (
	"fmt"
	"slices"
	"strings"
)

const (
	PRFeedbackPending         = "pending"
	PRFeedbackCompleted       = "completed"
	PRFeedbackSkipped         = "skipped-with-reason"
	PRFeedbackProviderFailure = "provider-failure"
	PRFeedbackHeadChanged     = "head-changed"
	PRFeedbackUnavailable     = "unavailable"
	PRFeedbackTimedOut        = "timed-out"
	PRFeedbackRateLimited     = "rate-limited"
)

type PRFeedbackContract struct {
	SchemaVersion            int                  `yaml:"schema_version"`
	Modes                    []string             `yaml:"modes"`
	WatcherKeyFields         []string             `yaml:"watcher_key_fields"`
	WakeEvents               []string             `yaml:"wake_events"`
	StatusQueryFields        []string             `yaml:"status_query_fields"`
	ProviderContext          string               `yaml:"provider_context"`
	CompletedDescription     string               `yaml:"completed_description"`
	SkippedDescriptionPrefix string               `yaml:"skipped_description_prefix"`
	StatusScheduleSeconds    []int                `yaml:"status_schedule_seconds"`
	QuietWindowSeconds       int                  `yaml:"quiet_window_seconds"`
	DefaultTimeoutSeconds    int                  `yaml:"default_timeout_seconds"`
	MaxTimeoutSeconds        int                  `yaml:"max_timeout_seconds"`
	JitterPercent            int                  `yaml:"jitter_percent"`
	MaxStatusRequestsPerHead int                  `yaml:"max_status_requests_per_head"`
	RequestBudgetPerHead     int                  `yaml:"request_budget_per_head"`
	RateReservePoints        int                  `yaml:"rate_reserve_points"`
	RateReservePercent       int                  `yaml:"rate_reserve_percent"`
	MaxHeadEpochs            int                  `yaml:"max_head_epochs"`
	MaxRepairPasses          int                  `yaml:"max_repair_passes"`
	Collection               PRFeedbackCollection `yaml:"collection"`
}

type PRFeedbackCollection struct {
	PageSize             int      `yaml:"page_size"`
	MaxPages             int      `yaml:"max_pages"`
	Sources              []string `yaml:"sources"`
	ExcludeResolved      bool     `yaml:"exclude_resolved"`
	ExcludeOutdated      bool     `yaml:"exclude_outdated"`
	IncludeHumanThreads  bool     `yaml:"include_human_threads"`
	PromptMarker         string   `yaml:"prompt_marker"`
	TrustedCommentMarker string   `yaml:"trusted_comment_marker"`
	FingerprintFields    []string `yaml:"fingerprint_fields"`
}

type PRFeedbackObservation struct {
	ExpectedHead              string
	ObservedHead              string
	PullRequestState          string
	ContextPresent            bool
	ContextState              string
	Description               string
	ProviderUnavailableReason string
	TimedOut                  bool
	HTTPStatus                int
	RateCost                  int
	RateRemaining             int
	RateLimit                 int
	RetryAfter                string
	ResetAt                   string
}

type PRFeedbackResult struct {
	State   string
	Reason  string
	RetryAt string
}

func ValidatePRFeedbackContract(contract PRFeedbackContract) error {
	if contract.SchemaVersion != 1 {
		return fmt.Errorf("schema_version must be 1")
	}
	for _, mode := range []string{"await", "collect"} {
		if !slices.Contains(contract.Modes, mode) {
			return fmt.Errorf("mode %q is required", mode)
		}
	}
	if !sameStrings(contract.WatcherKeyFields, []string{"repository", "pull_request", "head_sha"}) {
		return fmt.Errorf("watcher key must be repository, pull request, and head SHA")
	}
	for _, event := range []string{"status", "review", "review-comment"} {
		if !slices.Contains(contract.WakeEvents, event) {
			return fmt.Errorf("wake event %q is required", event)
		}
	}
	wantStatusFields := []string{"pull_request_state", "head_ref_oid", "provider_state", "provider_description", "rate_limit"}
	if !sameStrings(contract.StatusQueryFields, wantStatusFields) {
		return fmt.Errorf("compact status query fields are incomplete")
	}
	if contract.ProviderContext == "" || contract.CompletedDescription == "" || contract.SkippedDescriptionPrefix == "" {
		return fmt.Errorf("provider context and completion descriptions are required")
	}
	if err := validateSchedule(contract); err != nil {
		return err
	}
	if contract.RateReservePoints < 1 || contract.RateReservePercent < 1 || contract.RateReservePercent > 100 {
		return fmt.Errorf("rate reserves must be positive and percent must not exceed 100")
	}
	if contract.MaxHeadEpochs < 1 || contract.MaxHeadEpochs > 2 ||
		contract.MaxRepairPasses < 1 || contract.MaxRepairPasses > 2 {
		return fmt.Errorf("head epochs and repair passes must be bounded")
	}
	collection := contract.Collection
	if collection.PageSize < 1 || collection.PageSize > 100 || collection.MaxPages < 1 || collection.MaxPages > 20 {
		return fmt.Errorf("collection pagination must use page size at most 100 and at most 20 pages")
	}
	for _, source := range []string{"review-threads", "requested-change-reviews", "trusted-top-level-comments"} {
		if !slices.Contains(collection.Sources, source) {
			return fmt.Errorf("collection source %q is required", source)
		}
	}
	if !collection.ExcludeResolved || !collection.ExcludeOutdated || !collection.IncludeHumanThreads {
		return fmt.Errorf("collection must include active human threads and exclude resolved or outdated threads")
	}
	if collection.PromptMarker == "" || collection.TrustedCommentMarker == "" || len(collection.FingerprintFields) == 0 {
		return fmt.Errorf("collection markers and fingerprint fields are required")
	}
	if contract.RequestBudgetPerHead < contract.MaxStatusRequestsPerHead+collection.MaxPages {
		return fmt.Errorf("request budget cannot cover status and collection bounds")
	}
	if contract.RequestBudgetPerHead > 32 {
		return fmt.Errorf("request budget exceeds the workflow ceiling")
	}
	return nil
}

func validateSchedule(contract PRFeedbackContract) error {
	schedule := contract.StatusScheduleSeconds
	if len(schedule) == 0 || len(schedule) > 8 || schedule[0] != 0 {
		return fmt.Errorf("status schedule must begin at zero")
	}
	for index := 1; index < len(schedule); index++ {
		if schedule[index] <= schedule[index-1] {
			return fmt.Errorf("status schedule must be strictly increasing")
		}
	}
	if contract.QuietWindowSeconds < 1 || contract.DefaultTimeoutSeconds < schedule[len(schedule)-1] ||
		contract.MaxTimeoutSeconds < contract.DefaultTimeoutSeconds || contract.MaxTimeoutSeconds > 3600 {
		return fmt.Errorf("quiet window and timeout bounds are invalid")
	}
	if contract.JitterPercent < 0 || contract.JitterPercent > 50 {
		return fmt.Errorf("jitter_percent must be between 0 and 50")
	}
	if contract.MaxStatusRequestsPerHead != len(schedule)+1 {
		return fmt.Errorf("status request bound must include one quiet-window confirmation")
	}
	return nil
}

func (contract PRFeedbackContract) Classify(observation PRFeedbackObservation) PRFeedbackResult {
	if observation.HTTPStatus == 403 || observation.HTTPStatus == 429 {
		return PRFeedbackResult{
			State: PRFeedbackRateLimited, Reason: fmt.Sprintf("GitHub returned HTTP %d", observation.HTTPStatus),
			RetryAt: firstNonEmpty(observation.RetryAfter, observation.ResetAt),
		}
	}
	if observation.ExpectedHead != "" && observation.ObservedHead != "" && observation.ExpectedHead != observation.ObservedHead {
		return PRFeedbackResult{State: PRFeedbackHeadChanged, Reason: observation.ObservedHead}
	}
	if observation.TimedOut {
		return PRFeedbackResult{State: PRFeedbackTimedOut, Reason: "bounded await expired while feedback remained pending"}
	}
	if contract.belowRateReserve(observation) {
		return PRFeedbackResult{
			State: PRFeedbackRateLimited, Reason: "GitHub rate budget reached the configured reserve",
			RetryAt: observation.ResetAt,
		}
	}
	if observation.PullRequestState != "" && observation.PullRequestState != "OPEN" {
		return PRFeedbackResult{State: PRFeedbackUnavailable, Reason: "pull request is " + observation.PullRequestState}
	}
	if observation.ProviderUnavailableReason != "" {
		return PRFeedbackResult{State: PRFeedbackUnavailable, Reason: observation.ProviderUnavailableReason}
	}
	if !observation.ContextPresent {
		return PRFeedbackResult{State: PRFeedbackPending, Reason: "provider status context is not present yet"}
	}
	switch observation.ContextState {
	case "EXPECTED", "PENDING":
		return PRFeedbackResult{State: PRFeedbackPending, Reason: observation.Description}
	case "ERROR", "FAILURE":
		return PRFeedbackResult{State: PRFeedbackProviderFailure, Reason: observation.Description}
	case "SUCCESS":
		if observation.Description == contract.CompletedDescription {
			return PRFeedbackResult{State: PRFeedbackCompleted, Reason: observation.Description}
		}
		if strings.HasPrefix(observation.Description, contract.SkippedDescriptionPrefix) {
			return PRFeedbackResult{State: PRFeedbackSkipped, Reason: observation.Description}
		}
		return PRFeedbackResult{State: PRFeedbackUnavailable, Reason: observation.Description}
	default:
		return PRFeedbackResult{State: PRFeedbackUnavailable, Reason: observation.Description}
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (contract PRFeedbackContract) belowRateReserve(observation PRFeedbackObservation) bool {
	if observation.RateLimit <= 0 && observation.RateRemaining <= 0 {
		return false
	}
	percentReserve := (observation.RateLimit*contract.RateReservePercent + 99) / 100
	reserve := max(contract.RateReservePoints, percentReserve)
	return observation.RateRemaining < reserve
}

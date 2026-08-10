package registry

import (
	"os"
	"testing"
)

func TestPRFeedbackWorkflowContract(t *testing.T) {
	content, err := os.ReadFile("../../docs/references/workflows/pr-feedback-repair.md")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ParseMarkdown(string(content))
	if err != nil {
		t.Fatal(err)
	}
	artifact := CatalogArtifact{Kind: KindWorkflow, Slug: "pr-feedback-repair"}
	if err := ValidateDocument(doc, artifact); err != nil {
		t.Fatal(err)
	}
	contract := *doc.Metadata.PRFeedback
	wantSchedule := []int{0, 90, 180, 360, 600, 900, 1200, 1500}
	if !equalInts(contract.StatusScheduleSeconds, wantSchedule) || contract.QuietWindowSeconds != 60 {
		t.Fatalf("status schedule = %v, quiet = %d", contract.StatusScheduleSeconds, contract.QuietWindowSeconds)
	}
	if contract.MaxStatusRequestsPerHead != 9 || contract.RequestBudgetPerHead != 32 ||
		contract.DefaultTimeoutSeconds != 1500 || contract.MaxTimeoutSeconds != 3600 ||
		contract.JitterPercent != 10 || contract.MaxHeadEpochs != 2 || contract.MaxRepairPasses != 2 {
		t.Fatalf("workflow bounds = %#v", contract)
	}
	if !sameStrings(contract.WatcherKeyFields, []string{"repository", "pull_request", "head_sha"}) ||
		!sameStrings(contract.WakeEvents, []string{"status", "review", "review-comment"}) {
		t.Fatalf("watcher dedupe and wakeups = %#v", contract)
	}
	if !contract.Collection.IncludeHumanThreads || !contract.Collection.ExcludeResolved ||
		!contract.Collection.ExcludeOutdated || contract.Collection.PageSize != 50 ||
		contract.Collection.MaxPages != 20 {
		t.Fatalf("collection contract = %#v", contract.Collection)
	}
	wantFingerprint := []string{"kind", "node_id", "path", "line", "author", "url", "body"}
	if !sameStrings(contract.Collection.FingerprintFields, wantFingerprint) ||
		contract.Collection.PromptMarker != "Prompt for AI Agents" ||
		contract.Collection.TrustedCommentMarker != "<!-- kit:pr-feedback -->" {
		t.Fatalf("dedupe and prompt contract = %#v", contract.Collection)
	}
}

func TestPRFeedbackStatusClassification(t *testing.T) {
	contract := loadPRFeedbackContract(t)
	tests := []struct {
		name        string
		observation PRFeedbackObservation
		state       string
		reason      string
	}{
		{"pending", PRFeedbackObservation{ContextPresent: true, ContextState: "PENDING"}, PRFeedbackPending, ""},
		{"completed success", PRFeedbackObservation{ContextPresent: true, ContextState: "SUCCESS", Description: "Review completed"}, PRFeedbackCompleted, "Review completed"},
		{"skipped success", PRFeedbackObservation{ContextPresent: true, ContextState: "SUCCESS", Description: "Review skipped: 524 files exceed the limit of 300"}, PRFeedbackSkipped, "Review skipped: 524 files exceed the limit of 300"},
		{"unsafe unknown success", PRFeedbackObservation{ContextPresent: true, ContextState: "SUCCESS", Description: "OK"}, PRFeedbackUnavailable, "OK"},
		{"provider failure", PRFeedbackObservation{ContextPresent: true, ContextState: "FAILURE", Description: "provider failed"}, PRFeedbackProviderFailure, "provider failed"},
		{"head changed", PRFeedbackObservation{ExpectedHead: "aaa", ObservedHead: "bbb"}, PRFeedbackHeadChanged, "bbb"},
		{"timed out", PRFeedbackObservation{TimedOut: true}, PRFeedbackTimedOut, "bounded await expired while feedback remained pending"},
		{"provider unavailable", PRFeedbackObservation{ProviderUnavailableReason: "not installed"}, PRFeedbackUnavailable, "not installed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := contract.Classify(test.observation)
			if result.State != test.state || result.Reason != test.reason {
				t.Fatalf("result = %#v, want state %q reason %q", result, test.state, test.reason)
			}
		})
	}
}

func TestPRFeedbackRateBudgetAndHTTPBackoff(t *testing.T) {
	contract := loadPRFeedbackContract(t)
	for _, test := range []struct {
		status     int
		retryAfter string
		resetAt    string
		wantRetry  string
	}{
		{403, "120", "later", "120"},
		{429, "", "2026-08-10T20:00:00Z", "2026-08-10T20:00:00Z"},
	} {
		result := contract.Classify(PRFeedbackObservation{
			HTTPStatus: test.status, RetryAfter: test.retryAfter, ResetAt: test.resetAt,
		})
		if result.State != PRFeedbackRateLimited {
			t.Fatalf("HTTP %d state = %s", test.status, result.State)
		}
		if result.RetryAt != test.wantRetry {
			t.Fatalf("HTTP %d retry = %q, want %q", test.status, result.RetryAt, test.wantRetry)
		}
	}
	limited := contract.Classify(PRFeedbackObservation{RateLimit: 5000, RateRemaining: 499})
	if limited.State != PRFeedbackRateLimited {
		t.Fatalf("reserve state = %s", limited.State)
	}
	ready := contract.Classify(PRFeedbackObservation{ContextPresent: true, ContextState: "PENDING", RateLimit: 5000, RateRemaining: 500})
	if ready.State != PRFeedbackPending {
		t.Fatalf("at-reserve state = %s", ready.State)
	}
}

func TestPRFeedbackContractRejectsUnboundedOrIncompleteInput(t *testing.T) {
	contract := loadPRFeedbackContract(t)
	contract.MaxStatusRequestsPerHead = 100
	if err := ValidatePRFeedbackContract(contract); err == nil {
		t.Fatal("unbounded status requests were accepted")
	}
	contract = loadPRFeedbackContract(t)
	contract.Collection.IncludeHumanThreads = false
	if err := ValidatePRFeedbackContract(contract); err == nil {
		t.Fatal("human feedback exclusion was accepted")
	}
	contract = loadPRFeedbackContract(t)
	contract.MaxTimeoutSeconds = 3601
	if err := ValidatePRFeedbackContract(contract); err == nil {
		t.Fatal("timeout above the ceiling was accepted")
	}
}

func TestPRFeedbackWorkflowRejectsCatalogDependencyDrift(t *testing.T) {
	contract := loadPRFeedbackContract(t)
	doc := MarkdownDocument{Metadata: DocumentMetadata{
		Kind: KindWorkflow, Slug: "pr-feedback-repair", Description: "repair", PRFeedback: &contract,
		Dependencies: []string{"ruleset/safety-guardrails"},
	}}
	artifact := CatalogArtifact{
		Kind: KindWorkflow, Slug: "pr-feedback-repair",
		Dependencies: []string{"ruleset/agent-team-orchestration", "ruleset/safety-guardrails"},
	}
	if err := ValidateDocument(doc, artifact); err == nil {
		t.Fatal("catalog and front-matter dependency drift was accepted")
	}
}

func loadPRFeedbackContract(t *testing.T) PRFeedbackContract {
	t.Helper()
	content, err := os.ReadFile("../../docs/references/workflows/pr-feedback-repair.md")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ParseMarkdown(string(content))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Metadata.PRFeedback == nil {
		t.Fatal("pr_feedback contract is missing")
	}
	return *doc.Metadata.PRFeedback
}

func equalInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

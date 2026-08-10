package agentcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/prfix"
	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/pflag"
)

type fakePRGitHub struct {
	pullRequest prfix.PullRequest
	collection  prfix.Collection
	resolved    []string
	status      *registry.PRFeedbackObservation
}

func (fake *fakePRGitHub) CurrentRepository(context.Context, string) (prfix.Repository, error) {
	return prfix.Repository{Owner: "acme", Name: "app"}, nil
}
func (fake *fakePRGitHub) ListOpenPullRequests(context.Context, string, prfix.Repository) ([]prfix.OpenPullRequest, error) {
	return nil, nil
}
func (fake *fakePRGitHub) PullRequest(context.Context, string, prfix.Target) (prfix.PullRequest, error) {
	return fake.pullRequest, nil
}
func (fake *fakePRGitHub) Collect(context.Context, prfix.Target, registry.PRFeedbackContract, prfix.CollectionOptions, *prfix.Budget) (prfix.Collection, error) {
	return fake.collection, nil
}
func (fake *fakePRGitHub) Status(context.Context, prfix.Target, string) (registry.PRFeedbackObservation, error) {
	if fake.status == nil {
		return registry.PRFeedbackObservation{}, fmt.Errorf("unexpected status observation")
	}
	return *fake.status, nil
}
func (fake *fakePRGitHub) ResolveThread(_ context.Context, threadID string) error {
	fake.resolved = append(fake.resolved, threadID)
	return nil
}

type fakePRLane struct{ lane prfix.Lane }

func (fake fakePRLane) Resolve(context.Context, string, prfix.Target, prfix.PullRequest) (prfix.Lane, error) {
	return fake.lane, nil
}

type fakePRState struct{ saves int }

func (fake *fakePRState) Acquire(prfix.Target, string) (func(), error) { return func() {}, nil }
func (fake *fakePRState) Save(prfix.Target, string, prfix.AwaitResult, []prfix.Feedback) error {
	fake.saves++
	return nil
}

func TestPRFixCommandExposesOnlySafeFallbackControls(t *testing.T) {
	command := commandNamed(t, commandNamed(t, NewRoot(), "pr"), "fix")
	var got []string
	command.Flags().VisitAll(func(flag *pflag.Flag) { got = append(got, flag.Name) })
	sort.Strings(got)
	want := []string{"coderabbit", "copy", "edit", "editor", "exclude-dirty", "head", "include-dirty", "max-subagents", "output-only", "pr", "resolve", "thread", "timeout", "trusted-comment-author", "vim", "wait", "yes"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("flags = %v, want %v", got, want)
	}
}

func TestPRSelectionSupportsIndexAndExplicitNumber(t *testing.T) {
	items := []prfix.OpenPullRequest{
		{Number: 14, Title: "First", URL: "url-14", HeadRefName: "GH-14", BaseRefName: "main"},
		{Number: 1, Title: "Second", URL: "url-1", HeadRefName: "GH-1", BaseRefName: "main"},
	}
	for _, test := range []struct{ input, want string }{{"1\n", "url-14"}, {"#1\n", "url-1"}, {"14\n", "url-14"}} {
		output := &bytes.Buffer{}
		got, err := selectOpenPullRequest(strings.NewReader(test.input), output, items)
		if err != nil || got != test.want {
			t.Fatalf("input %q: got=%q err=%v output=%s", test.input, got, err, output)
		}
	}
}

func TestPRFixPromptOutputDoesNotMutateGitHub(t *testing.T) {
	target := prfix.Target{Repository: prfix.Repository{Owner: "acme", Name: "app"}, Number: 9}
	feedback := prfix.Feedback{Kind: "review-thread", ThreadID: "T1", Path: "a.go", Line: 2, Author: "human", URL: "url", Task: "Fix it", Fingerprint: "sha256:one"}
	github := &fakePRGitHub{
		pullRequest: prfix.PullRequest{URL: target.URL(), State: "OPEN", HeadRefName: "GH-9", HeadRefOID: "abc"},
		collection:  prfix.Collection{Items: []prfix.Feedback{feedback}},
	}
	state := &fakePRState{}
	runtime := prFixRuntime{github: github, lane: fakePRLane{lane: cleanPRLane(target)}, state: state}
	output := &bytes.Buffer{}
	command := newPRFixCommand()
	command.SetOut(output)
	command.SetErr(io.Discard)
	previousCopy := clipboardCopyFunc
	clipboardCopyFunc = func(string) error { return fmt.Errorf("clipboard must not be called") }
	t.Cleanup(func() { clipboardCopyFunc = previousCopy })
	err := runPRFixPrompt(command, runtime, "/repo", target, registry.PRFeedbackContract{RequestBudgetPerHead: 10}, prFixOptions{OutputOnly: true, MaxSubagents: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "kit contract resolve --workflow pr-feedback-repair --json") ||
		len(github.resolved) != 0 || state.saves != 1 {
		t.Fatalf("output=%s resolved=%v saves=%d", output, github.resolved, state.saves)
	}
}

func TestPRFixOutputOnlyKeepsNoFeedbackDiagnosticOffStdout(t *testing.T) {
	target := prfix.Target{Repository: prfix.Repository{Owner: "acme", Name: "app"}, Number: 9}
	github := &fakePRGitHub{pullRequest: prfix.PullRequest{
		URL: target.URL(), State: "OPEN", HeadRefName: "GH-9", HeadRefOID: "abc",
	}}
	runtime := prFixRuntime{github: github, lane: fakePRLane{lane: cleanPRLane(target)}, state: &fakePRState{}}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	command := newPRFixCommand()
	command.SetOut(stdout)
	command.SetErr(stderr)
	err := runPRFixPrompt(command, runtime, "/repo", target, registry.PRFeedbackContract{RequestBudgetPerHead: 10}, prFixOptions{OutputOnly: true, MaxSubagents: 3})
	if err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "No actionable") {
		t.Fatalf("stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestPRFixWaitFailureEmitsStructuredJSONToStdout(t *testing.T) {
	target := prfix.Target{Repository: prfix.Repository{Owner: "acme", Name: "app"}, Number: 9}
	status := registry.PRFeedbackObservation{
		ExpectedHead: "abc", ObservedHead: "abc", PullRequestState: "OPEN",
		ContextPresent: true, ContextState: "SUCCESS", Description: "Review skipped: quota",
		RateCost: 1, RateLimit: 5000, RateRemaining: 4500,
	}
	github := &fakePRGitHub{
		pullRequest: prfix.PullRequest{URL: target.URL(), State: "OPEN", HeadRefName: "GH-9", HeadRefOID: "abc"},
		status:      &status,
	}
	runtime := prFixRuntime{github: github, lane: fakePRLane{lane: cleanPRLane(target)}, monitor: prfix.NewMonitor(), state: &fakePRState{}}
	stdout := &bytes.Buffer{}
	command := newPRFixCommand()
	command.SetOut(stdout)
	command.SetErr(io.Discard)
	contract := registry.PRFeedbackContract{
		CompletedDescription: "Review completed", SkippedDescriptionPrefix: "Review skipped:",
		StatusScheduleSeconds: []int{0}, MaxTimeoutSeconds: 60, RequestBudgetPerHead: 2,
	}
	err := runPRFixPrompt(command, runtime, "/repo", target, contract, prFixOptions{Wait: true, Timeout: time.Second, OutputOnly: true, MaxSubagents: 3})
	if err == nil || !strings.Contains(stdout.String(), `"state": "skipped-with-reason"`) ||
		!strings.Contains(stdout.String(), `"rate_cost": 1`) {
		t.Fatalf("err=%v stdout=%q", err, stdout)
	}
}

func TestPRFixResolutionRequiresCurrentActiveThreadBeforeMutation(t *testing.T) {
	target := prfix.Target{Repository: prfix.Repository{Owner: "acme", Name: "app"}, Number: 9}
	github := &fakePRGitHub{
		pullRequest: prfix.PullRequest{URL: target.URL(), State: "OPEN", HeadRefName: "GH-9", HeadRefOID: "abc"},
		collection:  prfix.Collection{Items: []prfix.Feedback{{ThreadID: "T1"}}},
	}
	runtime := prFixRuntime{github: github, lane: fakePRLane{lane: cleanPRLane(target)}, state: &fakePRState{}}
	command := newPRFixCommand()
	command.SetOut(io.Discard)
	options := prFixOptions{Resolve: true, Yes: true, Head: "abc", Threads: []string{"missing"}, MaxSubagents: 3}
	if err := runPRFixResolution(command, runtime, "/repo", target, registry.PRFeedbackContract{RequestBudgetPerHead: 10}, options); err == nil {
		t.Fatal("missing active thread was accepted")
	}
	if len(github.resolved) != 0 {
		t.Fatalf("mutated GitHub for invalid input: %v", github.resolved)
	}
	options.Threads = []string{"T1"}
	if err := runPRFixResolution(command, runtime, "/repo", target, registry.PRFeedbackContract{RequestBudgetPerHead: 10}, options); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(github.resolved, []string{"T1"}) {
		t.Fatalf("resolved = %v", github.resolved)
	}
}

func TestPRFixPromptControlsPreserveClipboardFirstCompatibility(t *testing.T) {
	previousCopy := clipboardCopyFunc
	var copied string
	clipboardCopyFunc = func(content string) error {
		copied = content
		return nil
	}
	t.Cleanup(func() { clipboardCopyFunc = previousCopy })
	command := newPRFixCommand()
	stdout := &bytes.Buffer{}
	command.SetOut(stdout)
	if err := outputPRFixPrompt(command, "prompt", prFixOptions{}); err != nil {
		t.Fatal(err)
	}
	if copied != "prompt" || !strings.Contains(stdout.String(), "copied to clipboard") {
		t.Fatalf("copied=%q stdout=%q", copied, stdout)
	}
	copied, stdout = "", &bytes.Buffer{}
	command.SetOut(stdout)
	if err := outputPRFixPrompt(command, "raw prompt", prFixOptions{OutputOnly: true, Copy: true}); err != nil {
		t.Fatal(err)
	}
	if copied != "raw prompt" || stdout.String() != "raw prompt" {
		t.Fatalf("copied=%q stdout=%q", copied, stdout)
	}
}

func TestPRFixExplicitEditorDoesNotUseAShell(t *testing.T) {
	command := newPRFixCommand()
	command.SetContext(context.Background())
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	got, err := editPRFixTasks(command, prFixOptions{Editor: "true"}, "finding")
	if err != nil || got != "finding" {
		t.Fatalf("edited=%q err=%v", got, err)
	}
}

func cleanPRLane(target prfix.Target) prfix.Lane {
	return prfix.Lane{Repository: target.Slug(), PRNumber: target.Number, PRURL: target.URL(), HeadBranch: "GH-9", ExpectedHead: "abc", LocalHead: "abc", WorktreePath: "/repo", PushTarget: "origin/GH-9", DirtyOwnership: "none"}
}

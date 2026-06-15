package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCollectDispatchReviewResolutionCandidates(t *testing.T) {
	threads := []dispatchGitHubReviewThread{
		reviewThreadFixture("internal/app.go", 12, false, false, "coderabbitai", coderabbitCommentBody("Fix app routing."), "https://example.com/1"),
		reviewThreadFixture("internal/human.go", 4, false, false, "octocat", "Please update docs.", "https://example.com/2"),
		reviewThreadFixture("internal/old.go", 3, false, true, "coderabbitai", "Skip outdated.", "https://example.com/3"),
		reviewThreadFixture("internal/done.go", 7, true, false, "coderabbitai", "Skip resolved.", "https://example.com/4"),
	}
	threads[0].ID = "thread-1"
	threads[1].ID = "thread-2"
	threads[2].ID = "thread-3"
	threads[3].ID = "thread-4"

	all := collectDispatchReviewResolutionCandidates(threads, false)
	if len(all) != 2 {
		t.Fatalf("expected 2 all-author candidates, got %#v", all)
	}
	if all[0].ThreadID != "thread-1" || all[1].ThreadID != "thread-2" {
		t.Fatalf("unexpected candidates: %#v", all)
	}
	if all[0].Body != "Fix app routing." {
		t.Fatalf("candidate body = %q, want cleaned review summary", all[0].Body)
	}

	coderabbit := collectDispatchReviewResolutionCandidates(threads, true)
	if len(coderabbit) != 1 || coderabbit[0].ThreadID != "thread-1" {
		t.Fatalf("expected only CodeRabbit candidate, got %#v", coderabbit)
	}
}

func TestRunDispatchPRResolveRequiresExplicitYes(t *testing.T) {
	previousPR := dispatchPR
	previousResolve := dispatchResolve
	previousYes := dispatchYes
	defer func() {
		dispatchPR = previousPR
		dispatchResolve = previousResolve
		dispatchYes = previousYes
	}()

	dispatchPR = "Patient-Driven-Care/cortex#67"
	dispatchResolve = true
	dispatchYes = false

	err := runDispatchPRResolve(&cobra.Command{})
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("expected explicit --yes guard, got %v", err)
	}
}

func TestResolveDispatchReviewThreadUsesGraphQLMutation(t *testing.T) {
	previousExec := execCommand
	defer func() { execCommand = previousExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		cmdArgs := append([]string{"-test.run=TestDispatchReviewThreadResolverHelperProcess", "--", name}, args...)
		cmd := exec.Command(os.Args[0], cmdArgs...)
		cmd.Env = append(os.Environ(), "KIT_TEST_DISPATCH_RESOLVE_HELPER=1")
		return cmd
	}

	if err := resolveDispatchReviewThread("thread-1"); err != nil {
		t.Fatalf("resolveDispatchReviewThread() error = %v", err)
	}
}

func TestDispatchReviewThreadResolverHelperProcess(t *testing.T) {
	if os.Getenv("KIT_TEST_DISPATCH_RESOLVE_HELPER") != "1" {
		return
	}

	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "missing helper command args")
		os.Exit(2)
	}
	args = args[1:]
	joined := strings.Join(args, " ")
	if args[0] != "gh" ||
		!strings.Contains(joined, "api graphql") ||
		!strings.Contains(joined, "resolveReviewThread") ||
		!strings.Contains(joined, "threadId=thread-1") {
		fmt.Fprintf(os.Stderr, "unexpected helper args: %v\n", args)
		os.Exit(2)
	}

	fmt.Print(`{"data":{"resolveReviewThread":{"thread":{"id":"thread-1","isResolved":true}}}}`)
	os.Exit(0)
}

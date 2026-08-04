package worktree

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestListShowsOpenPullRequestNumbers(t *testing.T) {
	fixture := newGitFixture(t)
	runGit(t, fixture.primary, "branch", "--track", "GH-100", "origin/main")
	runWT(t, fixture.app, fixture.primary, "add", "GH-100", "--no-link-env")
	fixture.out.Reset()
	fixture.app.resolveListPRs = func(context.Context, string) listPRLookup {
		return successfulListPRLookup(map[string]listPRAnnotation{
			"GH-100": {numbers: "101", titles: "Add PR titles"},
		})
	}

	runWT(t, fixture.app, fixture.primary, "list", "--plain")
	output := fixture.out.String()
	if !strings.HasPrefix(output, "STATE\tHEAD\tPR#\tLAST UPDATED\tPATH\n") {
		t.Fatalf("list header:\n%s", output)
	}
	if strings.Contains(strings.SplitN(output, "\n", 2)[0], "TITLE") {
		t.Fatalf("plain list unexpectedly changed columns:\n%s", output)
	}
	if !strings.Contains(output, "\nclean\tmain\t-\t") {
		t.Fatalf("main row should show no open pull request:\n%s", output)
	}
	if !strings.Contains(output, "\nclean\tGH-100\t101\t") {
		t.Fatalf("issue row should show pull request 101:\n%s", output)
	}
}

func TestPopulateListPullRequestsAddsTitlesAndMirrorsFailures(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		lookup     listPRLookup
		wantNumber string
		wantTitle  string
	}{
		{
			name: "matching PR",
			lookup: successfulListPRLookup(map[string]listPRAnnotation{
				"GH-100": {numbers: "101", titles: "Add PR titles"},
			}),
			wantNumber: "101",
			wantTitle:  "Add PR titles",
		},
		{
			name:       "no PR",
			lookup:     successfulListPRLookup(nil),
			wantNumber: listPRNoneMarker,
			wantTitle:  listPRNoneMarker,
		},
		{
			name:       "lookup failure",
			lookup:     failedListPRLookup(listPRTimeoutMarker),
			wantNumber: listPRTimeoutMarker,
			wantTitle:  listPRTimeoutMarker,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			app := NewApp(io.Discard, io.Discard)
			app.resolveListPRs = func(context.Context, string) listPRLookup {
				return test.lookup
			}
			entries := []worktreeEntry{{branch: "GH-100"}}
			app.populateListPullRequests(context.Background(), t.TempDir(), entries)
			if got := entries[0].prText; got != test.wantNumber {
				t.Fatalf("PR# = %q, want %q", got, test.wantNumber)
			}
			if got := entries[0].prTitle; got != test.wantTitle {
				t.Fatalf("TITLE = %q, want %q", got, test.wantTitle)
			}
		})
	}
}

func TestListPullRequestFailuresRemainNonBlocking(t *testing.T) {
	for _, marker := range []string{
		listPRMissingGHMarker,
		listPRRateLimitMarker,
		listPRTimeoutMarker,
		listPRUnknownMarker,
	} {
		t.Run(marker, func(t *testing.T) {
			fixture := newGitFixture(t)
			fixture.app.resolveListPRs = func(context.Context, string) listPRLookup {
				return failedListPRLookup(marker)
			}

			runWT(t, fixture.app, fixture.primary, "list", "--plain")
			if !strings.Contains(fixture.out.String(), "\nclean\tmain\t"+marker+"\t") {
				t.Fatalf("list did not preserve failure marker %q:\n%s", marker, fixture.out.String())
			}
		})
	}
}

func TestResolveListPullRequestsUsesOneExactBranchBatch(t *testing.T) {
	app := NewApp(io.Discard, io.Discard)
	app.lookPath = func(name string) (string, error) {
		if name != "gh" {
			return "", fmt.Errorf("unexpected executable %s", name)
		}
		return "/usr/bin/gh", nil
	}
	var calls [][]string
	app.run = func(
		_ context.Context,
		_ string,
		name string,
		args ...string,
	) ([]byte, error) {
		if name != "gh" {
			return nil, fmt.Errorf("unexpected command %s", name)
		}
		calls = append(calls, append([]string(nil), args...))
		return json.Marshal([]listPullRequest{
			{Number: 12, HeadRefName: "GH-100", Title: "Second title"},
			{Number: 7, HeadRefName: "GH-100", Title: "First title"},
			{Number: 13, HeadRefName: "empty-title"},
			{Number: 5, HeadRefName: "fork-topic", IsCrossRepository: true, Title: "Fork"},
			{Number: 0, HeadRefName: "invalid"},
		})
	}

	lookup := app.resolveListPullRequests(context.Background(), t.TempDir())
	wantAnnotations := map[string]listPRAnnotation{
		"GH-100":      {numbers: "7,12", titles: "First title | Second title"},
		"empty-title": {numbers: "13", titles: listPRUnknownMarker},
	}
	if !reflect.DeepEqual(lookup.byBranch, wantAnnotations) {
		t.Fatalf("pull requests = %#v", lookup.byBranch)
	}
	if lookup.fallback != listPRNoneMarker {
		t.Fatalf("fallback = %q, want %q", lookup.fallback, listPRNoneMarker)
	}
	wantArgs := []string{
		"pr", "list", "--state", "open", "--limit", "1000",
		"--json", "number,headRefName,isCrossRepository,title",
	}
	if !reflect.DeepEqual(calls, [][]string{wantArgs}) {
		t.Fatalf("gh calls = %#v, want one batch %#v", calls, wantArgs)
	}
}

func TestHelpDocumentsListPullRequestMarkers(t *testing.T) {
	fixture := newGitFixture(t)
	runWT(t, fixture.app, fixture.primary, "help")
	output := fixture.out.String()
	for _, want := range []string{"List PR# markers:", "NG gh unavailable", "RL rate limited", "TO timed out", "?? other failure", "Interactive TITLE:", "truncates before PATH", "plain output is unchanged"} {
		if !strings.Contains(output, want) {
			t.Fatalf("help is missing %q:\n%s", want, output)
		}
	}
}

func TestResolveListPullRequestsClassifiesFailures(t *testing.T) {
	for _, test := range []struct {
		name      string
		output    string
		err       error
		want      string
		missing   bool
		malformed bool
	}{
		{name: "missing gh", want: listPRMissingGHMarker, missing: true},
		{name: "rate limited", output: "GraphQL: API rate limit exceeded", err: errors.New("exit status 1"), want: listPRRateLimitMarker},
		{name: "generic failure", output: "authentication failed", err: errors.New("exit status 1"), want: listPRUnknownMarker},
		{name: "malformed response", output: "{", want: listPRUnknownMarker, malformed: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			app := NewApp(io.Discard, io.Discard)
			if test.missing {
				app.lookPath = func(string) (string, error) {
					return "", errors.New("not found")
				}
				app.run = func(context.Context, string, string, ...string) ([]byte, error) {
					t.Fatal("gh should not run when it is unavailable")
					return nil, nil
				}
			} else {
				app.lookPath = func(string) (string, error) {
					return "/usr/bin/gh", nil
				}
				app.run = func(context.Context, string, string, ...string) ([]byte, error) {
					return []byte(test.output), test.err
				}
			}

			lookup := app.resolveListPullRequests(context.Background(), t.TempDir())
			if lookup.fallback != test.want {
				t.Fatalf("fallback = %q, want %q", lookup.fallback, test.want)
			}
			if test.malformed && lookup.byBranch != nil {
				t.Fatalf("malformed lookup unexpectedly returned branches: %#v", lookup.byBranch)
			}
		})
	}
}

func TestResolveListPullRequestsTimesOut(t *testing.T) {
	app := NewApp(io.Discard, io.Discard)
	app.lookPath = func(string) (string, error) {
		return "/usr/bin/gh", nil
	}
	app.listPRTimeout = 10 * time.Millisecond
	app.run = func(
		ctx context.Context,
		_ string,
		_ string,
		_ ...string,
	) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	started := time.Now()
	lookup := app.resolveListPullRequests(context.Background(), t.TempDir())
	if lookup.fallback != listPRTimeoutMarker {
		t.Fatalf("fallback = %q, want %q", lookup.fallback, listPRTimeoutMarker)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("timeout lookup took %s", elapsed)
	}
}

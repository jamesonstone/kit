package prfix

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/registry"
)

type scriptedRunner struct {
	run func(string, string, []string) ([]byte, error)
}

func (runner scriptedRunner) Run(_ context.Context, directory, name string, args ...string) ([]byte, error) {
	return runner.run(directory, name, args)
}

func TestCollectPaginatesThreadsAndIncludesLateHumanSources(t *testing.T) {
	threadPages := 0
	runner := scriptedRunner{run: func(_ string, name string, args []string) ([]byte, error) {
		if name != "gh" {
			return nil, fmt.Errorf("unexpected command %s", name)
		}
		query := argumentValue(args, "query")
		switch {
		case strings.Contains(query, "reviewThreads"):
			threadPages++
			if argumentValue(args, "cursor") == "" {
				return []byte(threadPageJSON("T1", "human.go", "octocat", "Human finding", true, "NEXT")), nil
			}
			return []byte(threadPageJSON("T2", "bot.go", "coderabbitai", coderabbitBody("Bot finding"), false, "")), nil
		case strings.Contains(query, "reviews(first"):
			return []byte(`{"data":{"repository":{"pullRequest":{"reviews":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[{"id":"R1","state":"CHANGES_REQUESTED","body":"Late review body","url":"review-url","author":{"login":"reviewer"}}]}}}}}`), nil
		case strings.Contains(query, "comments(first"):
			return []byte(`{"data":{"repository":{"pullRequest":{"comments":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[{"id":"C1","body":"<!-- kit:pr-feedback --> Late marked comment","url":"comment-url","authorAssociation":"MEMBER","author":{"login":"maintainer"}}]}}}}}`), nil
		default:
			return nil, fmt.Errorf("unexpected query: %s", query)
		}
	}}
	client := NewGitHubClient(runner)
	contract := collectionContract()
	budget := &Budget{Limit: 10}
	collection, err := client.Collect(context.Background(), Target{Repository{"acme", "app"}, 7}, contract, CollectionOptions{}, budget)
	if err != nil {
		t.Fatal(err)
	}
	if threadPages != 2 || budget.Used != 4 {
		t.Fatalf("thread pages=%d requests=%d", threadPages, budget.Used)
	}
	if len(collection.Items) != 4 {
		t.Fatalf("items = %#v", collection.Items)
	}
	joined := RenderFeedback(collection.Items)
	for _, want := range []string{"Human finding", "Bot finding", "Late review body", "Late marked comment"} {
		if !strings.Contains(joined, want) {
			t.Errorf("late one-shot collection missing %q:\n%s", want, joined)
		}
	}
}

func TestCollectPaginatesThreadCommentsAndHonorsBudget(t *testing.T) {
	commentPage := false
	runner := scriptedRunner{run: func(_ string, _ string, args []string) ([]byte, error) {
		query := argumentValue(args, "query")
		switch {
		case strings.Contains(query, "reviewThreads"):
			return []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[{"id":"T1","isResolved":false,"isOutdated":false,"path":"a.go","line":3,"startLine":0,"comments":{"pageInfo":{"hasNextPage":true,"endCursor":"C1"},"nodes":[{"id":"A","body":"Human root","url":"u","author":{"login":"human"}}]}}]}}}}}`), nil
		case strings.Contains(query, "node(id:$threadId)"):
			commentPage = true
			return []byte(`{"data":{"node":{"comments":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[{"id":"B","body":"reply","url":"v","author":{"login":"human"}}]}}}}`), nil
		case strings.Contains(query, "reviews(first"):
			return emptyReviews(), nil
		case strings.Contains(query, "comments(first"):
			return emptyComments(), nil
		default:
			return nil, fmt.Errorf("unexpected query")
		}
	}}
	client := NewGitHubClient(runner)
	budget := &Budget{Limit: 4}
	if _, err := client.Collect(context.Background(), Target{Repository{"acme", "app"}, 7}, collectionContract(), CollectionOptions{}, budget); err != nil {
		t.Fatal(err)
	}
	if !commentPage || budget.Used != 4 {
		t.Fatalf("commentPage=%t requests=%d", commentPage, budget.Used)
	}
	budget = &Budget{Limit: 3}
	if _, err := client.Collect(context.Background(), Target{Repository{"acme", "app"}, 7}, collectionContract(), CollectionOptions{}, budget); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("budget error = %v", err)
	}
}

func collectionContract() registry.PRFeedbackContract {
	return registry.PRFeedbackContract{RequestBudgetPerHead: 32, Collection: registry.PRFeedbackCollection{
		PageSize: 2, MaxPages: 3, TrustedCommentMarker: "<!-- kit:pr-feedback -->",
	}}
}

func argumentValue(args []string, key string) string {
	for index, value := range args {
		if (value == "-f" || value == "-F") && index+1 < len(args) {
			name, result, found := strings.Cut(args[index+1], "=")
			if found && name == key {
				return result
			}
		}
	}
	return ""
}

func threadPageJSON(id, path, author, body string, next bool, cursor string) string {
	return fmt.Sprintf(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":%t,"endCursor":%q},"nodes":[{"id":%q,"isResolved":false,"isOutdated":false,"path":%q,"line":3,"startLine":0,"comments":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[{"id":%q,"body":%q,"url":"url","author":{"login":%q}}]}}]}}}}}`,
		next, cursor, id, path, id+"C", body, author)
}

func emptyReviews() []byte {
	return []byte(`{"data":{"repository":{"pullRequest":{"reviews":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[]}}}}}`)
}

func emptyComments() []byte {
	return []byte(`{"data":{"repository":{"pullRequest":{"comments":{"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[]}}}}}`)
}

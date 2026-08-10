package prfix

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const resolveThreadMutation = `mutation($threadId:ID!){
  resolveReviewThread(input:{threadId:$threadId}){thread{id isResolved}}
}`

type resolveThreadResponse struct {
	Data struct {
		ResolveReviewThread struct {
			Thread struct {
				ID         string `json:"id"`
				IsResolved bool   `json:"isResolved"`
			} `json:"thread"`
		} `json:"resolveReviewThread"`
	} `json:"data"`
}

func ValidateResolution(head string, pullRequest PullRequest, requested []string, feedback []Feedback) ([]string, error) {
	if strings.TrimSpace(head) == "" {
		return nil, fmt.Errorf("--resolve requires --head with the verified pushed head SHA")
	}
	if head != pullRequest.HeadRefOID {
		return nil, fmt.Errorf("verified head %s does not match current PR head %s", head, pullRequest.HeadRefOID)
	}
	if len(requested) == 0 {
		return nil, fmt.Errorf("--resolve requires at least one explicit --thread ID")
	}
	active := map[string]bool{}
	for _, item := range feedback {
		if item.ThreadID != "" {
			active[item.ThreadID] = true
		}
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(requested))
	for _, id := range requested {
		id = strings.TrimSpace(id)
		if id == "" || !active[id] {
			return nil, fmt.Errorf("review thread %q is not current, unresolved, non-outdated feedback", id)
		}
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	sort.Strings(result)
	return result, nil
}

func (client *GitHubClient) ResolveThread(ctx context.Context, threadID string) error {
	args := graphQLArgs(resolveThreadMutation, map[string]string{"threadId": threadID})
	output, err := client.runner.Run(ctx, "", "gh", args...)
	if err != nil {
		return fmt.Errorf("resolve review thread %s: %w", threadID, err)
	}
	if err := validateGraphQL(output); err != nil {
		return err
	}
	var response resolveThreadResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return fmt.Errorf("decode review-thread resolution: %w", err)
	}
	thread := response.Data.ResolveReviewThread.Thread
	if thread.ID != threadID || !thread.IsResolved {
		return fmt.Errorf("GitHub did not confirm review thread %s was resolved", threadID)
	}
	return nil
}

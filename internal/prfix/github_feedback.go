package prfix

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jamesonstone/kit/internal/registry"
)

func (client *GitHubClient) Collect(
	ctx context.Context,
	target Target,
	contract registry.PRFeedbackContract,
	options CollectionOptions,
	budget *Budget,
) (Collection, error) {
	threads, err := client.collectThreads(ctx, target, contract, budget)
	if err != nil {
		return Collection{}, err
	}
	reviews, err := client.collectReviews(ctx, target, contract, budget)
	if err != nil {
		return Collection{}, err
	}
	comments, err := client.collectIssueComments(ctx, target, contract, budget)
	if err != nil {
		return Collection{}, err
	}
	items, active := normalizeFeedback(threads, reviews, comments, contract, options)
	return Collection{Items: items, ActiveCount: active, RequestCount: budgetUsed(budget)}, nil
}

func budgetUsed(budget *Budget) int {
	if budget == nil {
		return 0
	}
	return budget.Used
}

func (client *GitHubClient) collectThreads(
	ctx context.Context,
	target Target,
	contract registry.PRFeedbackContract,
	budget *Budget,
) ([]reviewThread, error) {
	var result []reviewThread
	cursor := ""
	for page := 0; page < contract.Collection.MaxPages; page++ {
		var response threadPageResponse
		if err := client.feedbackPage(ctx, threadQuery(contract.Collection.PageSize), target, cursor, budget, &response); err != nil {
			return nil, fmt.Errorf("fetch review threads: %w", err)
		}
		connection := response.Data.Repository.PullRequest.ReviewThreads
		for index := range connection.Nodes {
			if err := client.collectThreadComments(ctx, &connection.Nodes[index], contract, budget); err != nil {
				return nil, err
			}
		}
		result = append(result, connection.Nodes...)
		if !connection.PageInfo.HasNextPage {
			return result, nil
		}
		if connection.PageInfo.EndCursor == "" {
			return nil, fmt.Errorf("review thread pagination returned no cursor")
		}
		cursor = connection.PageInfo.EndCursor
	}
	return nil, fmt.Errorf("review thread pagination exceeded %d pages", contract.Collection.MaxPages)
}

func (client *GitHubClient) collectThreadComments(
	ctx context.Context,
	thread *reviewThread,
	contract registry.PRFeedbackContract,
	budget *Budget,
) error {
	cursor := thread.Comments.PageInfo.EndCursor
	for page := 1; thread.Comments.PageInfo.HasNextPage && page < contract.Collection.MaxPages; page++ {
		var response threadCommentPageResponse
		args := graphQLArgs(threadCommentQuery(contract.Collection.PageSize), map[string]string{
			"threadId": thread.ID, "cursor": cursor,
		})
		if err := client.graphQL(ctx, args, budget, &response); err != nil {
			return fmt.Errorf("fetch comments for review thread %s: %w", thread.ID, err)
		}
		connection := response.Data.Node.Comments
		thread.Comments.Nodes = append(thread.Comments.Nodes, connection.Nodes...)
		thread.Comments.PageInfo = connection.PageInfo
		cursor = connection.PageInfo.EndCursor
		if connection.PageInfo.HasNextPage && cursor == "" {
			return fmt.Errorf("review thread %s comment pagination returned no cursor", thread.ID)
		}
	}
	if thread.Comments.PageInfo.HasNextPage {
		return fmt.Errorf("review thread %s comments exceeded %d pages", thread.ID, contract.Collection.MaxPages)
	}
	return nil
}

func (client *GitHubClient) collectReviews(
	ctx context.Context,
	target Target,
	contract registry.PRFeedbackContract,
	budget *Budget,
) ([]pullRequestReview, error) {
	var result []pullRequestReview
	cursor := ""
	for page := 0; page < contract.Collection.MaxPages; page++ {
		var response reviewPageResponse
		if err := client.feedbackPage(ctx, reviewQuery(contract.Collection.PageSize), target, cursor, budget, &response); err != nil {
			return nil, fmt.Errorf("fetch requested-change reviews: %w", err)
		}
		connection := response.Data.Repository.PullRequest.Reviews
		result = append(result, connection.Nodes...)
		if !connection.PageInfo.HasNextPage {
			return result, nil
		}
		if connection.PageInfo.EndCursor == "" {
			return nil, fmt.Errorf("review pagination returned no cursor")
		}
		cursor = connection.PageInfo.EndCursor
	}
	return nil, fmt.Errorf("review pagination exceeded %d pages", contract.Collection.MaxPages)
}

func (client *GitHubClient) collectIssueComments(
	ctx context.Context,
	target Target,
	contract registry.PRFeedbackContract,
	budget *Budget,
) ([]issueComment, error) {
	var result []issueComment
	cursor := ""
	for page := 0; page < contract.Collection.MaxPages; page++ {
		var response issueCommentPageResponse
		if err := client.feedbackPage(ctx, issueCommentQuery(contract.Collection.PageSize), target, cursor, budget, &response); err != nil {
			return nil, fmt.Errorf("fetch trusted pull request comments: %w", err)
		}
		connection := response.Data.Repository.PullRequest.Comments
		result = append(result, connection.Nodes...)
		if !connection.PageInfo.HasNextPage {
			return result, nil
		}
		if connection.PageInfo.EndCursor == "" {
			return nil, fmt.Errorf("pull request comment pagination returned no cursor")
		}
		cursor = connection.PageInfo.EndCursor
	}
	return nil, fmt.Errorf("pull request comment pagination exceeded %d pages", contract.Collection.MaxPages)
}

func (client *GitHubClient) feedbackPage(
	ctx context.Context,
	query string,
	target Target,
	cursor string,
	budget *Budget,
	response interface{},
) error {
	variables := map[string]string{"owner": target.Owner, "name": target.Name, "number": fmt.Sprint(target.Number)}
	if cursor != "" {
		variables["cursor"] = cursor
	}
	return client.graphQL(ctx, graphQLArgs(query, variables), budget, response)
}

func (client *GitHubClient) graphQL(ctx context.Context, args []string, budget *Budget, response interface{}) error {
	if err := budget.Take(); err != nil {
		return err
	}
	output, err := client.runner.Run(ctx, "", "gh", args...)
	if err != nil {
		return err
	}
	if err := validateGraphQL(output); err != nil {
		return err
	}
	if err := json.Unmarshal(output, response); err != nil {
		return fmt.Errorf("decode GitHub feedback response: %w", err)
	}
	return nil
}

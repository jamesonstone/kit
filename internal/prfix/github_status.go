package prfix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/registry"
)

const statusQuery = `query($owner:String!,$name:String!,$number:Int!){
  repository(owner:$owner,name:$name){pullRequest(number:$number){state headRefOid commits(last:1){nodes{commit{statusCheckRollup{contexts(first:100){nodes{__typename ... on StatusContext{context state description}}}}}}}}}
  rateLimit{cost limit remaining resetAt}
}`

type statusResponse struct {
	Data struct {
		Repository struct {
			PullRequest *struct {
				State      string `json:"state"`
				HeadRefOID string `json:"headRefOid"`
				Commits    struct {
					Nodes []struct {
						Commit struct {
							Status struct {
								Contexts struct {
									Nodes []struct {
										Type        string `json:"__typename"`
										Context     string `json:"context"`
										State       string `json:"state"`
										Description string `json:"description"`
									} `json:"nodes"`
								} `json:"contexts"`
							} `json:"statusCheckRollup"`
						} `json:"commit"`
					} `json:"nodes"`
				} `json:"commits"`
			} `json:"pullRequest"`
		} `json:"repository"`
		RateLimit struct {
			Cost      int    `json:"cost"`
			Limit     int    `json:"limit"`
			Remaining int    `json:"remaining"`
			ResetAt   string `json:"resetAt"`
		} `json:"rateLimit"`
	} `json:"data"`
}

func (client *GitHubClient) Status(
	ctx context.Context,
	target Target,
	expectedHead string,
) (registry.PRFeedbackObservation, error) {
	args := []string{"api", "graphql", "--include", "-f", "query=" + statusQuery,
		"-f", "owner=" + target.Owner, "-f", "name=" + target.Name,
		"-F", fmt.Sprintf("number=%d", target.Number)}
	output, runErr := client.runner.Run(ctx, "", "gh", args...)
	body, metadata := splitIncludedResponse(output)
	var response statusResponse
	if len(strings.TrimSpace(string(body))) > 0 {
		_ = json.Unmarshal(body, &response)
	}
	resetAt := response.Data.RateLimit.ResetAt
	if err := classifyHTTPError(runErr, metadata, resetAt); err != nil {
		var httpErr *HTTPError
		if errors.As(err, &httpErr) {
			return registry.PRFeedbackObservation{
				ExpectedHead: expectedHead, HTTPStatus: httpErr.Status,
				RetryAfter: httpErr.RetryAfter, ResetAt: httpErr.ResetAt,
			}, nil
		}
		return registry.PRFeedbackObservation{}, fmt.Errorf("observe pull request status: %w", err)
	}
	if err := validateGraphQL(body); err != nil {
		return registry.PRFeedbackObservation{}, err
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return registry.PRFeedbackObservation{}, fmt.Errorf("decode status observation: %w", err)
	}
	pullRequest := response.Data.Repository.PullRequest
	if pullRequest == nil {
		return registry.PRFeedbackObservation{ExpectedHead: expectedHead,
			ProviderUnavailableReason: "pull request is unavailable"}, nil
	}
	observation := registry.PRFeedbackObservation{
		ExpectedHead: expectedHead, ObservedHead: pullRequest.HeadRefOID,
		PullRequestState: pullRequest.State, RateCost: response.Data.RateLimit.Cost,
		RateRemaining: response.Data.RateLimit.Remaining, RateLimit: response.Data.RateLimit.Limit,
		ResetAt: response.Data.RateLimit.ResetAt,
	}
	if len(pullRequest.Commits.Nodes) == 0 {
		return observation, nil
	}
	for _, context := range pullRequest.Commits.Nodes[0].Commit.Status.Contexts.Nodes {
		if context.Type == "StatusContext" && context.Context == "CodeRabbit" {
			observation.ContextPresent = true
			observation.ContextState = context.State
			observation.Description = context.Description
			break
		}
	}
	return observation, nil
}

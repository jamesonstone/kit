package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var dispatchReviewThreadResolver = resolveDispatchReviewThread

type dispatchReviewResolutionCandidate struct {
	Author   string
	Body     string
	Line     int
	Path     string
	ThreadID string
	URL      string
}

type dispatchReviewThreadMutationResponse struct {
	Data struct {
		ResolveReviewThread struct {
			Thread struct {
				ID         string `json:"id"`
				IsResolved bool   `json:"isResolved"`
			} `json:"thread"`
		} `json:"resolveReviewThread"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func runDispatchPRResolve(cmd *cobra.Command) error {
	if strings.TrimSpace(dispatchPR) == "" {
		return fmt.Errorf("--resolve requires --pr")
	}
	if strings.TrimSpace(dispatchFile) != "" {
		return fmt.Errorf("--file cannot be used with --resolve")
	}
	if dispatchWatch {
		return fmt.Errorf("--watch requires --loop")
	}
	if dispatchCopy || dispatchOutputOnly {
		return fmt.Errorf("--copy and --output-only cannot be used with --resolve")
	}
	if !dispatchYes {
		return fmt.Errorf("--resolve mutates GitHub review threads; rerun with --yes after confirming fixes or no-op decisions are complete")
	}

	target, err := resolveDispatchPRTarget(dispatchPR)
	if err != nil {
		return err
	}
	threads, err := fetchDispatchPRReviewThreads(target)
	if err != nil {
		return err
	}
	candidates := collectDispatchReviewResolutionCandidates(threads, dispatchCodeRabbit)
	if len(candidates) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "No unresolved PR review threads matched --resolve.")
		return err
	}

	for index, candidate := range candidates {
		if err := dispatchReviewThreadResolver(candidate.ThreadID); err != nil {
			return fmt.Errorf("failed to resolve review thread %s after %d/%d successful resolutions: %w",
				candidate.ThreadID,
				index,
				len(candidates),
				err,
			)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Resolved %d PR review thread(s):\n", len(candidates))
	for _, candidate := range candidates {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s (%s)\n", dispatchResolutionSourceLabel(candidate), candidate.ThreadID)
		if strings.TrimSpace(candidate.Body) != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Review: %s\n", candidate.Body)
		}
	}
	return nil
}

func collectDispatchReviewResolutionCandidates(
	threads []dispatchGitHubReviewThread,
	coderabbitOnly bool,
) []dispatchReviewResolutionCandidate {
	candidates := make([]dispatchReviewResolutionCandidate, 0, len(threads))
	seen := map[string]bool{}
	for _, thread := range threads {
		if thread.ID == "" || thread.IsResolved || thread.IsOutdated || seen[thread.ID] {
			continue
		}

		comment, ok := selectDispatchReviewComment(thread, coderabbitOnly)
		if !ok {
			continue
		}
		line := thread.Line
		if line == 0 {
			line = thread.StartLine
		}

		seen[thread.ID] = true
		candidates = append(candidates, dispatchReviewResolutionCandidate{
			Author:   comment.Author.Login,
			Body:     dispatchResolutionCommentSummary(comment.Body),
			Line:     line,
			Path:     thread.Path,
			ThreadID: thread.ID,
			URL:      comment.URL,
		})
	}
	return candidates
}

func resolveDispatchReviewThread(threadID string) error {
	query := `mutation($threadId:ID!) {
  resolveReviewThread(input:{threadId:$threadId}) {
    thread {
      id
      isResolved
    }
  }
}`
	output, err := commandOutput("gh", "api", "graphql", "-f", "query="+query, "-f", "threadId="+threadID)
	if err != nil {
		return fmt.Errorf("failed to resolve review thread: %w", err)
	}

	var response dispatchReviewThreadMutationResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return fmt.Errorf("failed to parse resolve review thread response: %w", err)
	}
	if len(response.Errors) > 0 {
		messages := make([]string, 0, len(response.Errors))
		for _, graphErr := range response.Errors {
			messages = append(messages, graphErr.Message)
		}
		return fmt.Errorf("GitHub GraphQL error: %s", strings.Join(messages, "; "))
	}
	resolved := response.Data.ResolveReviewThread.Thread
	if resolved.ID == "" || !resolved.IsResolved {
		return fmt.Errorf("GitHub did not confirm review thread %s was resolved", threadID)
	}
	return nil
}

func dispatchResolutionSourceLabel(candidate dispatchReviewResolutionCandidate) string {
	path := strings.TrimSpace(candidate.Path)
	if path == "" {
		path = "(no path)"
	}
	if candidate.Line > 0 {
		path = fmt.Sprintf("%s:%d", path, candidate.Line)
	}
	if strings.TrimSpace(candidate.Author) != "" {
		path = fmt.Sprintf("%s by %s", path, candidate.Author)
	}
	if strings.TrimSpace(candidate.URL) != "" {
		path = fmt.Sprintf("%s %s", path, candidate.URL)
	}
	return path
}

func dispatchResolutionCommentSummary(body string) string {
	cleaned, foundPrompt := extractPromptForAIAgents(body)
	if foundPrompt {
		cleaned, _ = stripCoderabbitSharedInstruction(cleaned)
	}
	if cleaned == "" {
		cleaned = cleanDispatchReviewComment(body)
	}
	if cleaned == "" {
		cleaned = normalizeDispatchRawInput(body)
	}
	lines := strings.Split(cleaned, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

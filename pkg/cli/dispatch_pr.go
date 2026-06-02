package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const coderabbitSharedReviewInstruction = `Verify each finding against current code. Fix only still-valid issues, skip the
rest with a brief reason, keep changes minimal, and validate.`

var (
	dispatchCurrentRepoResolver    = resolveCurrentGitHubRepo
	dispatchPRURLPattern           = regexp.MustCompile(`github\.com/([^/\s)]+)/([^/\s)]+)/pull/(\d+)`)
	dispatchPROwnerNumberPattern   = regexp.MustCompile(`^([^/\s#]+)/([^/\s#]+)#(\d+)$`)
	dispatchGitHubRemotePattern    = regexp.MustCompile(`github\.com[:/]([^/\s]+)/([^/\s]+?)(?:\.git)?$`)
	dispatchPromptDetailsPattern   = regexp.MustCompile(`(?is)<details>\s*<summary>[^<]*Prompt for AI Agents?[^<]*</summary>(.*?)</details>`)
	dispatchCodeFencePattern       = regexp.MustCompile("(?s)```(?:[a-zA-Z0-9_-]+)?\\s*\\n(.*?)\\n```")
	dispatchDetailsPattern         = regexp.MustCompile(`(?is)<details>.*?</details>`)
	dispatchSuggestionBlockPattern = regexp.MustCompile(`(?is)<!--\s*suggestion_start\s*-->.*?<!--\s*suggestion_end\s*-->`)
	dispatchHTMLCommentPattern     = regexp.MustCompile(`(?is)<!--.*?-->`)
	dispatchBoilerplatePattern     = regexp.MustCompile(`(?is)^\s*Verify each finding against current code\.\s*Fix only still-valid issues, skip the\s*rest with a brief reason, keep changes minimal, and validate\.\s*`)
	dispatchWhitespacePattern      = regexp.MustCompile(`\s+`)
)

type dispatchPRTarget struct {
	Owner  string
	Repo   string
	Number int
}

type dispatchReviewTask struct {
	Author string
	Body   string
	Line   int
	Path   string
	URL    string
}

type dispatchPRInput struct {
	CommonReviewInstruction string
	RawTasks                string
}

type dispatchGitHubReviewThread struct {
	IsOutdated bool   `json:"isOutdated"`
	IsResolved bool   `json:"isResolved"`
	Line       int    `json:"line"`
	StartLine  int    `json:"startLine"`
	Path       string `json:"path"`
	Comments   struct {
		Nodes []dispatchGitHubReviewComment `json:"nodes"`
	} `json:"comments"`
}

type dispatchGitHubReviewComment struct {
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	Body string `json:"body"`
	URL  string `json:"url"`
}

type dispatchGitHubReviewThreadResponse struct {
	Data struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					Nodes    []dispatchGitHubReviewThread `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func loadDispatchPRInput(
	prRef string,
	coderabbitOnly bool,
	inputCfg freeTextInputConfig,
) (dispatchPRInput, bool, error) {
	target, err := resolveDispatchPRTarget(prRef)
	if err != nil {
		return dispatchPRInput{}, false, err
	}

	threads, err := fetchDispatchPRReviewThreads(target)
	if err != nil {
		return dispatchPRInput{}, false, err
	}

	input := buildDispatchPRInput(threads, coderabbitOnly)
	if strings.TrimSpace(input.RawTasks) == "" {
		return dispatchPRInput{}, false, nil
	}

	initialContent := renderDispatchPRInputForEditor(input)
	edited, err := readEditorTextWithInitialContent(
		inputCfg,
		"dispatch review tasks",
		initialContent,
		false,
		false,
	)
	if err != nil {
		return dispatchPRInput{}, false, err
	}

	rawTasks, commonInstruction := splitDispatchPRInputFromEditor(edited, input.CommonReviewInstruction)
	if strings.TrimSpace(rawTasks) == "" {
		return dispatchPRInput{}, false, fmt.Errorf("dispatch review tasks cannot be empty")
	}

	return dispatchPRInput{
		CommonReviewInstruction: commonInstruction,
		RawTasks:                rawTasks,
	}, true, nil
}

func resolveDispatchPRTarget(raw string) (dispatchPRTarget, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return dispatchPRTarget{}, fmt.Errorf("--pr cannot be empty")
	}

	if match := dispatchPRURLPattern.FindStringSubmatch(value); match != nil {
		number, err := strconv.Atoi(match[3])
		if err != nil {
			return dispatchPRTarget{}, fmt.Errorf("invalid PR number %q: %w", match[3], err)
		}
		return dispatchPRTarget{Owner: match[1], Repo: match[2], Number: number}, nil
	}

	if parsed, err := url.Parse(value); err == nil && parsed.Host == "github.com" {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) >= 4 && parts[2] == "pull" {
			number, err := strconv.Atoi(parts[3])
			if err != nil {
				return dispatchPRTarget{}, fmt.Errorf("invalid PR number %q: %w", parts[3], err)
			}
			return dispatchPRTarget{Owner: parts[0], Repo: parts[1], Number: number}, nil
		}
	}

	if match := dispatchPROwnerNumberPattern.FindStringSubmatch(value); match != nil {
		number, err := strconv.Atoi(match[3])
		if err != nil {
			return dispatchPRTarget{}, fmt.Errorf("invalid PR number %q: %w", match[3], err)
		}
		return dispatchPRTarget{Owner: match[1], Repo: match[2], Number: number}, nil
	}

	if number, err := strconv.Atoi(value); err == nil {
		owner, repo, err := dispatchCurrentRepoResolver()
		if err != nil {
			return dispatchPRTarget{}, err
		}
		return dispatchPRTarget{Owner: owner, Repo: repo, Number: number}, nil
	}

	return dispatchPRTarget{}, fmt.Errorf("could not parse PR reference %q", raw)
}

func resolveCurrentGitHubRepo() (string, string, error) {
	output, err := commandOutput("git", "remote", "get-url", "origin")
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve current repo from git remote origin: %w", err)
	}

	owner, repo, err := parseGitHubRemoteURL(strings.TrimSpace(string(output)))
	if err != nil {
		return "", "", err
	}

	return owner, repo, nil
}

func parseGitHubRemoteURL(raw string) (string, string, error) {
	if match := dispatchGitHubRemotePattern.FindStringSubmatch(strings.TrimSpace(raw)); match != nil {
		return match[1], strings.TrimSuffix(match[2], ".git"), nil
	}

	return "", "", fmt.Errorf("remote origin is not a GitHub repository URL: %s", raw)
}

func fetchDispatchPRReviewThreads(target dispatchPRTarget) ([]dispatchGitHubReviewThread, error) {
	var all []dispatchGitHubReviewThread
	cursor := ""

	for {
		response, err := fetchDispatchPRReviewThreadPage(target, cursor)
		if err != nil {
			return nil, err
		}

		page := response.Data.Repository.PullRequest.ReviewThreads
		all = append(all, page.Nodes...)
		if !page.PageInfo.HasNextPage {
			return all, nil
		}
		cursor = page.PageInfo.EndCursor
		if cursor == "" {
			return nil, fmt.Errorf("GitHub reported another review-thread page without a cursor")
		}
	}
}

func fetchDispatchPRReviewThreadPage(
	target dispatchPRTarget,
	cursor string,
) (dispatchGitHubReviewThreadResponse, error) {
	query := dispatchPRReviewThreadsQuery(cursor != "")
	args := []string{
		"api", "graphql",
		"-f", "query=" + query,
		"-f", "owner=" + target.Owner,
		"-f", "name=" + target.Repo,
		"-F", fmt.Sprintf("number=%d", target.Number),
	}
	if cursor != "" {
		args = append(args, "-f", "cursor="+cursor)
	}

	output, err := commandOutput("gh", args...)
	if err != nil {
		return dispatchGitHubReviewThreadResponse{}, fmt.Errorf("failed to fetch PR review threads: %w", err)
	}

	var response dispatchGitHubReviewThreadResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return dispatchGitHubReviewThreadResponse{}, fmt.Errorf("failed to parse GitHub review thread response: %w", err)
	}
	if len(response.Errors) > 0 {
		messages := make([]string, 0, len(response.Errors))
		for _, graphErr := range response.Errors {
			messages = append(messages, graphErr.Message)
		}
		return dispatchGitHubReviewThreadResponse{}, fmt.Errorf("GitHub GraphQL error: %s", strings.Join(messages, "; "))
	}

	return response, nil
}

func dispatchPRReviewThreadsQuery(withCursor bool) string {
	after := ""
	if withCursor {
		after = ", after: $cursor"
	}
	cursorVar := ""
	if withCursor {
		cursorVar = ", $cursor: String!"
	}

	return fmt.Sprintf(`query($owner:String!, $name:String!, $number:Int!%s) {
  repository(owner:$owner, name:$name) {
    pullRequest(number:$number) {
      reviewThreads(first:100%s) {
        pageInfo { hasNextPage endCursor }
        nodes {
          isResolved
          isOutdated
          path
          line
          startLine
          comments(first:20) {
            nodes {
              author { login }
              body
              url
            }
          }
        }
      }
    }
  }
}`, cursorVar, after)
}

func commandOutput(name string, args ...string) ([]byte, error) {
	cmd := execCommand(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if ok := strings.TrimSpace(string(output)); ok != "" {
			if errors.As(err, &exitErr) {
				return nil, fmt.Errorf("%s: %s", err, ok)
			}
			return nil, fmt.Errorf("%w: %s", err, ok)
		}
		return nil, err
	}

	return output, nil
}

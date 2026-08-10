package agentcli

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/prfix"
)

func selectOpenPullRequest(input io.Reader, output io.Writer, pullRequests []prfix.OpenPullRequest) (string, error) {
	if len(pullRequests) == 0 {
		return "", fmt.Errorf("no open pull requests found in the current repository")
	}
	if _, err := fmt.Fprintln(output, "Open pull requests:"); err != nil {
		return "", err
	}
	for index, pullRequest := range pullRequests {
		state := "ready"
		if pullRequest.IsDraft {
			state = "draft"
		}
		if pullRequest.ReviewDecision != "" {
			state += ", " + strings.ToLower(pullRequest.ReviewDecision)
		}
		if _, err := fmt.Fprintf(output, "  %d. #%d %s [%s -> %s] %s\n",
			index+1, pullRequest.Number, strings.TrimSpace(pullRequest.Title),
			pullRequest.HeadRefName, pullRequest.BaseRefName, state); err != nil {
			return "", err
		}
	}
	if _, err := fmt.Fprintf(output, "Select PR (1-%d or #PR-number): ", len(pullRequests)); err != nil {
		return "", err
	}
	line, err := bufio.NewReader(input).ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return "", fmt.Errorf("read PR selection: %w", err)
	}
	rawChoice := strings.TrimSpace(line)
	explicitNumber := strings.HasPrefix(rawChoice, "#")
	choice, err := strconv.Atoi(strings.TrimPrefix(rawChoice, "#"))
	if err != nil {
		return "", fmt.Errorf("invalid PR selection %q", rawChoice)
	}
	if !explicitNumber && choice >= 1 && choice <= len(pullRequests) {
		return pullRequests[choice-1].URL, nil
	}
	for _, pullRequest := range pullRequests {
		if pullRequest.Number == choice {
			return pullRequest.URL, nil
		}
	}
	return "", fmt.Errorf("PR selection %d is not in the listed pull requests", choice)
}

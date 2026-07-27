package cli

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func selectPRFixOpenPullRequest(input io.Reader, output io.Writer) (string, error) {
	prs, err := prFixOpenPRLister()
	if err != nil {
		return "", err
	}
	if len(prs) == 0 {
		return "", fmt.Errorf("no open pull requests found in the current repository; pass --pr <url|owner/repo#number|number> to target another PR")
	}

	if _, err := fmt.Fprintln(output, "Open pull requests:"); err != nil {
		return "", err
	}
	for index, pr := range prs {
		if _, err := fmt.Fprintf(output, "  %d. #%d %s [%s -> %s] %s\n",
			index+1,
			pr.Number,
			strings.TrimSpace(pr.Title),
			strings.TrimSpace(pr.HeadRefName),
			strings.TrimSpace(pr.BaseRefName),
			prFixPRStateLabel(pr),
		); err != nil {
			return "", err
		}
	}
	if _, err := fmt.Fprintf(output, "Select PR (1-%d or PR number): ", len(prs)); err != nil {
		return "", err
	}

	reader := bufio.NewReader(input)
	line, err := reader.ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return "", fmt.Errorf("failed to read PR selection: %w", err)
	}
	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return "", fmt.Errorf("invalid PR selection %q", strings.TrimSpace(line))
	}

	for _, pr := range prs {
		if pr.Number == choice {
			return prFixPRRef(pr), nil
		}
	}
	if choice >= 1 && choice <= len(prs) {
		return prFixPRRef(prs[choice-1]), nil
	}
	return "", fmt.Errorf("PR selection %d is not in the listed PRs", choice)
}

func prFixPRRef(pr prFixOpenPullRequest) string {
	if strings.TrimSpace(pr.URL) != "" {
		return strings.TrimSpace(pr.URL)
	}
	return strconv.Itoa(pr.Number)
}

func prFixPRStateLabel(pr prFixOpenPullRequest) string {
	var parts []string
	if pr.IsDraft {
		parts = append(parts, "draft")
	} else {
		parts = append(parts, "ready")
	}
	if strings.TrimSpace(pr.ReviewDecision) != "" {
		parts = append(parts, strings.ToLower(strings.TrimSpace(pr.ReviewDecision)))
	}
	return strings.Join(parts, ", ")
}

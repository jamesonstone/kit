package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

var (
	reviewLoopRunner            reviewLoopCommandRunner = execCICommandRunner{}
	reviewLoopLocalRootResolver                         = resolveReviewLoopLocalRoot
	reviewLoopIssueHintPattern                          = regexp.MustCompile(`(?i)\b(?:close[sd]?|fixe[sd]?|resolve[sd]?|refs?)\s+#(\d+)\b`)
)

type reviewLoopCommandRunner interface {
	Output(dir string, name string, args ...string) ([]byte, error)
	OutputAllowError(dir string, name string, args ...string) ([]byte, error)
}

func fetchReviewLoopPRContext(prRef string) (reviewLoopPRContext, error) {
	target, err := resolveDispatchPRTarget(prRef)
	if err != nil {
		return reviewLoopPRContext{}, err
	}

	repo := target.Owner + "/" + target.Repo
	output, err := reviewLoopRunner.Output("", "gh", repoArgs(repo,
		"pr", "view", strconv.Itoa(target.Number),
		"--json", "number,url,title,body,headRefOid")...)
	if err != nil {
		return reviewLoopPRContext{}, fmt.Errorf("failed to fetch PR metadata: %w", err)
	}

	var payload struct {
		Number     int    `json:"number"`
		URL        string `json:"url"`
		Title      string `json:"title"`
		Body       string `json:"body"`
		HeadRefOID string `json:"headRefOid"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return reviewLoopPRContext{}, fmt.Errorf("failed to parse PR metadata: %w", err)
	}
	if payload.Number == 0 {
		payload.Number = target.Number
	}
	if strings.TrimSpace(payload.HeadRefOID) == "" {
		return reviewLoopPRContext{}, fmt.Errorf("PR metadata did not include current head SHA")
	}

	return reviewLoopPRContext{
		Target:       target,
		URL:          payload.URL,
		Title:        payload.Title,
		Body:         payload.Body,
		HeadRefOID:   payload.HeadRefOID,
		IssueHints:   extractReviewLoopIssueHints(payload.Title + "\n" + payload.Body),
		RepoFullName: repo,
		LocalRoot:    reviewLoopLocalRootResolver(),
	}, nil
}

func resolveReviewLoopLocalRoot() string {
	if projectRoot, found, err := config.FindProjectRootOptional(); err == nil && found {
		return projectRoot
	}

	output, err := commandOutput("git", "rev-parse", "--show-toplevel")
	if err == nil {
		if root := strings.TrimSpace(string(output)); root != "" {
			return root
		}
	}

	workingDirectory, err := os.Getwd()
	if err == nil && strings.TrimSpace(workingDirectory) != "" {
		return workingDirectory
	}
	return "."
}

func resolveReviewLoopLocalPath(ctx reviewLoopPRContext, repoRelativePath string) string {
	cleanPath := strings.TrimSpace(repoRelativePath)
	if cleanPath == "" || filepath.IsAbs(cleanPath) {
		return cleanPath
	}

	root := strings.TrimSpace(ctx.LocalRoot)
	if root == "" {
		root = "."
	}
	return filepath.Join(root, filepath.FromSlash(cleanPath))
}

func fetchReviewLoopChecks(ctx reviewLoopPRContext) ([]reviewLoopCheck, error) {
	output, err := reviewLoopRunner.OutputAllowError("", "gh", repoArgs(ctx.RepoFullName,
		"pr", "checks", strconv.Itoa(ctx.Target.Number),
		"--json", "bucket,completedAt,description,link,name,state,workflow")...)
	if err != nil && len(output) == 0 {
		return nil, fmt.Errorf("failed to fetch PR checks: %w", err)
	}

	var checks []reviewLoopCheck
	if err := json.Unmarshal(output, &checks); err != nil {
		return nil, fmt.Errorf("failed to parse PR checks: %w", err)
	}
	return checks, nil
}

func extractReviewLoopIssueHints(text string) []string {
	matches := reviewLoopIssueHintPattern.FindAllStringSubmatch(text, -1)
	seen := map[string]bool{}
	var hints []string
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		hint := "#" + match[1]
		if seen[hint] {
			continue
		}
		seen[hint] = true
		hints = append(hints, hint)
	}
	return hints
}

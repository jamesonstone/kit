package cli

import (
	"fmt"
	"os"
	"strings"
)

func classifyReviewLoopFindings(
	ctx reviewLoopPRContext,
	tasks []dispatchReviewTask,
) []reviewLoopClassifiedFinding {
	classified := make([]reviewLoopClassifiedFinding, 0, len(tasks))
	for _, task := range tasks {
		finding := reviewLoopFinding{Task: task}
		kind, reason := classifyReviewLoopTask(ctx, task)
		classified = append(classified, reviewLoopClassifiedFinding{
			Finding: finding,
			Kind:    kind,
			Reason:  reason,
		})
	}
	return classified
}

func classifyReviewLoopTask(
	ctx reviewLoopPRContext,
	task dispatchReviewTask,
) (reviewLoopClassification, string) {
	path := strings.TrimSpace(task.Path)
	if path == "" {
		return reviewLoopNeedsHuman, "review thread did not include a file path"
	}
	localPath := resolveReviewLoopLocalPath(ctx, path)
	if _, err := os.Stat(localPath); err != nil {
		if os.IsNotExist(err) {
			return reviewLoopStale, fmt.Sprintf("source file %s no longer exists locally", path)
		}
		return reviewLoopNeedsHuman, fmt.Sprintf("could not inspect source file %s: %v", path, err)
	}
	if task.Line <= 0 {
		return reviewLoopNeedsHuman, "review thread did not include a concrete line number"
	}
	if !reviewLoopLineExists(localPath, task.Line) {
		return reviewLoopStale, fmt.Sprintf("source line %s:%d no longer exists locally", path, task.Line)
	}

	body := strings.ToLower(task.Body)
	switch {
	case containsAnyReviewLoopToken(body, "false positive", "not actually an issue", "not reproducible"):
		return reviewLoopFalsePositive, "finding text indicates the issue is already disproven"
	case containsAnyReviewLoopToken(body, "out of scope", "outside scope", "outside this pr", "separate pr"):
		return reviewLoopValidOutOfScope, "finding appears valid but outside the current PR scope"
	case containsAnyReviewLoopToken(body, "needs human", "human decision", "ambiguous", "manual decision"):
		return reviewLoopNeedsHuman, "finding needs a human scope or product decision"
	case len(ctx.IssueHints) == 0 && containsAnyReviewLoopToken(body, "scope", "requirement"):
		return reviewLoopNeedsHuman, "finding references scope but no linked issue hint was available"
	case looksReviewLoopActionable(body):
		return reviewLoopFix, "source evidence is current and the finding appears actionable"
	default:
		return reviewLoopNeedsHuman, "source evidence is current but the requested action is unclear"
	}
}

func reviewLoopLineExists(path string, line int) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if line <= 0 {
		return false
	}
	lines := strings.Count(string(content), "\n")
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		lines++
	}
	return line <= lines
}

func containsAnyReviewLoopToken(text string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func looksReviewLoopActionable(body string) bool {
	normalized := " " + dispatchWhitespacePattern.ReplaceAllString(strings.ToLower(body), " ") + " "
	return containsAnyReviewLoopToken(normalized,
		" fix ",
		" add ",
		" update ",
		" use ",
		" replace ",
		" remove ",
		" rename ",
		" handle ",
		" guard ",
		" validate ",
		" ensure ",
		" narrow ",
		" prefer ",
		" avoid ",
	)
}

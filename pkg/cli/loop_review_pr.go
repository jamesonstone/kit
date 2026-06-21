package cli

import (
	"sort"
	"strconv"
	"strings"
)

func pollLoopReviewPRFeedback(ctx reviewLoopPRContext) (loopReviewPRFeedback, error) {
	tasks, commonInstruction, found, err := reviewLoopLoadReviewTasks(reviewLoopTargetRef(ctx.Target), true)
	if err != nil {
		return loopReviewPRFeedback{}, err
	}

	checks, err := fetchReviewLoopChecks(ctx)
	status := reviewLoopCheckUnavailable
	statusLabel := "CodeRabbit unavailable"
	if err == nil {
		status = summarizeReviewLoopCodeRabbitChecks(checks)
		statusLabel = loopReviewPRStatusLabel(status)
	}

	feedback := loopReviewPRFeedback{
		Status:            status,
		StatusLabel:       statusLabel,
		Found:             found,
		CommonInstruction: commonInstruction,
		Pending:           status == reviewLoopCheckPending,
	}
	if !found {
		return feedback, nil
	}
	feedback.Fingerprint = loopReviewFeedbackFingerprint(tasks)
	feedback.RenderedTasks = renderDispatchReviewTasks(tasks)
	return feedback, nil
}

func loopReviewPRStatusLabel(status reviewLoopCheckStatus) string {
	switch status {
	case reviewLoopCheckPending:
		return "CodeRabbit pending"
	case reviewLoopCheckComplete:
		return "CodeRabbit complete"
	default:
		return "CodeRabbit unavailable"
	}
}

func loopReviewFeedbackFingerprint(tasks []dispatchReviewTask) string {
	var parts []string
	for _, task := range tasks {
		parts = append(parts, strings.Join([]string{
			task.Path,
			strconv.Itoa(task.Line),
			task.URL,
			normalizeLoopReviewFeedbackBody(task.Body),
		}, "\x00"))
	}
	sort.Strings(parts)
	return strings.Join(parts, "\x01")
}

func normalizeLoopReviewFeedbackBody(body string) string {
	return strings.ToLower(dispatchWhitespacePattern.ReplaceAllString(strings.TrimSpace(body), " "))
}

func renderLoopReviewPRFeedback(feedback loopReviewPRFeedback) string {
	var builder strings.Builder
	if strings.TrimSpace(feedback.CommonInstruction) != "" {
		builder.WriteString(strings.TrimSpace(feedback.CommonInstruction))
		builder.WriteString("\n\n")
	}
	builder.WriteString(strings.TrimSpace(feedback.RenderedTasks))
	return builder.String()
}

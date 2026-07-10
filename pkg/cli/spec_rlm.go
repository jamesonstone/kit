package cli

import (
	"os"
	"strings"
)

var rlmPromptKeywords = []string{
	"codebase-wide",
	"codebase wide",
	"analyze all",
	"scan repository",
	"scan all",
	"large repository",
	"large repo",
	"recursive language model",
	"rlm",
}

func specNeedsRLM(featureSlug, specPath, brainstormPath string, answers *specAnswers) bool {
	var candidates []string
	candidates = append(candidates, featureSlug)
	candidates = append(candidates, readPromptContextFile(specPath))
	candidates = append(candidates, readPromptContextFile(brainstormPath))

	if answers != nil {
		candidates = append(
			candidates,
			answers.Problem,
			answers.Goals,
			answers.NonGoals,
			answers.Users,
			answers.Requirements,
			answers.Acceptance,
			answers.EdgeCases,
		)
	}

	haystack := strings.ToLower(strings.Join(candidates, "\n"))
	for _, keyword := range rlmPromptKeywords {
		if strings.Contains(haystack, keyword) {
			return true
		}
	}

	return false
}

func readPromptContextFile(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return string(content)
}

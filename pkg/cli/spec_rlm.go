package cli

import (
	"fmt"
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

func rlmSpecGuidanceStepText(specPath string) string {
	return strings.Join([]string{
		"Treat this feature as broad or noisy-context work and route discovery through the `rlm` skill:",
		"- start from `docs/agents/RLM.md` when that repo-local guide exists and preserve its progressive-disclosure model",
		fmt.Sprintf("- add `rlm` to the `## SKILLS` table in `%s` when full-context loading would be noisy or wasteful", specPath),
		"- record `parallelization_mode: \"rlm\"` in downstream planning notes or execution metadata so later stages preserve the dispatch strategy",
		"- use the trigger phrases `analyze codebase`, `scan all files`, `large repository analysis`, `scan repository`, and `recursive language model`",
		"- structure discovery as immediate decision → smallest artifact → required facts → act or recurse",
		"- keep map workers file-scoped so the synthesis step stays deterministic and source-attributed",
		"- record the docs, skills, and references that materially shaped the work in the feature dependency tables",
	}, "\n")
}

func appendRLMSpecGuidanceStep(sb *strings.Builder, step int, specPath string) int {
	sb.WriteString("\n# Use RLM Pattern\n")
	sb.WriteString(fmt.Sprintf("%d. %s\n", step, strings.ReplaceAll(rlmSpecGuidanceStepText(specPath), "\n", "\n   ")))
	return step + 1
}

package cli

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestModernizedPromptsAvoidAPIOnlyFeaturesAndRoutineApproval(t *testing.T) {
	prompts := map[string]string{
		"spec": buildSpecV2SupervisorPrompt(specV2PromptInput{
			SpecPath:    "/repo/docs/specs/0001-alpha/SPEC.md",
			FeatureSlug: "alpha",
			ProjectRoot: "/repo",
			Config:      config.Default(),
		}),
		"loop": buildLoopEngineeringPrompt(loopPromptInput{
			Title:      "alpha",
			Source:     "SPEC.md",
			Scope:      "all accepted work",
			Validation: "go test ./...",
			Docs:       "affected docs",
			Delivery:   "repo-local rules",
		}),
		"code-review": codeReviewInstructions(),
		"dispatch": buildDispatchPrompt(
			[]dispatchTask{{ID: "D001", Body: "Implement alpha"}},
			3,
			"/repo",
			dispatchInputSourceEditor,
			dispatchPromptOptions{},
		),
		"toolbox-short":        codingAgentShortPrompt,
		"toolbox-long":         codingAgentLongPrompt,
		"toolbox-instructions": codingAgentInstructionsPrompt,
	}

	for name, prompt := range prompts {
		for _, forbidden := range []string{
			"Programmatic Tool Calling",
			"persisted reasoning",
			"Pro mode",
			"text.verbosity",
		} {
			if strings.Contains(prompt, forbidden) {
				t.Fatalf("%s prompt contains API-only instruction %q", name, forbidden)
			}
		}
	}

	for name, required := range map[string]string{
		"spec":        "do not re-ask settled questions or request routine permission",
		"loop":        "need no routine approval",
		"code-review": "This is read-only",
		"dispatch":    "mutate Git/GitHub delivery state unless explicitly assigned and authorized",
	} {
		if !strings.Contains(prompts[name], required) {
			t.Fatalf("%s prompt missing autonomy boundary %q", name, required)
		}
	}
}

func TestModernizedSharedPromptLayersStayCompact(t *testing.T) {
	limits := map[string]struct {
		prompt string
		words  int
	}{
		"skills":    {skillPromptSuffix(), 150},
		"subagents": {subagentPromptSuffix(), 180},
		"frontend":  {frontendPromptProfileSuffix(), 220},
		"review":    {codeReviewInstructions(), 500},
	}
	for name, limit := range limits {
		if words := len(strings.Fields(limit.prompt)); words > limit.words {
			t.Fatalf("%s prompt layer has %d words, maximum %d", name, words, limit.words)
		}
	}
}

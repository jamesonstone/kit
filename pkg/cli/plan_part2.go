package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

// selectFeatureForPlan shows an interactive numbered list of features
// that have SPEC.md but no PLAN.md yet.
func selectFeatureForPlan(specsDir string) (*feature.Feature, error) {
	candidates, err := workflowStageCandidates(specsDir, workflowSelectionStagePlan)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features ready for planning (need SPEC.md without PLAN.md)\n\nRun 'kit spec <feature>' to create a new feature first")
	}

	printSelectionHeader("Select a feature to plan:")
	for i, f := range candidates {
		fmt.Printf("  [%d] %s\n", i+1, f.DirName)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(candidates) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := candidates[num-1]
	return &selected, nil
}

func selectFeatureForPlanPromptOnly(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "SPEC.md")) &&
			document.Exists(filepath.Join(f.Path, "PLAN.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no plans available to regenerate prompts for\n\nRun 'kit legacy plan <feature>' first")
	}

	printSelectionHeader("Select a feature to regenerate the plan prompt for:")
	for i, f := range candidates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, f.DirName, f.Phase)
	}
	fmt.Println()
	fmt.Print(selectionPrompt(os.Stdout))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(candidates) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	selected := candidates[num-1]
	return &selected, nil
}

func planDependencyInventoryStepText(planPath, specPath, brainstormPath string, hasBrainstorm bool) string {
	lines := []string{
		fmt.Sprintf("Populate or refresh canonical front matter `references` in `%s` before sign-off:", planPath),
		fmt.Sprintf("- carry forward still-relevant references from `%s`", specPath),
	}
	if hasBrainstorm {
		lines = append(lines, fmt.Sprintf("- carry forward still-relevant references from `%s`", brainstormPath))
	}
	lines = append(lines,
		"- include skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, assets, and other resources that shape the implementation strategy",
		"- use `name`, `type`, `target`, `relation`, `read_policy`, `used_for`, and `status`",
		"- add a stable `id` when the reference may need to be updated later",
		"- `selector_type` must be one of `artifact`, `heading`, `symbol`, `command`, `url`, or `node_id` when `selector` is set",
		"- `relation` describes the referenced target's role relative to the source artifact, such as `constrains`, `guides`, `informs`, `implements`, `verifies`, or `uses`",
		"- `read_policy` must be one of `must`, `conditional`, `evidence`, or `skip`",
		"- `status` must be one of `active`, `optional`, or `stale`",
		"- for Figma or MCP-driven design references, store the exact design URL or file/node reference in `target` and use stable selectors when needed",
		"- if a reference influenced the implementation strategy but is no longer current, keep it with `status: stale` and `read_policy: skip`",
		"- if no additional references apply, leave front matter references empty and keep the body `## DEPENDENCIES` section prose-only",
	)
	return strings.Join(lines, "\n")
}

// outputStandardPlanPrompt outputs the standard coding agent prompt.
func outputStandardPlanPrompt(planPath, specPath, brainstormPath string, feat *feature.Feature, cfg *config.Config, outputOnly bool) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	prompt := buildStandardPlanPrompt(planPath, specPath, brainstormPath, feat, cfg, projectRoot)

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, planCopy); err != nil {
		return fmt.Errorf("failed to output prompt: %w", err)
	}

	return nil
}

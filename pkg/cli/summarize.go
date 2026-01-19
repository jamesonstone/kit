// package cli implements the Kit command-line interface.
package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize [feature]",
	Short: "Output context summarization instructions",
	Long: `Output instructions for context window summarization that focus on
retaining facts necessary for strategy, implementation, and process.

When a feature is specified, instructions are scoped to that feature's context.
Without a feature argument, outputs generic best-practice instructions.

Use with coding agents: /compact (Warp), /summarize (Claude), etc.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSummarize,
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}

func runSummarize(cmd *cobra.Command, args []string) error {
	instructions := genericSummarizeInstructions()

	if len(args) == 1 {
		featureRef := args[0]

		projectRoot, err := config.FindProjectRoot()
		if err != nil {
			return err
		}

		cfg, err := config.Load(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		specsDir := cfg.SpecsPath(projectRoot)
		feat, err := feature.Resolve(specsDir, featureRef)
		if err != nil {
			return fmt.Errorf("failed to resolve feature: %w", err)
		}

		instructions = featureScopedSummarizeInstructions(feat.Slug, feat.Path)
	}

	fmt.Println(instructions)
	return nil
}

func genericSummarizeInstructions() string {
	return `## Context Summarization Instructions

Summarize the current context window using the following principles:

### Fact Retention Protocol
- Extract and retain ONLY facts strictly necessary for:
  - **Strategy**: architectural decisions, design patterns, system constraints
  - **Implementation**: code structure, dependencies, APIs, data models
  - **Process**: workflow state, pending tasks, blockers, decisions made

### What to KEEP
- Concrete decisions and their rationale
- File paths, function names, type definitions
- API contracts and data schemas
- Configuration values and environment requirements
- Error states and their resolutions
- Dependencies and version constraints
- Test requirements and acceptance criteria

### What to DISCARD
- Conversational pleasantries and acknowledgments
- Redundant explanations of the same concept
- Speculative discussions that were not acted upon
- Verbose error messages (keep only the root cause)
- Code that was shown but not modified
- Historical context superseded by later decisions

### Output Format
Structure the summary as:
1. **Current State**: what exists now (files, functions, configs)
2. **Active Work**: what is being implemented or modified
3. **Decisions Made**: concrete choices with brief rationale
4. **Pending Items**: unresolved questions or next steps
5. **Constraints**: hard limits, invariants, non-negotiables

### Rules
- Use bullet points, not prose
- One fact per line
- No filler words or qualifiers
- Quantify where possible (lines, counts, versions)
- If a fact cannot be verified from context, mark it [ASSUMED]`
}

func featureScopedSummarizeInstructions(featureSlug, featurePath string) string {
	return fmt.Sprintf(`## Context Summarization Instructions — Feature: %s

Summarize the current context window, focusing on feature **%s**.

### Feature Documents
Review and extract facts from:
- %s/SPEC.md — requirements and acceptance criteria
- %s/PLAN.md — implementation approach and components
- %s/TASKS.md — work units and their status
- %s/ANALYSIS.md — understanding state and open questions

### Fact Retention Protocol
- Extract and retain ONLY facts strictly necessary for:
  - **Strategy**: architectural decisions, design patterns, system constraints
  - **Implementation**: code structure, dependencies, APIs, data models
  - **Process**: workflow state, pending tasks, blockers, decisions made

### What to KEEP
- Concrete decisions and their rationale
- File paths, function names, type definitions
- API contracts and data schemas
- Configuration values and environment requirements
- Error states and their resolutions
- Dependencies and version constraints
- Test requirements and acceptance criteria
- Feature-specific constraints from SPEC.md
- Implementation choices from PLAN.md
- Task status and dependencies from TASKS.md

### What to DISCARD
- Conversational pleasantries and acknowledgments
- Redundant explanations of the same concept
- Speculative discussions that were not acted upon
- Verbose error messages (keep only the root cause)
- Code that was shown but not modified
- Historical context superseded by later decisions
- Information unrelated to feature %s

### Output Format
Structure the summary as:
1. **Feature Intent**: one-sentence purpose from SPEC.md
2. **Current State**: what exists now for this feature
3. **Active Work**: what is being implemented or modified
4. **Decisions Made**: concrete choices with brief rationale
5. **Pending Items**: unresolved questions, open tasks
6. **Constraints**: hard limits, invariants, non-negotiables

### Rules
- Use bullet points, not prose
- One fact per line
- No filler words or qualifiers
- Quantify where possible (lines, counts, versions)
- If a fact cannot be verified from context, mark it [ASSUMED]
- Prioritize facts from feature documents over conversation`, featureSlug, featureSlug,
		featurePath, featurePath, featurePath, featurePath, featureSlug)
}

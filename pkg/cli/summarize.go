// package cli implements the Kit command-line interface.
package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
	"github.com/spf13/cobra"
)

var summarizeCopy bool
var summarizeOutputOnly bool

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
	summarizeCmd.Flags().BoolVar(&summarizeCopy, "copy", false, "copy output to clipboard even with --output-only")
	summarizeCmd.Flags().BoolVar(&summarizeOutputOnly, "output-only", false, "output text to stdout instead of copying it to the clipboard")
	rootCmd.AddCommand(summarizeCmd)
}

func runSummarize(cmd *cobra.Command, args []string) error {
	instructions := genericSummarizeInstructions()
	featurePath := ""

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

		instructions = featureScopedSummarizeInstructions(projectRoot, feat.Slug, feat.Path)
		featurePath = feat.Path
	}

	outputOnly, _ := cmd.Flags().GetBool("output-only")

	if !outputOnly {
		printWorkflowInstructions("summarize context (supporting step)", []string{
			"resume your active phase: brainstorm -> spec -> plan -> tasks -> implement -> reflect",
		})
	}

	if featurePath != "" {
		if err := outputPromptForFeatureWithClipboardDefault(instructions, featurePath, outputOnly, summarizeCopy); err != nil {
			return err
		}
		return nil
	}
	if err := outputPromptWithClipboardDefault(instructions, outputOnly, summarizeCopy); err != nil {
		return err
	}

	return nil
}

func genericSummarizeInstructions() string {
	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Heading(2, "Context Summarization Instructions")
		doc.Paragraph("Summarize the current context window using the following principles:")
		doc.Heading(3, "Fact Retention Protocol")
		doc.BulletList(
			"Extract and retain ONLY facts strictly necessary for:\n  - **Strategy**: architectural decisions, design patterns, system constraints\n  - **Implementation**: code structure, dependencies, APIs, data models\n  - **Process**: workflow state, pending tasks, blockers, decisions made",
		)
		doc.Heading(3, "What to KEEP")
		doc.BulletList(
			"Concrete decisions and their rationale",
			"File paths, function names, type definitions",
			"API contracts and data schemas",
			"Configuration values and environment requirements",
			"Error states and their resolutions",
			"Dependencies and version constraints",
			"Test requirements and acceptance criteria",
		)
		doc.Heading(3, "What to DISCARD")
		doc.BulletList(
			"Conversational pleasantries and acknowledgments",
			"Redundant explanations of the same concept",
			"Speculative discussions that were not acted upon",
			"Verbose error messages (keep only the root cause)",
			"Code that was shown but not modified",
			"Historical context superseded by later decisions",
		)
		doc.Heading(3, "Output Format")
		doc.Paragraph("Structure the summary as:")
		doc.OrderedList(1,
			"**Current State**: what exists now (files, functions, configs)",
			"**Active Work**: what is being implemented or modified",
			"**Decisions Made**: concrete choices with brief rationale",
			"**Pending Items**: unresolved questions or next steps",
			"**Constraints**: hard limits, invariants, non-negotiables",
		)
		doc.Heading(3, "Rules")
		doc.BulletList(
			"Use bullet points, not prose",
			"One fact per line",
			"No filler words or qualifiers",
			"Quantify where possible (lines, counts, versions)",
			"If a fact cannot be verified from context, mark it [ASSUMED]",
		)
	})
}

func featureScopedSummarizeInstructions(projectRoot, featureSlug, featurePath string) string {
	cfg, _ := loadRepoInstructionContext(projectRoot)
	repoAgentsPath := repoKnowledgeEntrypointPath(projectRoot, cfg)
	repoReferencesPath := repoReferencesEntrypointPath(projectRoot, cfg)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Heading(2, fmt.Sprintf("Context Summarization Instructions — Feature: %s", featureSlug))
		doc.Paragraph(fmt.Sprintf("Summarize the current context window, focusing on feature **%s**.", featureSlug))
		doc.Heading(3, "Feature Documents")
		doc.Paragraph("Review and extract facts from:")
		items := []string{}
		if repoAgentsPath != "" {
			items = append(items, fmt.Sprintf("%s — repo-local entrypoint and workflow map", repoAgentsPath))
		}
		if repoReferencesPath != "" {
			items = append(items, fmt.Sprintf("%s — durable repo-wide references when they materially shape the feature", repoReferencesPath))
		}
		items = append(items,
			fmt.Sprintf("%s/BRAINSTORM.md — optional research findings, affected files, and strategy", featurePath),
			fmt.Sprintf("%s/SPEC.md — requirements and acceptance criteria", featurePath),
			fmt.Sprintf("%s/PLAN.md — implementation approach and components", featurePath),
			fmt.Sprintf("%s/TASKS.md — work units and their status", featurePath),
			fmt.Sprintf("%s/ANALYSIS.md — understanding state and open questions", featurePath),
		)
		doc.BulletList(items...)
		doc.Heading(3, "Fact Retention Protocol")
		doc.BulletList("Extract and retain ONLY facts strictly necessary for:\n  - **Strategy**: architectural decisions, design patterns, system constraints\n  - **Implementation**: code structure, dependencies, APIs, data models\n  - **Process**: workflow state, pending tasks, blockers, decisions made")
		doc.Heading(3, "What to KEEP")
		doc.BulletList(
			"Concrete decisions and their rationale",
			"File paths, function names, type definitions",
			"API contracts and data schemas",
			"Configuration values and environment requirements",
			"Error states and their resolutions",
			"Dependencies and version constraints",
			"Test requirements and acceptance criteria",
			"Research findings and recommended strategy from BRAINSTORM.md when present",
			"Feature-specific constraints from SPEC.md",
			"Implementation choices from PLAN.md",
			"Task status and dependencies from TASKS.md",
		)
		doc.Heading(3, "What to DISCARD")
		doc.BulletList(
			"Conversational pleasantries and acknowledgments",
			"Redundant explanations of the same concept",
			"Speculative discussions that were not acted upon",
			"Verbose error messages (keep only the root cause)",
			"Code that was shown but not modified",
			"Historical context superseded by later decisions",
			fmt.Sprintf("Information unrelated to feature %s", featureSlug),
		)
		doc.Heading(3, "Output Format")
		doc.Paragraph("Structure the summary as:")
		doc.OrderedList(1,
			"**Feature Intent**: one-sentence purpose from SPEC.md",
			"**Current State**: what exists now for this feature",
			"**Active Work**: what is being implemented or modified",
			"**Decisions Made**: concrete choices with brief rationale",
			"**Pending Items**: unresolved questions, open tasks",
			"**Constraints**: hard limits, invariants, non-negotiables",
		)
		doc.Heading(3, "Rules")
		doc.BulletList(
			"Use bullet points, not prose",
			"One fact per line",
			"No filler words or qualifiers",
			"Quantify where possible (lines, counts, versions)",
			"If a fact cannot be verified from context, mark it [ASSUMED]",
			"Prioritize facts from repository documents over conversation",
		)
	})
}

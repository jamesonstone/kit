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

func outputExistingBrainstormPrompt(args []string, projectRoot string, cfg *config.Config, outputOnly bool) error {
	if brainstormOutput != "" {
		return fmt.Errorf("--prompt-only cannot be used with --output because it writes a file")
	}

	specsDir := cfg.SpecsPath(projectRoot)

	var (
		feat *feature.Feature
		err  error
	)

	if len(args) == 1 {
		feat, err = loadFeatureWithState(specsDir, cfg, args[0])
		if err != nil {
			return fmt.Errorf("feature '%s' not found: %w", args[0], err)
		}
	} else {
		feat, err = selectFeatureForBrainstormPromptOnly(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	if !document.Exists(brainstormPath) {
		return fmt.Errorf("BRAINSTORM.md not found for %s. Run 'kit brainstorm %s' first", feat.Slug, feat.Slug)
	}

	thesis := existingBrainstormThesis(brainstormPath)
	prompt := buildBrainstormPrompt(brainstormPath, feat.Slug, projectRoot, thesis, cfg.GoalPercentage)
	preparedPrompt := prepareAgentPrompt(prompt)

	if brainstormOutput != "" {
		if err := document.Write(brainstormOutput, preparedPrompt); err != nil {
			return fmt.Errorf("failed to write prompt file: %w", err)
		}
		if !outputOnly {
			fmt.Printf("✓ Written prompt to %s\n", brainstormOutput)
		}
	}

	if err := writePromptWithClipboardDefault(preparedPrompt, outputOnly, brainstormCopy); err != nil {
		return err
	}

	if !outputOnly {
		printWorkflowInstructions("brainstorm (existing feature prompt)", []string{
			fmt.Sprintf("review and refine %s", brainstormPath),
			fmt.Sprintf("run kit spec %s when the brainstorm is complete", feat.Slug),
			"no repository docs were mutated by this prompt-only run",
		})
	}

	return nil
}

func selectFeatureForBrainstormPromptOnly(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		if document.Exists(filepath.Join(f.Path, "BRAINSTORM.md")) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features with BRAINSTORM.md available\n\nRun 'kit brainstorm <feature>' first")
	}

	printSelectionHeader("Select a feature to regenerate the brainstorm prompt for:")
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

func existingBrainstormThesis(brainstormPath string) string {
	doc, err := document.ParseFile(brainstormPath, document.TypeBrainstorm)
	if err != nil {
		return "Continue the existing brainstorm using the current file contents as the source of truth."
	}

	if section := doc.GetSection("USER THESIS"); section != nil {
		if thesis := document.ExtractFirstParagraph(section); thesis != "" {
			return thesis
		}
	}
	if section := doc.GetSection("SUMMARY"); section != nil {
		if summary := document.ExtractFirstParagraph(section); summary != "" {
			return summary
		}
	}

	return "Continue the existing brainstorm using the current file contents as the source of truth."
}

func buildBrainstormPrompt(brainstormPath, featureSlug, projectRoot, thesis string, goalPct int) string {
	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")

	var sb strings.Builder
	sb.WriteString("/plan\n\n")
	sb.WriteString(fmt.Sprintf(`You are in planning mode for feature: **%s**

You MUST update the brainstorm file at:
- **BRAINSTORM**: %s
- **Feature**: %s
- **Project Root**: %s

## User Thesis

%s

## Context Docs (read first)
| File | Purpose |
|------|---------|
| CONSTITUTION | %s |
| BRAINSTORM | %s |
| Project Root | %s |

## Your Task

1. Stay in planning and information-gathering mode only
2. Do NOT implement code, write production changes, or move into execution
3. Read CONSTITUTION.md first to understand project constraints and workflow rules
4. Read the current BRAINSTORM.md template and treat it as the source of truth for this research phase
5. Research the entire codebase at %s to identify relevant files, patterns, constraints, interfaces, and adjacent workflows
`, featureSlug, brainstormPath, featureSlug, projectRoot, thesis, constitutionPath, brainstormPath, projectRoot, projectRoot))

	sb.WriteString(fmt.Sprintf("6. Keep the `## DEPENDENCIES` table in %s current throughout the research phase:\n", brainstormPath))
	sb.WriteString("   - include every dependency that materially shapes the feature definition, such as skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, and assets\n")
	sb.WriteString("   - use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n")
	sb.WriteString("   - `Status` must be one of `active`, `optional`, or `stale`\n")
	sb.WriteString("   - for Figma or other MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n")
	sb.WriteString("   - if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale` instead of deleting it\n")
	sb.WriteString(fmt.Sprintf("7. Keep the `## RELATIONSHIPS` section in %s current throughout the research phase:\n", brainstormPath))
	sb.WriteString("   - use `none` when this feature does not build on an existing feature\n")
	sb.WriteString("   - otherwise record one bullet per explicit feature relationship\n")
	sb.WriteString("   - supported labels are `builds on`, `depends on`, and `related to`\n")
	sb.WriteString("   - use canonical feature directory identifiers such as `0007-catchup-command`\n")

	nextStep := appendNumberedSteps(
		&sb,
		8,
		clarificationLoopSteps(
			goalPct,
			fmt.Sprintf(
				"Reassess, update %s with durable findings, and continue with "+
					"additional batches of up to 10 questions until the specification "+
					"is precise enough to produce a correct, production-quality solution",
				brainstormPath,
			),
		),
	)

	sb.WriteString(fmt.Sprintf(`%d. Keep every finding filepath-specific whenever possible
%d. If you create a tentative plan in chat, fold the durable conclusions back into %s so the file stays current
%d. Stop before implementation. The next workflow step after this research phase is usually kit spec %s

## BRAINSTORM.md Requirements

The final BRAINSTORM.md must be a detailed, informational, filepath-specific document with:
- SUMMARY
- USER THESIS
- RELATIONSHIPS
- CODEBASE FINDINGS
- AFFECTED FILES
- DEPENDENCIES
- QUESTIONS
- OPTIONS
- RECOMMENDED STRATEGY
- NEXT STEP

## Rules

- planning only — no implementation
- no build or execution work intended to advance code changes
- the purpose of this phase is understanding, not code output
- use numbered lists for clarifying questions and progress updates
- continue the clarification loop until confidence reaches ≥%d%% and the specification is precise enough for a correct, production-quality solution
- preserve facts in BRAINSTORM.md, not just in chat
- make the final document dense, explicit, and easy for a coding agent to use when drafting SPEC.md
- keep the ## DEPENDENCIES table aligned with the tools, docs, and design references used during the phase
- keep the ## RELATIONSHIPS section aligned with any explicit dependency on previously shipped features
`, nextStep, nextStep+1, brainstormPath, nextStep+2, featureSlug, goalPct))
	appendNonEmptySectionRules(&sb, "`BRAINSTORM.md`")

	return sb.String()
}

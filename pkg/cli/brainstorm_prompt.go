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
	"github.com/jamesonstone/kit/internal/promptdoc"
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
	notesPath := featureNotesPath(projectRoot, featureNotesDirName(brainstormPath, featureSlug))

	taskSteps := []string{
		"Stay in planning and information-gathering mode only",
		"Do NOT implement code, write production changes, or move into execution",
		"Read CONSTITUTION.md first to understand project constraints and workflow rules",
		"Read the current BRAINSTORM.md template and treat it as the source of truth for this research phase",
		fmt.Sprintf(
			"Inspect the feature notes directory at %s for optional pre-brainstorm inputs:\n"+
				"- ignore `.gitkeep` and empty placeholder files\n"+
				"- if files other than `.gitkeep` exist, read only the notes relevant to the user thesis\n"+
				"- copy durable conclusions into BRAINSTORM.md instead of leaving them only in notes or chat\n"+
				"- record specific note files that shaped the brainstorm in `## DEPENDENCIES` with `Status` = `active`\n"+
				"- leave the notes directory dependency as `optional` when no usable note files exist",
			notesPath,
		),
		relatedFeatureContextStepText(projectRoot, brainstormPath),
		fmt.Sprintf(
			"Research the filtered relevant areas of the codebase at %s to identify the files, patterns, constraints, interfaces, and adjacent workflows that matter to this feature; expand beyond that set only when the evidence requires it",
			projectRoot,
		),
		fmt.Sprintf(
			"Keep the `## DEPENDENCIES` table in %s current throughout the research phase:\n"+
				"- include every dependency that materially shapes the feature definition, such as skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, and assets\n"+
				"- use the columns `Dependency`, `Type`, `Location`, `Used For`, and `Status`\n"+
				"- `Status` must be one of `active`, `optional`, or `stale`\n"+
				"- for Figma or other MCP-driven design dependencies, store the exact design URL or file/node reference in `Location`\n"+
				"- if a dependency influenced decisions but is no longer current, keep it in the table with `Status` = `stale` instead of deleting it",
			brainstormPath,
		),
		fmt.Sprintf(
			"Keep the `## RELATIONSHIPS` section in %s current throughout the research phase:\n"+
				"- use `none` when this feature does not build on an existing feature\n"+
				"- otherwise record one bullet per explicit feature relationship\n"+
				"- supported labels are `builds on`, `depends on`, and `related to`\n"+
				"- use canonical feature directory identifiers such as `0007-catchup-command`",
			brainstormPath,
		),
	}
	taskSteps = append(taskSteps, clarificationLoopSteps(
		goalPct,
		fmt.Sprintf(
			"Reassess, update %s with durable findings, and continue with additional batches of up to 10 questions until the specification is precise enough to produce a correct, production-quality solution",
			brainstormPath,
		),
	)...)
	taskSteps = append(taskSteps,
		"Keep every finding filepath-specific whenever possible",
		fmt.Sprintf(
			"If you create a tentative plan in chat, fold the durable conclusions back into %s so the file stays current",
			brainstormPath,
		),
		fmt.Sprintf(
			"Stop before implementation. The next workflow step after this research phase is usually kit spec %s",
			featureSlug,
		),
	)

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Raw("/plan")
		doc.Paragraph(fmt.Sprintf("You are in planning mode for feature: **%s**", featureSlug))
		doc.Paragraph("You MUST update the brainstorm file at:")
		doc.BulletList(
			fmt.Sprintf("**BRAINSTORM**: %s", brainstormPath),
			fmt.Sprintf("**Feature Notes**: %s", notesPath),
			fmt.Sprintf("**Feature**: %s", featureSlug),
			fmt.Sprintf("**Project Root**: %s", projectRoot),
		)
		doc.Heading(2, "User Thesis")
		doc.Paragraph(thesis)
		doc.Heading(2, "Context Docs (read first)")
		doc.Table(
			[]string{"File", "Purpose"},
			[][]string{
				{"CONSTITUTION", constitutionPath},
				{"BRAINSTORM", brainstormPath},
				{"FEATURE NOTES", notesPath},
				{"Project Root", projectRoot},
			},
		)
		doc.Heading(2, "Your Task")
		doc.OrderedList(1, taskSteps...)
		doc.Heading(2, "BRAINSTORM.md Requirements")
		doc.Paragraph("The final BRAINSTORM.md must be a detailed, informational, filepath-specific document with:")
		doc.BulletList(
			"SUMMARY",
			"USER THESIS",
			"RELATIONSHIPS",
			"CODEBASE FINDINGS",
			"AFFECTED FILES",
			"DEPENDENCIES",
			"QUESTIONS",
			"OPTIONS",
			"RECOMMENDED STRATEGY",
			"NEXT STEP",
		)
		doc.Heading(2, "Rules")
		doc.BulletList(
			"planning only — no implementation",
			"no build or execution work intended to advance code changes",
			"the purpose of this phase is understanding, not code output",
			"use numbered lists for clarifying questions and progress updates",
			fmt.Sprintf(
				"continue the clarification loop until confidence reaches ≥%d%% and the specification is precise enough for a correct, production-quality solution",
				goalPct,
			),
			"preserve facts in BRAINSTORM.md, not just in chat",
			"make the final document dense, explicit, and easy for a coding agent to use when drafting SPEC.md",
			"keep the ## DEPENDENCIES table aligned with the tools, docs, and design references used during the phase",
			"keep the ## RELATIONSHIPS section aligned with any explicit dependency on previously shipped features",
		)
		doc.Raw(renderNonEmptySectionRules("`BRAINSTORM.md`"))
	})
}

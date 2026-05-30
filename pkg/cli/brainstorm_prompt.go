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
	preparedPrompt := prepareAgentPromptForFeature(prompt, feat.Path)

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
	featureDirName := featureNotesDirName(brainstormPath, featureSlug)
	notesPath := featureNotesPath(projectRoot, featureDirName)
	designPath := featureDesignMaterialsPath(projectRoot, featureDirName)
	frontendProfileActive := effectivePromptProfile(filepath.Dir(brainstormPath)) == promptProfileFrontend

	taskSteps := []string{
		"Stay in research and information-gathering workflow only",
		"Do NOT implement code, write production changes, or move into execution",
		"Read CONSTITUTION.md first to understand project constraints and workflow rules",
		"Read the current BRAINSTORM.md template and treat it as the source of truth for this research phase",
		fmt.Sprintf(
			"Inspect the feature notes directory at %s for optional pre-brainstorm inputs:\n"+
				"- ignore `.gitkeep` and empty placeholder files\n"+
				"- if files other than `.gitkeep` exist, read only the notes relevant to the user thesis\n"+
				"- copy durable conclusions into BRAINSTORM.md instead of leaving them only in notes or chat\n"+
				"- record specific note files that shaped the brainstorm in canonical front matter references with `status: active`\n"+
				"- leave the notes directory reference as `optional` when no usable note files exist",
			notesPath,
		),
	}
	if frontendProfileActive {
		taskSteps = append(taskSteps, fmt.Sprintf(
			"Inspect optional frontend design materials at %s just in time:\n"+
				"- ignore `.gitkeep` and empty placeholder files\n"+
				"- list available files only to decide whether any are relevant to the user thesis\n"+
				"- read only the specific screenshots, references, or design notes needed for the current decision\n"+
				"- copy durable design conclusions into BRAINSTORM.md instead of leaving them only in notes or chat\n"+
				"- record specific design files or external design refs that shaped the brainstorm in canonical front matter references with `status: active`\n"+
				"- leave the design materials directory reference as `optional` when no usable design files exist",
			designPath,
		))
	}
	taskSteps = append(taskSteps,
		relatedFeatureContextStepText(projectRoot, brainstormPath),
		fmt.Sprintf(
			"Research the filtered relevant areas of the codebase at %s to identify the files, patterns, constraints, interfaces, and adjacent workflows that matter to this feature; expand beyond that set only when the evidence requires it",
			projectRoot,
		),
		fmt.Sprintf(
			"Keep canonical front matter references in %s current throughout the research phase:\n"+
				"- include every reference that materially shapes the feature definition, such as skills, MCP tools, repo docs, design refs, APIs, libraries, datasets, and assets\n"+
				"- use `name`, `type`, `target`, `relation`, `read_policy`, `used_for`, and `status`\n"+
				"- add a stable `id` when the reference may need to be updated later\n"+
				"- `selector_type` must be one of `artifact`, `heading`, `symbol`, `command`, `url`, or `node_id` when `selector` is set\n"+
				"- `relation` describes the referenced target's role relative to the source artifact, such as `constrains`, `guides`, `informs`, `implements`, `verifies`, or `uses`\n"+
				"- `read_policy` must be one of `must`, `conditional`, `evidence`, or `skip`\n"+
				"- `status` must be one of `active`, `optional`, or `stale`\n"+
				"- for Figma or other MCP-driven design references, store the exact design URL or file/node reference in `target` and use stable selectors when needed\n"+
				"- if a reference influenced decisions but is no longer current, keep it with `status: stale` and `read_policy: skip` instead of deleting it",
			brainstormPath,
		),
		fmt.Sprintf(
			"Keep canonical front matter relationships in %s current throughout the research phase, using the legacy `## RELATIONSHIPS` section only when front matter is absent:\n"+
				"- omit relationships or use `none` only in legacy body metadata when this feature does not build on an existing feature\n"+
				"- otherwise record one entry per explicit feature relationship\n"+
				"- supported front matter types are `builds_on`, `depends_on`, and `related_to`; supported legacy labels are `builds on`, `depends on`, and `related to`\n"+
				"- use canonical feature directory identifiers such as `0007-catchup-command`",
			brainstormPath,
		),
	)
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
			"If you create a tentative approach in chat, fold the durable conclusions back into %s so the file stays current",
			brainstormPath,
		),
		fmt.Sprintf(
			"Stop before implementation. The next workflow step after this research phase is usually kit spec %s",
			featureSlug,
		),
	)

	contextRows := [][]string{
		{"CONSTITUTION", constitutionPath},
		{"BRAINSTORM", brainstormPath},
		{"FEATURE NOTES", notesPath},
	}
	if frontendProfileActive {
		contextRows = append(contextRows, []string{"DESIGN MATERIALS", designPath})
	}
	contextRows = append(contextRows, []string{"Project Root", projectRoot})

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Research and document feature: **%s**", featureSlug))
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
			contextRows,
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
			docsOnlyWorkflowRule("BRAINSTORM.md and supporting documentation"),
			"research and documentation only; no implementation",
			"no build or execution work intended to advance code changes",
			"the purpose of this phase is understanding, not code output",
			"use numbered lists for clarifying questions and progress updates",
			fmt.Sprintf(
				"continue the clarification loop until confidence reaches ≥%d%% and the specification is precise enough for a correct, production-quality solution",
				goalPct,
			),
			"preserve facts in BRAINSTORM.md, not just in chat",
			"make the final document dense, explicit, and easy for a coding agent to use when drafting SPEC.md",
			"keep canonical front matter references aligned with the tools, docs, and design references used during the phase",
			"keep canonical front matter relationships aligned with any explicit dependency on previously shipped features; preserve legacy body fallback only for documents without front matter",
		)
		doc.Raw(renderNonEmptySectionRules("`BRAINSTORM.md`"))
		addFinalResponseContract(doc, brainstormFinalResponseContract(featureSlug)...)
	})
}

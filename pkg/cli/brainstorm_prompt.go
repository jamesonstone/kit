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
		return fmt.Errorf("BRAINSTORM.md not found for %s. Run 'kit legacy brainstorm %s' first", feat.Slug, feat.Slug)
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
		return nil, fmt.Errorf("no features with BRAINSTORM.md available\n\nRun 'kit legacy brainstorm <feature>' first")
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

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Paragraph(fmt.Sprintf("Research feature `%s` and keep `%s` as the durable research record. Do not implement product code.", featureSlug, brainstormPath))
		doc.Heading(2, "User Thesis")
		doc.Paragraph(thesis)

		doc.Heading(2, "Context")
		rows := [][]string{
			{"BRAINSTORM.md", brainstormPath, "Artifact to update"},
			{"Feature notes", notesPath, "Optional user research inputs"},
			{"Constitution", constitutionPath, "Project constraints"},
			{"Project root", projectRoot, "Ground findings in current code/docs"},
		}
		if frontendProfileActive {
			rows = append(rows, []string{"DESIGN MATERIALS", designPath, "Optional relevant screenshots/references"})
		}
		doc.Table([]string{"Input", "Path", "Use"}, rows)

		doc.Heading(2, "Research Contract")
		doc.OrderedList(1,
			"Read repository routing and the Constitution, then inspect only the notes, design materials, code, tests, docs, interfaces, and prior features relevant to the thesis. Ignore placeholders such as .gitkeep.",
			"Ground material claims in exact paths, symbols, commands, URLs, or user decisions. Keep front matter references and feature relationships current; mark optional or stale inputs honestly.",
			"Document the problem, users, constraints, affected files/interfaces, dependencies, viable options, tradeoffs, edge cases, and a recommended strategy without designing implementation code.",
			"Resolve discoverable ambiguity from repository evidence. Ask concise numbered questions only for material non-discoverable choices, with a recommended default and why the answer changes the result.",
			"After each answer or research discovery, update BRAINSTORM.md rather than leaving durable conclusions only in chat. Stop before implementation.",
		)

		doc.Heading(2, "Success Criteria")
		doc.BulletList(
			fmt.Sprintf("Confidence is at least %d, or remaining material questions are explicitly recorded as blockers.", goalPct),
			"BRAINSTORM.md is filepath-specific and contains thesis, findings, affected surfaces, dependencies/references, questions, options, recommendation, and next step.",
			"Research introduces no production changes and does not invent facts, scope, or tool requirements.",
			"Empty optional sections state `not applicable`; placeholder comments are removed.",
		)

		doc.Heading(2, "Output")
		doc.BulletList(
			"Update BRAINSTORM.md and supporting research notes only.",
			fmt.Sprintf("Report the research outcome, key decisions, exact open questions, and next step `kit spec %s`.", featureSlug),
		)
		addFinalResponseContract(doc, brainstormFinalResponseContract(featureSlug)...)
	})
}

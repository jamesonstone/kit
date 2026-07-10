package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

var (
	loopPromptCopy       bool
	loopPromptOutputOnly bool
)

type loopPromptInput struct {
	Title        string
	Source       string
	Scope        string
	Validation   string
	Docs         string
	Delivery     string
	FeatureSlug  string
	FeatureDir   string
	FeaturePhase string
	SpecPath     string
}

func newLoopPromptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "prompt [feature]",
		Short:         "Create a work-to-completion loop prompt",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Create a work-to-completion loop prompt for a coding agent.

With a feature argument, Kit reads the existing SPEC.md and renders a
feature-scoped implementation loop prompt. Without a feature argument, Kit asks
for a small ad hoc intake and renders a flexible loop prompt that is not tied to
any feature directory.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runLoopPrompt,
	}
	cmd.Flags().BoolVar(&loopPromptCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	cmd.Flags().BoolVar(&loopPromptOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	return cmd
}

func runLoopPrompt(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	outputOnly, _ := cmd.Flags().GetBool("output-only")
	if len(args) == 1 {
		prompt, feat, err := buildFeatureLoopPrompt(projectRoot, cfg, args[0])
		if err != nil {
			return err
		}
		if !outputOnly {
			printSectionBanner("🔁", "Feature Loop Prompt")
			fmt.Printf("Feature: %s\n", feat.DirName)
			fmt.Printf("Source of truth: %s\n\n", filepath.Join(feat.Path, "SPEC.md"))
		}
		return outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, loopPromptCopy)
	}

	out := io.Writer(os.Stdout)
	if outputOnly {
		out = os.Stderr
	}
	input, err := readAdHocLoopPromptInput(out, os.Stdin)
	if err != nil {
		return err
	}
	prompt := buildLoopEngineeringPrompt(input)
	if !outputOnly {
		printSectionBanner("🔁", "Ad Hoc Loop Prompt")
	}
	return outputPromptWithClipboardDefault(prompt, outputOnly, loopPromptCopy)
}

func buildFeatureLoopPrompt(projectRoot string, cfg *config.Config, featureRef string) (string, *feature.Feature, error) {
	feat, err := loadFeatureWithState(cfg.SpecsPath(projectRoot), cfg, featureRef)
	if err != nil {
		return "", nil, fmt.Errorf("feature '%s' not found: %w", featureRef, err)
	}
	specPath := filepath.Join(feat.Path, "SPEC.md")
	if !document.Exists(specPath) {
		return "", nil, fmt.Errorf("SPEC.md not found. Run 'kit spec %s' first", feat.Slug)
	}

	input := loopPromptInput{
		Title:        fmt.Sprintf("feature `%s`", feat.Slug),
		Source:       fmt.Sprintf("`%s`", specPath),
		Scope:        "every remaining in-scope task, acceptance criterion, validation obligation, documentation update, reflection gap, and delivery decision in SPEC.md",
		Validation:   "the validation map in SPEC.md, focused checks for changed behavior, relevant regressions, and the full required suite before delivery",
		Docs:         "documentation, API/OpenAPI, capability, README, ruleset, and project-reference updates required by changed behavior",
		Delivery:     "repo-local Kit/GitHub delivery rules after SPEC.md proves implementation, validation, reflection, and documentation are complete",
		FeatureSlug:  feat.Slug,
		FeatureDir:   feat.DirName,
		FeaturePhase: string(feat.Phase),
		SpecPath:     specPath,
	}
	return buildLoopEngineeringPrompt(input), feat, nil
}

func readAdHocLoopPromptInput(out io.Writer, in io.Reader) (loopPromptInput, error) {
	reader := bufio.NewReader(in)
	fmt.Fprintln(out, "Answer the loop-prompt intake. Press Enter to accept a default.")
	fmt.Fprintln(out)

	title, err := readLoopPromptField(reader, out, "Loop goal", "complete the requested work end to end")
	if err != nil {
		return loopPromptInput{}, err
	}
	source, err := readLoopPromptField(reader, out, "Source of truth", "the current user request plus repo-local instructions")
	if err != nil {
		return loopPromptInput{}, err
	}
	scope, err := readLoopPromptField(reader, out, "Scope", "all implementation, validation, review, and delivery steps required by the request")
	if err != nil {
		return loopPromptInput{}, err
	}
	validation, err := readLoopPromptField(reader, out, "Validation", "focused relevant checks, prior-scope regressions, and repo-local required validation")
	if err != nil {
		return loopPromptInput{}, err
	}
	docs, err := readLoopPromptField(reader, out, "Docs/API obligations", "update affected docs and API/OpenAPI docs when behavior changes")
	if err != nil {
		return loopPromptInput{}, err
	}
	delivery, err := readLoopPromptField(reader, out, "Delivery expectation", "follow repo-local GitHub delivery rules only if delivery is requested or already in scope")
	if err != nil {
		return loopPromptInput{}, err
	}

	return loopPromptInput{Title: title, Source: source, Scope: scope, Validation: validation, Docs: docs, Delivery: delivery}, nil
}

func readLoopPromptField(reader *bufio.Reader, out io.Writer, label, defaultValue string) (string, error) {
	fmt.Fprintf(out, "%s [%s]: ", label, defaultValue)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	value := strings.TrimSpace(line)
	if value == "" {
		value = defaultValue
	}
	return value, nil
}

func buildLoopEngineeringPrompt(input loopPromptInput) string {
	return renderPromptDocument(func(doc *promptdoc.Document) {
		title := loopPromptValue(input.Title, "the requested work")
		source := loopPromptValue(input.Source, "the current user request and repo-local instructions")
		scope := loopPromptValue(input.Scope, "all work required by the request")
		validation := loopPromptValue(input.Validation, "focused checks, relevant regressions, and repo-required validation")
		docs := loopPromptValue(input.Docs, "affected documentation when behavior changes")
		delivery := loopPromptValue(input.Delivery, "repo-local delivery rules when delivery is authorized and in scope")

		doc.Paragraph(fmt.Sprintf("Complete %s end to end.", title))
		doc.Heading(2, "Loop Goal")
		doc.BulletList(
			fmt.Sprintf("Treat %s as the implementation source of truth.", source),
			fmt.Sprintf("Finish or explicitly block this scope: %s.", scope),
			"Continue across coherent work slices until success; do not stop after a partial task when safe in-scope work remains.",
		)

		if input.SpecPath != "" {
			doc.Heading(2, "Feature Context")
			doc.BulletList(
				fmt.Sprintf("Feature: `%s`", input.FeatureSlug),
				fmt.Sprintf("Directory: `%s`", input.FeatureDir),
				fmt.Sprintf("Current phase: `%s`", loopPromptValue(input.FeaturePhase, "unknown")),
				fmt.Sprintf("SPEC.md: `%s`", input.SpecPath),
			)
		}

		doc.Heading(2, "Execution Contract")
		doc.OrderedList(1,
			"Read the current requirements, acceptance criteria, task/evidence state, and only the repository context needed for the next decision.",
			"Implement the smallest coherent slice using existing patterns; update focused tests and affected docs with the behavior.",
			fmt.Sprintf("Validate each slice using %s; fix failures and relevant regressions before continuing.", validation),
			"Review the integrated result for correctness, security, reliability, unnecessary scope, and stale contracts.",
			"Record exact files, checks/results, evidence, remaining risk, and durable workflow state, then continue until the success gate passes.",
		)

		doc.Heading(2, "Constraints And Boundaries")
		doc.BulletList(
			"Preserve unrelated or user-owned changes. Safe in-scope discovery and reversible edits need no routine approval; ask only for material non-discoverable choices or authority for external/irreversible action.",
			fmt.Sprintf("Documentation obligation: %s.", docs),
			"Never claim validation or review evidence that was not run or inspected. A skipped check records reason, risk, substitute evidence, and delivery impact.",
			"Use specialist agents only for low-overlap work and keep verification read-only; the supervisor owns integration and final evidence.",
			fmt.Sprintf("Delivery boundary: %s. Before mutation, load repo-local rules, establish the exact delivery contract, review/secret-scan/stage explicit files, and stop on any unknown field.", delivery),
		)

		doc.Heading(2, "Success And Output")
		doc.BulletList(
			"All in-scope acceptance criteria and tasks are complete or have a concrete blocker with owner and impact.",
			"Relevant validation, regression review, docs, evidence, reflection, and authorized delivery state agree with the final diff.",
			"Report outcome first, files changed, exact validation, evidence/delivery state, and only genuine blockers or residual risk.",
		)
	})
}

func loopPromptValue(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

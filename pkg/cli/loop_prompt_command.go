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
		Scope:        "every remaining phase, task-checklist item, acceptance criterion, validation-map obligation, documentation update, reflection gap, and delivery decision recorded in SPEC.md",
		Validation:   "the validation map in SPEC.md, focused checks for changed behavior, relevant prior-phase regressions, and the full relevant suite before delivery",
		Docs:         "all documentation, API/OpenAPI/Swagger, command capability, README, ruleset, and project-reference updates required by the implemented behavior",
		Delivery:     "follow repo-local Kit/GitHub delivery rules only after SPEC.md proves implementation, validation, reflection, and documentation are complete or explicitly gated",
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

	return loopPromptInput{
		Title:      title,
		Source:     source,
		Scope:      scope,
		Validation: validation,
		Docs:       docs,
		Delivery:   delivery,
	}, nil
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
		scope := loopPromptValue(input.Scope, "all implementation, validation, review, and delivery work required by the request")
		validation := loopPromptValue(input.Validation, "focused relevant checks, prior-scope regressions, and repo-local required validation")
		docs := loopPromptValue(input.Docs, "affected docs and API/OpenAPI docs when behavior changes")
		delivery := loopPromptValue(input.Delivery, "repo-local GitHub delivery rules only when delivery is requested or already in scope")

		doc.Paragraph(fmt.Sprintf("Proceed through the full implementation loop for %s.", title))

		doc.Heading(2, "Loop Goal")
		doc.BulletList(
			fmt.Sprintf("Treat %s as the implementation source of truth.", source),
			fmt.Sprintf("Continue until this scope is completed and validated, or explicitly marked blocked/gated with the reason recorded: %s.", scope),
			"Do not stop after an individual phase or task unless ambiguity, missing required source-contract evidence, an unresolvable validation blocker, or required user input prevents safe progress.",
			"If implementation reveals that the source of truth is wrong, stale, contradictory, or incomplete, update the durable source first, then continue only when the path is unambiguous.",
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

		doc.Heading(2, "Phase Execution Cycle")
		doc.Paragraph("For each phase or coherent work slice, repeat this loop:")
		doc.OrderedList(1,
			"Re-read the current requirements, acceptance criteria, validation map, task checklist, phase notes, and relevant evidence.",
			"Identify the concrete implementation scope for the phase and map it to source facts, acceptance criteria, tasks, expected files, and validation.",
			"Implement using existing repository patterns, naming, structure, service boundaries, auth patterns, tests, and documentation conventions.",
			"Add or update focused tests for the behavior changed in the phase, including repository, service, handler/API, integration, CLI, and documentation checks where applicable.",
			fmt.Sprintf("Update documentation obligations for the phase: %s.", docs),
			"Run the phase's relevant validation and fix failures before moving on.",
			"Run relevant regression checks for previously completed phases or related behavior.",
			"Review the completed phase for bugs, omissions, API/docs drift, broken assumptions, hidden scope creep, and repository-pattern violations.",
			"Re-review the immediately previous completed phase to ensure the new phase did not break or contradict it.",
			"Record phase evidence: files changed, tests run, validation results, review findings, remaining gates, blockers, and constraints discovered.",
			"Mark the phase complete only after implementation, focused tests, docs/API updates, validation, and review are complete.",
			"Continue to the next phase only after the phase completion gate passes.",
		)

		doc.Heading(2, "Phase Cycle Outputs")
		doc.BulletList(
			"implementation scope and touched-file mapping",
			"tests and validation commands run, with observed results",
			"documentation/API/OpenAPI/Swagger or explicit no-op decision",
			"regression checks for prior completed phases",
			"review findings and fixes",
			"remaining gates, blockers, assumptions, and owner for each blocker",
			"updated task, acceptance, validation, reflection, evidence, and delivery state in the durable artifact when one exists",
		)

		doc.Heading(2, "Validation And Regression Contract")
		doc.BulletList(
			fmt.Sprintf("Use this validation policy: %s.", validation),
			"Never claim validation passed unless the check actually ran or the evidence was directly inspected.",
			"Fix relevant test, lint, typecheck, build, runtime, API-doc, generated-doc, and review failures before continuing.",
			"Skipped or impossible validation must include reason, risk, substitute evidence, user-visible impact, owner or next action, and whether delivery is blocked.",
			"Any failed, blocked, partially proven, or verifier-disputed acceptance criterion routes back to implementation or clarification before reflection or delivery.",
		)

		doc.Heading(2, "Additional Loop Rules")
		doc.BulletList(
			"Preserve unrelated dirty worktree changes. Do not stage, modify, summarize, or revert unrelated files unless explicitly approved.",
			"Do not introduce new abstractions unless they match existing repository patterns or clearly reduce necessary complexity.",
			"Keep behavior local and reversible until the source of truth explicitly permits external production dispatch or irreversible changes.",
			"Use an accountable supervisor model. Spawn specialist or verification subagents only when the work separates cleanly and file overlap is controlled.",
			"Read-only verification must not edit files, stage changes, commit, push, mutate GitHub state, or mark its own findings closed.",
		)

		doc.Heading(2, "Final Integration Review")
		doc.OrderedList(1,
			"Review all completed phases together for consistency, regressions, missing tests, missing docs/API/OpenAPI updates, and repository-pattern drift.",
			"Run the full relevant validation suite, including repo-local checks required by the workflow.",
			"Review the full diff file by file for correctness, dead code, unnecessary public surfaces, stale docs, secrets, and machine-local config.",
			"Fix any issues found by final review or validation.",
			"Confirm every phase, task-checklist item, acceptance criterion, and validation-map obligation is complete and validated, or explicitly documented as blocked/gated.",
			"Do not mark the work complete while relevant tests or validations are failing unless the user explicitly approves a documented validation exception.",
		)

		doc.Heading(2, "GitHub Delivery Boundary")
		doc.BulletList(
			fmt.Sprintf("Delivery expectation: %s.", delivery),
			"Do not create issue, branch, commit, push, pull request, review-thread, or label mutations until the implementation loop is complete or intentionally gated.",
			"Before any GitHub delivery mutation, load repo-local Kit/GitHub delivery rules and run delivery recon: pwd, git status --short --branch, git remote -v, current branch, default/base branch, active PRs for the current branch, matching issues, and git author/committer identity.",
			"Create or reuse the correct GitHub issue before branching. Use the issue number as the branch name in exact `GH-123` form when repo-local rules require it.",
			"Refresh the base with fetch-only behavior before branching. Branch from the freshly fetched remote base, not from a stale local base.",
			"Review changes file by file before staging. Secret-scan before staging. Stage explicitly with `git add <file>` only.",
			"Review the staged diff before committing. Commit only after self-review confirms no known relevant errors remain.",
			"Push only after the commit is complete, then create or update the pull request using the repository PR template.",
			"Assign the issue and PR to the human user when repo-local rules require it. Create the PR ready for review unless the user explicitly asks for a draft.",
			"Report issue number, branch name, commit hash, PR URL, assignees, verification commands, observed PR/CI state, and residual risk.",
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

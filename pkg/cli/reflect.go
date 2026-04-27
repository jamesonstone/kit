// package cli implements the Kit command-line interface.
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
	"github.com/spf13/cobra"
)

var reflectCopy bool
var reflectOutputOnly bool

var reflectCmd = &cobra.Command{
	Use:   "reflect [feature]",
	Short: "Output reflection and verification instructions",
	Long: `Output instructions for reflecting on recent changes to ensure
implementation correctness.
When a feature is specified, instructions are scoped to that feature's context.
Without a feature argument, shows an interactive selection of features
that have SPEC.md, PLAN.md, and TASKS.md.
The reflection process uses git, lint, and tests to enforce a clean, working state.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReflect,
}

func init() {
	reflectCmd.Flags().BoolVar(&reflectCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	reflectCmd.Flags().BoolVar(&reflectOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	addPromptOnlyFlag(reflectCmd)
	rootCmd.AddCommand(reflectCmd)
}

func runReflect(cmd *cobra.Command, args []string) error {
	outputOnly, _ := cmd.Flags().GetBool("output-only")

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := cfg.SpecsPath(projectRoot)

	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	var feat *feature.Feature

	if len(args) == 1 {
		featureRef := args[0]
		feat, err = loadFeatureWithState(specsDir, cfg, featureRef)
		if err != nil {
			return fmt.Errorf("failed to resolve feature: %w", err)
		}
	} else {
		feat, err = selectFeatureForReflect(specsDir)
		if err != nil {
			return err
		}
		feature.ApplyLifecycleState(feat, cfg)
	}

	brainstormPath := filepath.Join(feat.Path, "BRAINSTORM.md")
	specPath := filepath.Join(feat.Path, "SPEC.md")
	planPath := filepath.Join(feat.Path, "PLAN.md")
	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	wasPaused := feat.Paused
	if err := clearPausedForExplicitResume(projectRoot, cfg, feat); err != nil {
		return err
	}
	if err := updateRollupForResume(projectRoot, cfg, feat.DirName, wasPaused); err != nil {
		return err
	}
	prompt := buildReflectPrompt(projectRoot, constitutionPath, summaryPath, brainstormPath, specPath, planPath, tasksPath, feat.Slug)
	if !outputOnly {
		if wasPaused {
			fmt.Println("  ✓ Cleared paused state")
		}
		printWorkflowInstructions("reflect", []string{
			"if issues remain, return to implement",
			"if clean, mark reflection complete",
		})
	}

	if err := outputPromptForFeatureWithClipboardDefault(prompt, feat.Path, outputOnly, reflectCopy); err != nil {
		return err
	}

	return nil
}

// selectFeatureForReflect shows an interactive numbered list of features
// that have SPEC.md, PLAN.md, and TASKS.md.
func selectFeatureForReflect(specsDir string) (*feature.Feature, error) {
	features, err := feature.ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	var candidates []feature.Feature
	for _, f := range features {
		specPath := filepath.Join(f.Path, "SPEC.md")
		planPath := filepath.Join(f.Path, "PLAN.md")
		tasksPath := filepath.Join(f.Path, "TASKS.md")
		if document.Exists(specPath) && document.Exists(planPath) && document.Exists(tasksPath) {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no features ready for reflection (need SPEC.md + PLAN.md + TASKS.md)\n\nRun 'kit tasks <feature>' to create tasks first")
	}

	printSelectionHeader("Select a feature to reflect on:")
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

// buildReflectPrompt builds the unified reflection prompt.
func buildReflectPrompt(projectRoot, constitutionPath, summaryPath, brainstormPath, specPath, planPath, tasksPath, featureSlug string) string {
	featureScoped := featureSlug != ""
	hasBrainstorm := brainstormPath != "" && document.Exists(brainstormPath)
	cfg, _ := loadRepoInstructionContext(projectRoot)
	repoAgentsPath := repoKnowledgeEntrypointPath(projectRoot, cfg)
	repoReferencesPath := repoReferencesEntrypointPath(projectRoot, cfg)

	goal := "ensure changes are correct, minimal, and consistent"
	if featureScoped {
		goal = "ensure changes match SPEC/PLAN/TASKS and are correct, minimal, and consistent"
	}

	contextDocs := []string{fmt.Sprintf("CONSTITUTION: %s", constitutionPath)}
	if repoAgentsPath != "" {
		contextDocs = append(contextDocs, fmt.Sprintf("AGENTS DOCS: %s", repoAgentsPath))
	}
	if repoReferencesPath != "" {
		contextDocs = append(contextDocs, fmt.Sprintf("REFERENCES: %s", repoReferencesPath))
	}
	contextDocs = append(contextDocs, fmt.Sprintf("PROJECT SUMMARY: %s", summaryPath))
	if featureScoped {
		if hasBrainstorm {
			contextDocs = append(contextDocs, fmt.Sprintf("BRAINSTORM: %s", brainstormPath))
		}
		contextDocs = append(contextDocs,
			fmt.Sprintf("SPEC: %s", specPath),
			fmt.Sprintf("PLAN: %s", planPath),
			fmt.Sprintf("TASKS: %s", tasksPath),
		)
	}

	steps := []string{
		"Snapshot the change set (do not skip)\n- git status\n- git diff\n- git diff --staged\n- git log -n 20 --oneline --decorate",
		"Build a review map\n- list changed files\n- for each file, state the intent in one line\n- identify risk areas (parsing, IO, error handling, concurrency, CLI UX)",
	}
	if featureScoped {
		verifyStep := "Verify correctness against docs\n"
		if hasBrainstorm {
			verifyStep += "- BRAINSTORM: ensure the implementation still aligns with the researched problem framing and identified constraints\n"
		}
		verifyStep += "- SPEC: ensure requirements + acceptance are fully satisfied\n" +
			"- PLAN: ensure decisions were followed\n" +
			"- TASKS: ensure every task marked done is actually done\n" +
			"- ensure no scope creep"
		steps = append(steps, verifyStep)
	} else {
		steps = append(steps, "Verify correctness against docs\n- ensure decisions in code respect CONSTITUTION.md\n- ensure no scope creep")
	}
	steps = append(steps,
		"Quality gates (hard checks)\n- zero-known-defects gate: do not mark reflection complete until all gates pass with evidence\n- required evidence gates: unresolved assumptions = 0; acceptance criteria mapped 1:1 to outputs; build/compile succeeds; lint/typecheck/test failures = 0; unrelated diff scope = 0\n- lint + tests are absolute gates: fix ALL failures before completion, including pre-existing and out-of-scope failures\n- if any gate fails: stop, report the exact failure, and propose the next fix\n- correctness: no panics, no silent failures\n- errors: wrapped/propagated with context, no swallowed errors\n- IO: paths resolved safely, no surprising writes\n- determinism: stable ordering in outputs\n- regression tests: add comprehensive tests for all completed work to prevent future bugs\n  - test happy path, error cases, edge cases, boundary conditions\n  - ensure tests fail without the implementation (tests validate the test itself)\n- docs: update only if behavior changed\n- agent-readability: code optimized for agent understanding and future iteration",
	)
	if featureScoped {
		steps = append(steps, "Correctness checklist\n- [ ] Code compiles without errors\n- [ ] Changes implement the intended task(s)\n- [ ] Implementation matches PLAN.md approach\n- [ ] Requirements from SPEC.md are satisfied\n- [ ] Changes respect CONSTITUTION.md constraints\n- [ ] No syntax errors or typos\n- [ ] Variable and function names are consistent\n- [ ] Imports are correct and used\n- [ ] Error handling is complete\n- [ ] Edge cases from SPEC.md are handled\n- [ ] No debug code or TODOs left behind\n- [ ] Style matches project conventions\n- [ ] Tests added/updated for all completed work\n- [ ] Tests cover happy path, error cases, and edge cases\n- [ ] Tests validate the implementation, not just pass trivially\n- [ ] Test names clearly describe what is being tested\n- [ ] All lint and test failures are fixed, including failures outside the feature scope\n- [ ] Code is written for agent readability and future iteration")
	} else {
		steps = append(steps, "Correctness checklist\n- [ ] Code compiles without errors\n- [ ] No syntax errors or typos\n- [ ] Variable and function names are consistent\n- [ ] Imports are correct and used\n- [ ] Error handling is complete\n- [ ] Edge cases are handled\n- [ ] Changes match stated intent\n- [ ] Changes respect CONSTITUTION.md constraints\n- [ ] No debug code or TODOs left behind\n- [ ] Style matches project conventions\n- [ ] Tests added/updated for all completed work\n- [ ] Tests cover happy path, error cases, and edge cases\n- [ ] All lint and test failures are fixed, including failures outside the immediate scope\n- [ ] Code is written for agent readability")
	}
	steps = append(steps,
		"Agent-optimized code structure\nCode should be built for agent readability and understanding, enabling both current and future agents to:\n- understand intent quickly: clear names, single responsibility, minimal nesting\n- modify safely: explicit error handling, testable design, clear contracts\n- extend effectively: composable pieces, discoverable patterns, good examples\nChecks:\n- [ ] Function/method names clearly describe what they do\n- [ ] Functions have single, well-defined responsibility\n- [ ] Complex logic is broken into named helper functions\n- [ ] Type names and fields describe their purpose\n- [ ] Public interfaces are documented with clear examples\n- [ ] Error paths are explicit, not silent\n- [ ] Dependencies are injected, not hidden in closures\n- [ ] Code avoids clever tricks; readability wins over cleverness\n- [ ] Configuration and magic numbers are named constants\n- [ ] Similar patterns use consistent approaches across codebase",
		"Cleanliness\n- remove dead code\n- remove unused exports and any public surface that is not strictly necessary\n- if an exported symbol is only used locally, reduce its visibility instead of keeping it exported\n- remove debug prints\n- remove unused flags/options\n- keep public surfaces small\n- ensure code is written for agent and human understanding",
	)
	if featureScoped {
		steps = append(steps, "Documentation generation\n- if exists, use the repositories documentation generation tools to update any affected documentation\n- always update affected documentation and ensure all touched documents are current and properly formatted\n- ensure documentation is agent-readable: clear structure, explicit examples, complete contracts\n- document public APIs with examples showing both normal usage and error handling")
	}
	steps = append(steps, "Final pass\n- rerun:\n  - git status\n  - git diff\n  - git diff --staged\n- summarize remaining issues, if any\n- propose next steps")
	if featureScoped {
		steps = append(steps, "Mark reflection complete\n- once all issues are resolved and confidence is 100%\n- append the following marker to the end of TASKS.md:\n  <!-- REFLECTION_COMPLETE -->\n- this marker signals that the feature has completed the full development cycle")
	} else {
		steps = append(steps, "Mark reflection complete (feature-scoped only)\n- if this is a feature-scoped reflection with a TASKS.md file\n- and all issues are resolved with 100% confidence\n- append to TASKS.md: <!-- REFLECTION_COMPLETE -->")
	}

	return renderPromptDocument(func(doc *promptdoc.Document) {
		if featureScoped {
			doc.Heading(2, fmt.Sprintf("Reflection — Feature: %s", featureSlug))
		} else {
			doc.Heading(2, "Reflection")
		}
		doc.Paragraph(fmt.Sprintf("You are in the REFLECT phase for this repo at %s.", projectRoot))
		doc.Paragraph("Goal:")
		doc.BulletList(
			"perform a strict code review of the current change set",
			goal,
		)
		doc.Paragraph("Context docs (read first):")
		doc.BulletList(contextDocs...)
		doc.Paragraph("Steps:")
		doc.OrderedList(1, steps...)
		doc.Paragraph("Output format:")
		outputSections := []string{"CHANGESET\n- files changed: <list>\n- key diffs: <tight bullets>"}
		if featureScoped {
			docTrace := "DOC TRACE"
			if hasBrainstorm {
				docTrace += "\n- BRAINSTORM: pass/fail + notes"
			}
			docTrace += "\n- SPEC: pass/fail + notes\n- PLAN: pass/fail + notes\n- TASKS: pass/fail + notes"
			outputSections = append(outputSections, docTrace)
		}
		outputSections = append(outputSections, "REFLECTION NOTES\n- risks remaining\n- follow-ups")
		doc.OrderedList(1, outputSections...)
		doc.Heading(2, "Rules")
		doc.BulletList(
			"be strict",
			"no fluff",
			`fix issues before reporting them as "known"`,
			"keep diffs minimal",
			"PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times",
		)
	})
}

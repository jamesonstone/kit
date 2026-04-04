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
	summaryPath := filepath.Join(projectRoot, "PROJECT_PROGRESS_SUMMARY.md")
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

	if err := outputPromptWithClipboardDefault(prompt, outputOnly, reflectCopy); err != nil {
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

	var sb strings.Builder
	step := 0
	nextStep := func() int { step++; return step }
	section := byte('A')
	nextSection := func() string { s := string(section); section++; return s }

	if featureScoped {
		sb.WriteString(fmt.Sprintf("## Reflection — Feature: %s\n\n", featureSlug))
	} else {
		sb.WriteString("## Reflection\n\n")
	}

	sb.WriteString(fmt.Sprintf("You are in the REFLECT phase for this repo at %s.\n\nGoal:\n- perform a strict code review of the current change set\n", projectRoot))

	if featureScoped {
		sb.WriteString("- ensure changes match SPEC/PLAN/TASKS and are correct, minimal, and consistent\n")
		sb.WriteString("\nContext docs (read first):\n")
		sb.WriteString(fmt.Sprintf("- CONSTITUTION: %s\n", constitutionPath))
		sb.WriteString(fmt.Sprintf("- PROJECT SUMMARY: %s\n", summaryPath))
		if hasBrainstorm {
			sb.WriteString(fmt.Sprintf("- BRAINSTORM: %s\n", brainstormPath))
		}
		sb.WriteString(fmt.Sprintf("- SPEC: %s\n", specPath))
		sb.WriteString(fmt.Sprintf("- PLAN: %s\n", planPath))
		sb.WriteString(fmt.Sprintf("- TASKS: %s\n", tasksPath))
	} else {
		sb.WriteString("- ensure changes are correct, minimal, and consistent\n")
		sb.WriteString(fmt.Sprintf("\nContext docs (read first):\n- CONSTITUTION: %s\n- PROJECT SUMMARY: %s\n", constitutionPath, summaryPath))
	}

	sb.WriteString("\nSteps:\n")

	sb.WriteString(fmt.Sprintf(`
%d) Snapshot the change set (do not skip)
- git status
- git diff
- git diff --staged
- git log -n 20 --oneline --decorate
`, nextStep()))

	sb.WriteString(fmt.Sprintf(`
%d) Build a review map
- list changed files
- for each file, state the intent in one line
- identify risk areas (parsing, IO, error handling, concurrency, CLI UX)
`, nextStep()))

	if featureScoped {
		sb.WriteString(fmt.Sprintf("\n%d) Verify correctness against docs\n", nextStep()))
		if hasBrainstorm {
			sb.WriteString("- BRAINSTORM: ensure the implementation still aligns with the researched problem framing and identified constraints\n")
		}
		sb.WriteString("- SPEC: ensure requirements + acceptance are fully satisfied\n")
		sb.WriteString("- PLAN: ensure decisions were followed\n")
		sb.WriteString("- TASKS: ensure every task marked done is actually done\n")
		sb.WriteString("- ensure no scope creep\n")
	} else {
		sb.WriteString(fmt.Sprintf(`
%d) Verify correctness against docs
- ensure decisions in code respect CONSTITUTION.md
- ensure no scope creep
`, nextStep()))
	}

	sb.WriteString(fmt.Sprintf(`
%d) Quality gates (hard checks)
- zero-known-defects gate: do not mark reflection complete until all gates pass with evidence
- required evidence gates: unresolved assumptions = 0; acceptance criteria mapped 1:1 to outputs; build/compile succeeds; lint/typecheck/test failures = 0; unrelated diff scope = 0
- lint + tests are absolute gates: fix ALL failures before completion, including pre-existing and out-of-scope failures
- if any gate fails: stop, report the exact failure, and propose the next fix
- correctness: no panics, no silent failures
- errors: wrapped/propagated with context, no swallowed errors
- IO: paths resolved safely, no surprising writes
- determinism: stable ordering in outputs
- regression tests: add comprehensive tests for all completed work to prevent future bugs
  - test happy path, error cases, edge cases, boundary conditions
  - ensure tests fail without the implementation (tests validate the test itself)
- docs: update only if behavior changed
- agent-readability: code optimized for agent understanding and future iteration
`, nextStep()))

	if featureScoped {
		sb.WriteString(fmt.Sprintf(`
%d) Correctness checklist
- [ ] Code compiles without errors
- [ ] Changes implement the intended task(s)
- [ ] Implementation matches PLAN.md approach
- [ ] Requirements from SPEC.md are satisfied
- [ ] Changes respect CONSTITUTION.md constraints
- [ ] No syntax errors or typos
- [ ] Variable and function names are consistent
- [ ] Imports are correct and used
- [ ] Error handling is complete
- [ ] Edge cases from SPEC.md are handled
- [ ] No debug code or TODOs left behind
- [ ] Style matches project conventions
- [ ] Tests added/updated for all completed work
- [ ] Tests cover happy path, error cases, and edge cases
- [ ] Tests validate the implementation, not just pass trivially
- [ ] Test names clearly describe what is being tested
- [ ] All lint and test failures are fixed, including failures outside the feature scope
- [ ] Code is written for agent readability and future iteration
`, nextStep()))
	} else {
		sb.WriteString(fmt.Sprintf(`
%d) Correctness checklist
- [ ] Code compiles without errors
- [ ] No syntax errors or typos
- [ ] Variable and function names are consistent
- [ ] Imports are correct and used
- [ ] Error handling is complete
- [ ] Edge cases are handled
- [ ] Changes match stated intent
- [ ] Changes respect CONSTITUTION.md constraints
- [ ] No debug code or TODOs left behind
- [ ] Style matches project conventions
- [ ] Tests added/updated for all completed work
- [ ] Tests cover happy path, error cases, and edge cases
- [ ] All lint and test failures are fixed, including failures outside the immediate scope
- [ ] Code is written for agent readability
`, nextStep()))
	}

	sb.WriteString(fmt.Sprintf(`
%d) Agent-optimized code structure
Code should be built for agent readability and understanding, enabling both current and future agents to:
- understand intent quickly: clear names, single responsibility, minimal nesting
- modify safely: explicit error handling, testable design, clear contracts
- extend effectively: composable pieces, discoverable patterns, good examples
Checks:
- [ ] Function/method names clearly describe what they do
- [ ] Functions have single, well-defined responsibility
- [ ] Complex logic is broken into named helper functions
- [ ] Type names and fields describe their purpose
- [ ] Public interfaces are documented with clear examples
- [ ] Error paths are explicit, not silent
- [ ] Dependencies are injected, not hidden in closures
- [ ] Code avoids clever tricks; readability wins over cleverness
- [ ] Configuration and magic numbers are named constants
- [ ] Similar patterns use consistent approaches across codebase
`, nextStep()))

	sb.WriteString(fmt.Sprintf(`
%d) Cleanliness
- remove dead code
- remove debug prints
- remove unused flags/options
- keep public surfaces small
- ensure code is written for agent and human understanding
`, nextStep()))

	if featureScoped {
		sb.WriteString(fmt.Sprintf(`
%d) Documentation generation
- if exists, use the repositories documentation generation tools to update any affected documentation
- ensure documentation is agent-readable: clear structure, explicit examples, complete contracts
- document public APIs with examples showing both normal usage and error handling
`, nextStep()))
	}

	sb.WriteString(fmt.Sprintf(`
%d) Final pass
- rerun:
  - git status
  - git diff
  - git diff --staged
- summarize remaining issues, if any
- propose next steps
`, nextStep()))

	if featureScoped {
		sb.WriteString(fmt.Sprintf(`
%d) Mark reflection complete
- once all issues are resolved and confidence is 100%%
- append the following marker to the end of TASKS.md:
  <!-- REFLECTION_COMPLETE -->
- this marker signals that the feature has completed the full development cycle
`, nextStep()))
	} else {
		sb.WriteString(fmt.Sprintf(`
%d) Mark reflection complete (feature-scoped only)
- if this is a feature-scoped reflection with a TASKS.md file
- and all issues are resolved with 100%% confidence
- append to TASKS.md: <!-- REFLECTION_COMPLETE -->
`, nextStep()))
	}

	sb.WriteString(fmt.Sprintf(`
Output format:

%s) CHANGESET
- files changed: <list>
- key diffs: <tight bullets>
`, nextSection()))

	if featureScoped {
		sb.WriteString(fmt.Sprintf("\n%s) DOC TRACE\n", nextSection()))
		if hasBrainstorm {
			sb.WriteString("- BRAINSTORM: pass/fail + notes\n")
		}
		sb.WriteString("- SPEC: pass/fail + notes\n")
		sb.WriteString("- PLAN: pass/fail + notes\n")
		sb.WriteString("- TASKS: pass/fail + notes\n")
	}

	sb.WriteString(fmt.Sprintf(`
%s) REFLECTION NOTES
- risks remaining
- follow-ups

Rules:
- be strict
- no fluff
- fix issues before reporting them as "known"
- keep diffs minimal
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times
`, nextSection()))

	return sb.String()
}

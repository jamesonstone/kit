// package cli implements the Kit command-line interface.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/spf13/cobra"
)

var reflectCopy bool

var reflectCmd = &cobra.Command{
	Use:   "reflect [feature]",
	Short: "Output reflection and verification instructions",
	Long: `Output instructions for reflecting on recent changes to ensure
implementation correctness.

When a feature is specified, instructions are scoped to that feature's context.
Without a feature argument, outputs generic verification instructions.

The reflection process uses git to analyze changes and optionally runs coderabbit
for additional validation.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReflect,
}

func init() {
	reflectCmd.Flags().Bool("no-coderabbit", false, "skip CodeRabbit config creation and instructions")
	reflectCmd.Flags().BoolVar(&reflectCopy, "copy", false, "copy agent prompt to clipboard (suppresses stdout)")
	rootCmd.AddCommand(reflectCmd)
}

func runReflect(cmd *cobra.Command, args []string) error {
	noCodeRabbit, _ := cmd.Flags().GetBool("no-coderabbit")

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	if !noCodeRabbit {
		ensureCodeRabbitConfig(projectRoot)
	}

	constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
	summaryPath := filepath.Join(projectRoot, "PROJECT_PROGRESS_SUMMARY.md")

	var prompt string

	if len(args) == 1 {
		featureRef := args[0]

		cfg, err := config.Load(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		specsDir := cfg.SpecsPath(projectRoot)
		feat, err := feature.Resolve(specsDir, featureRef)
		if err != nil {
			return fmt.Errorf("failed to resolve feature: %w", err)
		}

		specPath := filepath.Join(feat.Path, "SPEC.md")
		planPath := filepath.Join(feat.Path, "PLAN.md")
		tasksPath := filepath.Join(feat.Path, "TASKS.md")
		prompt = buildReflectPrompt(projectRoot, constitutionPath, summaryPath, specPath, planPath, tasksPath, feat.Slug, noCodeRabbit)
	} else {
		prompt = buildReflectPrompt(projectRoot, constitutionPath, summaryPath, "", "", "", "", noCodeRabbit)
	}

	if reflectCopy {
		if err := copyToClipboard(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("✓ Copied agent prompt to clipboard")
		return nil
	}

	fmt.Print(prompt)
	return nil
}

// buildReflectPrompt builds the unified reflection prompt.
func buildReflectPrompt(projectRoot, constitutionPath, summaryPath, specPath, planPath, tasksPath, featureSlug string, noCodeRabbit bool) string {
	featureScoped := featureSlug != ""

	var sb strings.Builder
	step := 0
	nextStep := func() int { step++; return step }
	section := byte('A')
	nextSection := func() string { s := string(section); section++; return s }

	// header
	if featureScoped {
		sb.WriteString(fmt.Sprintf("## Reflection — Feature: %s\n\n", featureSlug))
	} else {
		sb.WriteString("## Reflection\n\n")
	}

	// goal
	goalExtra := ""
	if !noCodeRabbit {
		goalExtra = "\n- run CodeRabbit in prompt-only mode and address all findings"
	}
	sb.WriteString(fmt.Sprintf("You are in the REFLECT phase for this repo at %s.\n\nGoal:\n- perform a strict code review of the current change set%s\n", projectRoot, goalExtra))

	if featureScoped {
		sb.WriteString("- ensure changes match SPEC/PLAN/TASKS and are correct, minimal, and consistent\n")
		sb.WriteString(fmt.Sprintf(`
Context docs (read first):
- CONSTITUTION: %s
- PROJECT SUMMARY: %s
- SPEC: %s
- PLAN: %s
- TASKS: %s
`, constitutionPath, summaryPath, specPath, planPath, tasksPath))
	} else {
		sb.WriteString("- ensure changes are correct, minimal, and consistent\n")
		sb.WriteString(fmt.Sprintf(`
Context docs (read first):
- CONSTITUTION: %s
- PROJECT SUMMARY: %s
`, constitutionPath, summaryPath))
	}

	// steps
	sb.WriteString("\nSteps:\n")

	// snapshot
	sb.WriteString(fmt.Sprintf(`
%d) Snapshot the change set (do not skip)
- git status
- git diff
- git diff --staged
- git log -n 20 --oneline --decorate
`, nextStep()))

	// review map
	sb.WriteString(fmt.Sprintf(`
%d) Build a review map
- list changed files
- for each file, state the intent in one line
- identify risk areas (parsing, IO, error handling, concurrency, CLI UX)
`, nextStep()))

	// coderabbit (optional)
	if !noCodeRabbit {
		sb.WriteString(fmt.Sprintf(`
%d) Run CodeRabbit (prompt-only)
- coderabbit --prompt-only
- treat the output as review findings
- fix all issues that are valid
- if you disagree with a finding, document why in a short bullet under REFLECTION NOTES (below)
`, nextStep()))
	}

	// verify correctness against docs
	if featureScoped {
		sb.WriteString(fmt.Sprintf(`
%d) Verify correctness against docs
- SPEC: ensure requirements + acceptance are fully satisfied
- PLAN: ensure decisions were followed
- TASKS: ensure every task marked done is actually done
- ensure no scope creep
`, nextStep()))
	} else {
		sb.WriteString(fmt.Sprintf(`
%d) Verify correctness against docs
- ensure decisions in code respect CONSTITUTION.md
- ensure no scope creep
`, nextStep()))
	}

	// quality gates
	sb.WriteString(fmt.Sprintf(`
%d) Quality gates (hard checks)
- correctness: no panics, no silent failures
- errors: wrapped/propagated with context, no swallowed errors
- IO: paths resolved safely, no surprising writes
- determinism: stable ordering in outputs
- tests: add or update only what is required to prove correctness
- docs: update only if behavior changed
`, nextStep()))

	// correctness checklist
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
`, nextStep()))
	}

	// cleanliness
	sb.WriteString(fmt.Sprintf(`
%d) Cleanliness
- remove dead code
- remove debug prints
- remove unused flags/options
- keep public surfaces small
`, nextStep()))

	// documentation generation (feature-scoped only)
	if featureScoped {
		sb.WriteString(fmt.Sprintf(`
%d) Documentation generation
- if exists, use the repositories documentation generation tools to update any affected documentation
`, nextStep()))
	}

	// final pass
	sb.WriteString(fmt.Sprintf(`
%d) Final pass
- rerun:
  - git status
  - git diff
  - git diff --staged
- summarize remaining issues, if any
- propose next steps
`, nextStep()))

	// mark reflection complete
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

	// output format
	sb.WriteString(fmt.Sprintf(`
Output format:

%s) CHANGESET
- files changed: <list>
- key diffs: <tight bullets>
`, nextSection()))

	if !noCodeRabbit {
		sb.WriteString(fmt.Sprintf(`
%s) CODERABBIT FINDINGS
- accepted + fixed: <list>
- rejected: <list with reason>
`, nextSection()))
	}

	if featureScoped {
		sb.WriteString(fmt.Sprintf(`
%s) DOC TRACE
- SPEC: pass/fail + notes
- PLAN: pass/fail + notes
- TASKS: pass/fail + notes
`, nextSection()))
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

// ensureCodeRabbitConfig creates .coderabbit.yaml if it doesn't exist.
func ensureCodeRabbitConfig(projectRoot string) {
	configPath := filepath.Join(projectRoot, ".coderabbit.yaml")
	if _, err := os.Stat(configPath); err == nil {
		// file exists, nothing to do
		return
	}

	const coderabbitConfig = `# yaml-language-server: $schema=https://coderabbit.ai/integrations/schema.v2.json
language: "en-US"
reviews:
  profile: "assertive"
  high_level_summary: true
  collapse_walkthrough: true
  path_filters:
    - "!docs/**"
    - "!.specify/**"
    - "!**/*.md"
    - "!**/mock-data/**"
  auto_review:
    enabled: true
    drafts: true
chat:
  auto_reply: true
`

	if err := os.WriteFile(configPath, []byte(coderabbitConfig), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not create .coderabbit.yaml: %v\n", err)
		return
	}
	fmt.Printf("%s✓ Created .coderabbit.yaml%s\n", plan, reset)
}

// package cli implements the Kit command-line interface.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/spf13/cobra"
)

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

	instructions := genericReflectInstructions(noCodeRabbit)

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

		instructions = featureScopedReflectInstructions(feat.Slug, feat.Path, noCodeRabbit)

		fmt.Println(instructions)

		// output easy-to-copy instruction for coding agents
		constitutionPath := filepath.Join(projectRoot, "docs", "CONSTITUTION.md")
		summaryPath := filepath.Join(projectRoot, "PROJECT_PROGRESS_SUMMARY.md")
		specPath := filepath.Join(feat.Path, "SPEC.md")
		planPath := filepath.Join(feat.Path, "PLAN.md")
		tasksPath := filepath.Join(feat.Path, "TASKS.md")

		// build conditional CodeRabbit content
		goalExtra := ""
		coderabbitStep := ""
		coderabbitOutput := ""
		if !noCodeRabbit {
			goalExtra = "\n- run CodeRabbit in prompt-only mode and address all findings"
			coderabbitStep = `
3) Run CodeRabbit (prompt-only)
- coderabbit --prompt-only
- treat the output as review findings
- fix all issues that are valid
- if you disagree with a finding, document why in a short bullet under REFLECTION NOTES (below)
`
			coderabbitOutput = `B) CODERABBIT FINDINGS
- accepted + fixed: <list>
- rejected: <list with reason>

`
		}

		fmt.Println("\n" + dim + "────────────────────────────────────────────────────────────────────────" + reset)
		fmt.Println(whiteBold + "Copy this prompt to your coding agent:" + reset)
		fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
		fmt.Printf(`
You are in the REFLECT phase for this repo at %s.

Goal:
- perform a strict code review of the current change set%s
- ensure changes match SPEC/PLAN/TASKS and are correct, minimal, and consistent

Context docs (read first):
- CONSTITUTION: %s
- PROJECT SUMMARY: %s
- SPEC: %s
- PLAN: %s
- TASKS: %s

Steps:

1) Snapshot the change set (do not skip)
- git status
- git diff
- git diff --staged
- git log -n 20 --oneline --decorate

2) Build a review map
- list changed files
- for each file, state the intent in one line
- identify risk areas (parsing, IO, error handling, concurrency, CLI UX)
%s
3) Verify correctness against docs
- SPEC: ensure requirements + acceptance are fully satisfied
- PLAN: ensure decisions were followed
- TASKS: ensure every task marked done is actually done
- ensure no scope creep

4) Quality gates (hard checks)
- correctness: no panics, no silent failures
- errors: wrapped/propagated with context, no swallowed errors
- IO: paths resolved safely, no surprising writes
- determinism: stable ordering in outputs (rollup tables, etc.)
- tests: add or update only what is required to prove correctness
- docs: update only if behavior changed

5) Cleanliness
- remove dead code
- remove debug prints
- remove unused flags/options
- keep public surfaces small

6) Documentation Generation
- if exists, use the repositories documentation generation tools to update any affected documentation

7) Final pass
- rerun:
  - git status
  - git diff
  - git diff --staged
- summarize remaining issues, if any
- propose next steps

Output format:

A) CHANGESET
- files changed: <list>
- key diffs: <tight bullets>

%sB) DOC TRACE
- SPEC: pass/fail + notes
- PLAN: pass/fail + notes
- TASKS: pass/fail + notes

C) REFLECTION NOTES
- risks remaining
- follow-ups

Rules:
- be strict
- no fluff
- fix issues before reporting them as "known"
- keep diffs minimal
- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

`, projectRoot, goalExtra, constitutionPath, summaryPath, specPath, planPath, tasksPath, coderabbitStep, coderabbitOutput)
		fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

		return nil
	}

	fmt.Println(instructions)
	return nil
}

func genericReflectInstructions(noCodeRabbit bool) string {
	coderabbitStep := `
### Step 5: Run CodeRabbit Validation
Once you reach 100% confidence in correctness:
` + "```bash" + `
coderabbit --prompt-only
` + "```" + `

Fix any issues found by this command before considering the reflection complete.
`
	if noCodeRabbit {
		coderabbitStep = ""
	}

	return `## Reflection Instructions

Reflect on all recent changes to ensure 100% implementation correctness.

**Important**: Before reviewing changes, ensure you've read docs/CONSTITUTION.md for project-wide constraints and principles.

### Step 1: Analyze Git State
Review staged and unstaged changes:
` + "```bash" + `
git status
git diff --stat
git diff
git diff --cached
` + "```" + `

### Step 2: Understand the Delta
For each changed file:
- What was the intent of the change?
- Does the change align with project architecture?
- Are there unintended side effects?
- Is error handling complete?
- Are edge cases covered?

### Step 3: Cross-Reference Context
Combine change analysis with:
- CONSTITUTION.md — project-wide constraints and principles
- Repository structure and conventions
- Related files that may need updates
- Test files that should cover the changes
- Documentation that may need updates

### Step 4: Verify Correctness Checklist
Confirm each item before proceeding:
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
` + coderabbitStep + `
### Output
After reflection, provide:
1. **Summary**: one-sentence description of changes
2. **Files Changed**: list with brief rationale for each
3. **Confidence Level**: percentage (target: 100%)
4. **Issues Found**: any problems discovered and how they were resolved
5. **Remaining Concerns**: anything that needs human review`
}

func featureScopedReflectInstructions(featureSlug, featurePath string, noCodeRabbit bool) string {
	coderabbitStep := `
### Step 5: Run CodeRabbit Validation
Once you reach 100%% confidence in correctness:
` + "```bash" + `
coderabbit --prompt-only
` + "```" + `

Fix any issues found by this command before considering the reflection complete.
`
	if noCodeRabbit {
		coderabbitStep = ""
	}

	return fmt.Sprintf(`## Reflection Instructions — Feature: %s

Reflect on all recent changes for feature **%s** to ensure 100%%%% implementation correctness.

### Required Reading
Verify changes align with:
- docs/CONSTITUTION.md — project-wide constraints and principles
- %s/SPEC.md — requirements and acceptance criteria
- %s/PLAN.md — implementation approach
- %s/TASKS.md — task definitions and dependencies

### Step 1: Analyze Git State
Review staged and unstaged changes:
`+"```bash"+`
git status
git diff --stat
git diff
git diff --cached
`+"```"+`

### Step 2: Understand the Delta
For each changed file:
- Does it implement a task from TASKS.md?
- Does it follow the approach defined in PLAN.md?
- Does it satisfy requirements from SPEC.md?
- Are there unintended side effects?
- Is error handling complete?

### Step 3: Cross-Reference Feature Context
Verify changes against:
- Feature specification requirements
- Implementation plan components
- Task dependencies and order
- Acceptance criteria from SPEC.md

### Step 4: Verify Correctness Checklist
Confirm each item before proceeding:
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
%s
### Output
After reflection, provide:
1. **Feature**: %s
2. **Tasks Completed**: which TASKS.md items are done
3. **Summary**: one-sentence description of changes
4. **Files Changed**: list with brief rationale for each
5. **Spec Compliance**: which SPEC.md requirements are now satisfied
6. **Confidence Level**: percentage (target: 100%%%%)  
7. **Issues Found**: any problems discovered and how they were resolved
8. **Remaining Concerns**: anything that needs human review`, featureSlug, featureSlug,
		featurePath, featurePath, featurePath, coderabbitStep, featureSlug)
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

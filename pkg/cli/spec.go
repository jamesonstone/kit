package cli

import (
	"github.com/spf13/cobra"
)

var specCopy bool

var specEditor string

var specInline bool

var specOutputOnly bool

var specReviseThesis bool

var specUseVim bool

var promptSpecFeatureRef = readSpecFeatureRef

var promptSpecSetupGate = readSpecSetupGateDecision

var specCmd = &cobra.Command{
	Use:   "spec [feature]",
	Short: "Start or resume the Kit v2 SPEC.md workflow",
	Long: `Start or resume Kit v2 feature work from one durable SPEC.md.

🧭 Human flow
  1. Pick or provide a feature slug/name.
  2. Enter one thesis/goal in an editor.
  3. Choose delivery intent: no, yes, or continue.
  4. Paste the copied v2 supervisor prompt into your coding agent.

🧠 Agent workflow
  idea → clarification loop → agent-team implementation → reflection →
  validation/verification → evidence + delivery gate

📦 What Kit writes
  - docs/specs/<feature>/SPEC.md as the single durable v2 feature artifact
  - docs/notes/<feature>/ reference-material directories for supporting inputs
  - PROJECT_PROGRESS_SUMMARY.md after creation or adoption

🧱 Setup gate
  Before writing feature artifacts, Kit checks whether project setup appears
  complete. If .kit.yaml, docs/CONSTITUTION.md, or required instruction docs
  are missing or the Constitution still looks like an unfilled starter, you
  can continue into the spec or copy the kit init prompt and stop.

🔁 Modes
  New SPEC.md       One thesis/goal entry + delivery intent, then prompt output
  Existing SPEC.md  Preserve content and regenerate/copy the supervisor prompt
  --revise-thesis   Append a dated thesis note; never silently replace the old one
  --prompt-only     Read existing SPEC.md and print/copy the prompt without writes

🧱 The generated prompt is the v2 supervisor contract. It keeps ideation,
clarification, implementation planning, task tracking, implementation,
reflection, validation/verification, documentation updates, and delivery
gating inside SPEC.md. It does not require BRAINSTORM.md, PLAN.md, TASKS.md,
implement, reflect, or standalone verification commands in the normal v2 path.

🚫 Git/GitHub safety
  kit spec records delivery intent only. It does not create issues, branches,
  commits, pushes, pull requests, or review-thread mutations.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSpec,
}

func init() {
	addFreeTextInputFlags(specCmd, &specUseVim, &specEditor)
	addInlineTextInputFlag(specCmd, &specInline)
	specCmd.Flags().Bool("template", false, "(deprecated) output empty template and prompt without interactive questions")
	specCmd.Flags().Bool("interactive", false, "prompt user for spec details interactively")
	specCmd.Flags().BoolVar(&specCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	specCmd.Flags().BoolVar(&specOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	specCmd.Flags().BoolVar(&specReviseThesis, "revise-thesis", false, "append a dated thesis note and refresh delivery intent before prompt output")
	addPromptOnlyFlag(specCmd)
	_ = specCmd.Flags().MarkHidden("template")
	_ = specCmd.Flags().MarkHidden("interactive")
	rootCmd.AddCommand(specCmd)
}

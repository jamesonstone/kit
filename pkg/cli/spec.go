package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/rollup"
	"github.com/jamesonstone/kit/internal/templates"
)

var specCopy bool
var specEditor string
var specInline bool
var specOutputOnly bool
var specReviseThesis bool
var specUseVim bool
var specLegacySupervisor bool

var promptSpecFeatureRef = readSpecFeatureRef
var promptSpecSetupGate = readSpecSetupGateDecision

var specCmd = &cobra.Command{
	Use:   "spec [feature]",
	Short: "Scaffold or orient a living feature specification",
	Long: `Scaffold, adopt, or orient repository-backed feature memory.

Native agent planning owns research, clarification, design, and implementation
planning. Kit creates the durable place where the accepted plan, material
decisions, discoveries, validation, outcome, and repository-memory curation
survive after the agent session.

Plain kit spec is non-interactive. Provide a feature name, then use your host
agent's native planning capability (for example /plan in Codex). When the work
contains material rationale, capture the accepted plan in SPEC.md before code.

Existing V1 and V2 specs are preserved. Use --legacy-supervisor temporarily to
run the deprecated V2 lifecycle supervisor.

Compatibility supervisor details:

🧭 Human flow
  1. Pick or provide a feature slug/name.
  2. Enter one thesis/goal in an editor.
  3. Choose delivery intent: 1 creates a later issue/branch/PR, 2 captures only, 3 continues current work.
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
	specCmd.Flags().BoolVar(&specLegacySupervisor, "legacy-supervisor", false, "run the deprecated V2 lifecycle supervisor")
	addPromptOnlyFlag(specCmd)
	_ = specCmd.Flags().MarkHidden("template")
	_ = specCmd.Flags().MarkHidden("interactive")
	rootCmd.AddCommand(specCmd)
}

func runSpec(cmd *cobra.Command, args []string) error {
	if specLegacySupervisor || specUsesLegacySupervisorFlags(cmd) {
		if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "Warning: the V2 compatibility lifecycle supervisor is deprecated; use native agent planning with plain `kit spec <feature>`."); err != nil {
			return err
		}
		return runLegacySpecSupervisor(cmd, args)
	}
	return runNativePlanSpec(cmd, args)
}

func specUsesLegacySupervisorFlags(cmd *cobra.Command) bool {
	for _, name := range []string{"template", "interactive", "copy", "output-only", "revise-thesis", "prompt-only", "vim", "editor", "inline"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func runNativePlanSpec(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("feature name required for non-interactive specification scaffolding: use `kit spec <feature>`")
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("Kit project not initialized: run `kit init` before `kit spec`: %w", err)
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}
	specsDir := cfg.SpecsPath(projectRoot)
	if err := ensureDir(specsDir); err != nil {
		return err
	}

	feat, created, err := feature.EnsureExists(cfg, projectRoot, specsDir, args[0])
	if err != nil {
		return err
	}
	feature.ApplyLifecycleState(feat, cfg)

	specPath := filepath.Join(feat.Path, "SPEC.md")
	specCreated := false
	if !document.Exists(specPath) {
		content := templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir(feat.DirName))
		if err := document.Write(specPath, content); err != nil {
			return fmt.Errorf("create SPEC.md: %w", err)
		}
		specCreated = true
	}
	_, notesRelPath, err := ensureFeatureNotesDir(projectRoot, feat.DirName)
	if err != nil {
		return err
	}

	doc, err := document.ParseFile(specPath, document.TypeSpec)
	if err != nil {
		return err
	}
	workflowVersion := 0
	if doc.Metadata != nil {
		workflowVersion = doc.Metadata.WorkflowVersion
	}
	if workflowVersion == document.WorkflowVersionV3 {
		updated, changed, updateErr := document.UpsertMetadata(doc.Content, document.TypeSpec, document.MetadataUpsert{
			Feature:         document.FeatureMetadataFromDir(feat.DirName),
			WorkflowVersion: document.WorkflowVersionV3,
			Phase:           doc.Metadata.Phase,
			DeliveryIntent:  doc.Metadata.DeliveryIntent,
			References:      referencesForMetadataUpsert(doc.Content, document.TypeSpec, []document.MetadataReference{featureNotesReference(notesRelPath)}),
		})
		if updateErr != nil {
			return updateErr
		}
		if changed {
			if err := document.Write(specPath, updated); err != nil {
				return err
			}
		}
		if effectivePromptProfile(feat.Path) == promptProfileFrontend {
			if _, _, err := ensureFeatureDesignMaterialsDirs(projectRoot, feat.DirName); err != nil {
				return err
			}
			if _, err := ensureFrontendProfileDependencyRows(specPath, document.TypeSpec, feat.DirName); err != nil {
				return err
			}
		}
	}

	if err := rollup.Update(projectRoot, cfg); err != nil {
		return fmt.Errorf("update PROJECT_PROGRESS_SUMMARY.md: %w", err)
	}

	out := cmd.OutOrStdout()
	action := "adopted"
	if created || specCreated {
		action = "created"
	}
	if _, err := fmt.Fprintf(out, "Specification %s: %s\n", action, displayProjectRelativePath(projectRoot, specPath)); err != nil {
		return err
	}
	if workflowVersion != document.WorkflowVersionV3 {
		if _, err := fmt.Fprintf(out, "Compatibility: preserved workflow_version %d without rewriting it. V2-to-V3 migration requires semantic curation.\n", workflowVersion); err != nil {
			return err
		}
	}
	if _, err = fmt.Fprintf(out, "Next: use native agent planning (for example `/plan` in Codex). If material rationale exists, capture the accepted plan in SPEC.md before implementation, keep consequential decisions current, and curate repository memory after validation.\nNotes: %s\n", notesRelPath); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "Managed guidance: run `kit status` and follow any Kit-managed refresh action before implementation.")
	return err
}

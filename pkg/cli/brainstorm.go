package cli

import (
	"github.com/spf13/cobra"
)

var (
	brainstormCopy       bool
	brainstormBacklog    bool
	brainstormInline     bool
	brainstormEditor     string
	brainstormOutput     string
	brainstormOutputOnly bool
	brainstormPrepare    bool
	brainstormUseVim     bool
)

var brainstormCmd = &cobra.Command{
	Use:   "brainstorm [feature]",
	Short: "Deprecated v1 staged workflow: create BRAINSTORM.md or backlog research",
	Long: `Deprecated v1 staged workflow: create or update a feature's BRAINSTORM.md document and output a
research and documentation prompt for a coding agent.

The default v2 feature workflow starts with kit spec <feature>. Use brainstorm
when intentionally working in the legacy staged artifact flow or capturing a
deferred backlog research item.

Creates:
	- Feature directory (e.g., docs/specs/0001-my-feature/)
	- Feature notes directory (e.g., docs/notes/0001-my-feature/.gitkeep)
	- BRAINSTORM.md as the first feature-scoped artifact

Interactive flow:
	1. Ask for a feature/project name (unless provided as an argument)
	2. Open $EDITOR for the multiline issue/feature thesis by default, falling back to a vim-compatible editor when $EDITOR is unset

The command never implements code. It outputs a prompt that instructs the
coding agent to research the codebase, use numbered lists for clarifying
questions, show percentage progress, and persist findings to BRAINSTORM.md.

Examples:
	kit legacy brainstorm
	kit legacy brainstorm --inline
	kit legacy brainstorm --editor nvim
	kit legacy brainstorm patient-intake-redesign
	kit legacy brainstorm patient-intake-redesign --output-only
	kit legacy brainstorm -o docs/brainstorm-prompt.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBrainstorm,
}

func init() {
	addFreeTextInputFlags(brainstormCmd, &brainstormUseVim, &brainstormEditor)
	addInlineTextInputFlag(brainstormCmd, &brainstormInline)
	brainstormCmd.Flags().BoolVar(&brainstormBacklog, "backlog", false, "capture a deferred brainstorm item and leave it paused")
	brainstormCmd.Flags().BoolVar(&brainstormCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	brainstormCmd.Flags().StringVarP(&brainstormOutput, "output", "o", "", "write output to file")
	brainstormCmd.Flags().BoolVar(&brainstormOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	brainstormCmd.Flags().BoolVar(&brainstormPrepare, "prepare", false, "create brainstorm directories and files without outputting the brainstorm prompt")
	addPromptOnlyFlag(brainstormCmd)
	legacyCmd.AddCommand(brainstormCmd)
}

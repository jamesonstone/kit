// package cli implements the Kit command-line interface.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/promptdoc"
)

var codeReviewCopy bool
var codeReviewOutputOnly bool

var codeReviewCmd = &cobra.Command{
	Use:   "code-review",
	Short: "Output coding agent instructions for branch code review",
	Long: `Output instructions that guide a coding agent through a systematic
code review of changes on the current branch compared to main/master.

The agent identifies the remote default branch, reviews the merge-base change
set and relevant contracts, and reports evidence-backed findings in severity
order. The generated prompt is read-only and does not mutate code or GitHub.`,
	Args: cobra.NoArgs,
	RunE: runCodeReview,
}

func init() {
	codeReviewCmd.Flags().BoolVarP(&codeReviewCopy, "copy", "c", false, "copy output to clipboard even with --output-only")
	codeReviewCmd.Flags().BoolVar(&codeReviewOutputOnly, "output-only", false, "output text to stdout instead of copying it to the clipboard")
	rootCmd.AddCommand(codeReviewCmd)
}

func runCodeReview(cmd *cobra.Command, args []string) error {
	output := codeReviewInstructions()

	outputOnly, _ := cmd.Flags().GetBool("output-only")

	if err := outputPromptWithoutSubagentsWithClipboardDefault(output, outputOnly, codeReviewCopy); err != nil {
		return err
	}
	if !outputOnly {
		style := styleForStdout()
		fmt.Println()
		fmt.Println(style.title("🔍", "Optional: Add specific concerns or goals"))
		fmt.Println(style.muted("After pasting the copied prompt, consider adding a follow-up message:"))
		fmt.Println()
		fmt.Println("  \"I have the following specific concerns for this review:")
		fmt.Println("    - [e.g., performance impact of the new caching layer]")
		fmt.Println("    - [e.g., error handling in the API endpoints]")
		fmt.Println("    - [e.g., backward compatibility with existing clients]\"")
		fmt.Println()
	}
	return nil
}

func codeReviewInstructions() string {
	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Heading(2, "Code Review")
		doc.Paragraph("Review the current branch against its remote mainline and report actionable findings. This is read-only: do not edit files or mutate Git/GitHub state.")

		doc.Heading(3, "Scope")
		doc.BulletList(
			"Identify the repository's remote default branch and compute the merge-base diff to `HEAD`; include staged, unstaged, and untracked in-scope changes when the user asked to review the working tree.",
			"Review changed files and only the minimal dependencies needed to establish whether a change is correct. Do not turn the review into a general repository audit.",
			"Read relevant tests, contracts, generated sources, and repo-local rules when they determine expected behavior.",
		)

		doc.Heading(3, "Review Priorities")
		doc.BulletList(
			"Correctness: broken behavior, edge cases, error handling, concurrency, compatibility, and violated invariants.",
			"Security and data safety: trust boundaries, authorization, secrets, injection, destructive behavior, and unsafe defaults.",
			"Performance and reliability: material complexity, resource leaks, blocking I/O, retries, timeouts, races, and failure recovery.",
			"Tests and evidence: missing coverage only when it leaves changed behavior unproven; flag assertions that do not test the intended contract.",
			"Documentation and interfaces: stale public docs, API/config/CLI drift, migration gaps, and generated-artifact inconsistency.",
		)

		doc.Heading(3, "Finding Standard")
		doc.BulletList(
			"Report only issues introduced or exposed by the changed scope and supported by concrete evidence.",
			"Order findings by severity. For each, give severity, `file:line`, failure mode, impact, evidence or reproduction, and the smallest credible fix direction.",
			"Do not add praise, style preferences, speculative refactors, or vague best-practice claims. Verify external framework behavior only when it is material to a finding.",
			"If no actionable findings exist, say `No findings.` and list only residual risks or tests you could not verify.",
		)

		doc.Heading(3, "Output")
		doc.BulletList(
			"`Findings`: ordered actionable items, or `No findings.`",
			"`Validation reviewed`: commands/evidence inspected and anything not run.",
			"`Residual risk`: concise unverified behavior, or `none`.",
		)
	})
}

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

The agent will:
  1. Identify the base branch (main or master)
  2. Diff all changes on the current branch
  3. Verify best practices using MCP tools (Context7)
  4. Analyze each change with thumbs up/down assessment
  5. Output a markdown table of findings
  6. Provide a summary with overall approval recommendation`,
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

	if err := outputPromptWithClipboardDefault(output, outputOnly, codeReviewCopy); err != nil {
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
		doc.Heading(2, "Code Review Agent Instructions")
		doc.Paragraph("**IMPORTANT: This is an INFORMATIONAL REVIEW ONLY.**")
		doc.BulletList(
			"Do NOT modify, edit, or change any code",
			"Do NOT create new files or update existing files",
			"Your sole task is to ANALYZE and EXPLAIN the changes",
			"Output your findings as documentation only",
		)
		doc.Paragraph("**REVIEW OBJECTIVES: Maximize for 100% CORRECTNESS and PERFORMANCE.**")
		doc.Raw("---")
		doc.Heading(3, "Step 1: Get the Changed Files List")
		doc.Paragraph("Run: `git diff --name-only main..HEAD` (or master if main doesn't exist)")
		doc.Paragraph("This output is the **ONLY** list of files you will review.")
		doc.Paragraph("**CRITICAL RULES:**")
		doc.BulletList(
			"If the output is empty, STOP — there are no changes to review",
			"**ONLY** analyze files that appear in this output",
			"Do **NOT** review files not in this list",
			"Do **NOT** explore or analyze other files in the codebase",
		)
		doc.Raw("---")
		doc.Heading(3, "Step 2: For Each Changed File")
		doc.Paragraph("For **each file in the list above** (and ONLY those files):")
		doc.OrderedList(1,
			"**Read the file** as it exists now",
			"**Understand** what it does and how it fits in the codebase",
			"**Analyze for correctness:**\n- Does the logic work correctly?\n- Are edge cases handled? (nil, empty, boundaries)\n- Are errors handled properly? (no swallowed errors)\n- Any potential panics or crashes?\n- Any race conditions?",
			"**Analyze for performance:**\n- Is the algorithm efficient?\n- Any unnecessary allocations or copies?\n- Any N+1 queries or unbatched I/O?",
			"**Check best practices:**\n- Use MCP tools (Context7) to verify framework best practices\n- Flag deprecated patterns or anti-patterns",
			"**Assess:** Is this change 👍 (net-positive) or 👎 (net-negative)?",
		)
		doc.Raw("---")
		doc.Heading(3, "Step 3: Analyze Test Files (if any changed)")
		doc.BulletList(
			"For test files in the changed list: do they test the right behavior?",
			"Do assertions match the application code changes?",
			"Do NOT recommend adding more tests",
		)
		doc.Raw("---")
		doc.Heading(3, "Step 4: Output Your Analysis")
		doc.Paragraph("**Format:**")
		doc.CodeBlock("markdown", `## Code Review: [branch] → main

### Files Reviewed
[List ONLY the files from the git diff output]

### Analysis

| File | Summary | Correctness | Performance | Assessment |
|------|---------|-------------|-------------|------------|
| file.go | what changed | ✅/⚠️/❌ | ✅/⚠️/❌ | 👍/👎 |

### Correctness Summary
**Confidence: [0-100]%**
- Bugs found: [list or "None"]
- Edge cases missed: [list or "None"]

### Performance Summary
**Assessment: [Optimal/Acceptable/Needs Work/Critical]**
- Issues: [list or "None"]

### Recommendation
[✅ APPROVE | ⚠️ APPROVE WITH NOTES | ❌ REQUEST CHANGES]
[If not clean approve, list specific issues]`)
		doc.Raw("---")
		doc.Heading(3, "Rules")
		doc.BulletList(
			"**INFORMATIONAL ONLY** — do not modify any code",
			"**ONLY REVIEW CHANGED FILES** — files from git diff output, nothing else",
			"**DO NOT EXPLORE** other files unless directly imported by a changed file",
			"**MAXIMIZE CORRECTNESS** — find all bugs, edge cases, error handling issues",
			"**MAXIMIZE PERFORMANCE** — identify inefficiencies",
			"Be thorough but stay focused on the changed files",
		)
	})
}

// package cli implements the Kit command-line interface.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
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
	codeReviewCmd.Flags().BoolVarP(&codeReviewCopy, "copy", "c", false, "copy output to clipboard")
	codeReviewCmd.Flags().BoolVar(&codeReviewOutputOnly, "output-only", false, "output text only, suppressing status messages")
	rootCmd.AddCommand(codeReviewCmd)
}

func runCodeReview(cmd *cobra.Command, args []string) error {
	output := codeReviewInstructions()

	outputOnly, _ := cmd.Flags().GetBool("output-only")

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	if codeReviewCopy {
		fmt.Println(whiteBold + "Agent prompt copied to clipboard" + reset)
	} else {
		fmt.Println(whiteBold + "Copy this prompt to your coding agent:" + reset)
	}
	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)

	if err := outputPrompt(output, outputOnly, codeReviewCopy); err != nil {
		return err
	}

	fmt.Println(dim + "────────────────────────────────────────────────────────────────────────" + reset)
	fmt.Println()
	fmt.Println(whiteBold + "Optional: Add specific concerns or goals" + reset)
	fmt.Println(dim + "After pasting the prompt above, consider adding a follow-up message:" + reset)
	fmt.Println()
	fmt.Println("  \"I have the following specific concerns for this review:")
	fmt.Println("    - [e.g., performance impact of the new caching layer]")
	fmt.Println("    - [e.g., error handling in the API endpoints]")
	fmt.Println("    - [e.g., backward compatibility with existing clients]\"")
	fmt.Println()
	return nil
}

func codeReviewInstructions() string {
	return `## Code Review Agent Instructions

**IMPORTANT: This is an INFORMATIONAL REVIEW ONLY.**
- Do NOT modify, edit, or change any code
- Do NOT create new files or update existing files
- Your sole task is to ANALYZE and EXPLAIN the changes
- Output your findings as documentation only

**REVIEW OBJECTIVES: Maximize for 100% CORRECTNESS and PERFORMANCE.**

---

### Step 1: Get the Changed Files List

Run: ` + "`git diff --name-only main..HEAD`" + ` (or master if main doesn't exist)

This output is the **ONLY** list of files you will review.

**CRITICAL RULES:**
- If the output is empty, STOP — there are no changes to review
- **ONLY** analyze files that appear in this output
- Do **NOT** review files not in this list
- Do **NOT** explore or analyze other files in the codebase

---

### Step 2: For Each Changed File

For **each file in the list above** (and ONLY those files):

1. **Read the file** as it exists now
2. **Understand** what it does and how it fits in the codebase
3. **Analyze for correctness:**
   - Does the logic work correctly?
   - Are edge cases handled? (nil, empty, boundaries)
   - Are errors handled properly? (no swallowed errors)
   - Any potential panics or crashes?
   - Any race conditions?

4. **Analyze for performance:**
   - Is the algorithm efficient?
   - Any unnecessary allocations or copies?
   - Any N+1 queries or unbatched I/O?

5. **Check best practices:**
   - Use MCP tools (Context7) to verify framework best practices
   - Flag deprecated patterns or anti-patterns

6. **Assess:** Is this change 👍 (net-positive) or 👎 (net-negative)?

---

### Step 3: Analyze Test Files (if any changed)

For test files in the changed list:
- Do they test the right behavior?
- Do assertions match the application code changes?
- Do NOT recommend adding more tests

---

### Step 4: Output Your Analysis

**Format:**

` + "```" + `markdown
## Code Review: [branch] → main

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
[If not clean approve, list specific issues]
` + "```" + `

---

### Rules

- **INFORMATIONAL ONLY** — do not modify any code
- **ONLY REVIEW CHANGED FILES** — files from git diff output, nothing else
- **DO NOT EXPLORE** other files unless directly imported by a changed file
- **MAXIMIZE CORRECTNESS** — find all bugs, edge cases, error handling issues
- **MAXIMIZE PERFORMANCE** — identify inefficiencies
- Be thorough but stay focused on the changed files`
}

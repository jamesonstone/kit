package cli

import (
	"fmt"
	"strings"
)

type implementDocumentReferenceRow struct {
	document string
	lines    []string
}

const (
	implementDocumentColumnWidth = 10
	implementPurposeColumnWidth  = 58
)

func printImplementDocumentReferenceTable() {
	rows := []implementDocumentReferenceRow{
		{
			document: "SPEC.md",
			lines: []string{
				"WHAT: Requirements, constraints, acceptance criteria",
				"- Consult when unsure if something is in scope",
				"- Do NOT add features not specified here",
			},
		},
		{
			document: "PLAN.md",
			lines: []string{
				"HOW: Architecture, components, data structures",
				"- Follow the design decisions made here",
				"- If blocked, check RISKS section for mitigations",
			},
		},
		{
			document: "TASKS.md",
			lines: []string{
				"EXECUTE: Ordered task list with acceptance criteria",
				"- Do NOT execute tasks until the readiness gate passes",
				"- Work through tasks in order (respect dependencies)",
				"- Mark tasks complete with [x] when acceptance met",
			},
		},
	}

	printImplementTableBorder("┌", "┬", "┐")
	printImplementTableRow("Document", "Purpose & Usage")
	printImplementTableBorder("├", "┼", "┤")

	for i, row := range rows {
		printImplementWrappedRow(row)
		if i < len(rows)-1 {
			printImplementTableBorder("├", "┼", "┤")
		}
	}

	printImplementTableBorder("└", "┴", "┘")
}

func printImplementWrappedRow(row implementDocumentReferenceRow) {
	for i, line := range row.lines {
		document := ""
		if i == 0 {
			document = row.document
		}
		printImplementTableRow(document, line)
	}
}

func printImplementTableBorder(left, middle, right string) {
	line := left +
		strings.Repeat("─", implementDocumentColumnWidth+2) +
		middle +
		strings.Repeat("─", implementPurposeColumnWidth+2) +
		right
	fmt.Println(dim + line + reset)
}

func printImplementTableRow(document, purpose string) {
	fmt.Printf(
		"%s│%s %-*s %s│%s %-*s %s│%s\n",
		dim,
		reset,
		implementDocumentColumnWidth,
		document,
		dim,
		reset,
		implementPurposeColumnWidth,
		purpose,
		dim,
		reset,
	)
}

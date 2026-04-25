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
			document: "TASKS.md",
			lines: []string{
				"EXECUTE: Next action and acceptance criteria",
				"- Read first; select the next incomplete task",
				"- Mark tasks complete with [x] only when acceptance is met",
			},
		},
		{
			document: "PLAN.md",
			lines: []string{
				"HOW: Architecture, components, data structures",
				"- Load only the section linked from the selected task",
				"- If blocked, check RISKS section for mitigations",
			},
		},
		{
			document: "SPEC.md",
			lines: []string{
				"WHAT: Requirements, constraints, acceptance criteria",
				"- Load only the requirement linked from PLAN.md",
				"- Do NOT add features not specified here",
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

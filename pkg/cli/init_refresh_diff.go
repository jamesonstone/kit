package cli

import (
	"fmt"
	"strings"
)

const initRefreshDiffContext = 3

type initRefreshDiffOp struct {
	kind string
	line string
}

func printInitRefreshDryRun(changes []initRefreshFileChange, stats initRefreshStats, opts initRefreshOptions) {
	if opts.diff {
		diff := renderInitRefreshDiff(changes)
		if strings.TrimSpace(diff) == "" {
			fmt.Println("No Kit-managed file changes planned.")
		} else {
			fmt.Print(diff)
		}
	}

	if opts.outputOnly {
		return
	}

	if !opts.diff {
		fmt.Println("🔎 Kit project refresh dry run:")
		for _, change := range changes {
			if change.result == instructionFileSkipped {
				continue
			}
			fmt.Printf("   %s %s\n", dryRunActionLabel(change.result), change.relativePath)
		}
		if stats.created+stats.updated+stats.merged == 0 {
			fmt.Println("   No Kit-managed file changes planned.")
		}
	}

	fmt.Printf(
		"\nDry run complete. Planned Created: %d, Updated: %d, Merged: %d, Skipped: %d\n",
		stats.created,
		stats.updated,
		stats.merged,
		stats.skipped,
	)
}

func dryRunActionLabel(result instructionFileWriteResult) string {
	switch result {
	case instructionFileCreated:
		return "create"
	case instructionFileUpdated:
		return "update"
	case instructionFileMerged:
		return "merge"
	default:
		return "skip"
	}
}

func renderInitRefreshDiff(changes []initRefreshFileChange) string {
	var builder strings.Builder
	for _, change := range changes {
		if change.result == instructionFileSkipped {
			continue
		}
		if change.result != instructionFileCreated && change.before == change.after {
			continue
		}
		builder.WriteString(renderInitRefreshFileDiff(change))
	}
	return builder.String()
}

func renderInitRefreshFileDiff(change initRefreshFileChange) string {
	oldPath := "a/" + change.relativePath
	newPath := "b/" + change.relativePath

	var builder strings.Builder
	fmt.Fprintf(&builder, "diff --git %s %s\n", oldPath, newPath)
	if change.result == instructionFileCreated {
		builder.WriteString("new file mode 100644\n")
		builder.WriteString("--- /dev/null\n")
	} else {
		fmt.Fprintf(&builder, "--- %s\n", oldPath)
	}
	fmt.Fprintf(&builder, "+++ %s\n", newPath)

	for _, hunk := range unifiedDiffHunks(splitDiffLines(change.before), splitDiffLines(change.after), initRefreshDiffContext) {
		builder.WriteString(hunk)
	}
	return builder.String()
}

func splitDiffLines(content string) []string {
	if content == "" {
		return nil
	}
	lines := strings.SplitAfter(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for i, line := range lines {
		lines[i] = strings.TrimSuffix(line, "\n")
	}
	return lines
}

func unifiedDiffHunks(oldLines, newLines []string, context int) []string {
	ops := buildDiffOps(oldLines, newLines)
	var changed []int
	for i, op := range ops {
		if op.kind != "equal" {
			changed = append(changed, i)
		}
	}
	if len(changed) == 0 {
		return nil
	}

	var hunks []string
	for i := 0; i < len(changed); {
		changeStart := changed[i]
		changeEnd := changeStart
		i++
		for i < len(changed) && changed[i]-changeEnd <= context*2 {
			changeEnd = changed[i]
			i++
		}

		start := changeStart - context
		if start < 0 {
			start = 0
		}
		end := changeEnd + context
		if end >= len(ops) {
			end = len(ops) - 1
		}
		hunks = append(hunks, renderDiffHunk(ops, start, end))
	}
	return hunks
}

func buildDiffOps(oldLines, newLines []string) []initRefreshDiffOp {
	oldLen := len(oldLines)
	newLen := len(newLines)
	table := make([][]int, oldLen+1)
	for i := range table {
		table[i] = make([]int, newLen+1)
	}
	for i := oldLen - 1; i >= 0; i-- {
		for j := newLen - 1; j >= 0; j-- {
			if oldLines[i] == newLines[j] {
				table[i][j] = table[i+1][j+1] + 1
				continue
			}
			if table[i+1][j] >= table[i][j+1] {
				table[i][j] = table[i+1][j]
			} else {
				table[i][j] = table[i][j+1]
			}
		}
	}

	var ops []initRefreshDiffOp
	i, j := 0, 0
	for i < oldLen && j < newLen {
		if oldLines[i] == newLines[j] {
			ops = append(ops, initRefreshDiffOp{kind: "equal", line: oldLines[i]})
			i++
			j++
			continue
		}
		if table[i+1][j] >= table[i][j+1] {
			ops = append(ops, initRefreshDiffOp{kind: "delete", line: oldLines[i]})
			i++
		} else {
			ops = append(ops, initRefreshDiffOp{kind: "insert", line: newLines[j]})
			j++
		}
	}
	for i < oldLen {
		ops = append(ops, initRefreshDiffOp{kind: "delete", line: oldLines[i]})
		i++
	}
	for j < newLen {
		ops = append(ops, initRefreshDiffOp{kind: "insert", line: newLines[j]})
		j++
	}
	return ops
}

func renderDiffHunk(ops []initRefreshDiffOp, start int, end int) string {
	oldBefore, newBefore := diffLineCounts(ops[:start])
	oldCount, newCount := diffLineCounts(ops[start : end+1])
	oldStart := oldBefore + 1
	newStart := newBefore + 1
	if oldCount == 0 {
		oldStart = oldBefore
	}
	if newCount == 0 {
		newStart = newBefore
	}

	var builder strings.Builder
	fmt.Fprintf(
		&builder,
		"@@ -%s +%s @@\n",
		diffRange(oldStart, oldCount),
		diffRange(newStart, newCount),
	)
	for _, op := range ops[start : end+1] {
		switch op.kind {
		case "equal":
			builder.WriteString(" ")
		case "delete":
			builder.WriteString("-")
		case "insert":
			builder.WriteString("+")
		}
		builder.WriteString(op.line)
		builder.WriteString("\n")
	}
	return builder.String()
}

func diffLineCounts(ops []initRefreshDiffOp) (oldCount int, newCount int) {
	for _, op := range ops {
		switch op.kind {
		case "equal":
			oldCount++
			newCount++
		case "delete":
			oldCount++
		case "insert":
			newCount++
		}
	}
	return oldCount, newCount
}

func diffRange(start int, count int) string {
	if count == 1 {
		return fmt.Sprintf("%d", start)
	}
	return fmt.Sprintf("%d,%d", start, count)
}

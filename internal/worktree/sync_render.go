package worktree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

func writeSyncJSON(writer io.Writer, report SyncReport) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func writeSyncHuman(writer io.Writer, report SyncReport) error {
	var output bytes.Buffer
	fmt.Fprintf(
		&output,
		"SYNC %s (%s)\n",
		detailOr(report.Repository, "unknown repository"),
		report.Result,
	)
	fmt.Fprintf(
		&output,
		"FETCH\t%s\t%s\n",
		report.Fetch.Status,
		report.Fetch.Detail,
	)
	fmt.Fprintf(
		&output,
		"DEFAULT\t%s\t%s\t%s",
		report.DefaultBranch.Branch,
		report.DefaultBranch.State,
		report.DefaultBranch.Action,
	)
	if report.DefaultBranch.Detail != "" {
		fmt.Fprintf(&output, "\t%s", singleLine(report.DefaultBranch.Detail))
	}
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "ACTION\tBRANCH\tREASON\tPATH")
	for _, lane := range report.Lanes {
		fmt.Fprintf(
			&output,
			"%s\t%s\t%s\t%s\n",
			lane.Action,
			detailOr(lane.Branch, displayDetachedOID(lane.HeadOID)),
			lane.Reason,
			lane.Path,
		)
	}
	fmt.Fprintf(
		&output,
		"PRUNE\t%s\t%s\n",
		report.Prune.Status,
		singleLine(report.Prune.Detail),
	)
	if len(report.Failures) != 0 {
		fmt.Fprintln(&output, "FAILURE\tOPERATION\tPATH\tERROR")
		for _, failure := range report.Failures {
			fmt.Fprintf(
				&output,
				"failure\t%s\t%s\t%s\n",
				failure.Operation,
				failure.Path,
				singleLine(failure.Error),
			)
		}
	}
	fmt.Fprintln(&output, "STATE\tHEAD\tLAST UPDATED\tPATH")
	for _, worktree := range report.Worktrees {
		fmt.Fprintf(
			&output,
			"%s\t%s\t%s\t%s\n",
			worktree.State,
			worktree.Head,
			worktree.LastUpdated,
			worktree.Path,
		)
	}
	_, err := writer.Write(output.Bytes())
	return err
}

func sortSyncLanes(lanes []SyncLaneDecision) {
	sort.SliceStable(lanes, func(i, j int) bool {
		return lanes[i].Path < lanes[j].Path
	})
}

func sortSyncFailures(failures []SyncFailure) {
	sort.SliceStable(failures, func(i, j int) bool {
		if failures[i].Operation != failures[j].Operation {
			return failures[i].Operation < failures[j].Operation
		}
		if failures[i].Path != failures[j].Path {
			return failures[i].Path < failures[j].Path
		}
		return failures[i].Error < failures[j].Error
	})
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func detailOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func displayDetachedOID(oid string) string {
	if oid == "" {
		return "detached"
	}
	return "detached@" + shortOID(oid)
}

func singleLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

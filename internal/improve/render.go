package improve

import (
	"fmt"
	"strings"
)

func uniqueTaskIDs(traces []Trace) []string {
	seen := map[string]struct{}{}
	var ids []string
	for _, trace := range traces {
		if _, ok := seen[trace.TaskID]; ok {
			continue
		}
		seen[trace.TaskID] = struct{}{}
		ids = append(ids, trace.TaskID)
	}
	return ids
}

func traceIDs(traces []Trace) []string {
	ids := make([]string, 0, len(traces))
	for _, trace := range traces {
		ids = append(ids, fmt.Sprintf("%s:%d", trace.TaskID, trace.RepeatIndex))
	}
	return ids
}

func confidenceFor(traces []Trace) string {
	if len(traces) >= 2 {
		return "high"
	}
	return "medium"
}

func candidatePrompt(candidate Candidate, cluster WeaknessCluster) string {
	var builder strings.Builder
	builder.WriteString("# Kit Improve Candidate\n\n")
	builder.WriteString("Verify each finding against current code. Fix only still-valid issues, skip the rest with a brief reason, keep changes minimal, and validate.\n\n")
	builder.WriteString("## Candidate\n\n")
	builder.WriteString("- ID: `" + candidate.ID + "`\n")
	builder.WriteString("- Target cluster: `" + candidate.TargetCluster + "`\n")
	builder.WriteString("- Summary: " + candidate.Summary + "\n")
	builder.WriteString("- Expected effect: " + candidate.ExpectedEffect + "\n\n")
	builder.WriteString("## Evidence\n\n")
	builder.WriteString("- Affected tasks: `" + strings.Join(cluster.AffectedTasks, "`, `") + "`\n")
	builder.WriteString("- Representative traces: `" + strings.Join(cluster.RepresentativeTraces, "`, `") + "`\n")
	builder.WriteString("- Confidence: `" + cluster.Confidence + "`\n\n")
	builder.WriteString("## Boundaries\n\n")
	builder.WriteString("- Edit only allowed Kit harness surfaces.\n")
	builder.WriteString("- Do not weaken held-out tasks or delivery rules.\n")
	builder.WriteString("- Do not use worktrees.\n")
	return builder.String()
}

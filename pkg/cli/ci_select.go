package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func resolveCIWorkflow(ctx ciRepoContext, raw string) (ciWorkflow, error) {
	if strings.TrimSpace(raw) == "" {
		return ciWorkflow{}, nil
	}
	output, err := ciRunner.Output(ctx.Directory, "gh", repoArgs(ctx.Target.FullName,
		"workflow", "list", "--all", "--json", "id,name,path,state")...)
	if err != nil {
		return ciWorkflow{}, fmt.Errorf("failed to list GitHub workflows: %w", err)
	}
	var workflows []ciWorkflow
	if err := json.Unmarshal(output, &workflows); err != nil {
		return ciWorkflow{}, fmt.Errorf("failed to parse GitHub workflows: %w", err)
	}
	return matchCIWorkflow(raw, workflows)
}

func matchCIWorkflow(raw string, workflows []ciWorkflow) (ciWorkflow, error) {
	query := strings.TrimSpace(raw)
	for _, workflow := range workflows {
		if workflow.Path == query {
			return workflow, nil
		}
	}
	for _, workflow := range workflows {
		if workflow.Name == query {
			return workflow, nil
		}
	}
	queryLower := strings.ToLower(query)
	var matches []ciWorkflow
	for _, workflow := range workflows {
		if strings.Contains(strings.ToLower(workflow.Path), queryLower) ||
			strings.Contains(strings.ToLower(workflow.Name), queryLower) {
			matches = append(matches, workflow)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) == 0 {
		return ciWorkflow{}, fmt.Errorf("no workflow matched %q", raw)
	}
	labels := make([]string, 0, len(matches))
	for _, workflow := range matches {
		labels = append(labels, fmt.Sprintf("%s (%s)", workflow.Name, workflow.Path))
	}
	sort.Strings(labels)
	return ciWorkflow{}, fmt.Errorf("workflow %q is ambiguous; matched: %s", raw, strings.Join(labels, ", "))
}

func selectDefaultBranchFailures(ctx ciRepoContext, workflow ciWorkflow) ([]ciRun, error) {
	args := []string{"run", "list", "--branch", ctx.DefaultBranch, "--status", "completed", "--limit", "30", "--json", ciRunListJSONFields}
	if workflow.Path != "" {
		args = append(args, "--workflow", workflow.Path)
	}
	runs, err := fetchCIRunList(ctx, args...)
	if err != nil {
		return nil, err
	}
	selected := selectDiagnosableRuns(runs, false)
	if len(selected) == 0 {
		return nil, nil
	}
	return fetchFullCIRuns(ctx, []ciRun{selected[0]})
}

func selectPRFailures(
	ctx ciRepoContext,
	target dispatchPRTarget,
	workflow ciWorkflow,
	opts ciOptions,
) (ciPR, []ciRun, []ciCheck, []ciCheck, error) {
	pr, err := fetchCIPR(ctx, target)
	if err != nil {
		return ciPR{}, nil, nil, nil, err
	}
	checks, err := fetchCIPRChecks(ctx, target)
	if err != nil {
		return ciPR{}, nil, nil, nil, err
	}
	failing, external := splitFailingCIChecks(checks, workflow)
	runs, err := actionRunsForChecks(ctx, failing, workflow)
	if err != nil {
		return ciPR{}, nil, nil, nil, err
	}
	if len(runs) == 0 {
		runs, err = fallbackRunsForPRHead(ctx, pr, workflow)
		if err != nil {
			return ciPR{}, nil, nil, nil, err
		}
	}
	return pr, runs, failing, external, nil
}

func actionRunsForChecks(ctx ciRepoContext, checks []ciCheck, workflow ciWorkflow) ([]ciRun, error) {
	runIDs := uniqueActionRunIDs(checks)
	var runs []ciRun
	for _, runID := range runIDs {
		run, err := fetchCIRun(ctx, strconv.FormatInt(runID, 10))
		if err != nil {
			return nil, err
		}
		if workflow.Path == "" || sameWorkflow(run, workflow) {
			runs = append(runs, run)
		}
	}
	return runs, nil
}

func fallbackRunsForPRHead(ctx ciRepoContext, pr ciPR, workflow ciWorkflow) ([]ciRun, error) {
	args := []string{"run", "list", "--commit", pr.HeadRefOID, "--status", "completed", "--limit", "30", "--json", ciRunListJSONFields}
	if workflow.Path != "" {
		args = append(args, "--workflow", workflow.Path)
	}
	list, err := fetchCIRunList(ctx, args...)
	if err != nil {
		return nil, err
	}
	return fetchFullCIRuns(ctx, selectDiagnosableRuns(list, true))
}

func workflowDisplayName(workflow ciWorkflow) string {
	if workflow.Path == "" {
		return ""
	}
	if workflow.Name != "" {
		return workflow.Name
	}
	return workflow.Path
}

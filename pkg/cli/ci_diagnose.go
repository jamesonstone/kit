package cli

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	ciFailureLinePattern = regexp.MustCompile(`(?i)(::error::|^\s*error[:\s]|panic:|fatal:|failed|failure|exception|traceback|exit code|cannot find|not found|undefined|denied|timed out|timeout|segmentation fault)`)
)

func diagnoseRuns(ctx ciRepoContext, runs []ciRun, opts ciOptions) ([]ciRunFailure, error) {
	failures := make([]ciRunFailure, 0, len(runs))
	for _, run := range runs {
		jobs, err := selectFailedCIJobs(run, opts.JobRef)
		if err != nil {
			return nil, err
		}
		runFailure := ciRunFailure{
			RunID:      run.DatabaseID,
			Name:       firstNonEmpty(run.DisplayTitle, run.Name),
			Workflow:   run.WorkflowName,
			Conclusion: run.Conclusion,
			Status:     run.Status,
			HeadBranch: run.HeadBranch,
			HeadSHA:    run.HeadSHA,
			URL:        run.URL,
		}
		for _, job := range jobs {
			rawLog, err := fetchCILog(ctx, run, job)
			if err != nil {
				return nil, err
			}
			excerpt, truncated := extractRelevantCILogExcerpt(rawLog, opts.LogLines)
			runFailure.LogTruncated = runFailure.LogTruncated || truncated
			runFailure.FailedJobs = append(runFailure.FailedJobs, ciJobFailure{
				JobID:       job.DatabaseID,
				Name:        job.Name,
				Conclusion:  job.Conclusion,
				Status:      job.Status,
				URL:         job.URL,
				FailedSteps: failedStepNames(job),
				LogExcerpt:  excerpt,
			})
		}
		failures = append(failures, runFailure)
	}
	return failures, nil
}

func selectFailedCIJobs(run ciRun, jobRef string) ([]ciJob, error) {
	var jobs []ciJob
	if strings.TrimSpace(jobRef) != "" {
		for _, job := range run.Jobs {
			if ciJobMatches(job, jobRef) {
				jobs = append(jobs, job)
			}
		}
		if len(jobs) == 0 {
			return nil, fmt.Errorf("run %d does not include job %q", run.DatabaseID, jobRef)
		}
		return jobs, nil
	}

	for _, job := range run.Jobs {
		if isDiagnosableCIConclusion(job.Conclusion) || isDiagnosableCIConclusion(job.Status) {
			jobs = append(jobs, job)
		}
	}
	if len(jobs) == 0 && isDiagnosableCIConclusion(run.Conclusion) {
		jobs = append(jobs, ciJob{Name: firstNonEmpty(run.WorkflowName, run.Name), Status: run.Status, Conclusion: run.Conclusion})
	}
	return jobs, nil
}

func ciJobMatches(job ciJob, raw string) bool {
	ref := strings.TrimSpace(raw)
	if ref == "" {
		return false
	}
	if strconv.FormatInt(job.DatabaseID, 10) == ref {
		return true
	}
	return strings.EqualFold(job.Name, ref)
}

func failedStepNames(job ciJob) []string {
	var names []string
	for _, step := range job.Steps {
		if isDiagnosableCIConclusion(step.Conclusion) || isDiagnosableCIConclusion(step.Status) {
			names = append(names, step.Name)
		}
	}
	return names
}

func selectDiagnosableRuns(runs []ciRun, all bool) []ciRun {
	var failures []ciRun
	for _, run := range runs {
		if strings.EqualFold(run.Conclusion, "failure") {
			failures = append(failures, run)
			if !all {
				return failures
			}
		}
	}
	if len(failures) > 0 {
		return failures
	}
	for _, run := range runs {
		if isDiagnosableCIConclusion(run.Conclusion) || isDiagnosableCIConclusion(run.Status) {
			failures = append(failures, run)
			if !all {
				return failures
			}
		}
	}
	return failures
}

func isDiagnosableCIConclusion(value string) bool {
	return ciDiagnosableConclusions[strings.ToLower(strings.TrimSpace(value))]
}

func splitFailingCIChecks(checks []ciCheck, workflow ciWorkflow) ([]ciCheck, []ciCheck) {
	var failing []ciCheck
	var external []ciCheck
	for _, check := range checks {
		if !ciCheckFailed(check) {
			continue
		}
		if workflow.Path != "" && !checkMatchesWorkflow(check, workflow) {
			continue
		}
		failing = append(failing, check)
		if actionRunIDFromLink(check.Link) == 0 {
			external = append(external, check)
		}
	}
	return failing, external
}

func ciCheckFailed(check ciCheck) bool {
	return strings.EqualFold(check.Bucket, "fail") ||
		strings.EqualFold(check.State, "failure") ||
		strings.EqualFold(check.State, "failed") ||
		isDiagnosableCIConclusion(check.State)
}

func checkMatchesWorkflow(check ciCheck, workflow ciWorkflow) bool {
	if workflow.Path == "" {
		return true
	}
	return strings.EqualFold(check.Workflow, workflow.Name) ||
		strings.EqualFold(check.Workflow, workflow.Path) ||
		strings.Contains(strings.ToLower(check.Name), strings.ToLower(workflow.Name))
}

func sameWorkflow(run ciRun, workflow ciWorkflow) bool {
	if workflow.Path == "" {
		return true
	}
	return strings.EqualFold(run.WorkflowName, workflow.Name) ||
		strings.EqualFold(run.Name, workflow.Name) ||
		strings.EqualFold(run.WorkflowName, workflow.Path)
}

func uniqueActionRunIDs(checks []ciCheck) []int64 {
	seen := map[int64]bool{}
	var ids []int64
	for _, check := range checks {
		id := actionRunIDFromLink(check.Link)
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] > ids[j] })
	return ids
}

func actionRunIDFromLink(link string) int64 {
	match := ciActionRunURLPattern.FindStringSubmatch(link)
	if match == nil {
		return 0
	}
	id, _ := strconv.ParseInt(match[1], 10, 64)
	return id
}

func finalizeCIDiagnosis(diagnosis *ciDiagnosis, opts ciOptions) {
	diagnosis.FailureFound = len(diagnosis.Runs) > 0 || len(diagnosis.ExternalChecks) > 0
	if !diagnosis.FailureFound {
		diagnosis.RootCause = "No failed GitHub Actions runs or failed PR checks were found for the selected target."
		diagnosis.Recommendation = "No CI fix is currently required for this target."
		diagnosis.AgentPrompt = buildCIAgentPrompt(*diagnosis)
		return
	}

	firstEvidence := firstCIDiagnosticLine(diagnosis.Runs)
	if firstEvidence != "" {
		diagnosis.RootCause = "The first relevant failing log line is: " + firstEvidence
	} else if len(diagnosis.ExternalChecks) > 0 {
		diagnosis.RootCause = "Only external failing checks were found; GitHub Actions logs were not available through gh."
	} else {
		diagnosis.RootCause = "GitHub reported failed jobs, but no failed-step log excerpt was available."
	}
	diagnosis.Evidence = collectCIEvidence(diagnosis.Runs, opts.LogLines)
	diagnosis.Recommendation = buildCIRecommendation(*diagnosis)
	diagnosis.AgentPrompt = buildCIAgentPrompt(*diagnosis)
}

func firstCIDiagnosticLine(runs []ciRunFailure) string {
	for _, run := range runs {
		for _, job := range run.FailedJobs {
			for _, line := range job.LogExcerpt {
				trimmed := strings.TrimSpace(line)
				if ciFailureLinePattern.MatchString(trimmed) {
					return trimmed
				}
			}
		}
	}
	return ""
}

func collectCIEvidence(runs []ciRunFailure, maxLines int) []string {
	var evidence []string
	for _, run := range runs {
		for _, job := range run.FailedJobs {
			for _, line := range job.LogExcerpt {
				evidence = append(evidence, line)
				if len(evidence) >= maxLines {
					return evidence
				}
			}
		}
	}
	return evidence
}

func buildCIRecommendation(diagnosis ciDiagnosis) string {
	if len(diagnosis.Runs) == 0 {
		return "Inspect the failing external checks listed above; they are not GitHub Actions jobs with logs available through gh."
	}
	var parts []string
	for _, run := range diagnosis.Runs {
		for _, job := range run.FailedJobs {
			label := job.Name
			if run.Workflow != "" {
				label = run.Workflow + " / " + label
			}
			parts = append(parts, label)
		}
	}
	return "Fix the still-valid failure in " + strings.Join(parts, ", ") + ", keep the change minimal, and run the closest local verification before pushing."
}

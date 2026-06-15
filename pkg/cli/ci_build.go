package cli

import "strings"

func buildCIDiagnosis(opts ciOptions) (ciDiagnosis, error) {
	ctx, prTarget, err := resolveCIRepoContext(opts)
	if err != nil {
		return ciDiagnosis{}, err
	}
	if err := requireGHAuth(ctx.Directory); err != nil {
		return ciDiagnosis{}, err
	}

	if needsDefaultBranch(opts) {
		if err := discoverAndCacheDefaultBranch(&ctx); err != nil {
			return ciDiagnosis{}, err
		}
	}

	workflow := ciWorkflow{}
	if strings.TrimSpace(opts.RunID) == "" {
		workflow, err = resolveCIWorkflow(ctx, opts.WorkflowRef)
		if err != nil {
			return ciDiagnosis{}, err
		}
	}

	diagnosis := ciDiagnosis{
		Target: ciTarget{
			Repository: ctx.Target.FullName,
		},
		Copilot: ciCopilotInfo{
			Requested: opts.UseCopilot,
			Available: false,
			Used:      false,
			Message:   "No callable GitHub Copilot diagnosis API is available; using GitHub Actions logs.",
		},
	}

	switch {
	case strings.TrimSpace(opts.RunID) != "":
		err = buildCIRunDiagnosis(&diagnosis, ctx, opts)
	case strings.TrimSpace(opts.PRRef) != "":
		err = buildCIPRDiagnosis(&diagnosis, ctx, prTarget, workflow, opts)
	default:
		err = buildCIDefaultBranchDiagnosis(&diagnosis, ctx, workflow, opts)
	}
	if err != nil {
		return ciDiagnosis{}, err
	}

	finalizeCIDiagnosis(&diagnosis, opts)
	return diagnosis, nil
}

func buildCIRunDiagnosis(diagnosis *ciDiagnosis, ctx ciRepoContext, opts ciOptions) error {
	run, err := fetchCIRun(ctx, opts.RunID)
	if err != nil {
		return err
	}
	diagnosis.Target.Kind = "run"
	diagnosis.Target.RunID = run.DatabaseID
	diagnosis.Target.Workflow = run.WorkflowName
	diagnosis.Target.Job = opts.JobRef
	diagnosis.Runs, err = diagnoseRuns(ctx, []ciRun{run}, opts)
	return err
}

func buildCIPRDiagnosis(
	diagnosis *ciDiagnosis,
	ctx ciRepoContext,
	prTarget dispatchPRTarget,
	workflow ciWorkflow,
	opts ciOptions,
) error {
	pr, runs, checks, external, err := selectPRFailures(ctx, prTarget, workflow, opts)
	if err != nil {
		return err
	}
	diagnosis.Target.Kind = "pr"
	diagnosis.Target.PRNumber = pr.Number
	diagnosis.Target.PRURL = pr.URL
	diagnosis.Target.HeadSHA = pr.HeadRefOID
	diagnosis.Target.Workflow = workflowDisplayName(workflow)
	diagnosis.FailingChecks = checks
	diagnosis.ExternalChecks = external
	diagnosis.Runs, err = diagnoseRuns(ctx, runs, opts)
	return err
}

func buildCIDefaultBranchDiagnosis(
	diagnosis *ciDiagnosis,
	ctx ciRepoContext,
	workflow ciWorkflow,
	opts ciOptions,
) error {
	runs, err := selectDefaultBranchFailures(ctx, workflow)
	if err != nil {
		return err
	}
	diagnosis.Target.Kind = "branch"
	diagnosis.Target.Branch = ctx.DefaultBranch
	diagnosis.Target.Workflow = workflowDisplayName(workflow)
	diagnosis.Runs, err = diagnoseRuns(ctx, runs, opts)
	return err
}

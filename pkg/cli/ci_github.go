package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
)

var (
	ciRunner                 ciCommandRunner = execCICommandRunner{}
	ciActionRunURLPattern                    = regexp.MustCompile(`/actions/runs/(\d+)`)
	ciDiagnosableConclusions                 = map[string]bool{
		"failure":         true,
		"timed_out":       true,
		"action_required": true,
		"startup_failure": true,
	}
)

type ciCommandRunner interface {
	Output(dir string, name string, args ...string) ([]byte, error)
	OutputAllowError(dir string, name string, args ...string) ([]byte, error)
}

type execCICommandRunner struct{}

func (execCICommandRunner) Output(dir string, name string, args ...string) ([]byte, error) {
	output, err := execCICommandRunner{}.OutputAllowError(dir, name, args...)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (execCICommandRunner) OutputAllowError(dir string, name string, args ...string) ([]byte, error) {
	cmd := execCommand(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if text := strings.TrimSpace(string(output)); text != "" {
			if errors.As(err, &exitErr) {
				return output, fmt.Errorf("%s: %s", err, text)
			}
			return output, fmt.Errorf("%w: %s", err, text)
		}
		return nil, err
	}
	return output, nil
}

func resolveCIRepoContext(opts ciOptions) (ciRepoContext, dispatchPRTarget, error) {
	dir := strings.TrimSpace(opts.RepoPath)
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ciRepoContext{}, dispatchPRTarget{}, fmt.Errorf("failed to get working directory: %w", err)
		}
		dir = cwd
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ciRepoContext{}, dispatchPRTarget{}, fmt.Errorf("failed to resolve --repo-path: %w", err)
	}

	gitOwner, gitRepo, gitErr := resolveGitHubRepoFromDir(absDir)
	target := dispatchPRTarget{}
	if strings.TrimSpace(opts.PRRef) != "" {
		target, err = resolveCIPRTarget(opts.PRRef, absDir)
		if err != nil {
			return ciRepoContext{}, dispatchPRTarget{}, err
		}
	} else if gitErr != nil {
		return ciRepoContext{}, dispatchPRTarget{}, gitErr
	} else {
		target = dispatchPRTarget{Owner: gitOwner, Repo: gitRepo}
	}

	ctx := ciRepoContext{
		Directory: absDir,
		Target: ciRepoTarget{
			Owner:    target.Owner,
			Repo:     target.Repo,
			FullName: target.Owner + "/" + target.Repo,
		},
	}

	projectRoot, cfg, ok := loadCIProjectConfig(absDir)
	if ok && gitErr == nil && gitOwner == target.Owner && gitRepo == target.Repo {
		ctx.ProjectRoot = projectRoot
		ctx.ConfigEligible = true
		if cfg.GitHub.Repository == ctx.Target.FullName && cfg.GitHub.DefaultBranch != "" {
			ctx.DefaultBranch = cfg.GitHub.DefaultBranch
			ctx.DefaultBranchSrc = ".kit.yaml"
		}
	}

	return ctx, target, nil
}

func resolveGitHubRepoFromDir(dir string) (string, string, error) {
	output, err := ciRunner.Output(dir, "git", "remote", "get-url", "origin")
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve current repo from git remote origin: %w", err)
	}
	return parseGitHubRemoteURL(strings.TrimSpace(string(output)))
}

func resolveCIPRTarget(raw, dir string) (dispatchPRTarget, error) {
	if number, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
		owner, repo, err := resolveGitHubRepoFromDir(dir)
		if err != nil {
			return dispatchPRTarget{}, err
		}
		return dispatchPRTarget{Owner: owner, Repo: repo, Number: number}, nil
	}
	return resolveDispatchPRTarget(raw)
}

func loadCIProjectConfig(start string) (string, *config.Config, bool) {
	dir := start
	for {
		if config.Exists(dir) {
			cfg, err := config.Load(dir)
			if err != nil {
				return "", nil, false
			}
			return dir, cfg, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil, false
		}
		dir = parent
	}
}

func requireGHAuth(dir string) error {
	if _, err := ciRunner.Output(dir, "gh", "auth", "status"); err != nil {
		return fmt.Errorf("gh is unavailable or unauthenticated; run `gh auth login`: %w", err)
	}
	return nil
}

func needsDefaultBranch(opts ciOptions) bool {
	return strings.TrimSpace(opts.RunID) == "" && strings.TrimSpace(opts.PRRef) == ""
}

func discoverAndCacheDefaultBranch(ctx *ciRepoContext) error {
	output, err := ciRunner.Output(ctx.Directory, "gh", repoArgs(ctx.Target.FullName,
		"repo", "view", "--json", "nameWithOwner,defaultBranchRef")...)
	if err != nil {
		return fmt.Errorf("failed to discover GitHub default branch: %w", err)
	}
	var payload struct {
		NameWithOwner    string `json:"nameWithOwner"`
		DefaultBranchRef struct {
			Name string `json:"name"`
		} `json:"defaultBranchRef"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return fmt.Errorf("failed to parse GitHub default branch response: %w", err)
	}
	if payload.DefaultBranchRef.Name == "" {
		return fmt.Errorf("GitHub default branch response did not include a branch name")
	}
	ctx.DefaultBranch = payload.DefaultBranchRef.Name
	ctx.DefaultBranchSrc = "GitHub"
	if payload.NameWithOwner != "" {
		ctx.Target.FullName = payload.NameWithOwner
	}
	if ctx.ConfigEligible {
		if err := cacheCIDefaultBranch(*ctx); err != nil {
			return err
		}
	}
	return nil
}

func cacheCIDefaultBranch(ctx ciRepoContext) error {
	cfg, err := config.Load(ctx.ProjectRoot)
	if err != nil {
		return err
	}
	if cfg.GitHub.Repository == ctx.Target.FullName &&
		cfg.GitHub.DefaultBranch == ctx.DefaultBranch {
		return nil
	}
	cfg.GitHub.Repository = ctx.Target.FullName
	cfg.GitHub.DefaultBranch = ctx.DefaultBranch
	return config.Save(ctx.ProjectRoot, cfg)
}

func fetchCIPR(ctx ciRepoContext, target dispatchPRTarget) (ciPR, error) {
	ref := strconv.Itoa(target.Number)
	output, err := ciRunner.Output(ctx.Directory, "gh", repoArgs(ctx.Target.FullName,
		"pr", "view", ref, "--json", "number,headRefName,headRefOid,title,url")...)
	if err != nil {
		return ciPR{}, fmt.Errorf("failed to fetch PR metadata: %w", err)
	}
	var pr ciPR
	if err := json.Unmarshal(output, &pr); err != nil {
		return ciPR{}, fmt.Errorf("failed to parse PR metadata: %w", err)
	}
	return pr, nil
}

func fetchCIPRChecks(ctx ciRepoContext, target dispatchPRTarget) ([]ciCheck, error) {
	ref := strconv.Itoa(target.Number)
	output, err := ciRunner.OutputAllowError(ctx.Directory, "gh", repoArgs(ctx.Target.FullName,
		"pr", "checks", ref, "--json", "bucket,completedAt,description,event,link,name,startedAt,state,workflow")...)
	if err != nil && len(output) == 0 {
		return nil, fmt.Errorf("failed to fetch PR checks: %w", err)
	}
	var checks []ciCheck
	if err := json.Unmarshal(output, &checks); err != nil {
		return nil, fmt.Errorf("failed to parse PR checks: %w", err)
	}
	return checks, nil
}

const ciRunListJSONFields = "attempt,conclusion,createdAt,databaseId,displayTitle,event,headBranch,headSha,name,number,startedAt,status,updatedAt,url,workflowDatabaseId,workflowName"

func fetchCIRunList(ctx ciRepoContext, args ...string) ([]ciRun, error) {
	output, err := ciRunner.Output(ctx.Directory, "gh", repoArgs(ctx.Target.FullName, args...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to list GitHub Actions runs: %w", err)
	}
	var runs []ciRun
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub Actions runs: %w", err)
	}
	return runs, nil
}

func fetchCIRun(ctx ciRepoContext, runID string) (ciRun, error) {
	output, err := ciRunner.Output(ctx.Directory, "gh", repoArgs(ctx.Target.FullName,
		"run", "view", runID, "--json", "attempt,conclusion,createdAt,databaseId,displayTitle,event,headBranch,headSha,jobs,name,number,startedAt,status,updatedAt,url,workflowDatabaseId,workflowName")...)
	if err != nil {
		return ciRun{}, fmt.Errorf("failed to fetch GitHub Actions run %s: %w", runID, err)
	}
	var run ciRun
	if err := json.Unmarshal(output, &run); err != nil {
		return ciRun{}, fmt.Errorf("failed to parse GitHub Actions run %s: %w", runID, err)
	}
	return run, nil
}

func fetchFullCIRuns(ctx ciRepoContext, runs []ciRun) ([]ciRun, error) {
	full := make([]ciRun, 0, len(runs))
	for _, run := range runs {
		if len(run.Jobs) > 0 {
			full = append(full, run)
			continue
		}
		expanded, err := fetchCIRun(ctx, strconv.FormatInt(run.DatabaseID, 10))
		if err != nil {
			return nil, err
		}
		full = append(full, expanded)
	}
	return full, nil
}

func fetchCILog(ctx ciRepoContext, run ciRun, job ciJob) (string, error) {
	args := []string{"run", "view", strconv.FormatInt(run.DatabaseID, 10), "--log-failed"}
	if job.DatabaseID != 0 {
		args = append(args, "--job", strconv.FormatInt(job.DatabaseID, 10))
	}
	output, err := ciRunner.Output(ctx.Directory, "gh", repoArgs(ctx.Target.FullName, args...)...)
	if err != nil {
		return "", fmt.Errorf("failed to fetch failed log for run %d job %q: %w", run.DatabaseID, job.Name, err)
	}
	return string(output), nil
}

func repoArgs(repo string, args ...string) []string {
	result := append([]string{}, args...)
	if repo != "" {
		result = append(result, "--repo", repo)
	}
	return result
}

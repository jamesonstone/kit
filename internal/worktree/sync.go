package worktree

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const syncReportSchemaVersion = 1

// SyncOptions controls a worktree synchronization run.
type SyncOptions struct {
	DryRun bool `json:"dry_run"`
	JSON   bool `json:"-"`
}

// SyncOperation reports one command-level synchronization operation.
type SyncOperation struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// SyncDefaultDecision reports how the local origin default branch was handled.
type SyncDefaultDecision struct {
	Branch    string `json:"branch"`
	LocalOID  string `json:"local_oid,omitempty"`
	RemoteOID string `json:"remote_oid,omitempty"`
	Path      string `json:"path,omitempty"`
	State     string `json:"state"`
	Action    string `json:"action"`
	Detail    string `json:"detail,omitempty"`
}

// SyncPullRequest is the exact GitHub evidence used to decide a lane.
type SyncPullRequest struct {
	Number            int        `json:"number"`
	State             string     `json:"state"`
	MergedAt          *time.Time `json:"mergedAt"`
	BaseRefName       string     `json:"baseRefName"`
	HeadRefName       string     `json:"headRefName"`
	HeadRefOID        string     `json:"headRefOid"`
	IsCrossRepository bool       `json:"isCrossRepository"`
	URL               string     `json:"url"`
}

// SyncLaneDecision reports whether one registered worktree lane was preserved
// or removed, together with the exact safety proof or refusal reason.
type SyncLaneDecision struct {
	Path        string           `json:"path"`
	Branch      string           `json:"branch,omitempty"`
	HeadOID     string           `json:"head_oid,omitempty"`
	Action      string           `json:"action"`
	Reason      string           `json:"reason"`
	Detail      string           `json:"detail,omitempty"`
	PullRequest *SyncPullRequest `json:"pull_request,omitempty"`
}

// SyncWorktree summarizes one registered worktree after synchronization.
type SyncWorktree struct {
	State       string `json:"state"`
	Head        string `json:"head"`
	LastUpdated string `json:"last_updated"`
	Path        string `json:"path"`
}

// SyncFailure records an operation failure without hiding independent results.
type SyncFailure struct {
	Operation string `json:"operation"`
	Path      string `json:"path,omitempty"`
	Error     string `json:"error"`
}

// SyncReport is the single source for human and JSON sync output.
type SyncReport struct {
	SchemaVersion int                 `json:"schema_version"`
	Result        string              `json:"result"`
	DryRun        bool                `json:"dry_run"`
	Repository    string              `json:"repository"`
	DefaultBranch SyncDefaultDecision `json:"default_branch"`
	Fetch         SyncOperation       `json:"fetch"`
	Lanes         []SyncLaneDecision  `json:"lanes"`
	Prune         SyncOperation       `json:"prune"`
	Worktrees     []SyncWorktree      `json:"worktrees"`
	Failures      []SyncFailure       `json:"failures"`
}

type syncPRResolverFunc func(
	context.Context,
	string,
	string,
	[]string,
) (map[string][]SyncPullRequest, error)

type syncRunError struct {
	count int
}

func (err syncRunError) Error() string {
	return fmt.Sprintf("sync completed with %d operation failure(s)", err.count)
}

func parseSyncOptions(args []string) (SyncOptions, error) {
	var options SyncOptions
	seen := make(map[string]bool)
	for _, arg := range args {
		switch arg {
		case "--dry-run":
			if seen[arg] {
				return SyncOptions{}, fmt.Errorf("duplicate sync flag %q", arg)
			}
			options.DryRun = true
		case "--json":
			if seen[arg] {
				return SyncOptions{}, fmt.Errorf("duplicate sync flag %q", arg)
			}
			options.JSON = true
		default:
			return SyncOptions{}, fmt.Errorf(
				"unknown sync flag %q (usage: git wt sync [--dry-run] [--json])",
				arg,
			)
		}
		seen[arg] = true
	}
	return options, nil
}

func (a *App) sync(ctx context.Context, cwd string, args []string) error {
	options, err := parseSyncOptions(args)
	if err != nil {
		return err
	}
	report := a.runSync(ctx, cwd, options)
	sortSyncFailures(report.Failures)
	if len(report.Failures) == 0 {
		if options.DryRun {
			report.Result = "dry-run"
		} else {
			report.Result = "ok"
		}
	} else {
		report.Result = "failed"
	}

	var outputErr error
	if options.JSON {
		outputErr = writeSyncJSON(a.out, report)
	} else {
		outputErr = writeSyncHuman(a.out, report)
	}
	if outputErr != nil {
		outputErr = fmt.Errorf("write output: %w", outputErr)
	}
	var runErr error
	if len(report.Failures) != 0 {
		runErr = syncRunError{count: len(report.Failures)}
	}
	return errors.Join(runErr, outputErr)
}

func (report *SyncReport) addFailure(operation, path string, err error) {
	report.Failures = append(report.Failures, SyncFailure{
		Operation: operation,
		Path:      path,
		Error:     err.Error(),
	})
}

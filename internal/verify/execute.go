package verify

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RunStatus string

const (
	RunStatusPass             RunStatus = "pass"
	RunStatusFail             RunStatus = "fail"
	RunStatusDryRun           RunStatus = "dry_run"
	RunStatusNoDeclaredChecks RunStatus = "no_declared_checks"
)

type CommandResult struct {
	CommandID  string    `json:"command_id"`
	TaskID     string    `json:"task_id,omitempty"`
	Argv       []string  `json:"argv"`
	Raw        string    `json:"raw"`
	Shell      bool      `json:"shell"`
	CWD        string    `json:"cwd"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	DurationMS int64     `json:"duration_ms"`
	ExitCode   int       `json:"exit_code"`
	Status     string    `json:"status"`
	Error      string    `json:"error,omitempty"`
	TimedOut   bool      `json:"timed_out,omitempty"`
	StdoutPath string    `json:"stdout_path,omitempty"`
	StderrPath string    `json:"stderr_path,omitempty"`
	Redacted   bool      `json:"redacted,omitempty"`
	Stdout     string    `json:"-"`
	Stderr     string    `json:"-"`
}

type Run struct {
	SchemaVersion int             `json:"schema_version"`
	RunID         string          `json:"run_id"`
	ParentRunID   string          `json:"parent_run_id,omitempty"`
	Feature       FeatureRef      `json:"feature"`
	TaskIDs       []string        `json:"task_ids,omitempty"`
	ExpectedFiles []string        `json:"expected_files,omitempty"`
	Commands      []Command       `json:"commands"`
	Results       []CommandResult `json:"results"`
	Status        RunStatus       `json:"status"`
	StartedAt     time.Time       `json:"started_at"`
	EndedAt       time.Time       `json:"ended_at"`
	ArtifactDir   string          `json:"artifact_dir,omitempty"`
}

type RunOptions struct {
	ProjectRoot   string
	Feature       FeatureRef
	TaskIDs       []string
	ExpectedFiles []string
	Commands      []Command
	DryRun        bool
	Timeout       time.Duration
	ParentRunID   string
}

func NewRunID(now time.Time) string {
	return now.UTC().Format("20060102T150405.000000000Z") + "-" + randomSuffix()
}

func randomSuffix() string {
	var data [3]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "000000"
	}
	return hex.EncodeToString(data[:])
}

func ExecuteRun(ctx context.Context, opts RunOptions) Run {
	startedAt := time.Now().UTC()
	run := Run{
		SchemaVersion: SchemaVersion,
		RunID:         NewRunID(startedAt),
		ParentRunID:   opts.ParentRunID,
		Feature:       opts.Feature,
		TaskIDs:       append([]string(nil), opts.TaskIDs...),
		ExpectedFiles: append([]string(nil), opts.ExpectedFiles...),
		Commands:      append([]Command(nil), opts.Commands...),
		StartedAt:     startedAt,
	}

	if opts.DryRun {
		run.Status = RunStatusDryRun
		run.EndedAt = time.Now().UTC()
		return run
	}
	if len(opts.Commands) == 0 {
		run.Status = RunStatusNoDeclaredChecks
		run.EndedAt = time.Now().UTC()
		return run
	}

	status := RunStatusPass
	for _, command := range opts.Commands {
		result := executeCommand(ctx, opts.ProjectRoot, command, opts.Timeout)
		if result.Status != "pass" {
			status = RunStatusFail
		}
		run.Results = append(run.Results, result)
	}
	run.Status = status
	run.EndedAt = time.Now().UTC()
	return run
}

func executeCommand(ctx context.Context, projectRoot string, command Command, timeout time.Duration) CommandResult {
	startedAt := time.Now().UTC()
	result := CommandResult{
		CommandID: command.ID,
		TaskID:    command.TaskID,
		Argv:      append([]string(nil), command.Argv...),
		Raw:       command.Raw,
		Shell:     command.Shell,
		CWD:       command.CWD,
		StartedAt: startedAt,
		ExitCode:  -1,
		Status:    "fail",
	}
	if result.CWD == "" {
		result.CWD = projectRoot
	}
	if !filepath.IsAbs(result.CWD) {
		result.CWD = filepath.Join(projectRoot, result.CWD)
	}

	if len(command.Argv) == 0 {
		result.Error = "command argv is empty"
		result.EndedAt = time.Now().UTC()
		result.DurationMS = result.EndedAt.Sub(result.StartedAt).Milliseconds()
		return result
	}

	commandCtx := ctx
	cancel := func() {}
	if timeout > 0 {
		commandCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	cmd := exec.CommandContext(commandCtx, command.Argv[0], command.Argv[1:]...)
	cmd.Dir = result.CWD
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.EndedAt = time.Now().UTC()
	result.DurationMS = result.EndedAt.Sub(result.StartedAt).Milliseconds()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	if commandCtx.Err() != nil && errors.Is(commandCtx.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
		result.Error = "command timed out"
		return result
	}
	if err != nil {
		result.ExitCode = exitCode(err)
		result.Error = err.Error()
		return result
	}

	result.ExitCode = 0
	result.Status = "pass"
	return result
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func RunSummary(run Run) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("run %s: %s", run.RunID, run.Status))
	if len(run.TaskIDs) > 0 {
		builder.WriteString(fmt.Sprintf(" tasks=%s", strings.Join(run.TaskIDs, ",")))
	}
	if len(run.Results) > 0 {
		builder.WriteString(fmt.Sprintf(" commands=%d", len(run.Results)))
	}
	return builder.String()
}

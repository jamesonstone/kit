package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var loopResultPattern = regexp.MustCompile(`(?m)^KIT_LOOP_RESULT:\s*(\{.*\})\s*$`)

func runLoopAgent(ctx context.Context, opts loopOptions, stage loopStage, iteration int, prompt string) loopAgentExecution {
	cmd := exec.CommandContext(ctx, opts.Agent.Command, opts.Agent.Args...)
	cmd.Dir = opts.ProjectRoot
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = append(os.Environ(),
		"KIT_LOOP_STAGE="+string(stage),
		"KIT_LOOP_FEATURE="+opts.Feature.DirName,
		fmt.Sprintf("KIT_LOOP_MIN_CONFIDENCE=%d", opts.MinConfidence),
		fmt.Sprintf("KIT_LOOP_ITERATION=%d", iteration),
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return loopAgentExecution{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: commandExitCode(err),
		Err:      err,
	}
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func parseLoopAgentResult(stdout, stderr string) (*loopAgentResult, error) {
	matches := loopResultPattern.FindAllStringSubmatch(stdout+"\n"+stderr, -1)
	if len(matches) == 0 {
		return nil, errors.New("agent output did not include KIT_LOOP_RESULT JSON")
	}
	raw := matches[len(matches)-1][1]
	var result loopAgentResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("invalid KIT_LOOP_RESULT JSON: %w", err)
	}
	return &result, nil
}

func validateLoopAgentResult(result loopAgentResult, expected loopStage, minConfidence int) error {
	if result.Stage != expected {
		return fmt.Errorf("agent reported stage %q, expected %q", result.Stage, expected)
	}
	if result.Status != "done" {
		if len(result.Blockers) > 0 {
			return fmt.Errorf("agent blocked at %s: %s", expected, strings.Join(result.Blockers, "; "))
		}
		return fmt.Errorf("agent reported status %q at %s", result.Status, expected)
	}
	if len(result.Blockers) > 0 {
		return fmt.Errorf("agent reported blockers at %s: %s", expected, strings.Join(result.Blockers, "; "))
	}
	if result.Confidence < minConfidence {
		return fmt.Errorf("agent confidence %d is below required %d", result.Confidence, minConfidence)
	}
	return nil
}

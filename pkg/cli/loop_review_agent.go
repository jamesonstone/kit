package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jamesonstone/kit/internal/config"
)

func runLoopReviewAgent(ctx context.Context, opts loopReviewOptions, iteration int, prompt string) loopAgentExecution {
	cmd := exec.CommandContext(ctx, opts.Agent.Command, opts.Agent.Args...)
	cmd.Dir = opts.ProjectRoot
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = append(os.Environ(),
		"KIT_LOOP_MODE=review",
		fmt.Sprintf("KIT_LOOP_MIN_CONFIDENCE=%d", opts.MinConfidence),
		fmt.Sprintf("KIT_LOOP_ITERATION=%d", iteration),
	)
	if opts.Feature != nil {
		cmd.Env = append(cmd.Env, "KIT_LOOP_FEATURE="+opts.Feature.DirName)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if opts.Progress != nil {
		cmd.Stdout = io.MultiWriter(&stdout, newLoopReviewStreamWriter(opts.Progress, "agent stdout"))
		cmd.Stderr = io.MultiWriter(&stderr, newLoopReviewStreamWriter(opts.Progress, "agent stderr"))
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}
	startedAt := time.Now()
	if err := cmd.Start(); err != nil {
		return loopAgentExecution{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: commandExitCode(err),
			Err:      err,
		}
	}
	loopReviewProgress(opts, "iteration %d: agent process started pid=%d; waiting for completion", iteration, cmd.Process.Pid)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	ticker := time.NewTicker(loopReviewProgressEvery)
	defer ticker.Stop()

	var err error
	for {
		select {
		case err = <-done:
			return loopAgentExecution{
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				ExitCode: commandExitCode(err),
				Err:      err,
			}
		case <-ticker.C:
			loopReviewProgress(opts, "iteration %d: still waiting for agent (%s elapsed)", iteration, time.Since(startedAt).Round(time.Second))
		}
	}
}

func loopReviewAgentCommandSummary(agent config.LoopAgentConfig) string {
	parts := append([]string{agent.Command}, agent.Args...)
	for i, part := range parts {
		if strings.ContainsAny(part, " \t\n") {
			parts[i] = strconv.Quote(part)
		}
	}
	return strings.Join(parts, " ")
}

func loopReviewProgress(opts loopReviewOptions, format string, args ...any) {
	if opts.Progress == nil {
		return
	}
	message := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(opts.Progress, "%s%s %s\n", loopReviewProgressPrefix(), loopReviewProgressIcon(message), message)
}

func loopReviewProgressPrefix() string {
	return fmt.Sprintf("[loop-review %s] ", time.Now().UTC().Format("15:04:05Z"))
}

func loopReviewProgressIcon(message string) string {
	switch {
	case strings.Contains(message, "failed"), strings.Contains(message, "stopping"), strings.Contains(message, "stopped"):
		return "❌"
	case strings.Contains(message, "complete"):
		return "✅"
	case strings.Contains(message, "waiting"), strings.Contains(message, "pending"), strings.Contains(message, "still"):
		return "⏳"
	case strings.Contains(message, "running agent"), strings.Contains(message, "agent process"), strings.Contains(message, "single-agent"), strings.Contains(message, "subagent"):
		return "🤖"
	case strings.Contains(message, "checking"), strings.Contains(message, "fetching"):
		return "🔎"
	case strings.Contains(message, "artifacts"), strings.Contains(message, "prompt written"):
		return "📁"
	case strings.Contains(message, "building prompt"):
		return "📝"
	case strings.Contains(message, "target"), strings.Contains(message, "base="):
		return "🎯"
	case strings.Contains(message, "continuing"):
		return "🔁"
	default:
		return "▶️"
	}
}

type loopReviewSynchronizedWriter struct {
	mu     sync.Mutex
	writer io.Writer
}

func (writer *loopReviewSynchronizedWriter) Write(p []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return writer.writer.Write(p)
}

type loopReviewStreamWriter struct {
	writer      io.Writer
	label       string
	atLineStart bool
}

func newLoopReviewStreamWriter(writer io.Writer, label string) *loopReviewStreamWriter {
	return &loopReviewStreamWriter{
		writer:      writer,
		label:       label,
		atLineStart: true,
	}
}

func (writer *loopReviewStreamWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	var output bytes.Buffer
	for _, b := range p {
		if writer.atLineStart {
			output.WriteString(loopReviewProgressPrefix())
			output.WriteString(loopReviewStreamIcon(writer.label))
			output.WriteString(" ")
			output.WriteString(writer.label)
			output.WriteString(": ")
			writer.atLineStart = false
		}
		output.WriteByte(b)
		if b == '\n' {
			writer.atLineStart = true
		}
	}
	if _, err := writer.writer.Write(output.Bytes()); err != nil {
		return 0, err
	}
	return len(p), nil
}

func loopReviewStreamIcon(label string) string {
	if strings.Contains(label, "stderr") {
		return "⚠️"
	}
	return "💬"
}

func parseLoopReviewAgentResult(stdout string) loopReviewAgentResult {
	result := loopReviewAgentResult{RawSummary: strings.TrimSpace(stdout)}
	for _, match := range loopReviewCorrectnessPattern.FindAllStringSubmatch(stdout, -1) {
		value, err := strconv.Atoi(match[1])
		if err == nil {
			result.Correctness = clampPercentage(value)
		}
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			result.Bullets = append(result.Bullets, trimmed)
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		result.Done = strings.TrimSpace(lines[i]) == "done"
		break
	}
	return result
}

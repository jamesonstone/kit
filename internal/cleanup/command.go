package cleanup

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runCommand(ctx context.Context, cwd, name string, args ...string) ([]byte, error) {
	return runCommandInput(ctx, cwd, name, nil, args...)
}

func runCommandInput(ctx context.Context, cwd, name string, input []byte, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewReader(input)
	cmd.Dir = cwd
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "GIT_") {
			continue
		}
		cmd.Env = append(cmd.Env, entry)
	}
	cmd.Env = append(cmd.Env, "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0", "GIT_NO_REPLACE_OBJECTS=1")
	var diagnostic strings.Builder
	cmd.Stderr = &diagnostic
	output, err := cmd.Output()
	if err != nil || diagnostic.Len() != 0 {
		return nil, fmt.Errorf("%s failed or returned diagnostics", name)
	}
	return output, nil
}

func (s *Service) git(ctx context.Context, repo string, args ...string) ([]byte, error) {
	options := []string{"-c", "core.hooksPath=" + os.DevNull, "-c", "core.fsmonitor=false"}
	return s.run(ctx, repo, "git", append(options, args...)...)
}

func (s *Service) value(ctx context.Context, repo string, args ...string) (string, error) {
	out, err := s.git(ctx, repo, args...)
	return strings.TrimSpace(string(out)), err
}

// gitInput streams NUL-delimited path lists instead of imposing ARG_MAX limits.
func (s *Service) gitInput(ctx context.Context, repo string, input []byte, args ...string) ([]byte, error) {
	options := []string{"-c", "core.hooksPath=" + os.DevNull, "-c", "core.fsmonitor=false"}
	return runCommandInput(ctx, repo, "git", input, append(options, args...)...)
}

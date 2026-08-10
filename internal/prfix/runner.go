package prfix

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Runner interface {
	Run(context.Context, string, string, ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, directory, name string, args ...string) ([]byte, error) {
	// #nosec G204 -- callers select git/gh/editor executables and pass arguments without a shell.
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err == nil {
		return output, nil
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		return output, err
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return output, fmt.Errorf("%w: %s", err, message)
	}
	return output, fmt.Errorf("%w: %s", err, message)
}

type HTTPError struct {
	Status     int
	RetryAfter string
	ResetAt    string
	Cause      error
}

func (err *HTTPError) Error() string {
	return fmt.Sprintf("GitHub returned HTTP %d: %v", err.Status, err.Cause)
}

func (err *HTTPError) Unwrap() error { return err.Cause }

type responseMetadata struct {
	status     int
	retryAfter string
}

var httpStatusPattern = regexp.MustCompile(`(?m)^HTTP/\S+\s+([0-9]{3})\b`)

func splitIncludedResponse(output []byte) ([]byte, responseMetadata) {
	normalized := strings.ReplaceAll(string(output), "\r\n", "\n")
	metadata := responseMetadata{}
	for _, match := range httpStatusPattern.FindAllStringSubmatch(normalized, -1) {
		metadata.status, _ = strconv.Atoi(match[1])
	}
	separator := strings.LastIndex(normalized, "\n\n")
	if separator < 0 {
		return output, metadata
	}
	headers, body := normalized[:separator], normalized[separator+2:]
	for _, line := range strings.Split(headers, "\n") {
		name, value, found := strings.Cut(line, ":")
		if found && strings.EqualFold(strings.TrimSpace(name), "retry-after") {
			metadata.retryAfter = strings.TrimSpace(value)
		}
	}
	return []byte(body), metadata
}

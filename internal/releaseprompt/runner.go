package releaseprompt

import (
	"context"
	"fmt"
	"os/exec"
)

type Runner interface {
	Run(context.Context, string, string, ...string) ([]byte, error)
	LookPath(string) (string, error)
}

type SystemRunner struct{}

func (SystemRunner) Run(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = dir
	output, err := command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s %v: %s", name, args, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("%s %v: %w", name, args, err)
	}
	return output, nil
}

func (SystemRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

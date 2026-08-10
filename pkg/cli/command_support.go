package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
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
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err == nil {
		return output, nil
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return nil, err
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return output, fmt.Errorf("%s: %s", err, detail)
	}
	return output, fmt.Errorf("%w: %s", err, detail)
}

type ciTarget struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch,omitempty"`
	PRNumber   int    `json:"pr_number,omitempty"`
	HeadSHA    string `json:"head_sha,omitempty"`
}

type ciDiagnosis struct {
	Target ciTarget `json:"target"`
}

func outputJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func repoArgs(repository string, args ...string) []string {
	result := append([]string(nil), args...)
	if strings.TrimSpace(repository) != "" {
		result = append(result, "--repo", repository)
	}
	return result
}

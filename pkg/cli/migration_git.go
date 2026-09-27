package cli

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitRestorablePaths returns the paths Git can restore after deletion: tracked
// and identical to the committed version. Removing anything else would lose
// data, so migration keeps it. Outside a Git work tree nothing is restorable.
func gitRestorablePaths(projectRoot string, relativePaths []string) (map[string]bool, error) {
	restorable := map[string]bool{}
	if len(relativePaths) == 0 {
		return restorable, nil
	}
	if out, err := gitOutput(projectRoot, "rev-parse", "--is-inside-work-tree"); err != nil || strings.TrimSpace(out) != "true" {
		return restorable, nil
	}
	args := append([]string{"ls-files", "-z", "--"}, relativePaths...)
	tracked, err := gitOutput(projectRoot, args...)
	if err != nil {
		return nil, err
	}
	for _, path := range strings.Split(tracked, "\x00") {
		if path != "" {
			restorable[filepath.ToSlash(path)] = true
		}
	}
	// Staged or unstaged differences from HEAD, including files never committed.
	args = append([]string{"diff", "HEAD", "--name-only", "--relative", "-z", "--"}, relativePaths...)
	dirty, err := gitOutput(projectRoot, args...)
	if err != nil {
		return map[string]bool{}, nil // no commit yet: nothing is restorable
	}
	for _, path := range strings.Split(dirty, "\x00") {
		delete(restorable, filepath.ToSlash(path))
	}
	return restorable, nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	return stdout.String(), err
}

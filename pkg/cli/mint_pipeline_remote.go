package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// mintRemoteDefaultBranch queries origin's current symbolic HEAD without fetching
// or updating refs/config. Never substitute potentially stale origin/HEAD here.
func mintRemoteDefaultBranch(root string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-c", "core.askPass=", "ls-remote", "--symref", "--exit-code", "origin", "HEAD")
	cmd.Dir = root
	cmd.WaitDelay = time.Second
	// Keep configured remotes/credential helpers and ordinary SSH host settings.
	// Override interactive/custom SSH commands: remote discovery must never prompt.
	overrides := map[string]string{"GIT_TERMINAL_PROMPT": "0", "GCM_INTERACTIVE": "never", "GIT_ASKPASS": "", "SSH_ASKPASS_REQUIRE": "never", "GIT_SSH_VARIANT": "ssh", "GIT_SSH_COMMAND": "ssh -oBatchMode=yes -oNumberOfPasswordPrompts=0 -oConnectTimeout=10"}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if _, replaced := overrides[key]; !replaced {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	for key, value := range overrides {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	var output mintRemoteBuffer
	cmd.Stdout = &output
	// Do not echo stderr: Git transport diagnostics can contain credential URLs.
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("INCOMPLETE: could not verify current origin HEAD; check remote access and its default branch, then rerun Mint adoption (no refs or credentials were changed)")
	}
	return parseMintRemoteHEAD(output.Bytes())
}

type mintRemoteBuffer struct{ data bytes.Buffer }

func (b *mintRemoteBuffer) Bytes() []byte { return b.data.Bytes() }

func (b *mintRemoteBuffer) Write(p []byte) (int, error) {
	if b.data.Len()+len(p) > 64*1024 {
		return 0, fmt.Errorf("remote HEAD response exceeds limit")
	}
	return b.data.Write(p)
}

func parseMintRemoteHEAD(output []byte) (string, error) {
	branch := ""
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if fields[0] != "ref:" {
			continue
		}
		if len(fields) != 3 || fields[2] != "HEAD" || !strings.HasPrefix(fields[1], "refs/heads/") || branch != "" {
			return "", fmt.Errorf("INCOMPLETE: current origin HEAD has ambiguous symbolic branch metadata; verify the remote default branch before Mint adoption")
		}
		branch = strings.TrimPrefix(fields[1], "refs/heads/")
		if !safeMintPipelineBranch(branch) {
			return "", fmt.Errorf("INCOMPLETE: current origin HEAD is not a safe branch; verify the remote default branch before Mint adoption")
		}
	}
	if branch == "" {
		return "", fmt.Errorf("INCOMPLETE: origin did not advertise a symbolic default branch; verify current remote HEAD before Mint adoption")
	}
	return branch, nil
}

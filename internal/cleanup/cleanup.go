// Package cleanup proves bounded task-artifact recovery immediately before mutation.
package cleanup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type commandFunc func(context.Context, string, string, ...string) ([]byte, error)

// Receipt records task ownership at creation; it does not authorize deletion.
type Receipt struct {
	Version                                        int
	Kind, Path, Repository, Owner, Ref, RefOID     string
	Identity, GitDir, GitIdentity, GeneratedDigest string
	Links                                          map[string]string
}

// Options contains the current authorized owner's handoff, never inferred from age or names.
type Options struct {
	Owner, FinishedHead, StopEvidence, Remote string
	PR                                        int
	WritersStopped, Apply                     bool
}

// Result reports actual outcomes; Bytes is removed logical file bytes, not measured disk blocks.
type Result struct {
	State, Reason string
	Bytes         int64
}

// Service owns proof and action; injected commands are used only by isolated tests.
type Service struct {
	run              commandFunc
	beforeRevalidate func()
}

func New() *Service { return &Service{run: runCommand} }

type retained struct{ state, reason string }

func (r retained) Error() string      { return r.reason }
func keep(state, reason string) error { return retained{state, reason} }
func outcome(err error) Result {
	var retained retained
	if errors.As(err, &retained) {
		return Result{State: retained.state, Reason: retained.reason}
	}
	return Result{State: "UNAVAILABLE", Reason: "a required identity, filesystem, Git, remote or process check failed"}
}

type action struct {
	fingerprint string
	bytes       int64
	state       string
	apply       func() error
}

// Execute repeats every proof under a cleanup lease and acts only on unchanged evidence.
// The lease coordinates cooperating writers; it is not an OS sandbox for arbitrary external writers.
func (s *Service) Execute(ctx context.Context, r Receipt, o Options) Result {
	if r.Version != 1 || r.Owner == "" || r.Owner != o.Owner || !o.WritersStopped {
		return Result{State: "KEEP_OWNERSHIP", Reason: "creation ownership and a current exclusive owner handoff are required"}
	}
	if o.Remote == "" {
		o.Remote = "origin"
	}
	if strings.HasPrefix(o.Remote, "-") || strings.ContainsAny(o.Remote, "\r\n\x00") {
		return Result{State: "UNAVAILABLE", Reason: "invalid selected remote"}
	}
	// Reject aliases or replaced parents as well as a replaced final directory.
	for _, path := range []string{r.Repository, r.Path} {
		if _, err := canonical(path); err != nil {
			// Missing worktree paths are handled by the recorded registration proof.
			if path == r.Path && r.Kind == "worktree" && os.IsNotExist(err) {
				continue
			}
			return Result{State: "KEEP_IDENTITY", Reason: "recorded physical path changed or is unavailable"}
		}
	}
	common, err := s.value(ctx, r.Repository, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return outcome(err)
	}
	if err = s.recoveryStorage(ctx, r.Repository, common); err != nil {
		return outcome(err)
	}
	if o.Apply {
		lock := filepath.Join(common, "kit-cleanup.lock")
		file, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return Result{State: "KEEP_CONCURRENT", Reason: "cleanup lease is unavailable; never remove another owner's lease"}
		}
		owned, statErr := file.Stat()
		_ = file.Close()
		if statErr != nil {
			return outcome(statErr)
		}
		defer func() {
			if current, err := os.Lstat(lock); err == nil && os.SameFile(owned, current) {
				_ = os.Remove(lock)
			}
		}()
	}
	first, err := s.plan(ctx, r, o)
	if err != nil {
		return outcome(err)
	}
	if !o.Apply {
		return Result{State: "ELIGIBLE", Reason: "preview only; apply repeats all evidence checks"}
	}
	if s.beforeRevalidate != nil {
		s.beforeRevalidate()
	}
	second, err := s.plan(ctx, r, o)
	if err != nil {
		return outcome(err)
	}
	if first.fingerprint != second.fingerprint {
		return Result{State: "KEEP_CONCURRENT", Reason: "proof changed before mutation"}
	}
	if err = second.apply(); err != nil {
		return Result{State: "REFUSED_OR_PARTIAL", Reason: "native operation failed after mutation began; inspect actual state and use retained recovery evidence; no force/reset/clean fallback"}
	}
	return Result{State: second.state, Bytes: second.bytes}
}

func fingerprint(values ...string) string {
	data, _ := json.Marshal(values)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (s *Service) plan(ctx context.Context, r Receipt, o Options) (action, error) {
	if o.FinishedHead == "" {
		return action{}, keep("KEEP_COMPLETION", "exact completed source head is required")
	}
	if err := s.completed(ctx, r, o); err != nil {
		return action{}, err
	}
	if r.GitDir != "" && r.Kind != "worktree" {
		current, err := s.value(ctx, r.Path, "rev-parse", "--absolute-git-dir")
		if err != nil {
			return action{}, err
		}
		id, err := identity(current)
		if err != nil {
			return action{}, err
		}
		if current != r.GitDir || id != r.GitIdentity {
			return action{}, keep("KEEP_IDENTITY", "repository metadata identity changed")
		}
	}
	if r.Kind != "worktree" {
		id, err := identity(r.Path)
		if err != nil {
			return action{}, err
		}
		if id != r.Identity {
			return action{}, keep("KEEP_IDENTITY", "recorded target identity changed")
		}
	}
	if err := s.idle(ctx, r); err != nil {
		return action{}, err
	}
	switch r.Kind {
	case "worktree":
		return s.worktree(ctx, r, o)
	case "primary":
		return s.primary(ctx, r, o)
	case "branch", "ref":
		return s.reference(ctx, r, o)
	case "scratch", "cache":
		return s.generated(ctx, r)
	default:
		return action{}, keep("KEEP_UNSUPPORTED", "no recovery adapter for this category")
	}
}

func (s *Service) completed(ctx context.Context, r Receipt, o Options) error {
	if o.PR <= 0 {
		if strings.TrimSpace(o.StopEvidence) == "" {
			return keep("KEEP_COMPLETION", "fresh merged/closed PR or current human stop evidence is required")
		}
		return nil
	}
	remoteURL, err := s.value(ctx, r.Repository, "remote", "get-url", o.Remote)
	if err != nil {
		return err
	}
	if strings.HasPrefix(remoteURL, "git@") {
		host, path, ok := strings.Cut(strings.TrimPrefix(remoteURL, "git@"), ":")
		if ok {
			remoteURL = "https://" + host + "/" + path
		}
	}
	out, err := s.run(ctx, r.Repository, "gh", "pr", "view", fmt.Sprint(o.PR), "--repo", strings.TrimSuffix(remoteURL, ".git"), "--json", "state,headRefOid")
	if err != nil {
		return err
	}
	var pr struct{ State, HeadRefOid string }
	if json.Unmarshal(out, &pr) != nil {
		return keep("UNAVAILABLE", "invalid remote completion response")
	}
	if (pr.State != "MERGED" && pr.State != "CLOSED") || pr.HeadRefOid != o.FinishedHead {
		return keep("KEEP_COMPLETION", "remote completion does not match this handoff")
	}
	return nil
}

func (s *Service) idle(ctx context.Context, r Receipt) error {
	out, err := s.run(ctx, os.TempDir(), "lsof", "-nP", "-Fpn")
	if err != nil || len(out) == 0 {
		return keep("UNAVAILABLE", "complete process visibility is unavailable")
	}
	seenProcess := false
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		if line[0] == 'p' {
			if len(line) == 1 || strings.Trim(line[1:], "0123456789") != "" {
				return keep("UNAVAILABLE", "malformed process inventory")
			}
			seenProcess = true
		} else if (line[0] != 'n' && line[0] != 'f') || len(line) == 1 || !seenProcess {
			return keep("UNAVAILABLE", "malformed process inventory")
		}
		if strings.HasPrefix(line, "n") {
			path := strings.TrimSuffix(strings.TrimPrefix(line, "n"), " (deleted)")
			if filepath.IsAbs(path) && (contains(r.Path, path) || (r.GitDir != "" && contains(r.GitDir, path))) {
				return keep("KEEP_IN_USE", "a process uses the target or its Git metadata")
			}
		}
	}
	if !seenProcess {
		return keep("UNAVAILABLE", "process inventory is incomplete")
	}
	return nil
}

// DefaultTimeout bounds external evidence acquisition.
const DefaultTimeout = 2 * time.Minute

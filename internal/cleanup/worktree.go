package cleanup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) worktree(ctx context.Context, r Receipt, o Options) (action, error) {
	entries, err := s.worktrees(ctx, r.Repository)
	if err != nil {
		return action{}, err
	}
	var target *worktree
	for i := range entries {
		if entries[i].Path == r.Path {
			target = &entries[i]
		}
	}
	if target == nil || r.Path == entries[0].Path {
		return action{}, keep("KEEP_IDENTITY", "only an exact linked worktree registration can be removed")
	}
	if target.Locked {
		return action{}, keep("KEEP_IN_USE", "worktree is locked")
	}
	if target.Head != o.FinishedHead {
		return action{}, keep("KEEP_COMPLETION", "lane head changed after completion")
	}
	recovery, err := s.recoverHead(ctx, r.Repository, o.Remote, target.Head)
	if err != nil {
		return action{}, err
	}
	metadata, err := s.metadata(ctx, r, target.Head, target.Branch)
	if err != nil {
		return action{}, err
	}
	var tree treeSnapshot
	id, e := identity(r.Path)
	if e != nil && !os.IsNotExist(e) {
		return action{}, e
	}
	if e == nil {
		info, statErr := os.Stat(r.Path)
		if statErr != nil {
			return action{}, statErr
		}
		if info.Mode() != os.ModeDir|0755 {
			return action{}, keep("KEEP_LOCAL_STATE", "worktree root has noncanonical permissions")
		}
		control, controlErr := os.Lstat(filepath.Join(r.Path, ".git"))
		if controlErr != nil {
			return action{}, controlErr
		}
		if control.Mode() != 0644 {
			return action{}, keep("KEEP_LOCAL_STATE", "worktree control file has noncanonical permissions")
		}
		pointer, pointerErr := os.ReadFile(filepath.Join(r.Path, ".git"))
		if pointerErr != nil {
			return action{}, pointerErr
		}
		if string(pointer) != "gitdir: "+r.GitDir+"\n" {
			return action{}, keep("KEEP_IDENTITY", "worktree control pointer changed")
		}
		if err = s.cleanIndex(ctx, r.Path); err != nil {
			return action{}, err
		}

		if id != r.Identity {
			return action{}, keep("KEEP_IDENTITY", "worktree was replaced")
		}
		for name, link := range r.Links {
			if (name != ".env" && name != ".envrc") || link != filepath.Join(entries[0].Path, name) {
				return action{}, keep("KEEP_IDENTITY", "environment link receipt is outside primary recovery scope")
			}
			if info, e := os.Lstat(link); e != nil || !info.Mode().IsRegular() {
				return action{}, keep("KEEP_RECOVERY", "primary environment source is unavailable")
			}
		}
		tree, err = verifyTrackedTree(ctx, s.runGitCommand, r.Path, target.Head, r.Links)
		if err != nil {
			return action{}, keep("KEEP_LOCAL_STATE", "tracked bytes, ignored/untracked data or filesystem metadata are not fully recoverable")
		}
	}
	state := "REMOVED"
	if os.IsNotExist(e) {
		state = "PRUNED"
	}
	return action{fingerprint: fingerprint(id, target.Head, target.Branch, metadata, tree.Digest, recovery), bytes: tree.Bytes, state: state,
		apply: func() error {
			// Git refuses untracked environment links even after recovery proof.
			// Remove only these recorded links, never their primary source bytes.
			for name, expected := range r.Links {
				path := filepath.Join(r.Path, name)
				actual, e := os.Readlink(path)
				if os.IsNotExist(e) && state == "PRUNED" {
					continue
				}
				if e != nil || actual != expected {
					return keep("KEEP_CONCURRENT", "environment link changed")
				}
				if e = os.Remove(path); e != nil {
					return e
				}
			}
			_, e := s.git(ctx, r.Repository, "worktree", "remove", r.Path)
			return e
		}}, nil
}

// runGitCommand applies the same safe Git options to filesystem proof queries.
func (s *Service) runGitCommand(ctx context.Context, repo, name string, args ...string) ([]byte, error) {
	if name == "git" {
		return s.git(ctx, repo, args...)
	}
	return s.run(ctx, repo, name, args...)
}

// cleanIndex rejects hidden semantic state even when tracked bytes happen to match HEAD.
func (s *Service) cleanIndex(ctx context.Context, path string) error {
	flags, err := s.git(ctx, path, "ls-files", "-v", "-f", "-z")
	if err != nil {
		return err
	}
	for _, flag := range strings.Split(string(flags), "\x00") {
		if flag != "" && !strings.HasPrefix(flag, "H ") {
			return keep("KEEP_LOCAL_STATE", "nonstandard index flags require preservation")
		}
	}
	if _, err = s.git(ctx, path, "diff", "--cached", "--quiet", "--no-ext-diff", "--no-textconv", "HEAD", "--"); err != nil {
		return keep("KEEP_LOCAL_STATE", "index has staged or unknown state")
	}
	return nil
}

package cleanup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type worktree struct {
	Path, Head, Branch string
	Locked             bool
}

func (s *Service) worktrees(ctx context.Context, repo string) ([]worktree, error) {
	out, err := s.git(ctx, repo, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, err
	}
	var result []worktree
	for _, block := range strings.Split(string(out), "\x00\x00") {
		var entry worktree
		for _, field := range strings.Split(block, "\x00") {
			switch {
			case strings.HasPrefix(field, "worktree "):
				entry.Path = strings.TrimPrefix(field, "worktree ")
			case strings.HasPrefix(field, "HEAD "):
				entry.Head = strings.TrimPrefix(field, "HEAD ")
			case strings.HasPrefix(field, "branch "):
				entry.Branch = strings.TrimPrefix(field, "branch ")
			case strings.HasPrefix(field, "locked"):
				entry.Locked = true
			}
		}
		if entry.Path != "" {
			result = append(result, entry)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("missing worktree registration")
	}
	return result, nil
}

// liveHeads reads the selected remote directly. Cached remote-tracking refs never suffice.
func (s *Service) liveHeads(ctx context.Context, repo, remote string) (map[string]string, error) {
	out, err := s.git(ctx, repo, "ls-remote", remote, "refs/heads/*", "refs/pull/*/head")
	if err != nil {
		return nil, err
	}
	refs := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			refs[fields[1]] = fields[0]
		}
	}
	if len(refs) == 0 {
		return nil, keep("KEEP_RECOVERY", "selected remote has no readable recovery refs")
	}
	return refs, nil
}

func (s *Service) recoverHead(ctx context.Context, repo, remote, head string) (string, error) {
	refs, err := s.liveHeads(ctx, repo, remote)
	if err != nil {
		return "", err
	}
	names := []string{}
	for name, oid := range refs {
		if oid == head {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "", keep("KEEP_RECOVERY", "no live selected-remote ref holds the exact head")
	}
	sort.Strings(names)
	return names[0] + ":" + head, nil
}

func (s *Service) ancestors(ctx context.Context, repo string, oids, keepers []string) error {
	for _, oid := range oids {
		if strings.Trim(oid, "0") == "" {
			continue
		}
		if err := requireCanonicalOID(oid); err != nil {
			return keep("KEEP_RECOVERY", "invalid local recovery object")
		}
		protected := false
		for _, tip := range keepers {
			if _, err := s.git(ctx, repo, "merge-base", "--is-ancestor", oid, tip); err == nil {
				protected = true
				break
			}
		}
		if !protected {
			return keep("KEEP_RECOVERY", "a local reflog or reference commit lacks retained recovery")
		}
	}
	return nil
}

func reflogObjects(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("unreadable reflog")
		}
		result = append(result, fields[0], fields[1])
	}
	return result, nil
}

// metadata permits only known reconstructible Git control state; anything else is kept.
func (s *Service) metadata(ctx context.Context, r Receipt, head, branch string) (string, error) {
	id, err := identity(r.GitDir)
	if err != nil {
		return "", err
	}
	if id != r.GitIdentity {
		return "", keep("KEEP_IDENTITY", "Git metadata identity changed")
	}
	common, err := s.value(ctx, r.Repository, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	if r.GitDir != common {
		allowed := map[string]bool{".": true, "HEAD": true, "index": true, "commondir": true, "gitdir": true, "logs": true, "logs/HEAD": true, "ORIG_HEAD": true, "COMMIT_EDITMSG": true, "refs": true}
		err = filepath.WalkDir(r.GitDir, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, e := filepath.Rel(r.GitDir, path)
			if e != nil {
				return e
			}
			info, e := entry.Info()
			if e != nil {
				return e
			}
			mode := os.FileMode(0644)
			if entry.IsDir() {
				mode = os.ModeDir | 0755
			}
			if info.Mode() != mode {
				return keep("KEEP_LOCAL_STATE", "Git control modes are not reconstructible")
			}
			if !allowed[filepath.ToSlash(rel)] {
				return keep("KEEP_LOCAL_STATE", "worktree metadata contains local config, operations or unknown state")
			}
			return nil
		})
		if err != nil {
			return "", err
		}
	} else {
		// Primary/shared metadata cannot be deleted; block operations that make sync unsafe.
		for _, name := range []string{"index.lock", "HEAD.lock", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "BISECT_LOG", "rebase-merge", "rebase-apply", "sequencer", "config.worktree"} {
			if _, e := os.Lstat(filepath.Join(common, name)); e == nil {
				return "", keep("KEEP_LOCAL_STATE", "Git operation or worktree configuration is active")
			} else if !os.IsNotExist(e) {
				return "", e
			}
		}
	}
	if r.GitDir != common {
		if data, e := os.ReadFile(filepath.Join(r.GitDir, "ORIG_HEAD")); e == nil {
			if e = s.ancestors(ctx, r.Repository, []string{strings.TrimSpace(string(data))}, []string{head}); e != nil {
				return "", e
			}
		} else if !os.IsNotExist(e) {
			return "", e
		}
		if data, e := os.ReadFile(filepath.Join(r.GitDir, "COMMIT_EDITMSG")); e == nil {
			message, e := s.git(ctx, r.Repository, "log", "-1", "--format=%B", head)
			if e != nil {
				return "", e
			}
			if strings.TrimSpace(string(data)) != strings.TrimSpace(string(message)) {
				return "", keep("KEEP_LOCAL_STATE", "local commit draft requires preservation")
			}
		} else if !os.IsNotExist(e) {
			return "", e
		}
	}
	if r.GitDir == common && r.Kind != "branch" && r.Kind != "ref" {
		for _, name := range []string{"HEAD", "index", "ORIG_HEAD", "logs/HEAD", "logs/" + branch, branch, "packed-refs"} {
			if _, e := controlFingerprint(ctx, filepath.Join(common, name)); e != nil {
				return "", e
			}
		}
		if data, e := os.ReadFile(filepath.Join(common, "ORIG_HEAD")); e == nil {
			if e = s.ancestors(ctx, r.Repository, []string{strings.TrimSpace(string(data))}, []string{head}); e != nil {
				return "", e
			}
		} else if !os.IsNotExist(e) {
			return "", e
		}
	}
	var logs []string
	var paths []string
	if r.Kind != "branch" && r.Kind != "ref" {
		paths = append(paths, filepath.Join(r.GitDir, "logs", "HEAD"))
	}
	if branch != "" {
		paths = append(paths, filepath.Join(common, "logs", filepath.FromSlash(branch)))
	}
	for _, path := range paths {
		oids, e := reflogObjects(path)
		if e != nil {
			return "", e
		}
		logs = append(logs, oids...)
	}
	if err = s.ancestors(ctx, r.Repository, logs, []string{head}); err != nil {
		return "", err
	}
	config, err := s.git(ctx, r.Repository, "config", "--null", "--list")
	if err != nil {
		return "", err
	}
	prefix := "branch." + strings.TrimPrefix(branch, "refs/heads/") + "."
	for _, entry := range strings.Split(string(config), "\x00") {
		key, value, _ := strings.Cut(entry, "\n")
		if branch != "" && strings.HasPrefix(key, prefix) {
			suffix := strings.TrimPrefix(key, prefix)
			if suffix != "remote" && suffix != "merge" {
				return "", keep("KEEP_LOCAL_STATE", "branch-specific configuration requires preservation")
			}
			if suffix == "merge" && value != branch {
				return "", keep("KEEP_LOCAL_STATE", "nonstandard branch merge configuration")
			}
			if suffix == "remote" && value != "origin" {
				return "", keep("KEEP_LOCAL_STATE", "nonstandard branch recovery configuration")
			}
		}
	}
	if r.GitDir != common {
		snap, e := snapshotTree(r.GitDir, ctx)
		if e != nil {
			return "", e
		}
		return fingerprint(snap.Digest, string(config), strings.Join(logs, ",")), nil
	}
	if r.Kind == "branch" || r.Kind == "ref" {
		return fingerprint(string(config), strings.Join(logs, ",")), nil
	}
	// Primary index/config/HEAD are sufficient alongside tracked-tree and operation checks.
	values := []string{string(config), strings.Join(logs, ",")}
	state, e := controlStates(ctx, common, "HEAD", "index", "config", "ORIG_HEAD", "logs/HEAD", "logs/"+branch, branch, "packed-refs")
	if e != nil {
		return "", e
	}
	values = append(values, state)
	return fingerprint(values...), nil
}

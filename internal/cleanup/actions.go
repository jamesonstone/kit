package cleanup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) reference(ctx context.Context, r Receipt, o Options) (action, error) {
	entries, err := s.worktrees(ctx, r.Repository)
	if err != nil {
		return action{}, err
	}
	if r.Ref == "" {
		return action{}, keep("KEEP_IDENTITY", "an exact recorded reference is required")
	}
	if _, err = s.git(ctx, r.Repository, "check-ref-format", r.Ref); err != nil {
		return action{}, err
	}
	for _, entry := range entries {
		if entry.Branch == r.Ref {
			return action{}, keep("KEEP_IN_USE", "branch is checked out")
		}
	}
	head, err := s.value(ctx, r.Repository, "rev-parse", "--verify", r.Ref)
	if err != nil {
		return action{}, err
	}
	common, err := s.value(ctx, r.Repository, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return action{}, err
	}
	var recovery string
	var keepers []string
	if r.Kind == "branch" {
		if !strings.HasPrefix(r.Ref, "refs/heads/") {
			return action{}, keep("KEEP_IDENTITY", "branch receipt must name refs/heads")
		}
		base, _, e := s.remoteDefault(ctx, r.Repository, o.Remote)
		if e != nil {
			return action{}, e
		}
		if r.Ref == base {
			return action{}, keep("KEEP_DEFAULT", "the live remote default branch is protected")
		}
		if head != o.FinishedHead {
			return action{}, keep("KEEP_COMPLETION", "branch head changed after completion")
		}
		recovery, err = s.recoverHead(ctx, r.Repository, o.Remote, head)
		if err != nil {
			return action{}, err
		}
		keepers = []string{head}
		if _, err = s.metadata(ctx, r, head, r.Ref); err != nil {
			return action{}, err
		}
	} else {
		if r.RefOID == "" || head != r.RefOID {
			return action{}, keep("KEEP_CONCURRENT", "recorded tracking reference changed")
		}
		prefix := "refs/remotes/" + o.Remote + "/"
		if !strings.HasPrefix(r.Ref, prefix) || r.Ref == prefix+"HEAD" {
			return action{}, keep("KEEP_IDENTITY", "only an exact selected-remote tracking ref can be pruned")
		}
		live, e := s.liveHeads(ctx, r.Repository, o.Remote)
		if e != nil {
			return action{}, e
		}
		if _, exists := live["refs/heads/"+strings.TrimPrefix(r.Ref, prefix)]; exists {
			return action{}, keep("KEEP_CURRENT", "remote branch still exists")
		}
		out, e := s.git(ctx, r.Repository, "for-each-ref", "--format=%(refname) %(objectname)", "refs/heads/", "refs/tags/")
		if e != nil {
			return action{}, e
		}
		recovery = string(out)
		for _, line := range strings.Split(recovery, "\n") {
			fields := strings.Fields(line)
			if len(fields) == 2 {
				keepers = append(keepers, fields[1])
			}
		}
	}
	objects, err := reflogObjects(filepath.Join(common, "logs", filepath.FromSlash(r.Ref)))
	if err != nil {
		return action{}, err
	}
	objects = append(objects, head)
	if err = s.ancestors(ctx, r.Repository, objects, keepers); err != nil {
		return action{}, err
	}
	// Retain configuration in place; CAS deletion never discards branch-specific config.
	logPath := filepath.Join(common, "logs", filepath.FromSlash(r.Ref))
	logState, err := controlStates(ctx, common, r.Ref, "logs/"+r.Ref, "packed-refs")
	if err != nil {
		return action{}, err
	}
	log, err := os.ReadFile(logPath)
	if err != nil && !os.IsNotExist(err) {
		return action{}, err
	}
	return action{fingerprint: fingerprint(head, recovery, string(log), logState), state: "PRUNED", apply: func() error {
		_, e := s.git(ctx, r.Repository, "update-ref", "-d", r.Ref, head)
		return e
	}}, nil
}

func (s *Service) generated(ctx context.Context, r Receipt) (action, error) {
	if err := s.generatedScope(ctx, r); err != nil {
		return action{}, err
	}

	// Never treat a nested checkout or a cache tool's unbounded selection as scratch.
	err := filepath.WalkDir(r.Path, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == ".git" {
			return keep("KEEP_LOCAL_STATE", "generated target contains Git state")
		}
		return nil
	})
	if err != nil {
		return action{}, err
	}
	tree, err := snapshotTree(r.Path, ctx)
	if err != nil {
		return action{}, err
	}
	if r.GeneratedDigest == "" || tree.Digest != r.GeneratedDigest {
		return action{}, keep("KEEP_LOCAL_STATE", "task-generated output changed since its provenance receipt")
	}

	return action{fingerprint: tree.Digest, bytes: tree.Bytes, state: "REMOVED", apply: func() error {
		return os.RemoveAll(r.Path)
	}}, nil
}

func (s *Service) primary(ctx context.Context, r Receipt, o Options) (action, error) {
	entries, err := s.worktrees(ctx, r.Repository)
	if err != nil {
		return action{}, err
	}
	if entries[0].Path != r.Path {
		return action{}, keep("KEEP_IDENTITY", "sync target is not the primary checkout")
	}
	base, remoteHead, err := s.remoteDefault(ctx, r.Repository, o.Remote)
	if err != nil {
		return action{}, err
	}
	if r.Ref != "" && r.Ref != base {
		return action{}, keep("KEEP_IDENTITY", "requested sync ref is not the live remote default")
	}
	head, err := s.value(ctx, r.Repository, "rev-parse", "--verify", base)
	if err != nil {
		return action{}, err
	}
	// Fetch objects only for the live default; no prune, reset, stash or arbitrary-base update.
	if o.Apply {
		if _, err = s.git(ctx, r.Repository, "fetch", "--quiet", "--no-prune", "--no-prune-tags", "--recurse-submodules=no", "--no-auto-maintenance", "--refmap=", "--no-tags", "--no-write-fetch-head", o.Remote, remoteHead); err != nil {
			return action{}, err
		}
	}
	if _, err = s.git(ctx, r.Repository, "cat-file", "-e", remoteHead+"^{commit}"); err != nil {
		return action{}, keep("KEEP_NEEDS_FETCH", "live default commit is not available locally; apply acquires objects before proving ancestry")
	}
	if err = s.ancestors(ctx, r.Repository, []string{head}, []string{remoteHead}); err != nil {
		return action{}, keep("KEEP_DIVERGED", "local default has commits outside the live remote default")
	}
	current := entries[0].Branch
	for _, entry := range entries[1:] {
		if entry.Branch == base {
			return action{}, keep("KEEP_IN_USE", "default branch is checked out elsewhere")
		}
	}
	if current == "" {
		return action{}, keep("KEEP_DETACHED", "primary checkout is detached")
	}
	metadata, err := s.metadata(ctx, r, head, base)
	if err != nil {
		return action{}, err
	}
	var tree treeSnapshot
	if current == base {
		if err = s.checkoutFilters(ctx, r.Repository, remoteHead); err != nil {
			return action{}, err
		}
		if err = s.cleanIndex(ctx, r.Path); err != nil {
			return action{}, err
		}
		tree, err = verifyTrackedTree(ctx, s.runGitCommand, r.Path, head)
		if err != nil {
			return action{}, keep("KEEP_LOCAL_STATE", "primary working tree is not fully recoverable")
		}
	}
	state := "UPDATED"
	if head == remoteHead {
		state = "CURRENT"
	}
	return action{fingerprint: fingerprint(base, head, remoteHead, current, metadata, tree.Digest), state: state, apply: func() error {
		if head == remoteHead {
			return nil
		}
		if current == base {
			_, e := s.git(ctx, r.Path, "merge", "--quiet", "--ff-only", "--no-overwrite-ignore", remoteHead)
			return e
		}
		_, e := s.git(ctx, r.Repository, "update-ref", base, remoteHead, head)
		return e
	}}, nil
}

// generatedScope permits bounded task output outside checkouts or inside its linked lane,
// never primary, Git metadata, tracked source, or a directory enclosing any checkout.
func (s *Service) generatedScope(ctx context.Context, r Receipt) error {
	entries, err := s.worktrees(ctx, r.Repository)
	if err != nil {
		return err
	}
	for i, entry := range entries {
		if contains(r.Path, entry.Path) {
			return keep("KEEP_IDENTITY", "generated output encloses a checkout")
		}
		if contains(entry.Path, r.Path) {
			if i == 0 {
				return keep("KEEP_IDENTITY", "primary checkout is not disposable scratch")
			}
			rel, e := filepath.Rel(entry.Path, r.Path)
			if e != nil {
				return e
			}
			for _, part := range strings.Split(rel, string(filepath.Separator)) {
				if strings.EqualFold(part, ".git") {
					return keep("KEEP_IDENTITY", "Git metadata is not scratch")
				}
			}
			tracked, e := s.git(ctx, entry.Path, "ls-files", "-z", "--", filepath.ToSlash(rel))
			if e != nil {
				return e
			}
			if len(tracked) != 0 {
				return keep("KEEP_LOCAL_STATE", "generated output overlaps tracked source")
			}
		}
	}
	common, err := s.value(ctx, r.Repository, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return err
	}
	if contains(common, r.Path) || contains(r.Path, common) {
		return keep("KEEP_IDENTITY", "generated output overlaps Git metadata")
	}
	return nil
}

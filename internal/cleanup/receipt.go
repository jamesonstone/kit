package cleanup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func canonical(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	if filepath.Clean(absolute) != resolved {
		return "", fmt.Errorf("target must use its physical path")
	}
	return resolved, nil
}

func identity(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("target is not a physical directory")
	}
	return treeIdentity(info)
}

// Record must be called when the task creates its artifact (or first observes primary sync scope).
// The caller retains the receipt outside the target. Existing shared lanes cannot be adopted by name.
func (s *Service) Record(ctx context.Context, kind, path, repository, owner, ref string) (Receipt, error) {
	receipt := Receipt{Version: 1, Kind: kind, Owner: owner, Ref: ref}
	if owner == "" {
		return receipt, fmt.Errorf("current task owner is required")
	}
	var err error
	receipt.Path, err = canonical(path)
	if err != nil {
		return receipt, err
	}
	receipt.Repository, err = canonical(repository)
	if err != nil {
		return receipt, err
	}
	top, topErr := s.value(ctx, receipt.Repository, "rev-parse", "--show-toplevel")
	if topErr != nil {
		return receipt, topErr
	}
	if top != receipt.Repository {
		return receipt, fmt.Errorf("repository must be its exact checkout root")
	}
	receipt.Identity, err = identity(receipt.Path)
	if err != nil {
		return receipt, err
	}
	if (kind == "branch" || kind == "ref") && receipt.Path != receipt.Repository {
		return receipt, fmt.Errorf("reference scope must equal repository root")
	}
	if kind == "branch" || kind == "ref" {
		prefix := "refs/heads/"
		if kind == "ref" {
			prefix = "refs/remotes/"
		}
		if !strings.HasPrefix(ref, prefix) {
			return receipt, fmt.Errorf("exact canonical reference required")
		}
		if _, err = s.git(ctx, receipt.Repository, "check-ref-format", ref); err != nil {
			return receipt, err
		}
		receipt.RefOID, err = s.value(ctx, receipt.Repository, "rev-parse", "--verify", ref)
		if err != nil {
			return receipt, err
		}
	}
	switch kind {
	case "worktree", "primary", "branch", "ref":
		receipt.GitDir, err = s.value(ctx, receipt.Repository, "rev-parse", "--absolute-git-dir")
		if kind == "worktree" || kind == "primary" {
			receipt.GitDir, err = s.value(ctx, receipt.Path, "rev-parse", "--absolute-git-dir")
		}
		if err != nil {
			return receipt, err
		}
		receipt.GitIdentity, err = identity(receipt.GitDir)
		if err == nil && kind == "worktree" {
			entries, e := s.worktrees(ctx, receipt.Repository)
			if e != nil {
				return receipt, e
			}
			receipt.Links = map[string]string{}
			for _, name := range []string{".env", ".envrc"} {
				target, e := os.Readlink(filepath.Join(receipt.Path, name))
				if e == nil && target == filepath.Join(entries[0].Path, name) {
					receipt.Links[name] = target
				}
			}
		}
	case "scratch", "cache":
		if err = s.generatedScope(ctx, receipt); err != nil {
			return receipt, err
		}

		var snapshot treeSnapshot
		snapshot, err = snapshotTree(receipt.Path, ctx)
		receipt.GeneratedDigest = snapshot.Digest
	default:
		err = fmt.Errorf("unsupported cleanup category")
	}
	return receipt, err
}

func contains(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

package cleanup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type treeSnapshot struct {
	Digest string
	Bytes  int64
}

type treeEntry struct {
	info                                fs.FileInfo
	data                                []byte // Only symlink targets and bounded root .git control files.
	contentDigest, blobSHA1, blobSHA256 string
	metadata                            string
}

// readTree never follows symlinks. Metadata and identity are sampled around each
// read, so a changed entry invalidates this observation. Callers must separately
// exclude concurrent writers for the entire proof/action interval.
func readTree(root string, skipGitDirectory bool) (map[string]treeEntry, treeSnapshot, error) {
	entries := make(map[string]treeEntry)
	digest := sha256.New()
	var snapshot treeSnapshot
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		before, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if path == root && !before.IsDir() {
			return fmt.Errorf("tree root is not a physical directory")
		}
		metadata, err := treeMetadata(path, before)
		if err != nil {
			return fmt.Errorf("metadata %q: %w", path, err)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		entry := treeEntry{info: before, metadata: metadata}
		switch {
		case before.Mode().IsRegular():
			entry, err = streamPhysicalFile(path, before, rel == ".git")
			entry.metadata = metadata
		case before.Mode()&os.ModeSymlink != 0:
			var target string
			target, err = os.Readlink(path)
			entry.data = []byte(target)
			entry.contentDigest, entry.blobSHA1, entry.blobSHA256 = treeByteDigests(entry.data)
		case before.IsDir():
		default:
			return fmt.Errorf("special file %q", path)
		}
		if err != nil {
			return err
		}
		after, err := os.Lstat(path)
		if err != nil {
			return err
		}
		afterMetadata, err := treeMetadata(path, after)
		if err != nil {
			return err
		}
		if !os.SameFile(before, after) || metadata != afterMetadata {
			return fmt.Errorf("entry changed during inspection: %q", path)
		}
		entries[rel] = entry
		_, _ = fmt.Fprintf(digest, "%d:%s%d:%s%d:%s", len(rel), rel, len(metadata), metadata, len(entry.contentDigest), entry.contentDigest)
		if before.Mode().IsRegular() {
			snapshot.Bytes += before.Size()
		}
		if skipGitDirectory && rel == ".git" && before.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, treeSnapshot{}, err
	}
	// Revalidate directories after enumeration as well as files: a new or removed
	// child after its parent callback must invalidate the observation.
	for name, entry := range entries {
		path := filepath.Join(root, filepath.FromSlash(name))
		current, err := os.Lstat(path)
		if err != nil {
			return nil, treeSnapshot{}, err
		}
		metadata, err := treeMetadata(path, current)
		if err != nil {
			return nil, treeSnapshot{}, err
		}
		if !os.SameFile(entry.info, current) || metadata != entry.metadata {
			return nil, treeSnapshot{}, fmt.Errorf("tree changed during enumeration: %q", name)
		}
	}
	snapshot.Digest = hex.EncodeToString(digest.Sum(nil))
	return entries, snapshot, nil
}

func snapshotTree(root string) (treeSnapshot, error) {
	_, snapshot, err := readTree(root, false)
	return snapshot, err
}

// verifyTrackedTree compares physical bytes with committed blobs, independently
// of the index (including assume-unchanged, skip-worktree, filters and stat cache).
func verifyTrackedTree(ctx context.Context, run commandFunc, root, head string, allowedLinks ...map[string]string) (treeSnapshot, error) {
	output, err := run(ctx, root, "git", "ls-tree", "-rz", "--full-tree", head)
	if err != nil {
		return treeSnapshot{}, err
	}
	expected := map[string]string{".": "040000"}
	ids := make(map[string]string)
	for _, record := range bytes.Split(output, []byte{0}) {
		if len(record) == 0 {
			continue
		}
		fields := bytes.SplitN(record, []byte{'\t'}, 2)
		if len(fields) != 2 {
			return treeSnapshot{}, fmt.Errorf("malformed Git tree record")
		}
		header := strings.Fields(string(fields[0]))
		if len(header) != 3 || header[1] != "blob" {
			return treeSnapshot{}, fmt.Errorf("unsupported Git tree entry")
		}
		mode, name := header[0], string(fields[1])
		if mode != "100644" && mode != "100755" && mode != "120000" {
			return treeSnapshot{}, fmt.Errorf("unsupported Git mode %s", mode)
		}
		if !fs.ValidPath(name) || name == "." || strings.Contains(name, "\\") {
			return treeSnapshot{}, fmt.Errorf("noncanonical Git path")
		}
		for _, part := range strings.Split(name, "/") {
			if strings.EqualFold(part, ".git") {
				return treeSnapshot{}, fmt.Errorf("nested Git control path")
			}
		}
		if _, exists := ids[name]; exists {
			return treeSnapshot{}, fmt.Errorf("duplicate Git path")
		}
		expected[name], ids[name] = mode, header[2]
		for parent := filepath.ToSlash(filepath.Dir(name)); parent != "."; parent = filepath.ToSlash(filepath.Dir(parent)) {
			expected[parent] = "040000"
		}
	}
	entries, snapshot, err := readTree(root, true)
	if err != nil {
		return treeSnapshot{}, err
	}
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entry := entries[name]
		if name == ".git" {
			if !entry.info.IsDir() && (!entry.info.Mode().IsRegular() || !bytes.HasPrefix(entry.data, []byte("gitdir: "))) {
				return treeSnapshot{}, fmt.Errorf("git control entry must be a directory or worktree file")
			}
			continue
		}
		if len(allowedLinks) == 1 {
			if target, allowed := allowedLinks[0][name]; allowed {
				if entry.info.Mode() != os.ModeSymlink|trackedSymlinkPermissions || string(entry.data) != target {
					return treeSnapshot{}, fmt.Errorf("registered environment link changed")
				}
				continue
			}
		}
		mode, ok := expected[name]
		if !ok {
			return treeSnapshot{}, fmt.Errorf("extra local entry %q", name)
		}
		switch mode {
		case "040000":
			if !entry.info.IsDir() || (name != "." && entry.info.Mode() != 0755|os.ModeDir) {
				return treeSnapshot{}, fmt.Errorf("noncanonical directory %q", name)
			}
		case "120000":
			if entry.info.Mode() != os.ModeSymlink|trackedSymlinkPermissions {
				return treeSnapshot{}, fmt.Errorf("symlink changed %q", name)
			}
		default:
			perm := os.FileMode(0644)
			if mode == "100755" {
				perm = 0755
			}
			if entry.info.Mode() != perm {
				return treeSnapshot{}, fmt.Errorf("noncanonical file mode %q", name)
			}
		}
		if id, ok := ids[name]; ok {
			actual := entry.blobSHA1
			switch len(id) {
			case 40:
			case 64:
				actual = entry.blobSHA256
			default:
				return treeSnapshot{}, fmt.Errorf("unsupported Git object format")
			}
			if actual != id {
				return treeSnapshot{}, fmt.Errorf("physical content differs from commit: %q", name)
			}
		}
		delete(expected, name)
	}
	if len(expected) != 0 {
		return treeSnapshot{}, fmt.Errorf("committed entries missing from physical tree")
	}
	if _, ok := entries[".git"]; !ok {
		return treeSnapshot{}, fmt.Errorf("missing Git control entry")
	}
	return snapshot, nil
}

//go:build darwin

package cleanup

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

const trackedSymlinkPermissions fs.FileMode = 0755

func treeMetadata(path string, info fs.FileInfo) (string, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("missing inode metadata")
	}
	if stat.Uid != uint32(os.Getuid()) || stat.Gid != uint32(os.Getegid()) {
		return "", fmt.Errorf("foreign filesystem ownership requires preservation")
	}
	if info.Mode().IsRegular() && stat.Nlink != 1 {
		return "", fmt.Errorf("hardlinked file")
	}
	if stat.Flags != 0 {
		return "", fmt.Errorf("file flags require preservation")
	}
	provenance, err := darwinProvenance(path)
	if err != nil {
		return "", err
	}
	// Darwin ACLs are separate from listxattr. The system ls reports extended ACLs
	// with '+' in the mode suffix; -d and default physical handling avoid following
	// symlinks. Only the fixed mode prefix is parsed, never a filename or ACL text.
	output, err := exec.Command("/bin/ls", "-lde", path).Output()
	if err != nil {
		return "", fmt.Errorf("ACL inspection failed: %w", err)
	}
	line := strings.SplitN(string(output), "\n", 2)[0]
	if bytes.Count(output, []byte{'\n'}) != 1 || len(line) < 11 || (line[10] != ' ' && line[10] != '\t' && line[10] != '@') {
		return "", fmt.Errorf("ACL or unknown permission metadata")
	}
	return fmt.Sprintf("%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d", stat.Dev, stat.Ino, stat.Mode, stat.Nlink, stat.Uid, stat.Gid, stat.Size, stat.Mtimespec.Sec, stat.Mtimespec.Nsec, stat.Ctimespec.Sec, stat.Ctimespec.Nsec, stat.Flags) + ":" + provenance, nil
}

func openPhysicalFile(path string, before fs.FileInfo) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	current, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if !current.Mode().IsRegular() || !os.SameFile(before, current) {
		_ = file.Close()
		return nil, fmt.Errorf("file identity changed before read")
	}
	return file, nil
}

// treeIdentity identifies a root without treating mutable recovery metadata as
// ownership evidence. Eligibility still requires the full metadata inspection.
func treeIdentity(info fs.FileInfo) (string, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("missing inode identity")
	}
	return fmt.Sprintf("%d:%d", stat.Dev, stat.Ino), nil
}

// macOS adds com.apple.provenance even to fresh files and may refuse its
// removal. It describes OS provenance, not task-authored work. This sole
// exception is bound into the observation; every other xattr must be retained.
func darwinProvenance(path string) (string, error) {
	size, err := unix.Llistxattr(path, nil)
	if err != nil {
		return "", err
	}
	names := make([]byte, size)
	count, err := unix.Llistxattr(path, names)
	if err != nil || count != size {
		return "", fmt.Errorf("xattr inspection changed or failed")
	}
	provenance := ""
	for _, name := range bytes.Split(names, []byte{0}) {
		if len(name) == 0 {
			continue
		}
		if string(name) != "com.apple.provenance" {
			return "", fmt.Errorf("extended attributes require preservation")
		}
		size, err := unix.Lgetxattr(path, string(name), nil)
		if err != nil {
			return "", err
		}
		value := make([]byte, size)
		count, err := unix.Lgetxattr(path, string(name), value)
		if err != nil || count != size {
			return "", fmt.Errorf("provenance changed or inspection failed")
		}
		provenance = fmt.Sprintf("%x", sha256.Sum256(value))
	}
	return provenance, nil
}

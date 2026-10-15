//go:build linux

package cleanup

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

const trackedSymlinkPermissions fs.FileMode = 0777

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
	// Linux POSIX access/default ACLs are xattrs, so rejecting all xattrs also
	// rejects ACLs, capabilities, security labels and other non-Git metadata.
	size, err := unix.Llistxattr(path, nil)
	if err != nil {
		return "", err
	}
	if size != 0 {
		return "", fmt.Errorf("extended attributes require preservation")
	}
	if info.Mode().IsRegular() || info.IsDir() {
		fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
		if err != nil {
			return "", err
		}
		flags, inspectErr := unix.IoctlGetInt(fd, unix.FS_IOC_GETFLAGS)
		_ = unix.Close(fd)
		if inspectErr != nil {
			return "", fmt.Errorf("file flag inspection unavailable: %w", inspectErr)
		}
		// EXTENTS and INDEX are storage implementation details, not user metadata.
		if flags & ^(0x00080000|0x00001000) != 0 {
			return "", fmt.Errorf("file flags require preservation")
		}
	}
	return fmt.Sprintf("%d:%d:%d:%d:%d:%d:%d:%d:%d:%d:%d", stat.Dev, stat.Ino, stat.Mode, stat.Nlink, stat.Uid, stat.Gid, stat.Size, stat.Mtim.Sec, stat.Mtim.Nsec, stat.Ctim.Sec, stat.Ctim.Nsec), nil
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

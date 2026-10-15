//go:build !darwin && !linux

package cleanup

import (
	"context"
	"fmt"
	"io/fs"
	"os"
)

const trackedSymlinkPermissions fs.FileMode = 0777

func treeMetadata(_ string, _ fs.FileInfo, _ ...context.Context) (string, error) {
	return "", fmt.Errorf("filesystem recovery metadata inspection unsupported on this platform")
}

func openPhysicalFile(_ string, _ fs.FileInfo) (*os.File, error) {
	return nil, fmt.Errorf("physical file read unsupported on this platform")
}

func treeIdentity(_ fs.FileInfo) (string, error) {
	return "", fmt.Errorf("filesystem identity unsupported on this platform")
}

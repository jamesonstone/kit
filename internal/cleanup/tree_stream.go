package cleanup

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
)

const maxGitControlBytes = 64 << 10

// streamPhysicalFile retains fixed-size hashes, not file contents. Generated
// caches and unique ignored databases may be larger than available memory.
func streamPhysicalFile(path string, before fs.FileInfo, control bool) (treeEntry, error) {
	entry := treeEntry{info: before}
	if control && before.Size() > maxGitControlBytes {
		return entry, fmt.Errorf("oversized Git control file")
	}
	file, err := openPhysicalFile(path, before)
	if err != nil {
		return entry, err
	}
	defer func() { _ = file.Close() }()
	content, blob1, blob256 := sha256.New(), sha1.New(), sha256.New()
	_, _ = fmt.Fprintf(blob1, "blob %d%c", before.Size(), byte(0))
	_, _ = fmt.Fprintf(blob256, "blob %d%c", before.Size(), byte(0))
	var controlBytes bytes.Buffer
	writers := []io.Writer{content, blob1, blob256}
	if control {
		writers = append(writers, &controlBytes)
	}
	// LimitReader also bounds a growing control file even if its initial stat was
	// small; one extra byte distinguishes a changed file from a complete read.
	var reader io.Reader = file
	if control {
		reader = io.LimitReader(file, maxGitControlBytes+1)
	}
	count, err := io.CopyBuffer(io.MultiWriter(writers...), reader, make([]byte, 32<<10))
	if err != nil {
		return entry, err
	}
	if count != before.Size() {
		return entry, fmt.Errorf("file size changed during streaming read")
	}
	entry.contentDigest = hex.EncodeToString(content.Sum(nil))
	entry.blobSHA1 = hex.EncodeToString(blob1.Sum(nil))
	entry.blobSHA256 = hex.EncodeToString(blob256.Sum(nil))
	if control {
		entry.data = controlBytes.Bytes()
	}
	return entry, nil
}

func treeByteDigests(data []byte) (content, blob1, blob256 string) {
	raw := sha256.Sum256(data)
	one, two := sha1.New(), sha256.New()
	_, _ = fmt.Fprintf(one, "blob %d%c", len(data), byte(0))
	_, _ = one.Write(data)
	_, _ = fmt.Fprintf(two, "blob %d%c", len(data), byte(0))
	_, _ = two.Write(data)
	return hex.EncodeToString(raw[:]), hex.EncodeToString(one.Sum(nil)), hex.EncodeToString(two.Sum(nil))
}

// physicalContentDigest binds streamed content to the same physical entry and
// recovery metadata before and after the read; it retains no file bytes.
func physicalContentDigest(path string, info fs.FileInfo) (string, error) {
	before, err := treeMetadata(path, info)
	if err != nil {
		return "", err
	}
	entry, err := streamPhysicalFile(path, info, false)
	if err != nil {
		return "", err
	}
	after, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	metadata, err := treeMetadata(path, after)
	if err != nil {
		return "", err
	}
	if !os.SameFile(info, after) || before != metadata {
		return "", fmt.Errorf("file changed during content proof")
	}
	return entry.contentDigest, nil
}

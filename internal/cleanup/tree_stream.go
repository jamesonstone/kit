package cleanup

import (
	"bytes"
	"context"
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
func streamPhysicalFile(path string, before fs.FileInfo, control bool, contexts ...context.Context) (treeEntry, error) {
	ctx, cancel := treeInspectionContext(contexts)
	defer cancel()
	entry := treeEntry{info: before}
	if err := ctx.Err(); err != nil {
		return entry, err
	}
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
	var reader io.Reader = treeContextReader{ctx: ctx, reader: file}
	if control {
		reader = io.LimitReader(reader, maxGitControlBytes+1)
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
func physicalContentDigest(path string, info fs.FileInfo, contexts ...context.Context) (string, error) {
	ctx, cancel := treeInspectionContext(contexts)
	defer cancel()
	before, err := treeMetadata(path, info, ctx)
	if err != nil {
		return "", err
	}
	entry, err := streamPhysicalFile(path, info, false, ctx)
	if err != nil {
		return "", err
	}
	after, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	metadata, err := treeMetadata(path, after, ctx)
	if err != nil {
		return "", err
	}
	if !os.SameFile(info, after) || before != metadata {
		return "", fmt.Errorf("file changed during content proof")
	}
	return entry.contentDigest, nil
}

// A shared deadline bounds the entire traversal rather than restarting the
// budget for each file. Optional contexts preserve existing internal callers.
func treeInspectionContext(contexts []context.Context) (context.Context, context.CancelFunc) {
	parent := context.Background()
	if len(contexts) == 1 && contexts[0] != nil {
		parent = contexts[0]
	}
	return context.WithTimeout(parent, DefaultTimeout)
}

type treeContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r treeContextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(buffer)
}

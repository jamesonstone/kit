package cleanup

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTrackedTreeRejectsIgnoredSubtreeBeforeInspection(t *testing.T) {
	_, lane := treeFixture(t, "sha1")
	ignored := filepath.Join(lane, "ignored")
	if err := os.Mkdir(ignored, 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(filepath.Join(ignored, "large-local-database"))
	if err != nil {
		t.Fatal(err)
	}
	// Sparse size is intentionally much larger than the proof's time budget. The
	// ignored directory must be rejected before any descendant read or metadata.
	if err := file.Truncate(64 << 30); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(ignored, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ignored, 0755) })
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = verifyTrackedTree(ctx, treeTestRun, lane, "HEAD")
	if err == nil || !strings.Contains(err.Error(), `extra local entry "ignored"`) {
		t.Fatalf("wanted early unknown-directory rejection, got %v", err)
	}
}

func TestTreeInspectionHonorsCanceledContext(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := snapshotTree(root, ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("snapshot cancellation: %v", err)
	}
	called := false
	run := func(context.Context, string, string, ...string) ([]byte, error) { called = true; return nil, nil }
	if _, err := verifyTrackedTree(ctx, run, root, "HEAD"); !errors.Is(err, context.Canceled) {
		t.Fatalf("tracked proof cancellation: %v", err)
	}
	if called {
		t.Fatal("started Git after cancellation")
	}
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return
	}
	info, err := os.Lstat(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := treeMetadata(root, info, ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("metadata cancellation: %v", err)
	}
}

func TestStreamingContentHonorsCancellationBetweenReads(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader := treeContextReader{ctx: ctx, reader: bytes.NewBufferString("physical bytes")}
	buffer := make([]byte, 1)
	if _, err := reader.Read(buffer); err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := reader.Read(buffer); !errors.Is(err, context.Canceled) {
		t.Fatalf("stream cancellation: %v", err)
	}
}

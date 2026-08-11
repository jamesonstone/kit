package usage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRefreshPrunesExpiredEventsAndPreservesCurrentEvents(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir, _ := Directory()
	if err := ensurePrivateDirectory(dir); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	path := filepath.Join(dir, now.Format("2006-01")+"-0001.jsonl")
	events := []Event{
		{SchemaVersion: SchemaVersion, Timestamp: now.Add(-DefaultRetention - time.Hour), Command: "status"},
		{SchemaVersion: SchemaVersion, Timestamp: now.Add(-time.Hour), Command: "health"},
	}
	if err := writeEvents(path, events); err != nil {
		t.Fatal(err)
	}

	preview, err := Refresh(true)
	if err != nil {
		t.Fatalf("Refresh(dry-run) error = %v", err)
	}
	if !preview.Changed || preview.Status.PrunedEvents != 1 {
		t.Fatalf("unexpected preview: %#v", preview)
	}
	var before []Event
	if err := readEvents(path, func(event Event) error { before = append(before, event); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(before) != 2 {
		t.Fatalf("dry-run mutated events: %d", len(before))
	}

	result, err := Refresh(false)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if !result.Changed || result.Status.PrunedEvents != 1 {
		t.Fatalf("unexpected refresh: %#v", result)
	}
	var after []Event
	if err := readEvents(path, func(event Event) error { after = append(after, event); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(after) != 1 || after[0].Command != "health" {
		t.Fatalf("unexpected retained events: %#v", after)
	}
}

func TestPruneTotalBytesRemovesOldestCompleteShards(t *testing.T) {
	dir := t.TempDir()
	prefix := `{"schema_version":"kit.usage/v1","timestamp":"2099-01-01T00:00:00Z","command":"`
	suffix := `","version":"v2.0.0","exit_code":0,"success":true,"elapsed_ms":1,"interactive":false}` + "\n"
	payload := []byte(prefix + strings.Repeat("x", int(MaxShardBytes)-len(prefix)-len(suffix)) + suffix)
	for index := 1; index <= 9; index++ {
		path := filepath.Join(dir, "2026-01-"+fourDigits(index)+".jsonl")
		if err := os.WriteFile(path, payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	status := StorageStatus{}
	if err := pruneTotalBytes(dir, false, &status); err != nil {
		t.Fatal(err)
	}
	shards, err := listShards(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(shards) != 8 || status.PrunedShards != 1 {
		t.Fatalf("shards=%d pruned=%d", len(shards), status.PrunedShards)
	}
	if shards[0].name != "2026-01-0002.jsonl" {
		t.Fatalf("oldest shard was not pruned: %s", shards[0].name)
	}
}

func TestRecordMaintainsTotalStorageBoundAfterAppend(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir, _ := Directory()
	if err := ensurePrivateDirectory(dir); err != nil {
		t.Fatal(err)
	}
	prefix := `{"schema_version":"kit.usage/v1","timestamp":"2099-01-01T00:00:00Z","command":"`
	suffix := `","version":"v2.0.0","exit_code":0,"success":true,"elapsed_ms":1,"interactive":false}` + "\n"
	payload := []byte(prefix + strings.Repeat("x", int(MaxShardBytes)-len(prefix)-len(suffix)) + suffix)
	for index := 1; index <= 8; index++ {
		path := filepath.Join(dir, "2025-01-"+fourDigits(index)+".jsonl")
		if err := os.WriteFile(path, payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := Record(RecordInput{Command: "status", Version: "v2.0.0"}); err != nil {
		t.Fatal(err)
	}
	shards, err := listShards(dir)
	if err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, shard := range shards {
		total += shard.size
	}
	if total > MaxTotalBytes {
		t.Fatalf("total bytes = %d, want at most %d", total, MaxTotalBytes)
	}
}

func TestRefreshDryRunCreatesNoStorage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir, _ := Directory()
	if _, err := Refresh(true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("dry-run created usage directory: %v", err)
	}
}

func fourDigits(value int) string {
	return fmt.Sprintf("%04d", value)
}

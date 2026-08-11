package usage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maintenanceMarker = ".maintenance"

func automaticMaintenance(dir string, now time.Time) error {
	marker := filepath.Join(dir, maintenanceMarker)
	if data, err := os.ReadFile(marker); err == nil && strings.TrimSpace(string(data)) == now.UTC().Format("2006-01-02") {
		return pruneTotalBytes(dir, false, nil)
	}
	if err := pruneStore(dir, now, false, nil); err != nil {
		return err
	}
	return os.WriteFile(marker, []byte(now.UTC().Format("2006-01-02")+"\n"), 0o600)
}

func Refresh(dryRun bool) (RefreshResult, error) {
	dir, err := Directory()
	if err != nil {
		return RefreshResult{}, err
	}
	result := RefreshResult{DryRun: dryRun}
	before, _ := Status("")
	if dryRun {
		result.Status = before
		if err := pruneStore(dir, time.Now().UTC(), true, &result.Status); err != nil {
			return result, err
		}
		result.Changed = result.Status.PrunedEvents > 0 || result.Status.PrunedShards > 0
		return result, nil
	}
	err = withStoreLock(dir, func() error {
		if err := pruneStore(dir, time.Now().UTC(), false, &result.Status); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, maintenanceMarker), []byte(time.Now().UTC().Format("2006-01-02")+"\n"), 0o600)
	})
	if err != nil {
		return result, err
	}
	result.Changed = result.Status.PrunedEvents > 0 || result.Status.PrunedShards > 0
	status, statusErr := Status("")
	status.PrunedEvents = result.Status.PrunedEvents
	status.PrunedShards = result.Status.PrunedShards
	result.Status = status
	return result, statusErr
}

func pruneStore(dir string, now time.Time, dryRun bool, status *StorageStatus) error {
	shards, err := listShards(dir)
	if err != nil {
		return err
	}
	cutoff := now.Add(-DefaultRetention)
	for _, item := range shards {
		var kept []Event
		expired := 0
		if err := readEvents(item.path, func(event Event) error {
			if event.Timestamp.Before(cutoff) {
				expired++
				return nil
			}
			kept = append(kept, event)
			return nil
		}); err != nil {
			return err
		}
		if expired == 0 {
			continue
		}
		if status != nil {
			status.PrunedEvents += expired
			if len(kept) == 0 {
				status.PrunedShards++
			}
		}
		if dryRun {
			continue
		}
		if len(kept) == 0 {
			if err := os.Remove(item.path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			continue
		}
		if err := writeEvents(item.path, kept); err != nil {
			return err
		}
	}
	return pruneTotalBytes(dir, dryRun, status)
}

func pruneTotalBytes(dir string, dryRun bool, status *StorageStatus) error {
	shards, err := listShards(dir)
	if err != nil {
		return err
	}
	var total int64
	for _, item := range shards {
		total += item.size
	}
	for len(shards) > 1 && total > MaxTotalBytes {
		oldest := shards[0]
		if !dryRun {
			if err := os.Remove(oldest.path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		total -= oldest.size
		shards = shards[1:]
		if status != nil {
			status.PrunedShards++
		}
	}
	return nil
}

func writeEvents(path string, events []Event) (resultErr error) {
	temp, err := os.CreateTemp(filepath.Dir(path), ".usage-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	closed := false
	defer func() {
		if !closed {
			if closeErr := temp.Close(); closeErr != nil && resultErr == nil {
				resultErr = closeErr
			}
		}
		_ = os.Remove(tempPath)
	}()
	if err := temp.Chmod(0o600); err != nil {
		return err
	}
	writer := bufio.NewWriter(temp)
	encoder := json.NewEncoder(writer)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	closed = true
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", filepath.Base(path), err)
	}
	return nil
}

func Clear(filter Filter, all bool) (int, error) {
	if !all && filter.Command == "" && filter.ProjectID == "" {
		return 0, fmt.Errorf("clear requires a command or project filter unless --all is used")
	}
	dir, err := Directory()
	if err != nil {
		return 0, err
	}
	removed := 0
	err = withStoreLock(dir, func() error {
		shards, err := listShards(dir)
		if err != nil {
			return err
		}
		for _, item := range shards {
			var kept []Event
			if err := readEvents(item.path, func(event Event) error {
				if all || matchesFilter(event, filter) {
					removed++
					return nil
				}
				kept = append(kept, event)
				return nil
			}); err != nil {
				return err
			}
			if len(kept) == 0 {
				if err := os.Remove(item.path); err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
			} else if err := writeEvents(item.path, kept); err != nil {
				return err
			}
		}
		return nil
	})
	return removed, err
}

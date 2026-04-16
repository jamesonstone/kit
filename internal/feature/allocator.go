package feature

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	allocatorDirName       = "kit"
	allocatorStateFileName = "feature-sequence.json"
	allocatorLockFileName  = "feature-sequence.lock"
	allocatorLockTimeout   = 5 * time.Second
	allocatorStaleLockAge  = 2 * time.Minute
)

var gitCommonDirResolver = resolveGitCommonDir

type featureSequenceState struct {
	LastReserved int       `json:"last_reserved"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

func reserveNextFeatureNumber(projectRoot string, localMax int) (int, error) {
	commonDir, err := gitCommonDirResolver(projectRoot)
	if err != nil || commonDir == "" {
		return localMax + 1, nil
	}

	allocatorDir := filepath.Join(commonDir, allocatorDirName)
	if err := os.MkdirAll(allocatorDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to prepare shared feature allocator: %w", err)
	}

	release, err := acquireAllocatorLock(filepath.Join(allocatorDir, allocatorLockFileName))
	if err != nil {
		return 0, err
	}
	defer release()

	statePath := filepath.Join(allocatorDir, allocatorStateFileName)
	state, err := readAllocatorState(statePath)
	if err != nil {
		return 0, err
	}

	next := max(localMax, state.LastReserved) + 1
	state.LastReserved = next
	state.UpdatedAt = time.Now().UTC()

	if err := writeAllocatorState(statePath, state); err != nil {
		return 0, err
	}

	return next, nil
}

func resolveGitCommonDir(projectRoot string) (string, error) {
	cmd := exec.Command("git", "-C", projectRoot, "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	commonDir := strings.TrimSpace(string(output))
	if commonDir == "" {
		return "", fmt.Errorf("git common dir is empty")
	}
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Clean(filepath.Join(projectRoot, commonDir))
	}

	return commonDir, nil
}

func acquireAllocatorLock(lockPath string) (func(), error) {
	deadline := time.Now().Add(allocatorLockTimeout)

	for {
		file, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err == nil {
			if _, writeErr := fmt.Fprintf(file, "pid=%d\ncreated_at=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano)); writeErr != nil {
				file.Close()
				_ = os.Remove(lockPath)
				return nil, fmt.Errorf("failed to initialize feature allocator lock: %w", writeErr)
			}
			if closeErr := file.Close(); closeErr != nil {
				_ = os.Remove(lockPath)
				return nil, fmt.Errorf("failed to finalize feature allocator lock: %w", closeErr)
			}

			return func() {
				_ = os.Remove(lockPath)
			}, nil
		}

		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("failed to acquire feature allocator lock: %w", err)
		}

		if staleErr := removeStaleAllocatorLock(lockPath); staleErr != nil {
			return nil, staleErr
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for shared feature allocator lock at %s", lockPath)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func removeStaleAllocatorLock(lockPath string) error {
	info, err := os.Stat(lockPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to inspect feature allocator lock: %w", err)
	}

	if time.Since(info.ModTime()) <= allocatorStaleLockAge {
		return nil
	}

	if err := os.Remove(lockPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to clear stale feature allocator lock: %w", err)
	}

	return nil
}

func readAllocatorState(path string) (featureSequenceState, error) {
	var state featureSequenceState

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state, nil
		}
		return state, fmt.Errorf("failed to read shared feature allocator state: %w", err)
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("failed to parse shared feature allocator state: %w", err)
	}

	return state, nil
}

func writeAllocatorState(path string, state featureSequenceState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode shared feature allocator state: %w", err)
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write shared feature allocator state: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to finalize shared feature allocator state: %w", err)
	}

	return nil
}

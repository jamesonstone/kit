package feature

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestNextNumber_UsesSharedAllocatorAcrossWorktrees(t *testing.T) {
	commonDir := t.TempDir()
	projectA := t.TempDir()
	projectB := t.TempDir()

	writeAllocatorFeatureDir(t, projectA, "docs/specs/0001-alpha")
	writeAllocatorFeatureDir(t, projectA, "docs/specs/0002-beta")
	writeAllocatorFeatureDir(t, projectB, "docs/specs/0001-alpha")
	writeAllocatorFeatureDir(t, projectB, "docs/specs/0002-beta")

	cfg := config.Default()
	specsA := cfg.SpecsPath(projectA)
	specsB := cfg.SpecsPath(projectB)

	previousResolver := gitCommonDirResolver
	gitCommonDirResolver = func(string) (string, error) { return commonDir, nil }
	defer func() { gitCommonDirResolver = previousResolver }()

	nextA, err := NextNumber(projectA, specsA)
	if err != nil {
		t.Fatalf("NextNumber(projectA) error = %v", err)
	}
	if nextA != 3 {
		t.Fatalf("NextNumber(projectA) = %d, want 3", nextA)
	}

	nextB, err := NextNumber(projectB, specsB)
	if err != nil {
		t.Fatalf("NextNumber(projectB) error = %v", err)
	}
	if nextB != 4 {
		t.Fatalf("NextNumber(projectB) = %d, want 4", nextB)
	}
}

func TestNextNumber_FallsBackToLocalSequenceWithoutSharedGitDir(t *testing.T) {
	projectRoot := t.TempDir()
	writeAllocatorFeatureDir(t, projectRoot, "docs/specs/0001-alpha")
	writeAllocatorFeatureDir(t, projectRoot, "docs/specs/0002-beta")

	cfg := config.Default()
	previousResolver := gitCommonDirResolver
	gitCommonDirResolver = func(string) (string, error) { return "", errors.New("git unavailable") }
	defer func() { gitCommonDirResolver = previousResolver }()

	next, err := NextNumber(projectRoot, cfg.SpecsPath(projectRoot))
	if err != nil {
		t.Fatalf("NextNumber() error = %v", err)
	}
	if next != 3 {
		t.Fatalf("NextNumber() = %d, want 3", next)
	}
}

func TestDuplicateNumberGroups_ReturnsOnlyConflicts(t *testing.T) {
	features := []Feature{
		{Number: 12, DirName: "0012-b"},
		{Number: 12, DirName: "0012-a"},
		{Number: 13, DirName: "0013-c"},
	}

	duplicates := DuplicateNumberGroups(features)
	if len(duplicates) != 1 {
		t.Fatalf("DuplicateNumberGroups() len = %d, want 1", len(duplicates))
	}
	group := duplicates[12]
	if len(group) != 2 {
		t.Fatalf("duplicate group len = %d, want 2", len(group))
	}
	if group[0].DirName != "0012-a" || group[1].DirName != "0012-b" {
		t.Fatalf("duplicate group order = [%s %s], want sorted dir names", group[0].DirName, group[1].DirName)
	}
}

func writeAllocatorFeatureDir(t *testing.T, projectRoot, relative string) {
	t.Helper()
	path := filepath.Join(projectRoot, filepath.FromSlash(relative))
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

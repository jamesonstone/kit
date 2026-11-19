package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/plugin"
)

func TestCopyMissingPluginFilesLeavesManifestAbsentWhenCopyFails(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	writeFixturePlugin(t, from, "roster\n")
	link := filepath.Join(from, plugin.RootDir, "fixture", "zzz-link")
	if err := os.Symlink(filepath.Join(t.TempDir(), "missing"), link); err != nil {
		t.Fatal(err)
	}
	if err := copyMissingPluginFiles(from, to); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatal(err)
	}
	manifest := filepath.Join(to, plugin.RootDir, "fixture", plugin.ManifestName)
	if _, err := os.Lstat(manifest); !os.IsNotExist(err) {
		t.Fatalf("partial manifest published: %v", err)
	}
	if !strings.Contains(readFile(t, filepath.Join(to, plugin.RootDir, "fixture", "instructions.md")), "roster") {
		t.Fatal("instructions were not copied before the failure")
	}
}

func TestCopyMissingPluginFilesRejectsDestinationSymlink(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	outside := t.TempDir()
	writeFixturePlugin(t, from, "roster\n")
	pluginDir := filepath.Join(to, plugin.RootDir, "fixture")
	if err := os.MkdirAll(filepath.Dir(pluginDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, pluginDir); err != nil {
		t.Fatal(err)
	}
	if err := copyMissingPluginFiles(from, to); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(outside, plugin.ManifestName)); !os.IsNotExist(err) {
		t.Fatalf("write escaped through the destination symlink: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(outside, "instructions.md")); !os.IsNotExist(err) {
		t.Fatalf("instructions escaped through the destination symlink: %v", err)
	}
}

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

func TestCopyMissingPluginFilesRejectsSymlinkManifest(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	secret := filepath.Join(t.TempDir(), "secret.yaml")
	if err := os.WriteFile(secret, []byte("secret: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeFixturePlugin(t, from, "roster\n")
	manifest := filepath.Join(from, plugin.RootDir, "fixture", plugin.ManifestName)
	if err := os.Remove(manifest); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, manifest); err != nil {
		t.Fatal(err)
	}
	if err := copyMissingPluginFiles(from, to); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatal(err)
	}
	copied := filepath.Join(to, plugin.RootDir, "fixture", plugin.ManifestName)
	if _, err := os.Lstat(copied); !os.IsNotExist(err) {
		t.Fatalf("symlink manifest was published: %v", err)
	}
	if readFile(t, secret) != "secret: value\n" {
		t.Fatal("source target changed")
	}
}

func TestCopyMissingPluginFilesKeepsInstructionNamedLikeTemp(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	writeFixturePlugin(t, from, "roster\n")
	tempName := filepath.Join(from, plugin.RootDir, "fixture", ".kit-plugin.yaml.tmp")
	if err := os.WriteFile(tempName, []byte("keep instruction\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyMissingPluginFiles(from, to); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(to, plugin.RootDir, "fixture")
	if readFile(t, filepath.Join(dir, ".kit-plugin.yaml.tmp")) != "keep instruction\n" {
		t.Fatal("instruction file was replaced by the temporary manifest")
	}
	if !strings.Contains(readFile(t, filepath.Join(dir, plugin.ManifestName)), "id: fixture") {
		t.Fatal("manifest was not published")
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
	outsideTemp := filepath.Join(outside, ".kit-plugin.yaml.tmp")
	if err := os.WriteFile(outsideTemp, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyMissingPluginFiles(from, to); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatal(err)
	}
	if readFile(t, outsideTemp) != "keep" {
		t.Fatal("cleanup removed a file through the destination symlink")
	}
	if _, err := os.Lstat(filepath.Join(outside, plugin.ManifestName)); !os.IsNotExist(err) {
		t.Fatalf("write escaped through the destination symlink: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(outside, "instructions.md")); !os.IsNotExist(err) {
		t.Fatalf("instructions escaped through the destination symlink: %v", err)
	}
}

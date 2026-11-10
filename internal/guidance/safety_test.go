package guidance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectRejectsEscapeAndSymlinks(t *testing.T) {
	root := project(t)
	outside := t.TempDir()
	write(t, outside, "private.md", "private marker")
	write(t, root, ".kit.yaml", "schema_version: 2\nagents: ['../private.md', '"+filepath.Join(outside, "private.md")+"']\n")
	write(t, root, "docs/references/rules/good.md", rule("good", "active"))
	if err := os.Symlink(filepath.Join(outside, "private.md"), filepath.Join(root, "docs/references/rules/link.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "docs/sops")); err != nil {
		t.Fatal(err)
	}
	report, err := Inspect(root)
	if err == nil {
		t.Fatal("unsafe paths accepted")
	}
	if !find(t, report, "docs/references/rules/good.md").Active {
		t.Fatal("valid lost")
	}
	if e := find(t, report, "docs/references/rules/link.md"); e.State != "invalid" || e.Metadata != nil {
		t.Fatalf("symlink read %+v", e)
	}
	for _, problem := range report.Errors {
		if strings.Contains(problem, "private marker") {
			t.Fatal("outside content disclosed")
		}
	}
}
func TestInspectRejectsSymlinkConfig(t *testing.T) {
	root := t.TempDir()
	outside := project(t)
	if err := os.Symlink(filepath.Join(outside, ".kit.yaml"), filepath.Join(root, ".kit.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(root); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("%v", err)
	}
}
func TestInspectPermissionFailures(t *testing.T) {
	root := project(t)
	path := filepath.Join(root, "docs/references/rules/blocked.md")
	write(t, root, "docs/references/rules/blocked.md", rule("blocked", "active"))
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0644) })
	if _, err := os.ReadFile(path); err == nil {
		t.Skip("process can read mode-000 files")
	}
	report, err := Inspect(root)
	if err == nil {
		t.Fatal("permission failure ignored")
	}
	if find(t, report, "docs/references/rules/blocked.md").State != "invalid" {
		t.Fatal("read failure reported missing")
	}
	if err := os.Chmod(filepath.Join(root, ".kit.yaml"), 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, ".kit.yaml"), 0644) })
	if _, err := Inspect(root); err == nil || strings.Contains(err.Error(), "not a kit project") {
		t.Fatalf("config permission: %v", err)
	}
}

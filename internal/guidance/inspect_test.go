package guidance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func write(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
func project(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, ".kit.yaml", "schema_version: 2\n")
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
func rule(slug, status string) string {
	return "---\nkind: ruleset\nslug: " + slug + "\ndescription: Sample rule\nstatus: " + status + "\napplies_to: [testing]\nread_policy_default: conditional\n---\n# Sample\n"
}
func find(t *testing.T, report Report, path string) Entry {
	t.Helper()
	for _, e := range report.Entries {
		if e.Path == path {
			return e
		}
	}
	t.Fatalf("missing %s in %+v", path, report)
	return Entry{}
}
func TestInspectLocalDeterministic(t *testing.T) {
	root := project(t)
	write(t, root, "docs/references/rules/z.md", rule("z", "optional"))
	write(t, root, "docs/references/rules/a.md", rule("a", "active"))
	write(t, root, "docs/sops/release/SOP.md", "---\nkind: sop\nslug: release\ndescription: Release procedure\nstatus: draft\nowner: Maintainers\n---\n")
	write(t, root, "AGENTS.md", "# Agent")
	write(t, root, "docs/specs/history/SPEC.md", "invalid: [")
	nested := filepath.Join(root, "src", "deep")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(root, ".kit.yaml"))
	a, err := Inspect(nested)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("nested report differs")
	}
	if !find(t, a, "docs/references/rules/a.md").Active {
		t.Fatal("active rule not active")
	}
	if find(t, a, "docs/references/rules/z.md").Active {
		t.Fatal("optional rule active")
	}
	if find(t, a, "docs/sops/release/SOP.md").State != "draft" {
		t.Fatal("draft missing")
	}
	first, _ := json.Marshal(a)
	second, _ := json.Marshal(b)
	if string(first) != string(second) {
		t.Fatal("unstable JSON")
	}
	after, _ := os.ReadFile(filepath.Join(root, ".kit.yaml"))
	if string(before) != string(after) {
		t.Fatal("config changed")
	}
}
func TestInspectInvalidPreservesGoodEntries(t *testing.T) {
	root := project(t)
	write(t, root, "docs/references/rules/good.md", rule("good", "active"))
	for _, tc := range []struct{ name, body string }{{"unknown", rule("unknown", "mystery")}, {"malformed", "---\nslug: [\n---\n"}, {"missing", "# No metadata"}} {
		write(t, root, "docs/references/rules/"+tc.name+".md", tc.body)
	}
	report, err := Inspect(root)
	if err == nil {
		t.Fatal("invalid inventory succeeded")
	}
	if !find(t, report, "docs/references/rules/good.md").Active {
		t.Fatal("valid entry lost")
	}
	for _, name := range []string{"unknown", "malformed", "missing"} {
		e := find(t, report, "docs/references/rules/"+name+".md")
		if e.Active || e.State != "invalid" || e.Error == "" {
			t.Fatalf("invalid entry %+v", e)
		}
	}
}
func TestResolveErrorsAndNearest(t *testing.T) {
	root := project(t)
	write(t, root, "nested/.kit.yaml", "schema_version: 2\n")
	got, err := Resolve(filepath.Join(root, "nested"))
	if err != nil || got != filepath.Join(root, "nested") {
		t.Fatalf("%s %v", got, err)
	}
	for _, tc := range []struct{ path, want string }{{filepath.Join(root, "missing"), "no such file"}, {filepath.Join(root, ".kit.yaml"), "not a directory"}, {t.TempDir(), "not a kit project"}} {
		_, err := Resolve(tc.path)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s: %v", tc.path, err)
		}
	}
	write(t, root, "nested/.kit.yaml", "schema_version: [")
	if _, err := Inspect(filepath.Join(root, "nested")); err == nil || strings.Contains(err.Error(), "not a kit project") {
		t.Fatalf("malformed config: %v", err)
	}
}
func TestInspectEmptyAndConfigErrors(t *testing.T) {
	root := project(t)
	report, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range report.Entries {
		if e.State != "missing" {
			t.Fatalf("unexpected entry %+v", e)
		}
	}
	for _, data := range []string{"schema_version: 999\n", "schema_version: 2\nconstitution_path: ''\n", "[]"} {
		write(t, root, ".kit.yaml", data)
		if _, err := Inspect(root); err == nil {
			t.Fatalf("accepted %s", data)
		}
	}
}
func TestInspectCwdAndRelative(t *testing.T) {
	root := project(t)
	write(t, root, "src/file.txt", "unrelated")
	t.Chdir(filepath.Join(root, "src"))
	a, err := Inspect("")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Inspect("..")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("cwd/relative mismatch")
	}
	cwd, _ := os.Getwd()
	if cwd != filepath.Join(root, "src") {
		t.Fatal("cwd changed")
	}
}

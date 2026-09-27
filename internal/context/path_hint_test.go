package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRecordsFutureAndDirectoryPathHintsWithoutBlocking(t *testing.T) {
	root := contextProject(t)
	writeContextFile(t, root, "docs/references/workflows/main.md", workflowDocument("main", nil, nil, nil))
	writeContextFile(t, root, "pkg/existing.go", "package pkg\n")

	result := Resolve(root, Request{Workflow: "main", Paths: []string{"pkg/new_file.go", "pkg", "pkg/existing.go"}})
	if result.Blocked {
		t.Fatalf("resolution blocked by valid path hints: %#v", result.Diagnostics)
	}
	states := map[string]EvidenceItem{}
	for _, item := range result.Evidence {
		states[item.Path] = item
	}
	if item := states["pkg/new_file.go"]; item.State != "absent" || item.Required {
		t.Fatalf("future path hint = %#v, want absent and not required", item)
	}
	if item := states["pkg"]; item.State != "directory" || item.Required {
		t.Fatalf("directory path hint = %#v, want directory and not required", item)
	}
	if item := states["pkg/existing.go"]; item.State != "present" || !item.Required {
		t.Fatalf("existing path hint = %#v, want present and required", item)
	}
}

func TestResolveStillBlocksPathHintsOutsideProject(t *testing.T) {
	root := contextProject(t)
	writeContextFile(t, root, "docs/references/workflows/main.md", workflowDocument("main", nil, nil, nil))

	result := Resolve(root, Request{Workflow: "main", Paths: []string{"../outside.go"}})
	if !result.Blocked {
		t.Fatalf("escaping path hint did not block: %#v", result.Evidence)
	}
}

func TestResolveBlocksMissingPathHintBelowEscapingSymlink(t *testing.T) {
	root := contextProject(t)
	writeContextFile(t, root, "docs/references/workflows/main.md", workflowDocument("main", nil, nil, nil))
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "outside")); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	result := Resolve(root, Request{Workflow: "main", Paths: []string{"outside/new_file.go"}})
	if !result.Blocked {
		t.Fatalf("missing hint below escaping symlink did not block: %#v", result.Evidence)
	}
}

func TestResolveBlocksDanglingSymlinkPathHint(t *testing.T) {
	root := contextProject(t)
	writeContextFile(t, root, "docs/references/workflows/main.md", workflowDocument("main", nil, nil, nil))
	target := filepath.Join(t.TempDir(), "missing.go")
	if err := os.Symlink(target, filepath.Join(root, "dangling.go")); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	result := Resolve(root, Request{Workflow: "main", Paths: []string{"dangling.go"}})
	if !result.Blocked {
		t.Fatalf("dangling symlink hint did not block: %#v", result.Evidence)
	}
}

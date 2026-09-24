package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

func TestRefreshDoesNotReinjectRetiredThreadLifecycle(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	leftover := `## Conversation Naming

- At creation, load ` + "`docs/references/thread-naming.md`" + `; if absent, use ` + "`kit instructions naming`" + `.

## Codex Thread Initialization Hard Gate

- First call ` + "`set_thread_title`" + `, then ` + "`set_thread_pinned`" + `.

`
	path := filepath.Join(root, "AGENTS.md")
	writeFile(t, path, leftover+templates.MemoryAgentsMD)
	opts := initRefreshOptions{files: []string{"AGENTS.md"}, outputOnly: true}
	if err := runInitRefresh(root, opts); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	for _, snippet := range []string{
		"## Conversation Naming",
		"kit instructions naming",
		"## Codex Thread Initialization Hard Gate",
		"set_thread_title",
		"set_thread_pinned",
	} {
		if !strings.Contains(got, snippet) {
			t.Fatalf("append-only refresh stripped leftover %q", snippet)
		}
		if strings.Contains(templates.MemoryAgentsMD, snippet) {
			t.Fatalf("current AGENTS template still contains %q", snippet)
		}
	}
	if err := runInitRefresh(root, opts); err != nil {
		t.Fatal(err)
	}
	if readFile(t, path) != got {
		t.Fatal("second refresh is not idempotent")
	}
}

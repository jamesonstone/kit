package worktree

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListInteractiveSelectorEntersSelectedWorktree(t *testing.T) {
	fixture := newGitFixture(t)
	fixture.app.isTerminal = func() bool { return true }
	fixture.app.selectList = func(_ context.Context, entries []worktreeEntry) (worktreeEntry, bool, error) {
		for _, entry := range entries {
			if samePath(entry.path, fixture.primary) {
				return entry, true, nil
			}
		}
		t.Fatalf("primary worktree was not offered: %#v", entries)
		return worktreeEntry{}, false, nil
	}
	var entered string
	fixture.app.runShell = func(_ context.Context, path string) error {
		entered = path
		return nil
	}

	runWT(t, fixture.app, fixture.primary, "list")
	if !samePath(entered, fixture.primary) {
		t.Fatalf("entered %q, want %q", entered, fixture.primary)
	}
	if fixture.out.Len() != 0 {
		t.Fatalf("interactive list unexpectedly wrote the plain table:\n%s", fixture.out.String())
	}
}

func TestListPlainBypassesInteractiveSelector(t *testing.T) {
	fixture := newGitFixture(t)
	fixture.app.isTerminal = func() bool { return true }
	fixture.app.selectList = func(context.Context, []worktreeEntry) (worktreeEntry, bool, error) {
		t.Fatal("--plain must bypass the interactive selector")
		return worktreeEntry{}, false, nil
	}

	runWT(t, fixture.app, fixture.primary, "list", "--plain")
	if !strings.Contains(fixture.out.String(), "STATE\tHEAD\tLAST UPDATED\tPATH") {
		t.Fatalf("plain list output:\n%s", fixture.out.String())
	}
}

func TestReadSelectorKeySupportsArrowsAndTab(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		input string
		want  selectorKey
	}{
		"down":      {input: "\x1b[B", want: selectorNext},
		"right":     {input: "\x1b[C", want: selectorNext},
		"tab":       {input: "\t", want: selectorNext},
		"up":        {input: "\x1b[A", want: selectorPrevious},
		"left":      {input: "\x1b[D", want: selectorPrevious},
		"shift-tab": {input: "\x1b[Z", want: selectorPrevious},
		"enter":     {input: "\r", want: selectorChoose},
		"cancel":    {input: "q", want: selectorCancel},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := readSelectorKey(strings.NewReader(test.input))
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("key = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRenderWorktreeSelectorUsesColorAndReadableDate(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "selector.txt")
	output, err := os.Create(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	entries := []worktreeEntry{
		{branch: "GH-86", state: "clean", updatedText: "Jul 26, 2026", path: "/tmp/GH-86"},
		{branch: "topic/dirty", state: "dirty", updatedText: "Jul 25, 2026", path: "/tmp/topic"},
	}
	if _, err := renderWorktreeSelector(output, entries, 0); err != nil {
		t.Fatal(err)
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(colorBrightCyan)) || !bytes.Contains(data, []byte(colorYellow)) {
		t.Fatalf("selector output is missing selection/state colors: %q", data)
	}
	for _, want := range []string{"GH-86", "topic/dirty", "Jul 26, 2026"} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("selector output is missing %q: %q", want, data)
		}
	}
}

func TestParseListOptionsRecognizesPlain(t *testing.T) {
	options, err := parseListOptions([]string{"--sort", "head", "--reverse", "--plain"})
	if err != nil {
		t.Fatal(err)
	}
	if options.sortBy != "head" || !options.reverse || !options.plain {
		t.Fatalf("options = %#v", options)
	}
}

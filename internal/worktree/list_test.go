package worktree

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode"
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
	if !strings.Contains(fixture.out.String(), "STATE\tHEAD\tPR#\tLAST UPDATED\tPATH") {
		t.Fatalf("plain list output:\n%s", fixture.out.String())
	}
}

func TestReadSelectorKeySupportsArrowsAndTab(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		input string
		want  selectorKey
	}{
		"down":       {input: "\x1b[B", want: selectorNext},
		"right":      {input: "\x1b[C", want: selectorNext},
		"tab":        {input: "\t", want: selectorNext},
		"up":         {input: "\x1b[A", want: selectorPrevious},
		"left":       {input: "\x1b[D", want: selectorPrevious},
		"shift-tab":  {input: "\x1b[Z", want: selectorPrevious},
		"enter":      {input: "\r", want: selectorChoose},
		"home":       {input: "h", want: selectorHome},
		"home-upper": {input: "H", want: selectorHome},
		"cancel":     {input: "q", want: selectorCancel},
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
		{branch: "main", primary: true, state: "clean", prText: "-", updatedText: "Jul 26, 2026 17:44", path: "/tmp/root"},
		{branch: "GH-86", state: "clean", prText: "94", updatedText: "Jul 26, 2026 17:43", path: "/tmp/GH-86"},
		{branch: "topic/dirty", state: "dirty", prText: "-", updatedText: "Jul 25, 2026 09:08", path: "/tmp/topic"},
	}
	if _, err := renderWorktreeSelector(output, entries, 1); err != nil {
		t.Fatal(err)
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, color := range []string{colorBrightMagenta, colorBrightCyan, colorYellow} {
		if !bytes.Contains(data, []byte(color)) {
			t.Fatalf("selector output is missing %q: %q", color, data)
		}
	}
	for _, want := range []string{"main [home]", "GH-86", "topic/dirty", "PR#", "94", "Jul 26, 2026 17:44", "h: home"} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("selector output is missing %q: %q", want, data)
		}
	}
}

func TestRenderWorktreeSelectorKeepsSelectedHomeBrightMagenta(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "selector.txt")
	output, err := os.Create(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	entry := worktreeEntry{
		branch: "main", primary: true, state: "clean",
		prText: "-", updatedText: "Jul 26, 2026 17:44", path: "/tmp/root",
	}
	if _, err := renderWorktreeSelector(output, []worktreeEntry{entry}, 0); err != nil {
		t.Fatal(err)
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	expectedLine := colorBrightMagenta + "> clean"
	if !bytes.Contains(data, []byte(expectedLine)) {
		t.Fatalf("selected primary row is not bright magenta: %q", data)
	}
	if bytes.Contains(data, []byte(colorBrightCyan+"> clean")) {
		t.Fatalf("selected primary row used the generic selection color: %q", data)
	}
}

func TestSelectorEntryColorKeepsMainBrightMagenta(t *testing.T) {
	t.Parallel()
	if got := selectorEntryColor(worktreeEntry{branch: "main"}, "clean", true); got != colorBrightMagenta {
		t.Fatalf("selected main color = %q, want %q", got, colorBrightMagenta)
	}
	if got := selectorEntryColor(worktreeEntry{primary: true, branch: "topic"}, "dirty", false); got != colorBrightMagenta {
		t.Fatalf("primary topic color = %q, want %q", got, colorBrightMagenta)
	}
	if got := selectorEntryColor(worktreeEntry{branch: "GH-95"}, "clean", true); got != colorBrightCyan {
		t.Fatalf("selected lane color = %q, want %q", got, colorBrightCyan)
	}
}

func TestSelectorDisplayHeadMarksIdentity(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		entry worktreeEntry
		want  string
	}{
		{name: "primary main", entry: worktreeEntry{primary: true, branch: "main"}, want: "main [home]"},
		{name: "primary topic", entry: worktreeEntry{primary: true, branch: "topic"}, want: "topic [home]"},
		{name: "linked main", entry: worktreeEntry{branch: "main"}, want: "main [main]"},
		{name: "ordinary lane", entry: worktreeEntry{branch: "GH-114"}, want: "GH-114"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := selectorDisplayHead(test.entry); got != test.want {
				t.Fatalf("selectorDisplayHead() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestPrimaryListEntry(t *testing.T) {
	t.Parallel()
	want := worktreeEntry{path: "/tmp/root", primary: true}
	got, ok := primaryListEntry([]worktreeEntry{{path: "/tmp/lane"}, want})
	if !ok || got != want {
		t.Fatalf("primaryListEntry() = %#v, %t, want %#v, true", got, ok, want)
	}
	if _, ok := primaryListEntry([]worktreeEntry{{path: "/tmp/lane"}}); ok {
		t.Fatal("primaryListEntry() found a missing primary")
	}
}

func TestPinPrimaryListEntry(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		position string
		want     []string
	}{
		{name: "top", position: "top", want: []string{"root", "one", "two"}},
		{name: "bottom", position: "bottom", want: []string{"one", "two", "root"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			entries := []worktreeEntry{
				{path: "one"},
				{path: "root", primary: true},
				{path: "two"},
			}
			pinPrimaryListEntry(entries, test.position)
			got := make([]string, len(entries))
			for i := range entries {
				got[i] = entries[i].path
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("paths = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestRenderWorktreeSelectorSanitizesDynamicFields(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "selector.txt")
	output, err := os.Create(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	entry := worktreeEntry{
		branch:      "topic/\x1b[31mred",
		state:       "dirty\x1b[2J",
		prText:      "12\x1b[2J",
		updatedText: "Jul 26,\r2026 17:44",
		path:        "/tmp/\x9b2Jowned\nlane",
	}
	if _, err := renderWorktreeSelector(output, []worktreeEntry{entry}, 0); err != nil {
		t.Fatal(err)
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	expectedLine := fmt.Sprintf(
		"%s%s%s\r\n",
		colorBrightCyan,
		fmt.Sprintf(
			"> %-8s %-24s %-6s %-18s %s",
			"dirty[2J",
			"topic/[31mred",
			"12[2J",
			"Jul 26,2026 17:44",
			"/tmp/2Jownedlane",
		),
		colorReset,
	)
	if !bytes.Contains(data, []byte(expectedLine)) {
		t.Fatalf("selector output does not preserve sanitized alignment:\ngot  %q\nwant %q", data, expectedLine)
	}

	unstyled := string(data)
	for _, sequence := range []string{colorReset, colorBold, colorBrightCyan, colorBrightMagenta, colorGreen, colorYellow, colorRed} {
		unstyled = strings.ReplaceAll(unstyled, sequence, "")
	}
	for _, char := range unstyled {
		if char != '\r' && char != '\n' && unicode.IsControl(char) {
			t.Fatalf("selector output contains injected control character %U: %q", char, data)
		}
	}
}

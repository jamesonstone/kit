package worktree

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode"

	"golang.org/x/term"
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
		{branch: "main", primary: true, state: "clean", updatedText: "Jul 26, 2026", path: "/tmp/root"},
		{branch: "GH-86", state: "clean", updatedText: "Jul 26, 2026", path: "/tmp/GH-86"},
		{branch: "topic/dirty", state: "dirty", updatedText: "Jul 25, 2026", path: "/tmp/topic"},
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
	for _, color := range []string{colorBrightGreen, colorBrightCyan, colorYellow} {
		if !bytes.Contains(data, []byte(color)) {
			t.Fatalf("selector output is missing %q: %q", color, data)
		}
	}
	for _, want := range []string{"main", "GH-86", "topic/dirty", "Jul 26, 2026", "h: home"} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("selector output is missing %q: %q", want, data)
		}
	}
}

func TestRenderWorktreeSelectorKeepsSelectedHomeBrightGreen(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "selector.txt")
	output, err := os.Create(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	entry := worktreeEntry{
		branch: "main", primary: true, state: "clean",
		updatedText: "Jul 26, 2026", path: "/tmp/root",
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
	expectedLine := colorBrightGreen + "> clean"
	if !bytes.Contains(data, []byte(expectedLine)) {
		t.Fatalf("selected primary row is not bright green: %q", data)
	}
	if bytes.Contains(data, []byte(colorBrightCyan+"> clean")) {
		t.Fatalf("selected primary row used the generic selection color: %q", data)
	}
}

func TestSelectorEntryColorKeepsMainBrightGreen(t *testing.T) {
	t.Parallel()
	if got := selectorEntryColor(worktreeEntry{branch: "main"}, "clean", true); got != colorBrightGreen {
		t.Fatalf("selected main color = %q, want %q", got, colorBrightGreen)
	}
	if got := selectorEntryColor(worktreeEntry{primary: true, branch: "topic"}, "dirty", false); got != colorBrightGreen {
		t.Fatalf("primary topic color = %q, want %q", got, colorBrightGreen)
	}
	if got := selectorEntryColor(worktreeEntry{branch: "GH-95"}, "clean", true); got != colorBrightCyan {
		t.Fatalf("selected lane color = %q, want %q", got, colorBrightCyan)
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
		updatedText: "Jul 26,\r2026",
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
			"> %-8s %-24s %-13s %s",
			"dirty[2J",
			"topic/[31mred",
			"Jul 26,2026",
			"/tmp/2Jownedlane",
		),
		colorReset,
	)
	if !bytes.Contains(data, []byte(expectedLine)) {
		t.Fatalf("selector output does not preserve sanitized alignment:\ngot  %q\nwant %q", data, expectedLine)
	}

	unstyled := string(data)
	for _, sequence := range []string{colorReset, colorBold, colorBrightCyan, colorBrightGreen, colorGreen, colorYellow, colorRed} {
		unstyled = strings.ReplaceAll(unstyled, sequence, "")
	}
	for _, char := range unstyled {
		if char != '\r' && char != '\n' && unicode.IsControl(char) {
			t.Fatalf("selector output contains injected control character %U: %q", char, data)
		}
	}
}

func TestSelectWorktreeTerminalCancellationRestoresPTY(t *testing.T) {
	script, err := exec.LookPath("script")
	if err != nil {
		t.Skip("script command is required for pseudo-terminal coverage")
	}
	args := []string{"-q"}
	switch runtime.GOOS {
	case "darwin":
		args = append(args, "/dev/null", os.Args[0], "-test.run=^TestSelectWorktreeTerminalCancellationPTYHelper$")
	case "linux":
		command := strings.Join([]string{
			shellQuote(os.Args[0]),
			"-test.run=^TestSelectWorktreeTerminalCancellationPTYHelper$",
		}, " ")
		args = append(args, "-c", command, "/dev/null")
	default:
		t.Skipf("script pseudo-terminal invocation is not configured for %s", runtime.GOOS)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, script, args...)
	command.Env = append(os.Environ(), "KIT_TEST_SELECTOR_IDLE_PTY=1")
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	sentinelWritten := make(chan error, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		_, writeErr := fmt.Fprintln(stdin, "sentinel-after-cancel")
		sentinelWritten <- writeErr
	}()
	err = command.Wait()
	_ = stdin.Close()
	if writeErr := <-sentinelWritten; writeErr != nil {
		t.Fatalf("write post-cancellation sentinel: %v\n%s", writeErr, output.String())
	}
	if ctx.Err() != nil {
		t.Fatalf("idle selector did not return promptly after cancellation: %v\n%s", ctx.Err(), output.String())
	}
	if err != nil {
		t.Fatalf("idle selector PTY helper failed: %v\n%s", err, output.String())
	}
	if !bytes.Contains(output.Bytes(), []byte(hideCursor)) || !bytes.Contains(output.Bytes(), []byte(showCursor)) {
		t.Fatalf("selector did not restore cursor visibility: %q", output.Bytes())
	}
	if !strings.Contains(output.String(), "selector PTY cancellation passed") {
		t.Fatalf("selector PTY helper did not confirm terminal restoration: %q", output.Bytes())
	}
}

func TestSelectWorktreeTerminalCancellationPTYHelper(t *testing.T) {
	if os.Getenv("KIT_TEST_SELECTOR_IDLE_PTY") != "1" {
		return
	}
	inputFD := int(os.Stdin.Fd())
	if !term.IsTerminal(inputFD) || !term.IsTerminal(int(os.Stdout.Fd())) {
		t.Fatal("PTY helper streams are not terminals")
	}
	before, err := term.GetState(inputFD)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)
	started := time.Now()
	_, ok, err := selectWorktreeTerminal(
		ctx,
		os.Stdin,
		os.Stdout,
		[]worktreeEntry{{branch: "GH-86", state: "clean", updatedText: "Jul 26, 2026", path: "/tmp/GH-86"}},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("selector error = %v, want context cancellation", err)
	}
	if ok {
		t.Fatal("cancelled selector unexpectedly returned a selection")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("selector cancellation took %s", elapsed)
	}
	after, err := term.GetState(inputFD)
	if err != nil {
		t.Fatal(err)
	}
	if beforeState, afterState := normalizedTerminalState(before), normalizedTerminalState(after); !reflect.DeepEqual(afterState, beforeState) {
		t.Fatalf("terminal state was not restored:\nbefore: %#v\nafter:  %#v", before, after)
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		t.Fatalf("read post-cancellation terminal input: %v", err)
	}
	if strings.TrimSpace(line) != "sentinel-after-cancel" {
		t.Fatalf("post-cancellation terminal input = %q, want sentinel", line)
	}
	fmt.Fprintln(os.Stderr, "selector PTY cancellation passed")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func normalizedTerminalState(state *term.State) []uint64 {
	var values []uint64
	var collect func(reflect.Value, string)
	collect = func(value reflect.Value, name string) {
		switch value.Kind() {
		case reflect.Pointer:
			collect(value.Elem(), name)
		case reflect.Struct:
			valueType := value.Type()
			for i := 0; i < value.NumField(); i++ {
				collect(value.Field(i), valueType.Field(i).Name)
			}
		case reflect.Array:
			for i := 0; i < value.Len(); i++ {
				collect(value.Index(i), name)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			fieldValue := value.Uint()
			if runtime.GOOS == "darwin" && name == "Lflag" {
				fieldValue &^= 0x20000000 // PENDIN is a transient input-reprint flag.
			}
			values = append(values, fieldValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			values = append(values, uint64(value.Int()))
		}
	}
	collect(reflect.ValueOf(state), "")
	return values
}

func TestParseListOptionsRecognizesPlain(t *testing.T) {
	options, err := parseListOptions([]string{"--sort", "head", "--root-position", "bottom", "--reverse", "--plain"})
	if err != nil {
		t.Fatal(err)
	}
	if options.sortBy != "head" || options.rootPosition != "bottom" || !options.reverse || !options.plain {
		t.Fatalf("options = %#v", options)
	}
}

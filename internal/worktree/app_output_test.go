package worktree

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestRootCommandRunsList(t *testing.T) {
	fixture := newGitFixture(t)

	runWT(t, fixture.app, fixture.primary)
	got := fixture.out.String()

	fixture.out.Reset()
	runWT(t, fixture.app, fixture.primary, "list")
	if want := fixture.out.String(); got != want {
		t.Fatalf("root command output differs from list:\nroot:\n%s\nlist:\n%s", got, want)
	}
}

func TestHelpCommandsShowUsage(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"--help"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out bytes.Buffer
			app := NewApp(&out, io.Discard)
			if err := app.Run(context.Background(), t.TempDir(), args); err != nil {
				t.Fatalf("Run(%q) error = %v", args, err)
			}
			if !strings.Contains(out.String(), "Usage: git wt [command] [arguments]") {
				t.Fatalf("Run(%q) output:\n%s", args, out.String())
			}
		})
	}
}

func TestUnknownCommandShowsHelp(t *testing.T) {
	fixture := newGitFixture(t)
	err := fixture.app.Run(context.Background(), fixture.primary, []string{"nope"})
	if err == nil || !strings.Contains(err.Error(), "Usage: git wt") {
		t.Fatalf("unknown command error = %v", err)
	}
}

func TestOutputFailureIsReturned(t *testing.T) {
	app := NewApp(failingWriter{}, io.Discard)
	err := app.Run(context.Background(), t.TempDir(), []string{"help"})
	if err == nil || !strings.Contains(err.Error(), "write output") {
		t.Fatalf("help output error = %v", err)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("write failed")
}

func Example() {
	fmt.Println("git wt issue 76")
	fmt.Println("git wt home")
	fmt.Println("git wt cd GH-76")
	fmt.Println(`cd "$(git wt path GH-76)"`)
	fmt.Println("git wt pr 77")
	fmt.Println("git wt repair 77")
	// Output:
	// git wt issue 76
	// git wt home
	// git wt cd GH-76
	// cd "$(git wt path GH-76)"
	// git wt pr 77
	// git wt repair 77
}

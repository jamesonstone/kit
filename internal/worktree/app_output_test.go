package worktree

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

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

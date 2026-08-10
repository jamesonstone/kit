package worktree

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/term"
)

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
		// Race-instrumented helper startup can exceed 200 ms. Keep the input
		// absent until well after the helper's 50 ms cancellation timer starts,
		// otherwise the queued newline is interpreted as a selection.
		time.Sleep(2 * time.Second)
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
		[]worktreeEntry{{branch: "GH-86", state: "clean", updatedText: "Jul 26, 2026 17:44", path: "/tmp/GH-86"}},
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

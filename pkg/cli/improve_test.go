package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/improve"
)

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestImproveCommandOverview(t *testing.T) {
	cmd := newImproveCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("kit improve error = %v", err)
	}
	content := out.String()
	for _, want := range []string{
		"Kit improve",
		"kit improve run --suite default",
		"kit improve mine --from .kit/improve/latest",
		"kit improve validate --candidate <path>",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected overview to contain %q, got %q", want, content)
		}
	}
}

func TestImproveRunSupportsBinaryOverrideAndFailsFailedManifest(t *testing.T) {
	cmd := newImproveRunCommand(&improveOptions{})
	if cmd.Flags().Lookup("kit-binary") == nil {
		t.Fatalf("expected --kit-binary flag")
	}
	err := improveRunFailure(improve.RunManifest{RunID: "run-1", RunDir: "/tmp/run-1", Status: "failed"})
	if err == nil || !strings.Contains(err.Error(), "benchmark run-1 failed") {
		t.Fatalf("improveRunFailure() = %v", err)
	}
	if err := improveRunFailure(improve.RunManifest{Status: "pass"}); err != nil {
		t.Fatalf("passing manifest returned error: %v", err)
	}
}

func TestImproveRunPropagatesHumanReadableWriteError(t *testing.T) {
	wantErr := errors.New("write failed")
	cmd := newImproveRunCommand(&improveOptions{})
	cmd.SetOut(errorWriter{err: wantErr})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--dry-run"})

	if err := cmd.Execute(); !errors.Is(err, wantErr) {
		t.Fatalf("kit improve run error = %v, want %v", err, wantErr)
	}
}

func TestImproveCommandRegistersSubcommands(t *testing.T) {
	for _, commandPath := range [][]string{
		{"improve", "run"},
		{"improve", "mine"},
		{"improve", "propose"},
		{"improve", "validate"},
		{"improve", "report"},
		{"improve", "pr"},
	} {
		cmd, _, err := rootCmd.Find(commandPath)
		if err != nil {
			t.Fatalf("rootCmd.Find(%v) error = %v", commandPath, err)
		}
		if cmd == nil {
			t.Fatalf("expected command %v to be registered", commandPath)
		}
	}
}

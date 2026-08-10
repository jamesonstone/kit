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
		"Run benchmark-backed Kit harness improvement workflows",
		"run",
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
	cmd := newImproveCommand()
	if len(cmd.Commands()) != 1 || cmd.Commands()[0].Name() != "run" {
		t.Fatalf("improve subcommands = %#v, want only run", cmd.Commands())
	}
}

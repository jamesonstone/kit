package cli

import (
	"bytes"
	"runtime/debug"
	"testing"
)

func TestCurrentVersion_PrefersLinkerInjectedVersion(t *testing.T) {
	originalVersion := Version
	originalBuildInfoReader := buildInfoReader
	Version = "v1.2.3"
	defer func() {
		Version = originalVersion
		buildInfoReader = originalBuildInfoReader
	}()

	buildInfoReader = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, true
	}

	if got, want := currentVersion(), "v1.2.3"; got != want {
		t.Fatalf("currentVersion() = %q, want %q", got, want)
	}
}

func TestCurrentVersion_FallsBackToBuildInfo(t *testing.T) {
	originalVersion := Version
	originalBuildInfoReader := buildInfoReader
	Version = "dev"
	defer func() {
		Version = originalVersion
		buildInfoReader = originalBuildInfoReader
	}()

	buildInfoReader = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, true
	}

	if got, want := currentVersion(), "v1.2.3"; got != want {
		t.Fatalf("currentVersion() = %q, want %q", got, want)
	}
}

func TestRunVersion_PrintsCurrentVersion(t *testing.T) {
	originalVersion := Version
	originalBuildInfoReader := buildInfoReader
	Version = "v1.2.3"
	defer func() {
		Version = originalVersion
		buildInfoReader = originalBuildInfoReader
	}()

	buildInfoReader = func() (*debug.BuildInfo, bool) {
		return nil, false
	}

	cmd := versionCmd
	output := &bytes.Buffer{}
	cmd.SetOut(output)

	if err := runVersion(cmd, nil); err != nil {
		t.Fatalf("runVersion() error = %v", err)
	}

	if got, want := output.String(), "v1.2.3\n"; got != want {
		t.Fatalf("runVersion() output = %q, want %q", got, want)
	}
}

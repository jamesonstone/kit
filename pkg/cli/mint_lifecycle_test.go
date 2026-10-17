package cli

import (
	"context"
	"os"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/minttool"
)

// Package tests must never maintain a developer's real Mint installation.
// Individual lifecycle tests replace this shared seam and restore it afterwards.
func TestMain(m *testing.M) {
	maintainInstalledMint = func(context.Context, minttool.Options, bool) (minttool.Report, error) {
		return minttool.Report{Installed: minttool.Installation{Status: "missing", Method: "none"}, VersionStatus: "missing"}, nil
	}
	os.Exit(m.Run())
}

func stubMintLifecycle(t *testing.T, fn func(context.Context, minttool.Options, bool) (minttool.Report, error)) {
	t.Helper()
	previous := maintainInstalledMint
	maintainInstalledMint = fn
	t.Cleanup(func() { maintainInstalledMint = previous })
}

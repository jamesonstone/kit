package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/minttool"
)

func TestMintToolsReadOnlyUpdateAndFailureJSON(t *testing.T) {
	root := t.TempDir()
	setWorkingDirectory(t, root)
	writeFile(t, filepath.Join(root, ".kit.yaml"), "schema_version: 2\n")
	writeFile(t, filepath.Join(root, "go.mod"), "module acme/app\nrequire github.com/jamesonstone/mint v0.5.0\n")
	oldCheck, oldUpdate, oldPreview := checkMintTool, updateMintTool, previewMintTool
	t.Cleanup(func() { checkMintTool, updateMintTool, previewMintTool = oldCheck, oldUpdate, oldPreview })
	checks, updates, previews := 0, 0, 0
	fixture := minttool.Report{Installed: minttool.Installation{Path: "/fixture/mint", Version: "v0.8.0", Status: "installed"}, Latest: "v0.8.0"}
	checkMintTool = func(context.Context, minttool.Options) (minttool.Report, error) { checks++; return fixture, nil }
	updateMintTool = func(context.Context, minttool.Options) (minttool.Report, error) { updates++; return fixture, nil }
	previewMintTool = func(context.Context, minttool.Options) (minttool.Report, error) { previews++; return fixture, nil }
	for _, tc := range []struct {
		update                             bool
		args                               []string
		wantCheck, wantUpdate, wantPreview int
	}{
		{false, []string{"--json"}, 1, 0, 0},
		{true, []string{"--json", "--dry-run"}, 1, 0, 1},
		{true, []string{"--json", "--version", "v0.8.0"}, 1, 1, 1},
	} {
		cmd := newMintToolsCommand(tc.update)
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(tc.args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		var result mintToolsReport
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatalf("%s %v", out.String(), err)
		}
		if len(result.RepositoryPins) != 1 || result.RepositoryPinsUpdated || result.RepositoryPins[0].Ref != "v0.5.0" {
			t.Fatalf("%#v", result)
		}
		if checks != tc.wantCheck || updates != tc.wantUpdate || previews != tc.wantPreview {
			t.Fatalf("%d %d %d", checks, updates, previews)
		}
	}
	checkMintTool = func(context.Context, minttool.Options) (minttool.Report, error) {
		return fixture, errors.New("release lookup failed")
	}
	cmd := newMintToolsCommand(false)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("failure exit expected")
	}
	var result mintToolsReport
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Local.Installed.Version != "v0.8.0" {
		t.Fatalf("%s %v", out.String(), err)
	}
}

func TestMintPinsAreSeparateAndRejectLinks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example\nrequire github.com/jamesonstone/mint v0.5.0\n")
	writeFile(t, filepath.Join(root, ".github/workflows/ci.yml"), "jobs:\n  go:\n    steps:\n      - uses: jamesonstone/mint@0123456789012345678901234567890123456789 # v0.5.0\n      - run: echo jamesonstone/mint@not-an-action\n")
	pins, err := inspectMintRepositoryPins(root)
	if err != nil || len(pins) != 2 || pins[1].VersionHint != "v0.5.0" || pins[1].Compatibility != "not_verified" {
		t.Fatalf("%#v %v", pins, err)
	}
}

func TestMintToolsInventoryFailureStillEmitsJSON(t *testing.T) {
	root := t.TempDir()
	setWorkingDirectory(t, root)
	writeFile(t, filepath.Join(root, ".kit.yaml"), "schema_version: 2\n")
	writeFile(t, filepath.Join(root, ".github/workflows/bad.yml"), "jobs: [\n")
	cmd := newMintToolsCommand(false)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("bad inventory accepted")
	}
	var report mintToolsReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil || report.Error == "" || report.RepositoryPins == nil {
		t.Fatalf("missing failure JSON: %s %v", out.String(), err)
	}
}

func TestMintPinsSymlinksFailClosed(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "go.mod"), "module elsewhere\n")
	if err := os.Symlink(filepath.Join(outside, "go.mod"), filepath.Join(root, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if _, err := inspectMintRepositoryPins(root); err == nil {
		t.Fatal("linked module accepted")
	}
	root = t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, ".github")); err != nil {
		t.Fatal(err)
	}
	if _, err := inspectMintRepositoryPins(root); err == nil {
		t.Fatal("linked workflow directory accepted")
	}
}

func TestMintModuleInventoryBothRequireFormsAndReplacement(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module app\nrequire github.com/jamesonstone/mint v0.5.0\nrequire (\n github.com/jamesonstone/mint v0.8.0 // note\n)\nreplace github.com/jamesonstone/mint => ../mint\n")
	pins, err := inspectMintRepositoryPins(root)
	if err != nil || len(pins) != 3 || pins[0].Ref != "v0.5.0" || pins[1].Ref != "v0.8.0" || pins[2].Kind != "go-module-replacement" {
		t.Fatalf("%#v %v", pins, err)
	}
}

func TestMintPinHintsDoNotCopyUnrelatedCommentText(t *testing.T) {
	for _, tc := range []struct{ comment, want string }{
		{"# v0.5.0 explanation", "v0.5.0"}, {"# unrelated private annotation", ""}, {"# v0.5.0-rc.1", ""},
	} {
		if got := mintPinVersionHint(tc.comment); got != tc.want {
			t.Fatalf("hint %q, want %q", got, tc.want)
		}
	}
}

func TestMintToolsOutsideProjectPinsAreAnArray(t *testing.T) {
	setWorkingDirectory(t, t.TempDir())
	previous := checkMintTool
	t.Cleanup(func() { checkMintTool = previous })
	checkMintTool = func(context.Context, minttool.Options) (minttool.Report, error) {
		return minttool.Report{Installed: minttool.Installation{Status: "missing"}}, nil
	}
	cmd := newMintToolsCommand(false)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("missing executable must exit nonzero")
	}
	var report mintToolsReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil || report.RepositoryPins == nil || len(report.RepositoryPins) != 0 {
		t.Fatalf("%s %v", out.String(), err)
	}
}

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/instructions"
)

func TestNamingReferenceRefreshIsIdempotentAndPreservesCustomContent(t *testing.T) {
	root := t.TempDir()
	setupInitHome(t)
	path := "docs/references/thread-naming.md"
	opts := initRefreshOptions{files: []string{path}, outputOnly: true}
	if err := runInitRefresh(root, opts); err != nil {
		t.Fatal(err)
	}
	full := filepath.Join(root, path)
	before, err := os.ReadFile(full)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != instructions.ThreadNamingPolicy {
		t.Fatal("missing installed canonical policy")
	}
	if err := runInitRefresh(root, opts); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(full)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("second refresh changed policy")
	}
	custom := "# User policy\nKeep my title\n"
	if err := os.WriteFile(full, []byte(custom), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runInitRefresh(root, opts); err == nil {
		t.Fatal("expected existing append-only conflict for unrecognized user sections")
	}
	after, err = os.ReadFile(full)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != custom {
		t.Fatal("refresh overwrote user-owned content")
	}
}

func TestInstructionsNamingPrintsCanonicalPolicy(t *testing.T) {
	cmd := newInstructionsCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"naming"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.String() != instructions.ThreadNamingPolicy {
		t.Fatal("CLI policy differs")
	}
}

package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func mintRemoteFixture(t *testing.T) (string, string) {
	t.Helper()
	root := mintPreparationRepository(t)
	remote := t.TempDir()
	mintPreparationGit(t, remote, "init", "--bare", "--initial-branch=trunk")
	mintPreparationGit(t, root, "commit", "--allow-empty", "-m", "feat: local fixture")
	mintPreparationGit(t, root, "remote", "add", "origin", remote)
	mintPreparationGit(t, root, "push", "origin", "HEAD:refs/heads/trunk")
	mintPreparationGit(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/old-default")
	return root, remote
}

func mintFilteredTrigger(t *testing.T, branch string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("on:\n  push:\n    branches: ["+branch+"]\n"), &doc); err != nil {
		t.Fatal(err)
	}
	return &doc
}

func TestMintRemoteHEADOverridesStaleLocalMetadata(t *testing.T) {
	root, remote := mintRemoteFixture(t)
	before := mintPreparationGit(t, root, "symbolic-ref", "refs/remotes/origin/HEAD")
	branch, err := mintRemoteDefaultBranch(root)
	if err != nil || branch != "trunk" {
		t.Fatalf("branch=%q err=%v", branch, err)
	}
	if err := validateMintPushTrigger(root, mintFilteredTrigger(t, "trunk")); err != nil {
		t.Fatal(err)
	}
	if err := validateMintPushTrigger(root, mintFilteredTrigger(t, "old-default")); err == nil {
		t.Fatal("stale local branch accepted")
	}
	mintPreparationGit(t, root, "push", "origin", "HEAD:refs/heads/release")
	mintPreparationGit(t, remote, "symbolic-ref", "HEAD", "refs/heads/release")
	if err := validateMintPushTrigger(root, mintFilteredTrigger(t, "trunk")); err == nil {
		t.Fatal("old remote branch accepted after remote HEAD changed")
	}
	if err := validateMintPushTrigger(root, mintFilteredTrigger(t, "release")); err != nil {
		t.Fatal(err)
	}
	after := mintPreparationGit(t, root, "symbolic-ref", "refs/remotes/origin/HEAD")
	if after != before {
		t.Fatal("read-only verification updated local origin/HEAD")
	}
	if _, err := os.Stat(filepath.Join(root, ".git/FETCH_HEAD")); !os.IsNotExist(err) {
		t.Fatal("remote verification fetched repository refs")
	}
}

func TestMintRemoteHEADFailsClosed(t *testing.T) {
	root, remote := mintRemoteFixture(t)
	mintPreparationGit(t, remote, "symbolic-ref", "HEAD", "refs/heads/missing")
	if _, err := mintRemoteDefaultBranch(root); err == nil || !strings.Contains(err.Error(), "INCOMPLETE") {
		t.Fatalf("missing remote branch: %v", err)
	}
	mintPreparationGit(t, root, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "does-not-exist"))
	if err := validateMintPushTrigger(root, mintFilteredTrigger(t, "old-default")); err == nil {
		t.Fatal("unavailable remote fell back to local metadata")
	}
	t.Setenv("PATH", t.TempDir())
	if _, err := mintRemoteDefaultBranch(root); err == nil || !strings.Contains(err.Error(), "INCOMPLETE") {
		t.Fatalf("missing git: %v", err)
	}
	for _, trigger := range []string{"push", "[push, pull_request]", "\n  push:"} {
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte("on: "+trigger+"\n"), &doc); err != nil {
			t.Fatal(err)
		}
		if err := validateMintPushTrigger(root, &doc); err != nil {
			t.Fatalf("unfiltered trigger unexpectedly needs Git: %v", err)
		}
	}
}

func TestParseMintRemoteHEAD(t *testing.T) {
	good := []byte("ref: refs/heads/trunk\tHEAD\n0123456789012345678901234567890123456789\tHEAD\n")
	if branch, err := parseMintRemoteHEAD(good); err != nil || branch != "trunk" {
		t.Fatalf("%q %v", branch, err)
	}
	for _, bad := range []string{
		"", "0123456789012345678901234567890123456789\tHEAD\n",
		"ref: refs/tags/v1.0.0\tHEAD\n", "ref: refs/heads/trunk\tOTHER\n",
		"ref: refs/heads/trunk\tHEAD\nref: refs/heads/other\tHEAD\n",
		"ref: refs/heads/trunk\tHEAD\nref: refs/heads/trunk\tHEAD\n",
		"ref: refs/heads/../unsafe\tHEAD\n", "ref: refs/heads/trunk extra\tHEAD\n",
	} {
		if _, err := parseMintRemoteHEAD([]byte(bad)); err == nil || !strings.Contains(err.Error(), "INCOMPLETE") {
			t.Errorf("accepted ambiguous/missing metadata %q: %v", bad, err)
		}
	}
}

func TestMintRemoteOutputBound(t *testing.T) {
	var buffer mintRemoteBuffer
	if _, err := buffer.Write(make([]byte, 64*1024+1)); err == nil {
		t.Fatal("unbounded remote output accepted")
	}
	// Match os/exec's io.Copy path without a source WriterTo optimization. An
	// embedded bytes.Buffer would promote ReadFrom and bypass the Write limit.
	reader := struct{ io.Reader }{strings.NewReader(strings.Repeat("x", 64*1024+1))}
	written, err := io.Copy(&buffer, reader)
	if err == nil || written > 64*1024 || len(buffer.Bytes()) > 64*1024 {
		t.Fatalf("io.Copy bypassed output bound: written=%d retained=%d error=%v", written, len(buffer.Bytes()), err)
	}
}

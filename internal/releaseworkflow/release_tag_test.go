package releaseworkflow_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseNextTag(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, repo string) string
		want  []string
	}{
		{
			name: "starts v3 at zero",
			setup: func(t *testing.T, repo string) string {
				return commit(t, repo, "initial")
			},
			want: []string{"next_tag=v3.0.0", "mode=create"},
		},
		{
			name: "ignores v2 history",
			setup: func(t *testing.T, repo string) string {
				head := commit(t, repo, "initial")
				tag(t, repo, "v2.0.6", head)
				return head
			},
			want: []string{"next_tag=v3.0.0", "mode=create"},
		},
		{
			name: "increments latest v3 patch",
			setup: func(t *testing.T, repo string) string {
				first := commit(t, repo, "first")
				tag(t, repo, "v3.2.7", first)
				return commit(t, repo, "second")
			},
			want: []string{"next_tag=v3.2.8", "mode=create"},
		},
		{
			name: "reuses same sha tag",
			setup: func(t *testing.T, repo string) string {
				head := commit(t, repo, "initial")
				tag(t, repo, "v3.0.0", head)
				return head
			},
			want: []string{"next_tag=v3.0.0", "mode=reuse"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := initRepository(t)
			head := tt.setup(t, repo)
			output := runSelector(t, repo, head, true)
			for _, want := range append(tt.want, "head_sha="+head) {
				if !strings.Contains(output, want) {
					t.Errorf("selector output missing %q:\n%s", want, output)
				}
			}
		})
	}
}

func TestReleaseNextTagRejectsAmbiguousOrStaleHeadTags(t *testing.T) {
	repo := initRepository(t)
	first := commit(t, repo, "first")
	tag(t, repo, "v3.0.0", first)
	tag(t, repo, "v3.0.1", first)
	output := runSelector(t, repo, first, false)
	if !strings.Contains(output, "Multiple v3 release tags") {
		t.Fatalf("selector did not reject ambiguous tags:\n%s", output)
	}

	repo = initRepository(t)
	first = commit(t, repo, "first")
	tag(t, repo, "v3.0.0", first)
	second := commit(t, repo, "second")
	tag(t, repo, "v3.0.1", second)
	output = runSelector(t, repo, first, false)
	if !strings.Contains(output, "is not the latest v3 release tag") {
		t.Fatalf("selector did not reject stale head tag:\n%s", output)
	}
}

func TestReleaseWorkflowOrdersAndRecoversSafely(t *testing.T) {
	workflow := readRepositoryFile(t, ".github/workflows/release-tag-main.yml")
	for _, want := range []string{
		"queue: max",
		"paths-ignore:",
		"- .kit.yaml",
		".github/scripts/release-next-tag.sh HEAD",
		"Reusing ${NEXT_TAG}",
		"refusing to move it",
		"needs.prepare-release.outputs.next_tag != ''",
		"GITHUB_TOKEN do not trigger a second",
		"gh release upload",
		"gh release create",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("release workflow missing %q", want)
		}
	}
	quality := strings.Index(workflow, "Run release quality gates")
	create := strings.Index(workflow, "Create and push release tag")
	if quality < 0 || create < 0 || quality > create {
		t.Error("release quality gates must run before tag creation")
	}
}

func TestExternalReleasePublisherAcceptsOnlyV3Tags(t *testing.T) {
	workflow := readRepositoryFile(t, ".github/workflows/release-publish.yml")
	for _, want := range []string{`- "v3.*"`, `^v3\.[0-9]+\.[0-9]+$`, "queue: max"} {
		if !strings.Contains(workflow, want) {
			t.Errorf("external publisher missing %q", want)
		}
	}
}

func initRepository(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init", "-b", "main")
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "user.email", "test@example.com")
	return repo
}

func commit(t *testing.T, repo, message string) string {
	t.Helper()
	path := filepath.Join(repo, "history")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(message + "\n"); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "history")
	runGit(t, repo, "commit", "-m", message)
	return strings.TrimSpace(runGit(t, repo, "rev-parse", "HEAD"))
}

func tag(t *testing.T, repo, name, ref string) {
	t.Helper()
	runGit(t, repo, "tag", "-a", name, ref, "-m", name)
}

func runSelector(t *testing.T, repo, head string, wantSuccess bool) string {
	t.Helper()
	script, err := filepath.Abs(filepath.Join("..", "..", ".github", "scripts", "release-next-tag.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(script, head)
	cmd.Dir = repo
	output, runErr := cmd.CombinedOutput()
	if wantSuccess && runErr != nil {
		t.Fatalf("selector failed: %v\n%s", runErr, output)
	}
	if !wantSuccess && runErr == nil {
		t.Fatalf("selector unexpectedly succeeded:\n%s", output)
	}
	return string(output)
}

func runGit(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

func readRepositoryFile(t *testing.T, relativePath string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(relativePath)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

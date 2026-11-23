package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise the release executable behind the composite action with a local
// Git fixture. The default suite stays offline and does not require Mint.
func TestMintPipelinePreparationActualMixedHistory(t *testing.T) {
	mint := os.Getenv("KIT_TEST_MINT")
	if !filepath.IsAbs(mint) {
		t.Skip("set KIT_TEST_MINT to the absolute path of the release executable")
	}
	root := mintPreparationRepository(t)
	git := func(args ...string) string { return mintPreparationGit(t, root, args...) }
	git("commit", "--allow-empty", "-m", "chore: baseline")
	git("tag", "v1.0.0")
	git("commit", "--allow-empty", "-m", "fix(GH-42): retain historical issue", "-m", "Closes #42")
	githubSHA := git("rev-parse", "HEAD")
	git("commit", "--allow-empty", "-m", "feat(ENG-123): :sparkles: add invitations", "-m", "Private ticket detail stays out of notes.\n\nRefs: https://linear.app/acme/issue/ENG-123/invitations")
	linearSHA := git("rev-parse", "HEAD")
	tagsBefore := git("show-ref", "--tags")
	artifactDir := t.TempDir()
	bindings := map[string]string{
		"github.sha": linearSHA, "github.repository_owner": "acme",
		"github.event.repository.name": "app", "runner.temp": artifactDir,
	}
	steps := mintReleasePreparationSteps()
	resolveInputs := mintPreparationInputs(t, steps[0], bindings)
	if resolveInputs["command"] != "release-resolve" {
		t.Fatalf("unexpected command: %v", resolveInputs)
	}
	resolve := func() map[string]string {
		outputPath := filepath.Join(t.TempDir(), "release-output")
		runMintPreparation(t, root, mint, "release", "resolve", "--commitish", resolveInputs["commitish"], "--github-output", outputPath)
		outputs := map[string]string{}
		for _, line := range strings.Split(readFile(t, outputPath), "\n") {
			if key, value, ok := strings.Cut(line, "="); ok && !strings.Contains(key, "<<") {
				outputs[key] = value
			}
		}
		return outputs
	}
	outputs := resolve()
	for key, want := range map[string]string{
		"base_tag": "v1.0.0", "version_tag": "v1.1.0", "version_bump": "minor",
		"commit_count": "2", "needs_git_tag": "true", "target_sha": linearSHA,
	} {
		if outputs[key] != want {
			t.Fatalf("unexpected release range: %+v", outputs)
		}
	}
	for key, value := range outputs {
		bindings["steps.kit_mint_release.outputs."+key] = value
	}
	if !mintPreparationGuard(t, steps[1], outputs) || !mintPreparationGuard(t, steps[2], outputs) {
		t.Fatal("new release must generate and retain its changelog")
	}
	inputs := mintPreparationInputs(t, steps[1], bindings)
	if inputs["command"] != "changelog" {
		t.Fatal(inputs)
	}
	runMintPreparation(t, root, mint,
		"--prev-tag", inputs["prev-tag"], "--current-tag", inputs["current-tag"],
		"--current-ref", inputs["current-ref"], "--owner", inputs["owner"],
		"--repo", inputs["repo"], "--output", inputs["output"])
	artifact := mintPreparationInputs(t, steps[2], bindings)
	if artifact["path"] != inputs["output"] || artifact["name"] != "mint-changelog-v1.1.0" || artifact["if-no-files-found"] != "error" {
		t.Fatalf("retention does not use generated artifact: %v", artifact)
	}
	data, err := os.ReadFile(artifact["path"])
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## [1.1.0](https://github.com/acme/app/releases/tag/v1.1.0)",
		"[ENG-123](https://linear.app/acme/issue/ENG-123/invitations)",
		"[#42](https://github.com/acme/app/issues/42)",
		"https://github.com/acme/app/commit/" + githubSHA[:7],
		"https://github.com/acme/app/commit/" + linearSHA[:7],
	} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("missing %q in %s", want, data)
		}
	}
	if strings.Contains(string(data), "Private ticket detail") || strings.Contains(string(data), "baseline") {
		t.Fatalf("private body or excluded history rendered: %s", data)
	}
	if got := git("show-ref", "--tags"); got != tagsBefore {
		t.Fatal("preparation published a tag")
	}
	if got := git("status", "--porcelain"); got != "" {
		t.Fatalf("preparation changed repository files: %s", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".version")); !os.IsNotExist(err) {
		t.Fatal("preparation wrote a runtime version file")
	}

	// A retry after publication must not generate a duplicate version block.
	// Only this fixture creates the tag; preparation itself remains read-only.
	git("tag", outputs["version_tag"])
	tagged := resolve()
	if tagged["needs_git_tag"] != "false" || tagged["version_bump"] != "already-tagged" ||
		mintPreparationGuard(t, steps[1], tagged) || mintPreparationGuard(t, steps[2], tagged) {
		t.Fatalf("tagged retry must skip generation and upload: %+v", tagged)
	}
	retained, err := os.ReadFile(artifact["path"])
	if err != nil || !bytes.Equal(data, retained) {
		t.Fatal("tagged retry changed retained artifact")
	}
}

func mintPreparationRepository(t *testing.T) string {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_NAME", "Mint Fixture")
	t.Setenv("GIT_AUTHOR_EMAIL", "fixture@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "Mint Fixture")
	t.Setenv("GIT_COMMITTER_EMAIL", "fixture@example.invalid")
	t.Setenv("GIT_AUTHOR_DATE", "2026-01-02T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-01-02T00:00:00Z")
	root := t.TempDir()
	mintPreparationGit(t, root, "init", "--initial-branch=main")
	return root
}

func mintPreparationGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	args = append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false", "-c", "core.hooksPath=" + os.DevNull}, args...)
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func runMintPreparation(t *testing.T, root, mint string, args ...string) {
	t.Helper()
	cmd := exec.Command(mint, args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("mint %v: %v: %s", args, err, out)
	}
}

func mintPreparationInputs(t *testing.T, raw any, bindings map[string]string) map[string]string {
	t.Helper()
	step := raw.(map[string]any)
	values := step["with"].(map[string]any)
	result := map[string]string{}
	for key, raw := range values {
		value := raw.(string)
		for name, replacement := range bindings {
			value = strings.ReplaceAll(value, "${{ "+name+" }}", replacement)
		}
		if strings.Contains(value, "${{") {
			t.Fatalf("unresolved renderer binding %s=%s", key, value)
		}
		result[key] = value
	}
	return result
}

func mintPreparationGuard(t *testing.T, raw any, outputs map[string]string) bool {
	t.Helper()
	// Evaluate the deliberately narrow generated predicate against actual Mint
	// outputs. Fail if the renderer changes rather than pretending to run Actions.
	if raw.(map[string]any)["if"] != "${{ steps.kit_mint_release.outputs.needs_git_tag == 'true' && steps.kit_mint_release.outputs.commit_count != '0' }}" {
		t.Fatal("unsupported preparation guard")
	}
	return outputs["needs_git_tag"] == "true" && outputs["commit_count"] != "0"
}

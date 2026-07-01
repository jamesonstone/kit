package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunInitRefresh_UsesOnlyWorkflowBadgeForPrivateRepository(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	withReadmeVisibility(t, "acme/private-widget", "PRIVATE")
	cfg := config.Default()
	cfg.GitHub.Repository = "acme/private-widget"
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, ".github", "workflows", "ci.yml"), "name: CI\n")
	writeFile(t, filepath.Join(tempDir, readmePath), "# Private Widget\n\nPrivate Widget handles internal jobs.\n\n"+readmeBadgeStart+"\n[![Last commit](https://img.shields.io/github/last-commit/acme/private-widget)](https://github.com/acme/private-widget/commits) [![Open issues](https://img.shields.io/github/issues/acme/private-widget)](https://github.com/acme/private-widget/issues) [![Pull requests](https://img.shields.io/github/issues-pr/acme/private-widget)](https://github.com/acme/private-widget/pulls) [![Release](https://img.shields.io/github/v/release/acme/private-widget)](https://github.com/acme/private-widget/releases)\n"+readmeBadgeEnd+"\n\n## Setup\n\nRun it.\n")

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{readmePath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, readmePath))
	for _, unexpected := range []string{
		"img.shields.io/github/last-commit/acme/private-widget",
		"img.shields.io/github/issues/acme/private-widget",
		"img.shields.io/github/issues-pr/acme/private-widget",
		"img.shields.io/github/v/release/acme/private-widget",
	} {
		if strings.Contains(content, unexpected) {
			t.Fatalf("did not expect private README to contain public Shields badge %q, got:\n%s", unexpected, content)
		}
	}
	for _, check := range []string{
		readmeBadgeStart,
		"github.com/acme/private-widget/actions/workflows/ci.yml/badge.svg",
		readmeBadgeEnd,
		"## Setup",
		"## Maintainers",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected private README to contain %q, got:\n%s", check, content)
		}
	}
	if got := lastReadmeH2(content); got != "## Maintainers" {
		t.Fatalf("last README H2 = %q, want ## Maintainers\n%s", got, content)
	}
}

func TestRunInitRefresh_RemovesManagedReadmeBadgesForPrivateRepositoryWithoutWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	withReadmeVisibility(t, "acme/private-widget", "PRIVATE")
	cfg := config.Default()
	cfg.GitHub.Repository = "acme/private-widget"
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, readmePath), "# Private Widget\n\nPrivate Widget handles internal jobs.\n\n"+readmeBadgeStart+"\n[![Last commit](https://img.shields.io/github/last-commit/acme/private-widget)](https://github.com/acme/private-widget/commits)\n"+readmeBadgeEnd+"\n\n## Setup\n\nRun it.\n")

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{readmePath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})

	content := readFile(t, filepath.Join(tempDir, readmePath))
	for _, unexpected := range []string{
		readmeBadgeStart,
		readmeBadgeEnd,
		"img.shields.io/github/last-commit/acme/private-widget",
	} {
		if strings.Contains(content, unexpected) {
			t.Fatalf("did not expect private README without workflow to contain %q, got:\n%s", unexpected, content)
		}
	}
	for _, check := range []string{"# Private Widget", "## Setup", "## Maintainers"} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected README to preserve %q, got:\n%s", check, content)
		}
	}
}

func withReadmeCommandRunner(t *testing.T, runner ciCommandRunner) {
	t.Helper()
	previous := readmeCommandRunner
	readmeCommandRunner = runner
	t.Cleanup(func() {
		readmeCommandRunner = previous
	})
}

func withReadmeVisibility(t *testing.T, repository string, visibility string) {
	t.Helper()
	withReadmeCommandRunner(t, fakeReadmeRunner{
		outputs: map[string][]byte{
			"gh repo view " + repository + " --json visibility -q .visibility": []byte(visibility + "\n"),
		},
	})
}

type fakeReadmeRunner struct {
	outputs map[string][]byte
}

func (f fakeReadmeRunner) Output(_ string, name string, args ...string) ([]byte, error) {
	key := name
	if len(args) > 0 {
		key += " " + strings.Join(args, " ")
	}
	if output, ok := f.outputs[key]; ok {
		return output, nil
	}
	return nil, fmt.Errorf("unexpected command: %s", key)
}

func (f fakeReadmeRunner) OutputAllowError(dir string, name string, args ...string) ([]byte, error) {
	return f.Output(dir, name, args...)
}

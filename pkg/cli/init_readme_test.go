package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunInitRefresh_AddsManagedReadmeBadgesAndStaysIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	cfg := config.Default()
	cfg.GitHub.Repository = "acme/widget"
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, ".github", "workflows", "ci.yml"), "name: CI\n")
	writeFile(t, filepath.Join(tempDir, readmePath), "```text\nWIDGET\n\n                         useful widget service\n```\n\nWidget runs useful jobs for the Acme platform.\n\n## Install\n\nRun it.\n")

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

	first := readFile(t, filepath.Join(tempDir, readmePath))
	for _, check := range []string{
		readmeBadgeStart,
		"img.shields.io/github/last-commit/acme/widget",
		"img.shields.io/github/issues/acme/widget",
		"img.shields.io/github/issues-pr/acme/widget",
		"github.com/acme/widget/actions/workflows/ci.yml/badge.svg",
		"img.shields.io/github/v/release/acme/widget",
		readmeBadgeEnd,
	} {
		if !strings.Contains(first, check) {
			t.Fatalf("expected README to contain %q, got:\n%s", check, first)
		}
	}
	if strings.Contains(strings.ToLower(first), "license") {
		t.Fatalf("README badges should not include a license badge, got:\n%s", first)
	}
	if !strings.Contains(first, "Widget runs useful jobs for the Acme platform.\n\n"+readmeBadgeStart+"\n") {
		t.Fatalf("expected badge block after opening paragraph, got:\n%s", first)
	}
	if !strings.Contains(first, "## Maintainers\n\nMaintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).") {
		t.Fatalf("expected managed maintainers section, got:\n%s", first)
	}
	if got := lastReadmeH2(first); got != "## Maintainers" {
		t.Fatalf("last README H2 = %q, want ## Maintainers\n%s", got, first)
	}

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{readmePath}

		_ = captureStdout(t, func() {
			if err := runInit(initCmd, nil); err != nil {
				t.Fatalf("second runInit() error = %v", err)
			}
		})
	})

	second := readFile(t, filepath.Join(tempDir, readmePath))
	if second != first {
		t.Fatalf("README refresh was not idempotent:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}

func TestRunInitRefresh_CreatesReadmeStarterWithManagedBadges(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	cfg := config.Default()
	cfg.GitHub.Repository = "acme/background-worker"
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

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
	for _, check := range []string{
		"```text\nBACKGROUND WORKER",
		readmeStarterTagline,
		"Background Worker is a Kit-managed project.",
		"img.shields.io/github/last-commit/acme/background-worker",
		"img.shields.io/github/v/release/acme/background-worker",
		"## Maintainers",
		"`jamesonstone`",
	} {
		if !strings.Contains(content, check) {
			t.Fatalf("expected README starter to contain %q, got:\n%s", check, content)
		}
	}
	if got := lastReadmeH2(content); got != "## Maintainers" {
		t.Fatalf("last README H2 = %q, want ## Maintainers\n%s", got, content)
	}
}

func TestRunInitRefresh_ReplacesMaintainerSectionAndKeepsMaintainersLast(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	cfg := config.Default()
	cfg.GitHub.Repository = "acme/widget"
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, readmePath), "# Widget\n\nWidget does work.\n\n## Maintainer\n\nOld maintainer copy.\n\n## License\n\nMIT\n")

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
	for _, unexpected := range []string{"## Maintainer\n", "Old maintainer copy"} {
		if strings.Contains(content, unexpected) {
			t.Fatalf("expected stale maintainer content %q to be removed, got:\n%s", unexpected, content)
		}
	}
	if !strings.Contains(content, "## License\n\nMIT\n\n## Maintainers\n\nMaintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).") {
		t.Fatalf("expected managed Maintainers section after License, got:\n%s", content)
	}
	if got := lastReadmeH2(content); got != "## Maintainers" {
		t.Fatalf("last README H2 = %q, want ## Maintainers\n%s", got, content)
	}
}

func TestRunInitRefresh_AddsMaintainersWithoutGitHubRemote(t *testing.T) {
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	if err := config.Save(tempDir, config.Default()); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	writeFile(t, filepath.Join(tempDir, readmePath), "# Local Tool\n\nLocal Tool is a local-only utility.\n")

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
	if strings.Contains(content, readmeBadgeStart) {
		t.Fatalf("did not expect badge block without GitHub repo, got:\n%s", content)
	}
	if !strings.Contains(content, "## Maintainers\n\nMaintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone) (`jamesonstone`).") {
		t.Fatalf("expected managed Maintainers section, got:\n%s", content)
	}
	if got := lastReadmeH2(content); got != "## Maintainers" {
		t.Fatalf("last README H2 = %q, want ## Maintainers\n%s", got, content)
	}
}

func lastReadmeH2(content string) string {
	last := ""
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### ") {
			last = trimmed
		}
	}
	return last
}

package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func runReadmeMaintainersRefresh(t *testing.T, repository, ownerType, readme string) (string, string) {
	t.Helper()
	tempDir := t.TempDir()
	setupInitHome(t)
	setWorkingDirectory(t, tempDir)
	withReadmeGitHub(t, repository, "PUBLIC", ownerType)
	cfg := config.Default()
	cfg.GitHub.Repository = repository
	if err := config.Save(tempDir, cfg); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	path := filepath.Join(tempDir, readmePath)
	writeFile(t, path, readme)

	withInitFlags(t, func() {
		initRefresh = true
		initOutputOnly = true
		initRefreshFiles = []string{readmePath}
		_ = captureStdout(t, func() {
			if err := runInitForTest(initCmd, nil); err != nil {
				t.Fatalf("runInit() error = %v", err)
			}
		})
	})
	return path, readFile(t, path)
}

func TestRunInitRefresh_UserOwnedRepoMaintainersDoNotSayTeam(t *testing.T) {
	_, content := runReadmeMaintainersRefresh(t, "octocat/widget", "User", "# Widget\n\nWidget does work.\n")

	if !strings.Contains(content, "## Maintainers\n\nMaintained by [octocat](https://github.com/octocat).\n") {
		t.Fatalf("expected user-owned maintainers wording, got:\n%s", content)
	}
	if strings.Contains(content, " team.") {
		t.Fatalf("user-owned repository must not be credited to a team, got:\n%s", content)
	}
}

func TestRunInitRefresh_UnknownOwnerTypeMaintainersDoNotSayTeam(t *testing.T) {
	_, content := runReadmeMaintainersRefresh(t, "octocat/widget", "", "# Widget\n\nWidget does work.\n")

	if !strings.Contains(content, "Maintained by [octocat](https://github.com/octocat).\n") {
		t.Fatalf("expected neutral maintainers wording when owner type is unknown, got:\n%s", content)
	}
}

func TestRunInitRefresh_ReplacesGeneratedTeamDefaultForUserOwner(t *testing.T) {
	readme := "# Widget\n\nWidget does work.\n\n## Maintainers\n\nMaintained by the [octocat](https://github.com/octocat) team.\n"
	_, content := runReadmeMaintainersRefresh(t, "octocat/widget", "User", readme)

	if strings.Contains(content, " team.") {
		t.Fatalf("expected Kit-generated team default to be corrected, got:\n%s", content)
	}
	if !strings.Contains(content, "Maintained by [octocat](https://github.com/octocat).\n") {
		t.Fatalf("expected user-owned maintainers wording, got:\n%s", content)
	}
}

func TestRunInitRefresh_PreservesHumanWrittenMaintainers(t *testing.T) {
	readme := readmeWithBadges("jamesonstone/rungrid") +
		"\n## Maintainers\n\nMaintained with 🪖 and ❤️ by [Jameson](https://github.com/jamesonstone)\n(`jamesonstone`).\n"
	path, first := runReadmeMaintainersRefresh(t, "jamesonstone/rungrid", "User", readme)

	if first != readme {
		t.Fatalf("expected human-written maintainers README to stay unchanged.\nwant:\n%s\ngot:\n%s", readme, first)
	}
	if strings.Contains(first, "team") {
		t.Fatalf("human-written maintainers line was replaced, got:\n%s", first)
	}
	if got := readFile(t, path); got != readme {
		t.Fatalf("README changed on disk:\n%s", got)
	}
}

func TestRunInitRefresh_PreservesCustomMaintainerSectionAndKeepsMaintainersLast(t *testing.T) {
	readme := "# Widget\n\nWidget does work.\n\n## Maintainer\n\nOld maintainer copy.\n\n### Emeritus\n\n- Someone\n\n## License\n\nMIT\n"
	_, content := runReadmeMaintainersRefresh(t, "acme/widget", readmeOwnerTypeOrganization, readme)

	if strings.Contains(content, "## Maintainer\n") {
		t.Fatalf("expected singular heading to be normalized, got:\n%s", content)
	}
	if !strings.Contains(content, "## License\n\nMIT\n\n## Maintainers\n\nOld maintainer copy.\n\n### Emeritus\n\n- Someone\n") {
		t.Fatalf("expected custom maintainers content preserved after License, got:\n%s", content)
	}
	if strings.Contains(content, "Maintained by the [acme]") {
		t.Fatalf("custom maintainers content was replaced by the default, got:\n%s", content)
	}
	if got := lastReadmeH2(content); got != "## Maintainers" {
		t.Fatalf("last README H2 = %q, want ## Maintainers\n%s", got, content)
	}
}

func readmeWithBadges(repository string) string {
	return "# Rungrid\n\nRungrid runs grids.\n\n" +
		managedReadmeBadgeBlock(repository, "", "PUBLIC")
}

func TestDefaultReadmeMaintainersSectionWording(t *testing.T) {
	cases := map[string]string{
		"Organization": "## Maintainers\n\nMaintained by the [acme](https://github.com/acme) team.\n",
		"User":         "## Maintainers\n\nMaintained by [acme](https://github.com/acme).\n",
		"":             "## Maintainers\n\nMaintained by [acme](https://github.com/acme).\n",
	}
	for ownerType, want := range cases {
		if got := defaultReadmeMaintainersSection("acme", ownerType); got != want {
			t.Errorf("defaultReadmeMaintainersSection(acme, %q) = %q, want %q", ownerType, got, want)
		}
	}
	if got := defaultReadmeMaintainersSection("", "User"); got != readmeMaintainersFallback {
		t.Errorf("empty owner = %q, want fallback", got)
	}
}

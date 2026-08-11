package releaseprompt

import "testing"

func TestCleanGitHubPathTrimsSlashBeforeGitSuffix(t *testing.T) {
	if got := cleanGitHubPath("/acme/service.git/"); got != "acme/service" {
		t.Fatalf("cleanGitHubPath() = %q, want %q", got, "acme/service")
	}
}

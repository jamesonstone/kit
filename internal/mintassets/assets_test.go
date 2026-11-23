package mintassets

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestResolveReleaseUsesOnlyAStableTag(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("workflow resolver requires Bash")
	}
	for _, tc := range []struct{ name, tag, status, want string }{
		{"current stable", "v1.9.1", "0", "ref=refs/tags/v1.9.1\n"},
		{"future stable", "v2.0.0", "0", "ref=refs/tags/v2.0.0\n"},
		{"prerelease", "v2.0.0-rc.1", "0", ""},
		{"unexpected ref", "main", "0", ""},
		{"output injection", "v2.0.0\nref=main", "0", ""},
		{"no release", "", "0", ""},
		{"API failure", "", "1", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			gh := `#!/bin/sh
test "$1 $2 $3 $4 $5" = 'api --hostname github.com repos/jamesonstone/mint/releases/latest --jq' || exit 2
printf '%s\n' "$MOCK_TAG"
exit "$MOCK_STATUS"
`
			if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(gh), 0o700); err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(dir, "output")
			if err := os.WriteFile(output, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(bash, "-c", ResolveRelease)
			cmd.Env = append(cmd.Environ(), "PATH="+dir, "BASH_ENV=", "GITHUB_OUTPUT="+output, "MOCK_TAG="+tc.tag, "MOCK_STATUS="+tc.status)
			if log, err := cmd.CombinedOutput(); (err != nil) != (tc.want == "") {
				t.Fatalf("resolver: %v, %s", err, log)
			}
			data, err := os.ReadFile(output)
			if err != nil || string(data) != tc.want {
				t.Fatalf("output: %q, %v", data, err)
			}
		})
	}
}

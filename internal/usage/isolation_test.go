package usage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Store tests below exercise real recording inside temporary homes.
	recordingInTests = true
	os.Exit(m.Run())
}

func recordedEvents(t *testing.T, home string) int {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(home, ".config", "kit", "usage", "*.jsonl"))
	return len(matches)
}

func TestRecordSkipsDevelopmentBuildsAndOptOut(t *testing.T) {
	for name, tc := range map[string]struct {
		version string
		env     string
		want    int
	}{
		"released build":    {version: "v3.1.0", want: 1},
		"development build": {version: "dev", want: 0},
		"unversioned build": {version: "", want: 0},
		"env opt-out":       {version: "v3.1.0", env: "1", want: 0},
	} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_CONFIG_HOME", "")
			t.Setenv(DisableEnv, tc.env)
			if err := Record(RecordInput{Command: "check", Version: tc.version, Elapsed: time.Millisecond}); err != nil {
				t.Fatalf("Record() error = %v", err)
			}
			if got := recordedEvents(t, home); got != tc.want {
				t.Fatalf("recorded shards = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRecordIsSuppressedUnderOtherPackagesTests(t *testing.T) {
	recordingInTests = false
	t.Cleanup(func() { recordingInTests = true })
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(DisableEnv, "")
	if err := Record(RecordInput{Command: "check", Version: "v3.1.0"}); err != nil {
		t.Fatal(err)
	}
	if got := recordedEvents(t, home); got != 0 {
		t.Fatalf("test binaries recorded %d shards into HOME", got)
	}
}

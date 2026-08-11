package usage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRecordStoresMinimalPrivateEventAndReport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectRoot := filepath.Join(t.TempDir(), "secret-repository-name")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(projectRoot, config.Default()); err != nil {
		t.Fatal(err)
	}

	err := Record(RecordInput{
		Command: "status", Version: "v2.0.0", ExitCode: 0,
		Elapsed: 20 * time.Millisecond, ProjectRoot: projectRoot, Interactive: true,
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	dir, _ := Directory()
	shards, err := listShards(dir)
	if err != nil || len(shards) != 1 {
		t.Fatalf("listShards() = %v, %v", shards, err)
	}
	raw, err := os.ReadFile(shards[0].path)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{projectRoot, "secret-repository-name", "--token", "cwd"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("usage event leaked %q: %s", forbidden, raw)
		}
	}

	report, err := BuildReport(Filter{Since: time.Now().Add(-time.Hour)})
	if err != nil {
		t.Fatalf("BuildReport() error = %v", err)
	}
	if report.TotalCalls != 1 || report.Successes != 1 || len(report.Commands) != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if report.Commands[0].Command != "status" || report.Commands[0].Interactive != 1 {
		t.Fatalf("unexpected command summary: %#v", report.Commands[0])
	}
}

func TestGlobalDisableSuppressesProjectEnable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	projectRoot := t.TempDir()
	project := config.Default()
	project.Usage = &config.UsageConfig{Enabled: boolPointer(true)}
	if err := config.Save(projectRoot, project); err != nil {
		t.Fatal(err)
	}
	globalPath, _ := config.GlobalConfigPath()
	if err := config.UpdateUsageEnabled(globalPath, false); err != nil {
		t.Fatal(err)
	}
	settings, err := EffectiveSettings(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if settings.Enabled || settings.ProjectState != "suppressed-by-global" {
		t.Fatalf("unexpected settings: %#v", settings)
	}
	if err := Record(RecordInput{Command: "status", ProjectRoot: projectRoot}); err != nil {
		t.Fatal(err)
	}
	dir, _ := Directory()
	shards, err := listShards(dir)
	if err != nil || len(shards) != 0 {
		t.Fatalf("disabled collection wrote shards: %v, %v", shards, err)
	}
}

func TestUsageCommandsDoNotRecordThemselves(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := Record(RecordInput{Command: "usage report"}); err != nil {
		t.Fatal(err)
	}
	dir, _ := Directory()
	shards, err := listShards(dir)
	if err != nil || len(shards) != 0 {
		t.Fatalf("usage command wrote shards: %v, %v", shards, err)
	}
}

func boolPointer(value bool) *bool { return &value }

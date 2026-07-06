package improve

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSuiteSelectsDefaultTasks(t *testing.T) {
	root := fixtureProjectRoot(t)
	suite, tasks, err := LoadSuite(root, "default")
	if err != nil {
		t.Fatalf("LoadSuite() error = %v", err)
	}
	if suite.ID != "default" {
		t.Fatalf("suite.ID = %q, want default", suite.ID)
	}
	if len(tasks) != 8 {
		t.Fatalf("len(tasks) = %d, want 8", len(tasks))
	}
}

func TestSelectTasksExcludesHeldOutWhenTagsOverlap(t *testing.T) {
	suite := Suite{
		HeldIn: TaskSelector{IncludeTags: []string{"shared", "held-in"}},
		HeldOut: TaskSelector{
			IncludeTags:        []string{"shared"},
			HiddenFromProposer: true,
		},
	}
	tasks := []Task{
		{ID: "held-in", RegressionTags: []string{"held-in"}},
		{ID: "overlap", RegressionTags: []string{"shared"}},
	}
	selected := selectTasks(suite, tasks)
	if len(selected) != 1 || selected[0].ID != "held-in" {
		t.Fatalf("selected = %#v, want only held-in", selected)
	}
}

func TestSelectTasksWithEmptyHeldInStillExcludesHeldOut(t *testing.T) {
	suite := Suite{
		HeldOut: TaskSelector{
			IncludeTags:        []string{"held-out"},
			HiddenFromProposer: true,
		},
	}
	tasks := []Task{
		{ID: "candidate", RegressionTags: []string{"review-loop"}},
		{ID: "hidden", RegressionTags: []string{"held-out"}},
	}
	selected := selectTasks(suite, tasks)
	if len(selected) != 1 || selected[0].ID != "candidate" {
		t.Fatalf("selected = %#v, want only candidate", selected)
	}
}

func TestRunWritesTraceArtifacts(t *testing.T) {
	root := fixtureProjectRoot(t)
	kitBinary := fakeKitBinary(t)
	manifest, err := Run(context.Background(), RunOptions{
		ProjectRoot: root,
		SuiteName:   "default",
		KitBinary:   kitBinary,
		KitVersion:  "test",
		GitCommit:   "abc123",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if manifest.Status != "pass" {
		t.Fatalf("manifest.Status = %q, want pass", manifest.Status)
	}
	if len(manifest.Traces) != 8 {
		t.Fatalf("len(manifest.Traces) = %d, want 8", len(manifest.Traces))
	}
	if got := manifest.Traces[0].Commands[0].Argv[0]; got != kitBinary {
		t.Fatalf("kit placeholder resolved to %q, want %q", got, kitBinary)
	}
	if _, err := os.Stat(filepath.Join(manifest.RunDir, "run.json")); err != nil {
		t.Fatalf("run.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ArtifactDir, "latest")); err != nil {
		if _, fallbackErr := os.Stat(filepath.Join(root, ArtifactDir, "latest.txt")); fallbackErr != nil {
			t.Fatalf("latest marker missing: symlink=%v fallback=%v", err, fallbackErr)
		}
	}
}

func TestMineProposeValidateReport(t *testing.T) {
	root := fixtureProjectRoot(t)
	manifest, err := Run(context.Background(), RunOptions{ProjectRoot: root, SuiteName: "default", KitBinary: fakeKitBinary(t)})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	report, err := Mine(root, manifest.RunDir)
	if err != nil {
		t.Fatalf("Mine() error = %v", err)
	}
	if len(report.Clusters) != 0 {
		t.Fatalf("expected passing suite to produce no clusters, got %#v", report.Clusters)
	}
	if err := writeJSON(filepath.Join(manifest.RunDir, "weakness-report.json"), WeaknessReport{
		SchemaVersion: SchemaVersion,
		Kind:          "weakness_report",
		SourceDir:     manifest.RunDir,
		Clusters: []WeaknessCluster{{
			Signature:            "github-delivery:generic-pr-default-used",
			AffectedTasks:        []string{"github-pr-delivery-contract"},
			RepresentativeTraces: []string{"github-pr-delivery-contract:1"},
			Confidence:           "high",
		}},
	}); err != nil {
		t.Fatalf("write weakness report: %v", err)
	}
	candidates, err := Propose(root, manifest.RunDir, 1)
	if err != nil {
		t.Fatalf("Propose() error = %v", err)
	}
	if len(candidates) != 1 || candidates[0].Status != "proposed" {
		t.Fatalf("unexpected candidates: %#v", candidates)
	}
	scorecard, err := Validate(root, filepath.Join(manifest.RunDir, "candidates", "candidate-001", "candidate.json"))
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if scorecard.Acceptance != "accepted-for-review" {
		t.Fatalf("scorecard.Acceptance = %q", scorecard.Acceptance)
	}
	markdown, err := Report(root, manifest.RunDir)
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if markdown == "" {
		t.Fatalf("expected report markdown")
	}
}

func TestProposeMinesWhenWeaknessReportIsMissing(t *testing.T) {
	root := fixtureProjectRoot(t)
	manifest, err := Run(context.Background(), RunOptions{ProjectRoot: root, SuiteName: "default", KitBinary: fakeKitBinary(t)})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	candidates, err := Propose(root, manifest.RunDir, 3)
	if err != nil {
		t.Fatalf("Propose() error = %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("expected passing suite to produce no candidates, got %#v", candidates)
	}
	if _, err := os.Stat(filepath.Join(manifest.RunDir, "weakness-report.json")); err != nil {
		t.Fatalf("weakness-report.json missing after Propose(): %v", err)
	}
}

func TestProposeReturnsCorruptWeaknessReportError(t *testing.T) {
	root := fixtureProjectRoot(t)
	manifest, err := Run(context.Background(), RunOptions{ProjectRoot: root, SuiteName: "default", KitBinary: fakeKitBinary(t)})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	reportPath := filepath.Join(manifest.RunDir, "weakness-report.json")
	invalidJSON := []byte("{not-json")
	if err := os.WriteFile(reportPath, invalidJSON, 0o644); err != nil {
		t.Fatalf("write corrupt weakness report: %v", err)
	}
	if _, err := Propose(root, manifest.RunDir, 3); err == nil {
		t.Fatalf("Propose() error = nil, want corrupt report error")
	}
	got, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read corrupt weakness report: %v", err)
	}
	if string(got) != string(invalidJSON) {
		t.Fatalf("corrupt weakness report was overwritten: %q", got)
	}
}

func TestStrictYAMLRejectsUnknownFields(t *testing.T) {
	var task Task
	err := decodeStrictYAML([]byte("schema_version: 1\nid: x\nunknown: true\n"), &task)
	if err == nil {
		t.Fatalf("expected strict YAML decode to reject unknown field")
	}
}

func TestAllowedSurfaceViolations(t *testing.T) {
	violations := allowedSurfaceViolations([]string{
		"docs/CONSTITUTION.md",
		"internal/app.go",
		"README.md",
	}, []string{
		"docs/**",
		"README.md",
	})
	if len(violations) != 1 || violations[0] != "internal/app.go" {
		t.Fatalf("violations = %#v, want internal/app.go", violations)
	}
}

func TestRedactOutput(t *testing.T) {
	token := "ghp_" + "abcdefghijklmnopqrstuvwxyz0123456789"
	password := "hunter" + "2"
	input := "token=" + token + "\npassword=" + password + "\n"
	out := redactOutput(input)
	if out == "" || out == input {
		t.Fatalf("expected output to be redacted, got %q", out)
	}
	if strings.Contains(out, "ghp_") || strings.Contains(out, password) {
		t.Fatalf("redacted output still contains secret material: %q", out)
	}
}

func fakeKitBinary(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kit")
	content := "#!/bin/sh\n" +
		"printf '%s\\n' '\"command\": \"capabilities\"'\n" +
		"printf '%s\\n' 'CodeRabbit'\n" +
		"printf '%s\\n' 'Kit-managed refresh state'\n" +
		"printf '%s\\n' 'github'\n" +
		"printf '%s\\n' '--refresh'\n" +
		"printf '%s\\n' '.kit/improve'\n" +
		"printf '%s\\n' 'private repositories skip'\n" +
		"printf '%s\\n' 'rules view'\n"
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake kit binary: %v", err)
	}
	return path
}

func fixtureProjectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		next := filepath.Dir(wd)
		if next == wd {
			t.Fatalf("could not find repo root")
		}
		wd = next
	}
}

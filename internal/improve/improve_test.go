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
	if manifest.Provenance.SuiteDefinitionSHA256 == "" || manifest.Provenance.KitBinarySHA256 == "" {
		t.Fatalf("expected benchmark provenance hashes, got %#v", manifest.Provenance)
	}
	if manifest.Metrics.TaskRuns != 8 || manifest.Metrics.PassedTaskRuns != 8 {
		t.Fatalf("unexpected run metrics: %#v", manifest.Metrics)
	}
	if manifest.Metrics.OutputCompleteness != 1 {
		t.Fatalf("output completeness = %v, want 1", manifest.Metrics.OutputCompleteness)
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
	if scorecard.Acceptance != "metadata-only" || scorecard.Score != 0 {
		t.Fatalf("scorecard.Acceptance = %q", scorecard.Acceptance)
	}
	body, err := PullRequestBody(root, manifest.RunDir, "#54")
	if err != nil {
		t.Fatalf("PullRequestBody() error = %v", err)
	}
	if !strings.Contains(body, "run provenance, aggregate metrics, and trace status") {
		t.Fatalf("expected PR body to describe included run evidence, got %q", body)
	}
	if strings.Contains(strings.ToLower(body), "includes scorecard") {
		t.Fatalf("PR body must not claim to include a scorecard: %q", body)
	}
	markdown, err := Report(root, manifest.RunDir)
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if markdown == "" {
		t.Fatalf("expected report markdown")
	}
}

func TestValidateRejectsIncompleteCandidateMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate.json")
	if err := os.WriteFile(path, []byte(`{"status":"proposed"}`), 0o644); err != nil {
		t.Fatalf("write candidate: %v", err)
	}
	_, err := Validate(t.TempDir(), path)
	if err == nil {
		t.Fatal("Validate() error = nil, want incomplete metadata error")
	}
	for _, want := range []string{"schema_version", "id is required", "target_cluster is required", "editable_surfaces", "regression_risks"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Validate() error = %q, want %q", err, want)
		}
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

func TestFailureClusterReportsActualSurfaceAndObservedFlakeRate(t *testing.T) {
	failures := []Trace{
		{TaskID: "prompt-model", FailureSignature: "assertion:prompt-model:stdout_contains:1", Status: "failed"},
		{TaskID: "prompt-model", FailureSignature: "assertion:prompt-model:stdout_contains:1", Status: "failed"},
	}
	all := append(append([]Trace(nil), failures...), Trace{TaskID: "prompt-model", Status: "passed"})
	if got := likelyHarnessSurface(failures[0].FailureSignature); got != "prompt-model" {
		t.Fatalf("likelyHarnessSurface() = %q", got)
	}
	if got := flakeRateFor(failures, all); got < 0.333 || got > 0.334 {
		t.Fatalf("flakeRateFor() = %v, want about 0.333", got)
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

func TestRunFailsWhenCommandFailsEvenIfStdoutAssertionPasses(t *testing.T) {
	root := fixtureProjectRoot(t)
	binary := filepath.Join(t.TempDir(), "kit")
	content := "#!/bin/sh\n" +
		"printf '%s\\n' '\"command\": \"capabilities\"' 'CodeRabbit' 'Kit-managed refresh state' 'github' '--refresh' '.kit/improve' 'private repositories skip' 'rules view'\n" +
		"exit 7\n"
	if err := os.WriteFile(binary, []byte(content), 0o755); err != nil {
		t.Fatalf("write failing Kit binary: %v", err)
	}
	manifest, err := Run(context.Background(), RunOptions{ProjectRoot: root, SuiteName: "default", KitBinary: binary})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if manifest.Status != "failed" || manifest.Metrics.FailedTaskRuns != 8 {
		t.Fatalf("expected every command failure to fail the run, got %#v", manifest.Metrics)
	}
	for _, trace := range manifest.Traces {
		if !strings.HasPrefix(trace.FailureSignature, "command:"+trace.TaskID+":0:exit-7") {
			t.Fatalf("trace %s failure signature = %q", trace.TaskID, trace.FailureSignature)
		}
	}
}

// package eval runs small local harness regression checks.
package eval

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/runstore"
	"github.com/jamesonstone/kit/internal/state"
	"github.com/jamesonstone/kit/internal/verify"
)

type CaseResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Report struct {
	SchemaVersion int          `json:"schema_version"`
	GeneratedAt   time.Time    `json:"generated_at"`
	Status        string       `json:"status"`
	Cases         []CaseResult `json:"cases"`
}

func Run(projectRoot string, cfg *config.Config) Report {
	report := Report{
		SchemaVersion: verify.SchemaVersion,
		GeneratedAt:   time.Now().UTC(),
		Status:        "pass",
	}
	report.Cases = append(report.Cases, evalRejectsShellSyntax())
	report.Cases = append(report.Cases, evalRejectsMessyCommand())
	report.Cases = append(report.Cases, evalIncompleteTaskNoDeclaredChecks())
	report.Cases = append(report.Cases, evalParsesCurrentTask(projectRoot, cfg))
	report.Cases = append(report.Cases, evalStateIsPointerOnly(projectRoot, cfg))
	report.Cases = append(report.Cases, evalStateHasStaleDetectionSources(projectRoot, cfg))
	report.Cases = append(report.Cases, evalTraceReplayEvidenceShape())
	report.Cases = append(report.Cases, evalReflectEvidenceLookup())
	for _, result := range report.Cases {
		if result.Status != "pass" {
			report.Status = "fail"
			break
		}
	}
	return report
}

func (r Report) Failed() bool {
	return r.Status != "pass"
}

func (r Report) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func evalRejectsShellSyntax() CaseResult {
	_, err := verify.ParseCommand("go test ./... && echo unsafe", "T000", 1, "fixture", false)
	if err == nil {
		return CaseResult{Name: "verify-parser-rejects-shell", Status: "fail", Message: "expected shell syntax rejection"}
	}
	return CaseResult{Name: "verify-parser-rejects-shell", Status: "pass"}
}

func evalRejectsMessyCommand() CaseResult {
	_, err := verify.ParseCommand("go test \"./...", "T000", 1, "fixture", false)
	if err == nil {
		return CaseResult{Name: "messy-verify-command-rejected", Status: "fail", Message: "expected unterminated quote rejection"}
	}
	return CaseResult{Name: "messy-verify-command-rejected", Status: "pass"}
}

func evalIncompleteTaskNoDeclaredChecks() CaseResult {
	run := verify.ExecuteRun(context.Background(), verify.RunOptions{
		ProjectRoot: "/tmp",
		Feature:     verify.FeatureRef{DirName: "fixture"},
		TaskIDs:     []string{"T001"},
	})
	if run.Status != verify.RunStatusNoDeclaredChecks {
		return CaseResult{Name: "incomplete-task-no-declared-checks", Status: "fail", Message: "expected no_declared_checks"}
	}
	return CaseResult{Name: "incomplete-task-no-declared-checks", Status: "pass"}
}

func evalParsesCurrentTask(projectRoot string, cfg *config.Config) CaseResult {
	featurePath := filepath.Join(projectRoot, cfg.SpecsDir, "0031-executable-verification-harness")
	tasksPath := filepath.Join(featurePath, "TASKS.md")
	bundles, err := verify.LoadTaskBundles(tasksPath, verify.FeatureRefFromDir(featurePath), false)
	if err != nil {
		return CaseResult{Name: "parse-executable-harness-tasks", Status: "fail", Message: err.Error()}
	}
	bundle, ok := verify.FindTaskBundle(bundles, "T001")
	if !ok || len(bundle.Verify) == 0 {
		return CaseResult{Name: "parse-executable-harness-tasks", Status: "fail", Message: "T001 verify command missing"}
	}
	return CaseResult{Name: "parse-executable-harness-tasks", Status: "pass"}
}

func evalStateIsPointerOnly(projectRoot string, cfg *config.Config) CaseResult {
	generated, err := state.Generate(projectRoot, cfg)
	if err != nil {
		return CaseResult{Name: "state-pointer-only", Status: "fail", Message: err.Error()}
	}
	data, err := json.Marshal(generated)
	if err != nil {
		return CaseResult{Name: "state-pointer-only", Status: "fail", Message: err.Error()}
	}
	if strings.Contains(string(data), "## REQUIREMENTS") || strings.Contains(string(data), "## TASK DETAILS") {
		return CaseResult{Name: "state-pointer-only", Status: "fail", Message: "state contains full document body markers"}
	}
	return CaseResult{Name: "state-pointer-only", Status: "pass"}
}

func evalStateHasStaleDetectionSources(projectRoot string, cfg *config.Config) CaseResult {
	generated, err := state.Generate(projectRoot, cfg)
	if err != nil {
		return CaseResult{Name: "state-stale-detection-sources", Status: "fail", Message: err.Error()}
	}
	needed := map[string]bool{
		"docs/PROJECT_PROGRESS_SUMMARY.md": false,
		filepath.ToSlash(filepath.Join(cfg.SpecsDir, "0031-executable-verification-harness", "TASKS.md")): false,
	}
	for _, source := range generated.Sources {
		if _, ok := needed[source.Path]; ok && source.Size > 0 && !source.ModTime.IsZero() {
			needed[source.Path] = true
		}
	}
	for path, ok := range needed {
		if !ok {
			return CaseResult{Name: "state-stale-detection-sources", Status: "fail", Message: "missing source fingerprint for " + path}
		}
	}
	return CaseResult{Name: "state-stale-detection-sources", Status: "pass"}
}

func evalTraceReplayEvidenceShape() CaseResult {
	projectRoot, err := os.MkdirTemp("", "kit-eval-runs-*")
	if err != nil {
		return CaseResult{Name: "trace-replay-evidence-shape", Status: "fail", Message: err.Error()}
	}
	defer os.RemoveAll(projectRoot)

	featureRef := verify.FeatureRef{ID: "0000", Slug: "fixture", DirName: "0000-fixture", Path: filepath.Join(projectRoot, "docs/specs/0000-fixture")}
	parent := verify.Run{
		SchemaVersion: verify.SchemaVersion,
		RunID:         verify.NewRunID(time.Now().UTC()),
		Feature:       featureRef,
		TaskIDs:       []string{"T001"},
		ExpectedFiles: []string{"fixture.go"},
		Commands:      []verify.Command{{ID: "T001-001", TaskID: "T001", Raw: "go test ./...", Argv: []string{"go", "test", "./..."}}},
		Results:       []verify.CommandResult{{CommandID: "T001-001", TaskID: "T001", Raw: "go test ./...", Status: "pass", ExitCode: 0, Stdout: "ok\n"}},
		Status:        verify.RunStatusPass,
		StartedAt:     time.Now().UTC(),
		EndedAt:       time.Now().UTC(),
	}
	if err := runstore.Write(projectRoot, &parent); err != nil {
		return CaseResult{Name: "trace-replay-evidence-shape", Status: "fail", Message: err.Error()}
	}
	child := parent
	child.RunID = verify.NewRunID(time.Now().UTC())
	child.ParentRunID = parent.RunID
	child.ArtifactDir = ""
	child.StartedAt = parent.StartedAt.Add(time.Nanosecond)
	child.EndedAt = parent.EndedAt.Add(time.Nanosecond)
	if err := runstore.Write(projectRoot, &child); err != nil {
		return CaseResult{Name: "trace-replay-evidence-shape", Status: "fail", Message: err.Error()}
	}
	latest, ok, err := runstore.LatestForFeature(projectRoot, featureRef.DirName)
	if err != nil {
		return CaseResult{Name: "trace-replay-evidence-shape", Status: "fail", Message: err.Error()}
	}
	if !ok || latest.ParentRunID != parent.RunID || latest.ExpectedFiles[0] != "fixture.go" {
		return CaseResult{Name: "trace-replay-evidence-shape", Status: "fail", Message: "latest linked run shape mismatch"}
	}
	return CaseResult{Name: "trace-replay-evidence-shape", Status: "pass"}
}

func evalReflectEvidenceLookup() CaseResult {
	projectRoot, err := os.MkdirTemp("", "kit-eval-reflect-*")
	if err != nil {
		return CaseResult{Name: "reflect-evidence-lookup", Status: "fail", Message: err.Error()}
	}
	defer os.RemoveAll(projectRoot)

	if _, ok, err := runstore.LatestForFeature(projectRoot, "0000-fixture"); err != nil || ok {
		return CaseResult{Name: "reflect-evidence-lookup", Status: "fail", Message: "expected missing evidence before any run"}
	}
	run := verify.Run{
		SchemaVersion: verify.SchemaVersion,
		RunID:         verify.NewRunID(time.Now().UTC()),
		Feature:       verify.FeatureRef{ID: "0000", Slug: "fixture", DirName: "0000-fixture"},
		Status:        verify.RunStatusPass,
		StartedAt:     time.Now().UTC(),
		EndedAt:       time.Now().UTC(),
	}
	if err := runstore.Write(projectRoot, &run); err != nil {
		return CaseResult{Name: "reflect-evidence-lookup", Status: "fail", Message: err.Error()}
	}
	if latest, ok, err := runstore.LatestForFeature(projectRoot, "0000-fixture"); err != nil || !ok || latest.RunID != run.RunID {
		return CaseResult{Name: "reflect-evidence-lookup", Status: "fail", Message: "expected latest evidence after run write"}
	}
	return CaseResult{Name: "reflect-evidence-lookup", Status: "pass"}
}

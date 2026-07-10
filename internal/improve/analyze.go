package improve

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Mine(projectRoot, from string) (WeaknessReport, error) {
	sourceDir := resolveArtifactDir(projectRoot, from)
	traceDir := filepath.Join(sourceDir, "traces")
	traces, err := readTraces(traceDir)
	if err != nil {
		return WeaknessReport{}, err
	}
	bySignature := map[string][]Trace{}
	for _, trace := range traces {
		if trace.Status == "passed" || strings.TrimSpace(trace.FailureSignature) == "" {
			continue
		}
		bySignature[trace.FailureSignature] = append(bySignature[trace.FailureSignature], trace)
	}
	report := WeaknessReport{
		SchemaVersion: SchemaVersion,
		Kind:          "weakness_report",
		SourceDir:     sourceDir,
	}
	for signature, values := range bySignature {
		cluster := WeaknessCluster{
			Signature:            signature,
			ObservedFailureMode:  signature,
			LikelyHarnessSurface: likelyHarnessSurface(signature),
			Actionability:        "needs-review",
			Confidence:           confidenceFor(values),
			ReproducibilityCount: len(values),
			FlakeRate:            flakeRateFor(values, traces),
			ProposedEvalTasks:    []string{"preserve " + signature},
			AffectedTasks:        uniqueTaskIDs(values),
			RepresentativeTraces: traceIDs(values),
		}
		if len(values) >= 2 {
			cluster.Actionability = "actionable"
		}
		report.Clusters = append(report.Clusters, cluster)
	}
	sort.Slice(report.Clusters, func(i, j int) bool {
		return report.Clusters[i].Signature < report.Clusters[j].Signature
	})
	if err := writeJSON(filepath.Join(sourceDir, "weakness-report.json"), report); err != nil {
		return WeaknessReport{}, err
	}
	return report, nil
}

func likelyHarnessSurface(signature string) string {
	parts := strings.Split(signature, ":")
	if len(parts) >= 2 && strings.TrimSpace(parts[1]) != "" {
		return parts[1]
	}
	return "unknown"
}

func flakeRateFor(failures, all []Trace) float64 {
	affected := map[string]struct{}{}
	for _, trace := range failures {
		affected[trace.TaskID] = struct{}{}
	}
	attempts := 0
	for _, trace := range all {
		if _, ok := affected[trace.TaskID]; ok {
			attempts++
		}
	}
	if attempts == 0 {
		return 0
	}
	return 1 - float64(len(failures))/float64(attempts)
}

func Propose(projectRoot, from string, maxCandidates int) ([]Candidate, error) {
	sourceDir := resolveArtifactDir(projectRoot, from)
	reportPath := filepath.Join(sourceDir, "weakness-report.json")
	report, err := readWeaknessReport(reportPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		report, err = Mine(projectRoot, sourceDir)
		if err != nil {
			return nil, err
		}
	}
	if maxCandidates <= 0 {
		maxCandidates = 3
	}
	var candidates []Candidate
	for i, cluster := range report.Clusters {
		if i >= maxCandidates {
			break
		}
		id := fmt.Sprintf("candidate-%03d", i+1)
		promptPath := filepath.Join(sourceDir, "candidates", id, "prompt.md")
		candidate := Candidate{
			SchemaVersion:    SchemaVersion,
			ID:               id,
			TargetCluster:    cluster.Signature,
			EditableSurfaces: []string{"docs/agents/**", "docs/references/rules/**", "pkg/cli/**"},
			PromptPath:       promptPath,
			Summary:          "Investigate and address " + cluster.Signature,
			ExpectedEffect:   "Improve held-in behavior without held-out regressions.",
			Rationale:        "Generated from reproducible weakness cluster evidence.",
			RegressionRisks:  []string{"Prompt bloat", "Overfitting to held-in tasks"},
			Rollback:         "Revert the candidate patch and rerun kit improve validate.",
			Status:           "proposed",
		}
		if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(promptPath, []byte(candidatePrompt(candidate, cluster)), 0o644); err != nil {
			return nil, err
		}
		if err := writeJSON(filepath.Join(sourceDir, "candidates", id, "candidate.json"), candidate); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func Validate(projectRoot, candidatePath string) (Scorecard, error) {
	candidate, err := readCandidate(candidatePath)
	if err != nil {
		return Scorecard{}, err
	}
	scorecard := Scorecard{
		SchemaVersion:      SchemaVersion,
		CandidateID:        candidate.ID,
		Score:              0,
		Acceptance:         "inconclusive",
		Reasons:            []string{"This command validates candidate metadata only; it does not run or compare benchmark behavior."},
		ValidationCommands: []string{"kit improve validate --candidate " + candidatePath + " --json"},
	}
	if candidate.Status == "proposed" {
		scorecard.Acceptance = "metadata-only"
		scorecard.Reasons = []string{"candidate metadata is well-formed", "score 0 is not a task-quality judgment", "run identical benchmark suites separately before review"}
	}
	return scorecard, nil
}

func Report(projectRoot, from string) (string, error) {
	sourceDir := resolveArtifactDir(projectRoot, from)
	manifest, err := readRunManifest(filepath.Join(sourceDir, "run.json"))
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("# Kit Improve Report\n\n")
	fmt.Fprintf(&builder, "- Suite: `%s`\n", manifest.Suite)
	fmt.Fprintf(&builder, "- Status: `%s`\n", manifest.Status)
	fmt.Fprintf(&builder, "- Traces: `%d`\n", len(manifest.Traces))
	builder.WriteString(fmt.Sprintf("- Suite definition SHA-256: `%s`\n", manifest.Provenance.SuiteDefinitionSHA256))
	builder.WriteString(fmt.Sprintf("- Benchmark runner SHA-256: `%s`\n", manifest.Provenance.RunnerBinarySHA256))
	builder.WriteString(fmt.Sprintf("- Kit binary SHA-256: `%s`\n", manifest.Provenance.KitBinarySHA256))
	builder.WriteString(fmt.Sprintf("- Task success: `%d/%d` (%.3f)\n", manifest.Metrics.PassedTaskRuns, manifest.Metrics.TaskRuns, manifest.Metrics.TaskSuccessRate))
	builder.WriteString(fmt.Sprintf("- Required-output assertions: `%d/%d` (%.3f)\n", manifest.Metrics.PassedAssertions, manifest.Metrics.Assertions, manifest.Metrics.OutputCompleteness))
	builder.WriteString(fmt.Sprintf("- Prompt/output size: `%d` lines, `%d` words, `%d` bytes, about `%d` tokens\n", manifest.Metrics.Stdout.Lines, manifest.Metrics.Stdout.Words, manifest.Metrics.Stdout.Bytes, manifest.Metrics.Stdout.EstimatedTokens))
	builder.WriteString(fmt.Sprintf("- Aggregate command duration: `%d ms`\n", manifest.Metrics.CommandDurationMS))
	builder.WriteString(fmt.Sprintf("- Repeated-output determinism: `%d/%d` (%.3f)\n", manifest.Metrics.StableRepeatedTasks, manifest.Metrics.RepeatedTasks, manifest.Metrics.DeterminismRate))
	builder.WriteString("- Unobservable without instrumented model calls: model latency, provider cost, conversational turns, and live tool/subagent routing.\n")
	builder.WriteString("\n## Tasks\n\n")
	for _, trace := range manifest.Traces {
		builder.WriteString(fmt.Sprintf("- `%s`: %s\n", trace.TaskID, trace.Status))
	}
	report := builder.String()
	if err := os.WriteFile(filepath.Join(sourceDir, "report.md"), []byte(report), 0o644); err != nil {
		return "", err
	}
	return report, nil
}

func PullRequestBody(projectRoot, from string, issue string) (string, error) {
	report, err := Report(projectRoot, from)
	if err != nil {
		return "", err
	}
	ticket := "Refs " + issue
	if strings.TrimSpace(issue) == "" {
		ticket = "Refs #[ticket number]"
	}
	return "## Description\n\n" +
		"- :sparkles: Adds benchmark-backed Kit harness improvement evidence.\n" +
		"- :white_check_mark: Includes run provenance, aggregate metrics, and trace status.\n" +
		"- :mag: Makes no candidate-acceptance claim; review candidate metadata and benchmark comparisons separately.\n" +
		"- :memo: Preserves Kit delivery gates and human review boundaries.\n\n" +
		"## How to Test\n\n" +
		"1. `kit improve run --suite default --json`\n" +
		"2. `kit improve mine --from .kit/improve/latest --json`\n" +
		"3. `kit improve report --from .kit/improve/latest`\n\n" +
		"## Ticket\n\n" + ticket + "\n\n" +
		"## Improvement Evidence\n\n" + report, nil
}

func readTraces(dir string) ([]Trace, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var traces []Trace
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		var trace Trace
		if err := readJSON(filepath.Join(dir, entry.Name()), &trace); err != nil {
			return nil, err
		}
		traces = append(traces, trace)
	}
	return traces, nil
}

func resolveArtifactDir(projectRoot, from string) string {
	if strings.TrimSpace(from) == "" {
		return filepath.Join(artifactRoot(projectRoot), "latest")
	}
	if filepath.IsAbs(from) {
		return from
	}
	return filepath.Join(projectRoot, from)
}

func readJSON(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func readWeaknessReport(path string) (WeaknessReport, error) {
	var report WeaknessReport
	return report, readJSON(path, &report)
}

func readCandidate(path string) (Candidate, error) {
	var candidate Candidate
	if err := readJSON(path, &candidate); err != nil {
		return Candidate{}, err
	}
	if err := validateCandidateMetadata(candidate); err != nil {
		return Candidate{}, err
	}
	return candidate, nil
}

func validateCandidateMetadata(candidate Candidate) error {
	var findings []string
	if candidate.SchemaVersion != SchemaVersion {
		findings = append(findings, fmt.Sprintf("schema_version must be %d", SchemaVersion))
	}
	requiredText := []struct {
		name  string
		value string
	}{
		{name: "id", value: candidate.ID},
		{name: "target_cluster", value: candidate.TargetCluster},
		{name: "summary", value: candidate.Summary},
		{name: "expected_effect", value: candidate.ExpectedEffect},
		{name: "rationale", value: candidate.Rationale},
		{name: "rollback", value: candidate.Rollback},
		{name: "status", value: candidate.Status},
	}
	for _, field := range requiredText {
		if strings.TrimSpace(field.value) == "" {
			findings = append(findings, field.name+" is required")
		}
	}
	if len(candidate.EditableSurfaces) == 0 {
		findings = append(findings, "editable_surfaces must not be empty")
	}
	if len(candidate.RegressionRisks) == 0 {
		findings = append(findings, "regression_risks must not be empty")
	}
	if strings.TrimSpace(candidate.Status) != "" && candidate.Status != "proposed" {
		findings = append(findings, "status must be proposed")
	}
	if len(findings) > 0 {
		return fmt.Errorf("invalid candidate metadata: %s", strings.Join(findings, "; "))
	}
	return nil
}

func readRunManifest(path string) (RunManifest, error) {
	var manifest RunManifest
	return manifest, readJSON(path, &manifest)
}

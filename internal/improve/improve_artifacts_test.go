package improve

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/verify"
)

func TestEvaluateAssertionsReportsOutputMetricsAndActualCause(t *testing.T) {
	results := []verify.CommandResult{{Status: "pass", ExitCode: 0, Stdout: "one two\nthree\n"}}
	task := Task{ID: "prompt", Assertions: []Assertion{
		{Type: "command_succeeds", CommandIndex: 0},
		{Type: "stdout_contains", CommandIndex: 0, Value: "missing"},
		{Type: "stdout_words_max", CommandIndex: 0, Max: 3},
	}}
	assertions := evaluateAssertions(task, results, nil)
	if assertions[1].Status != "failed" {
		t.Fatalf("missing output assertion = %#v", assertions[1])
	}
	if signature := failureSignature(task, results, assertions, nil); signature != "assertion:prompt:stdout_contains:1" {
		t.Fatalf("failure signature = %q", signature)
	}
	metrics := measureText(results[0].Stdout)
	if metrics.Lines != 2 || metrics.Words != 3 || metrics.Bytes != 14 || metrics.EstimatedTokens != 4 {
		t.Fatalf("text metrics = %#v", metrics)
	}
}

func TestNormalizeOutputForPersistenceRemovesDisposableWorkspace(t *testing.T) {
	got := normalizeOutputForPersistence("read /tmp/run/workspace/docs/SPEC.md\n", "/tmp/run/workspace")
	if got != "read {{workspace}}/docs/SPEC.md\n" {
		t.Fatalf("normalizeOutputForPersistence() = %q", got)
	}
}

func TestWriteCommandOutputRedactsPersistedMetadata(t *testing.T) {
	workspace := t.TempDir()
	token := "ghp_" + "abcdefghijklmnopqrstuvwxyz0123456789"
	password := "hunter" + "2"
	lines := make([]string, 205)
	for i := range lines {
		lines[i] = fmt.Sprintf("workspace=%s line=%d", workspace, i)
	}
	lines[0] += " token=" + token
	result := verify.CommandResult{
		CWD:    workspace,
		Stdout: strings.Join(lines, "\n") + "\n",
		Stderr: "token=" + token + "\n",
		Error:  "password=" + password,
	}

	traces, err := writeCommandOutput(t.TempDir(), "redaction", 1, []verify.CommandResult{result})
	if err != nil {
		t.Fatalf("writeCommandOutput() error = %v", err)
	}
	trace := traces[0]
	persistedStdout := limitLines(normalizeOutputForPersistence(redactOutput(result.Stdout), workspace), 200)
	if trace.Error != redactOutput(result.Error) {
		t.Fatalf("trace.Error = %q, want redacted error", trace.Error)
	}
	if strings.Contains(trace.Error, password) || !strings.Contains(trace.Error, "[REDACTED]") {
		t.Fatalf("trace.Error contains unredacted secret metadata: %q", trace.Error)
	}
	if trace.Stdout != measureText(persistedStdout) {
		t.Fatalf("trace.Stdout = %#v, want persisted metrics %#v", trace.Stdout, measureText(persistedStdout))
	}
	wantHash := hashText(persistedStdout)
	if trace.StdoutSHA256 != wantHash {
		t.Fatalf("trace.StdoutSHA256 = %q, want %q", trace.StdoutSHA256, wantHash)
	}
	for _, path := range []string{trace.StdoutPath, trace.StderrPath} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read output artifact %q: %v", path, err)
		}
		if strings.Contains(string(content), token) {
			t.Fatalf("output artifact %q contains secret material: %q", path, content)
		}
		if path == trace.StdoutPath && string(content) != persistedStdout {
			t.Fatalf("stdout artifact differs from measured and hashed content")
		}
	}
	if strings.Contains(persistedStdout, workspace) || !strings.Contains(persistedStdout, "{{workspace}}") {
		t.Fatalf("persisted stdout did not normalize workspace path: %q", persistedStdout)
	}
	if !strings.HasSuffix(persistedStdout, "[truncated]\n") {
		t.Fatalf("persisted stdout was not limited to 200 lines: %q", persistedStdout)
	}
	assertion := assertCommandSucceeds(
		Assertion{Type: "command_succeeds", CommandIndex: 0},
		[]verify.CommandResult{{Status: "failed", ExitCode: 1, Error: result.Error}},
	)
	if strings.Contains(assertion.Message, password) || !strings.Contains(assertion.Message, "[REDACTED]") {
		t.Fatalf("assertion metadata was not redacted: %q", assertion.Message)
	}
}

func TestAssertionResultPersistsZeroCommandIndex(t *testing.T) {
	result := passedAssertion(Assertion{Type: "command_succeeds", CommandIndex: 0})
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(data), `"command_index":0`) {
		t.Fatalf("command-scoped assertion JSON = %s, want command_index 0", data)
	}

	unscoped, err := json.Marshal(AssertionResult{Type: "git_diff_empty", Status: "passed"})
	if err != nil {
		t.Fatalf("json.Marshal() unscoped error = %v", err)
	}
	if strings.Contains(string(unscoped), "command_index") {
		t.Fatalf("unscoped assertion JSON = %s, want no command_index", unscoped)
	}
}

func TestLoadPromptSystemSuiteExercisesAgentContractSurfaces(t *testing.T) {
	root := fixtureProjectRoot(t)
	suite, tasks, err := LoadSuite(root, "prompt-system")
	if err != nil {
		t.Fatalf("LoadSuite(prompt-system) error = %v", err)
	}
	if suite.Repeat != 3 || len(tasks) != 8 {
		t.Fatalf("prompt-system suite repeat/tasks = %d/%d, want 3/8", suite.Repeat, len(tasks))
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
		"printf '%s\\n' '\"command\": \"status\"'\n" +
		"printf '%s\\n' 'github'\n" +
		"printf '%s\\n' '\"command\": \"init\"' '--refresh' '--dry-run'\n" +
		"printf '%s\\n' '.kit/improve'\n" +
		"printf '%s\\n' '\"command\": \"rules add\"' 'writes_files'\n"
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

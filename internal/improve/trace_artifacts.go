package improve

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/verify"
)

func writeCommandOutput(runDir, taskID string, repeat int, results []verify.CommandResult) ([]CommandTrace, error) {
	outDir := filepath.Join(runDir, "traces", "output", fmt.Sprintf("%s-%d", taskID, repeat))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	traces := make([]CommandTrace, 0, len(results))
	for i, result := range results {
		stdoutPath := filepath.Join(outDir, fmt.Sprintf("%d.stdout.txt", i+1))
		stderrPath := filepath.Join(outDir, fmt.Sprintf("%d.stderr.txt", i+1))
		if err := os.WriteFile(stdoutPath, []byte(limitLines(redactOutput(result.Stdout), 200)), 0o644); err != nil {
			return nil, err
		}
		if err := os.WriteFile(stderrPath, []byte(limitLines(redactOutput(result.Stderr), 200)), 0o644); err != nil {
			return nil, err
		}
		traces = append(traces, CommandTrace{
			Argv:         append([]string(nil), result.Argv...),
			ExitCode:     result.ExitCode,
			Status:       result.Status,
			DurationMS:   result.DurationMS,
			Error:        result.Error,
			TimedOut:     result.TimedOut,
			Stdout:       measureText(result.Stdout),
			StdoutSHA256: hashText(normalizeOutputForHash(result.Stdout, result.CWD)),
			StdoutPath:   stdoutPath,
			StderrPath:   stderrPath,
		})
	}
	return traces, nil
}

func evaluateAssertions(task Task, results []verify.CommandResult, changed []string) []AssertionResult {
	var out []AssertionResult
	for _, assertion := range task.Assertions {
		switch assertion.Type {
		case "command_succeeds":
			out = append(out, assertCommandSucceeds(assertion, results))
		case "stdout_contains":
			out = append(out, assertStdoutContains(assertion, results))
		case "stdout_not_contains":
			out = append(out, assertStdoutNotContains(assertion, results))
		case "stdout_nonempty":
			out = append(out, assertStdoutNonempty(assertion, results))
		case "stdout_lines_max", "stdout_words_max", "stdout_estimated_tokens_max":
			out = append(out, assertStdoutMaximum(assertion, results))
		case "git_diff_empty":
			if len(changed) == 0 {
				out = append(out, AssertionResult{Type: assertion.Type, Status: "passed"})
			} else {
				out = append(out, AssertionResult{Type: assertion.Type, Status: "failed", Message: "changed files: " + strings.Join(changed, ", ")})
			}
		default:
			out = append(out, AssertionResult{Type: assertion.Type, Status: "inconclusive", Message: "unsupported assertion type"})
		}
	}
	return out
}

func assertStdoutContains(assertion Assertion, results []verify.CommandResult) AssertionResult {
	result, failure := commandResult(assertion, results)
	if failure != nil {
		return *failure
	}
	if strings.Contains(result.Stdout, assertion.Value) {
		return passedAssertion(assertion)
	}
	return failedAssertion(assertion, fmt.Sprintf("stdout missing %q", assertion.Value))
}

func assertStdoutNotContains(assertion Assertion, results []verify.CommandResult) AssertionResult {
	result, failure := commandResult(assertion, results)
	if failure != nil {
		return *failure
	}
	if !strings.Contains(result.Stdout, assertion.Value) {
		return passedAssertion(assertion)
	}
	return failedAssertion(assertion, fmt.Sprintf("stdout unexpectedly contains %q", assertion.Value))
}

func assertStdoutNonempty(assertion Assertion, results []verify.CommandResult) AssertionResult {
	result, failure := commandResult(assertion, results)
	if failure != nil {
		return *failure
	}
	if strings.TrimSpace(result.Stdout) != "" {
		return passedAssertion(assertion)
	}
	return failedAssertion(assertion, "stdout is empty")
}

func assertCommandSucceeds(assertion Assertion, results []verify.CommandResult) AssertionResult {
	result, failure := commandResult(assertion, results)
	if failure != nil {
		return *failure
	}
	if result.Status == "pass" && result.ExitCode == 0 {
		return passedAssertion(assertion)
	}
	message := fmt.Sprintf("command exited %d", result.ExitCode)
	if result.TimedOut {
		message = "command timed out"
	} else if strings.TrimSpace(result.Error) != "" {
		message += ": " + result.Error
	}
	return failedAssertion(assertion, message)
}

func assertStdoutMaximum(assertion Assertion, results []verify.CommandResult) AssertionResult {
	result, failure := commandResult(assertion, results)
	if failure != nil {
		return *failure
	}
	metrics := measureText(result.Stdout)
	actual := 0
	switch assertion.Type {
	case "stdout_lines_max":
		actual = metrics.Lines
	case "stdout_words_max":
		actual = metrics.Words
	case "stdout_estimated_tokens_max":
		actual = metrics.EstimatedTokens
	}
	if actual <= assertion.Max {
		return passedAssertion(assertion)
	}
	return failedAssertion(assertion, fmt.Sprintf("stdout %s %d exceeds maximum %d", strings.TrimPrefix(assertion.Type, "stdout_"), actual, assertion.Max))
}

func commandResult(assertion Assertion, results []verify.CommandResult) (verify.CommandResult, *AssertionResult) {
	if assertion.CommandIndex < 0 || assertion.CommandIndex >= len(results) {
		failure := failedAssertion(assertion, "command_index out of range")
		return verify.CommandResult{}, &failure
	}
	return results[assertion.CommandIndex], nil
}

func passedAssertion(assertion Assertion) AssertionResult {
	return AssertionResult{Type: assertion.Type, CommandIndex: assertion.CommandIndex, Status: "passed"}
}

func failedAssertion(assertion Assertion, message string) AssertionResult {
	return AssertionResult{Type: assertion.Type, CommandIndex: assertion.CommandIndex, Status: "failed", Message: message}
}

func measureText(value string) TextMetrics {
	metrics := TextMetrics{
		Words: len(strings.Fields(value)),
		Bytes: len([]byte(value)),
	}
	if value != "" {
		metrics.Lines = strings.Count(value, "\n")
		if !strings.HasSuffix(value, "\n") {
			metrics.Lines++
		}
	}
	metrics.EstimatedTokens = (metrics.Bytes + 3) / 4
	return metrics
}

func hashText(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func normalizeOutputForHash(value, workspace string) string {
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return value
	}
	return strings.ReplaceAll(value, workspace, "{{workspace}}")
}

func failureSignature(task Task, commandResults []verify.CommandResult, assertions []AssertionResult, violations []string) string {
	for index, result := range commandResults {
		if result.Status == "pass" && result.ExitCode == 0 {
			continue
		}
		cause := fmt.Sprintf("exit-%d", result.ExitCode)
		if result.TimedOut {
			cause = "timeout"
		}
		return fmt.Sprintf("command:%s:%d:%s", task.ID, index, cause)
	}
	for index, assertion := range assertions {
		if assertion.Status != "passed" {
			return fmt.Sprintf("assertion:%s:%s:%d", task.ID, assertion.Type, index)
		}
	}
	if len(violations) > 0 {
		return "allowed-surface:" + task.ID
	}
	return ""
}

func writeJSON(filePath string, value any) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, append(data, '\n'), 0o644)
}

func limitLines(value string, max int) string {
	lines := strings.Split(value, "\n")
	if len(lines) <= max {
		return value
	}
	return strings.Join(lines[:max], "\n") + "\n[truncated]\n"
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{20,}`),
	regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password)=\S+`),
	regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`),
}

func redactOutput(value string) string {
	out := value
	for _, pattern := range secretPatterns {
		out = pattern.ReplaceAllString(out, "[REDACTED]")
	}
	return out
}

func updateLatest(root, runDir string) {
	latest := filepath.Join(root, "latest")
	_ = os.Remove(latest)
	if err := os.Symlink(runDir, latest); err != nil {
		_ = os.WriteFile(latest+".txt", []byte(runDir+"\n"), 0o644)
	}
}

func snapshotDir(root string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(root, func(filePath string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		files[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	return files, err
}

func changedFiles(before, after map[string]string) []string {
	seen := map[string]struct{}{}
	for filePath, hash := range after {
		if before[filePath] != hash {
			seen[filePath] = struct{}{}
		}
	}
	for filePath := range before {
		if _, ok := after[filePath]; !ok {
			seen[filePath] = struct{}{}
		}
	}
	var paths []string
	for filePath := range seen {
		paths = append(paths, filePath)
	}
	sort.Strings(paths)
	return paths
}

func allowedSurfaceViolations(changed, allowed []string) []string {
	if len(changed) == 0 || len(allowed) == 0 {
		return nil
	}
	var violations []string
	for _, filePath := range changed {
		if !matchesAllowedSurface(filePath, allowed) {
			violations = append(violations, filePath)
		}
	}
	return violations
}

func matchesAllowedSurface(filePath string, allowed []string) bool {
	filePath = filepath.ToSlash(filePath)
	for _, raw := range allowed {
		pattern := strings.TrimSpace(filepath.ToSlash(raw))
		if pattern == "" {
			continue
		}
		if pattern == filePath {
			return true
		}
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "**")
			if strings.HasPrefix(filePath, prefix) {
				return true
			}
		}
		if ok, _ := path.Match(pattern, filePath); ok {
			return true
		}
	}
	return false
}

package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Tests exercise the published recipes in disposable Git repositories. This
// record is a fixture reading aid, not a Kit schema or production validator.
type contextEvidenceFixture struct {
	Repository    string `json:"repository"`
	Commit        string `json:"commit"`
	Path          string `json:"path"`
	OriginalInput string `json:"original_input"`
	Question      string `json:"question"`
}

// contextEvidenceFence extracts a published recipe for direct fixture execution.
func contextEvidenceFence(t *testing.T, content, language string, index int) string {
	t.Helper()
	pattern := regexp.MustCompile("(?s)```" + language + "\n(.*?)\n```")
	matches := pattern.FindAllStringSubmatch(content, -1)
	if index >= len(matches) {
		t.Fatalf("missing %s recipe %d", language, index)
	}
	return matches[index][1]
}

// contextEvidenceRun runs a published POSIX recipe in an isolated repository.
func contextEvidenceRun(t *testing.T, recipe string, source contextEvidenceFixture) (string, error) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX recipe needs a shell; native Git operations remain available")
	}
	cmd := exec.Command("sh", "-c", recipe)
	cmd.Dir = source.Repository
	cmd.Env = append(os.Environ(), "source_commit="+source.Commit, "source_path="+source.Path)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// contextEvidenceSource creates a retained source revision with an uncaptured requirement.
func contextEvidenceSource(t *testing.T, path, content string) contextEvidenceFixture {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, path), content)
	initializeReconcileGitFixture(t, root)
	return contextEvidenceFixture{
		Repository: root, Commit: reconcileGitOutput(t, root, "rev-parse", "HEAD"),
		Path: path, OriginalInput: "UNCAPTURED", Question: "Which behavior does this source establish?",
	}
}

// TestContextEvidenceSourceRecoveryAndFreshness verifies historical recovery and current content checks.
func TestContextEvidenceSourceRecoveryAndFreshness(t *testing.T) {
	content := contextEvidenceRule(t).Content
	reopen := contextEvidenceFence(t, content, "sh", 0)
	freshness := contextEvidenceFence(t, content, "sh", 1)
	tests := []struct {
		name   string
		change string
		commit bool
		remove bool
		want   string
	}{
		{name: "normal code question", want: "SOURCE_UNCHANGED"},
		{name: "stale committed source", change: "package example\nconst phaseSource = \"metadata\"\n", commit: true, want: "NEEDS_REVALIDATION"},
		{name: "dirty source at identical HEAD", change: "package example\nconst phaseSource = \"working edit\"\n", want: "NEEDS_REVALIDATION"},
		{name: "missing working source", remove: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := contextEvidenceSource(t, "source with spaces.go", "package example\nconst phaseSource = \"tasks\"\n")
			if tt.change != "" {
				writeFile(t, filepath.Join(source.Repository, source.Path), tt.change)
				if tt.commit {
					runGitForSourceAuditTest(t, source.Repository, "add", "--", source.Path)
					runGitForSourceAuditTest(t, source.Repository, "commit", "-m", "change fixture")
				}
			}
			if tt.remove {
				if err := os.Remove(filepath.Join(source.Repository, source.Path)); err != nil {
					t.Fatal(err)
				}
			}
			got, err := contextEvidenceRun(t, reopen, source)
			if err != nil || got != "package example\nconst phaseSource = \"tasks\"" {
				t.Fatalf("original evidence not recovered: %q, %v", got, err)
			}
			got, err = contextEvidenceRun(t, freshness, source)
			if tt.want == "" {
				if err == nil || got == "SOURCE_UNCHANGED" {
					t.Fatalf("missing source was accepted: %q, %v", got, err)
				}
			} else if err != nil || got != tt.want {
				t.Fatalf("freshness = %q, %v; want %s", got, err, tt.want)
			}
		})
	}
}

// TestContextEvidenceHistoricalDecisionVersusCurrentImplementation keeps historical and current decisions separate.
func TestContextEvidenceHistoricalDecisionVersusCurrentImplementation(t *testing.T) {
	rule := contextEvidenceRule(t).Content
	source := contextEvidenceSource(t, "DECISIONS.md", "Decision: use task files for phase.\n")
	writeFile(t, filepath.Join(source.Repository, source.Path), "Decision superseded: use spec metadata.\n")
	runGitForSourceAuditTest(t, source.Repository, "add", "--", source.Path)
	runGitForSourceAuditTest(t, source.Repository, "commit", "-m", "supersede decision")
	recipe := contextEvidenceFence(t, rule, "sh", 0)
	historical, err := contextEvidenceRun(t, recipe, source)
	if err != nil || !strings.Contains(historical, "use task files") {
		t.Fatalf("historical evidence = %q, %v", historical, err)
	}
	source.Commit = reconcileGitOutput(t, source.Repository, "rev-parse", "HEAD")
	current, err := contextEvidenceRun(t, recipe, source)
	if err != nil || !strings.Contains(current, "use spec metadata") || current == historical {
		t.Fatalf("current evidence conflated with history: %q, %v", current, err)
	}
}

// TestContextEvidenceFreshProcessHandoffAndUncapturedRequirement verifies recovery without inventing absent input.
func TestContextEvidenceFreshProcessHandoffAndUncapturedRequirement(t *testing.T) {
	rule := contextEvidenceRule(t).Content
	source := contextEvidenceSource(t, "phase.go", "package example\n\nconst phaseSource = \"metadata\"\n")
	data, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(t.TempDir(), "handoff.json")
	if err := os.WriteFile(record, data, 0600); err != nil {
		t.Fatal(err)
	}
	// Read only the transferred record. A new shell receives repository identity
	// and source coordinates, without the old conversational context.
	data, err = os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	var received contextEvidenceFixture
	if err := json.Unmarshal(data, &received); err != nil {
		t.Fatal(err)
	}
	if received.OriginalInput != "UNCAPTURED" {
		t.Fatal("fixture incorrectly claims recovery of the original requirement")
	}
	got, err := contextEvidenceRun(t, contextEvidenceFence(t, rule, "sh", 0), received)
	if err != nil || !strings.Contains(got, "phaseSource = \"metadata\"") {
		t.Fatalf("fresh-process recovery = %q, %v", got, err)
	}
	// A summary with no source identity cannot reopen evidence. Do not treat
	// the resulting error as proof or invent a requirement to fill the gap.
	received.Commit = strings.Repeat("0", 40)
	if got, err := contextEvidenceRun(t, contextEvidenceFence(t, rule, "sh", 0), received); err == nil {
		t.Fatalf("unavailable source accepted: %q", got)
	}
}

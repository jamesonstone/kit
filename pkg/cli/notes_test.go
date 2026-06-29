package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jamesonstone/kit/internal/config"
)

func TestEnsureFeatureNotesDirCreatesFullScaffold(t *testing.T) {
	projectRoot := t.TempDir()

	notesPath, notesRelPath, err := ensureFeatureNotesDir(projectRoot, "0001-alpha")
	if err != nil {
		t.Fatalf("ensureFeatureNotesDir() error = %v", err)
	}

	if notesRelPath != "docs/notes/0001-alpha" {
		t.Fatalf("notesRelPath = %q, want docs/notes/0001-alpha", notesRelPath)
	}
	for _, path := range []string{
		filepath.Join(notesPath, ".gitkeep"),
		filepath.Join(notesPath, "README.md"),
		filepath.Join(notesPath, "inbox", ".gitkeep"),
		filepath.Join(notesPath, "references", ".gitkeep"),
		filepath.Join(notesPath, "responses", ".gitkeep"),
		filepath.Join(notesPath, "private", ".gitignore"),
		filepath.Join(notesPath, "private", "README.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected scaffold file %s: %v", path, err)
		}
	}

	privateIgnore := readFile(t, filepath.Join(notesPath, "private", ".gitignore"))
	for _, want := range []string{"*", "!.gitignore", "!README.md"} {
		if !strings.Contains(privateIgnore, want) {
			t.Fatalf("expected private .gitignore to contain %q, got:\n%s", want, privateIgnore)
		}
	}
}

func TestRunNotesAddsTrackedNoteWithFrontMatter(t *testing.T) {
	projectRoot := setupNotesTestProject(t)
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreNotesNow := stubNotesNow(t)
	defer restoreNotesNow()

	output, err := executeNotesCommand("alpha", "--add", "--source", "slack", "--title", "Customer Ask", "--json")
	if err != nil {
		t.Fatalf("kit notes error = %v\noutput:\n%s", err, output)
	}

	var result notesResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, output)
	}
	wantNotePath := filepath.Join("docs", "notes", "0001-alpha", "inbox", "2026-06-29-120000-customer-ask.md")
	if result.NotePath != filepath.ToSlash(wantNotePath) {
		t.Fatalf("note path = %q, want %q", result.NotePath, filepath.ToSlash(wantNotePath))
	}
	content := readFile(t, filepath.Join(projectRoot, wantNotePath))
	for _, want := range []string{
		"kind: note",
		"source: slack",
		"status: active",
		"sensitivity: internal",
		"captured_at: 2026-06-29T12:00:00Z",
		"feature: 0001-alpha",
		"# Customer Ask",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected note to contain %q, got:\n%s", want, content)
		}
	}
}

func TestRunNotesAddsPrivateIgnoredNote(t *testing.T) {
	projectRoot := setupNotesTestProject(t)
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreNotesNow := stubNotesNow(t)
	defer restoreNotesNow()

	output, err := executeNotesCommand("alpha", "--add", "--private", "--title", "Slack Conversation", "--json")
	if err != nil {
		t.Fatalf("kit notes private error = %v\noutput:\n%s", err, output)
	}

	var result notesResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v\noutput: %s", err, output)
	}
	if !result.Private {
		t.Fatalf("private = false, want true")
	}
	wantNotePath := filepath.Join("docs", "notes", "0001-alpha", "private", "2026-06-29-120000-slack-conversation.md")
	content := readFile(t, filepath.Join(projectRoot, wantNotePath))
	for _, want := range []string{
		"sensitivity: private",
		"feature: 0001-alpha",
		"# Slack Conversation",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected private note to contain %q, got:\n%s", want, content)
		}
	}

	privateIgnore := readFile(t, filepath.Join(projectRoot, "docs", "notes", "0001-alpha", "private", ".gitignore"))
	if !strings.Contains(privateIgnore, "*") || !strings.Contains(privateIgnore, "!README.md") {
		t.Fatalf("expected private ignore contract, got:\n%s", privateIgnore)
	}
}

func TestRunNotesInteractiveSelectsFeatureAndAddsPrivateNote(t *testing.T) {
	projectRoot := setupNotesTestProject(t)
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreNotesNow := stubNotesNow(t)
	defer restoreNotesNow()

	if _, err := executeNotesCommand("alpha"); err != nil {
		t.Fatalf("setup kit notes alpha error = %v", err)
	}

	output, err := executeNotesCommandWithInput("1\n5\nPrivate Chat\nslack\nactive\n")
	if err != nil {
		t.Fatalf("interactive kit notes error = %v\noutput:\n%s", err, output)
	}
	if !strings.Contains(output, "Created note: docs/notes/0001-alpha/private/2026-06-29-120000-private-chat.md") {
		t.Fatalf("expected private note creation output, got:\n%s", output)
	}
}

func executeNotesCommand(args ...string) (string, error) {
	return executeNotesCommandWithInput("", args...)
}

func executeNotesCommandWithInput(input string, args ...string) (string, error) {
	cmd := newNotesCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func setupNotesTestProject(t *testing.T) string {
	t.Helper()

	projectRoot := t.TempDir()
	if err := config.Save(projectRoot, config.Default()); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}
	return projectRoot
}

func stubNotesNow(t *testing.T) func() {
	t.Helper()

	previous := notesNow
	notesNow = func() time.Time {
		return time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	}
	return func() {
		notesNow = previous
	}
}

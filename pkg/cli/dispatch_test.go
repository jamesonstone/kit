package cli

import (
	"strings"
	"testing"
)

func TestBuildDispatchPrompt(t *testing.T) {
	tasks := []dispatchTask{
		{ID: "D001", Index: 1, Body: "Update middleware"},
		{ID: "D002", Index: 2, Body: "Refresh README"},
	}

	prompt := buildDispatchPrompt(tasks, 10, "/tmp/project", dispatchInputSourceEditor)
	checks := []string{
		"/plan",
		"Prepare a subagent dispatch plan",
		"Working directory: /tmp/project",
		"Input source: editor",
		"Effective max subagents: 10",
		"### D001",
		"### D002",
		"Do NOT launch any subagents yet",
		"anticipate which files are likely to change",
		"overlap clusters",
		"dispatch queue",
		"subagent assignments",
		"risks and unknowns",
		"Wait for explicit user approval",
		"launch at most 10 concurrent subagents",
		"git worktree add ~/worktrees/<repo>-<branch> <branch>",
		"keep all worktrees flat under `~/worktrees/`",
		"do not create worktrees inside the repository tree",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("expected prompt to contain %q", check)
		}
	}

	if !strings.HasPrefix(prompt, "/plan\n\nPrepare a subagent dispatch plan") {
		t.Fatalf("expected prompt to start with /plan dispatch header, got %q", prompt[:40])
	}
}

func TestNormalizeDispatchTasks(t *testing.T) {
	raw := strings.Join([]string{
		"Investigate auth failures",
		"on expired sessions",
		"",
		"- Update middleware",
		"  - preserve nested detail",
		"  - keep order",
		"",
		"1. Refresh CLI help",
		"2. Add README entry",
		"",
		"Confirm validation output",
	}, "\n")

	tasks, err := normalizeDispatchTasks(raw)
	if err != nil {
		t.Fatalf("expected task normalization to succeed: %v", err)
	}

	if len(tasks) != 5 {
		t.Fatalf("expected 5 normalized tasks, got %d", len(tasks))
	}

	wantBodies := []string{
		"Investigate auth failures\non expired sessions",
		"Update middleware\n  - preserve nested detail\n  - keep order",
		"Refresh CLI help",
		"Add README entry",
		"Confirm validation output",
	}

	for index, wantBody := range wantBodies {
		if tasks[index].ID != "D00"+string(rune('1'+index)) {
			t.Fatalf("expected stable task ID at index %d, got %q", index, tasks[index].ID)
		}
		if tasks[index].Body != wantBody {
			t.Fatalf("expected body %q, got %q", wantBody, tasks[index].Body)
		}
	}
}

func TestNormalizeDispatchTasksRejectsEmptyInput(t *testing.T) {
	if _, err := normalizeDispatchTasks(" \n\t "); err == nil {
		t.Fatalf("expected empty task input to fail")
	}
}

func TestResolveDispatchInputSourcePrecedence(t *testing.T) {
	if got := resolveDispatchInputSource("tasks.md", false); got != dispatchInputSourceFile {
		t.Fatalf("expected --file to win, got %s", got)
	}

	if got := resolveDispatchInputSource("", false); got != dispatchInputSourceStdin {
		t.Fatalf("expected stdin source, got %s", got)
	}

	if got := resolveDispatchInputSource("", true); got != dispatchInputSourceEditor {
		t.Fatalf("expected editor source, got %s", got)
	}
}

func TestValidateDispatchMaxSubagents(t *testing.T) {
	if err := validateDispatchMaxSubagents(1); err != nil {
		t.Fatalf("expected positive max-subagents to be valid: %v", err)
	}

	if err := validateDispatchMaxSubagents(0); err == nil {
		t.Fatalf("expected max-subagents validation to fail for zero")
	}
}

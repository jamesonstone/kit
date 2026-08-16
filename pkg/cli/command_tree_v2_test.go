package cli

import (
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/v3/internal/commandset"
)

func TestV2CommandTreeMatchesProtectedSurface(t *testing.T) {
	rootCmd.InitDefaultCompletionCmd()
	pruneCommandTree(rootCmd, "")

	wantSet := map[string]bool{}
	for _, path := range commandset.ProtectedPaths() {
		parts := strings.Fields(path)
		for index := range parts {
			wantSet[strings.Join(parts[:index+1], " ")] = true
		}
	}
	want := make([]string, 0, len(wantSet))
	for path := range wantSet {
		want = append(want, path)
	}
	want = append(want, "completion bash", "completion fish", "completion powershell", "completion zsh")
	sort.Strings(want)

	got := commandPaths(rootCmd, "")
	if !slices.Equal(got, want) {
		t.Fatalf("command tree mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestV2CommandTreeRejectsRemovedGroups(t *testing.T) {
	pruneCommandTree(rootCmd, "")
	for _, path := range []string{
		"backlog", "brainstorm", "ci", "complete", "feature", "handoff",
		"implement", "legacy", "loop", "map", "notes", "plan", "project",
		"prompt", "reflect", "replay", "scaffold", "skill", "state", "tasks",
		"trace", "verify",
	} {
		if commandPathPresent(rootCmd, path) {
			t.Errorf("removed command path %q remains available", path)
		}
	}
}

func commandPaths(parent *cobra.Command, prefix string) []string {
	var paths []string
	for _, child := range parent.Commands() {
		if !child.IsAvailableCommand() || child.Name() == "help" {
			continue
		}
		path := strings.TrimSpace(prefix + " " + child.Name())
		paths = append(paths, path)
		paths = append(paths, commandPaths(child, path)...)
	}
	sort.Strings(paths)
	return paths
}

func commandPathPresent(root *cobra.Command, path string) bool {
	parts := strings.Fields(path)
	command, remaining, err := root.Find(parts)
	return err == nil && command != root && len(remaining) == 0
}

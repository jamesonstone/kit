package cli

import (
	"strings"
	"testing"
)

func TestPromptProfileFlagRegisteredOnRootCommand(t *testing.T) {
	if flag := rootCmd.PersistentFlags().Lookup("profile"); flag == nil {
		t.Fatal("expected root command to register --profile")
	}
}

func TestPromptProfileFlagRejectsUnsupportedValues(t *testing.T) {
	var profile promptProfile
	if err := profile.Set(""); err != nil {
		t.Fatalf("Set(empty) error = %v", err)
	}
	if profile != promptProfileNone {
		t.Fatalf("expected empty profile, got %q", profile)
	}
	if err := profile.Set("frontend"); err != nil {
		t.Fatalf("Set(frontend) error = %v", err)
	}
	if profile != promptProfileFrontend {
		t.Fatalf("expected frontend profile, got %q", profile)
	}

	err := profile.Set("backend")
	if err == nil {
		t.Fatal("expected unsupported profile error")
	}
	if !strings.Contains(err.Error(), "frontend") {
		t.Fatalf("expected error to name supported frontend value, got %q", err.Error())
	}
}

func TestFrontendBooleanFlagIsNotRegistered(t *testing.T) {
	if flag := rootCmd.PersistentFlags().Lookup("frontend"); flag != nil {
		t.Fatal("did not expect root command to register --frontend")
	}
}

func TestPreparePromptWithoutFrontendProfileOmitsFrontendSection(t *testing.T) {
	restorePromptProfileState(t, promptProfileNone, false)
	restoreSingleAgent(t, false)

	got := prepareAgentPrompt("Please implement the API.\n")
	if strings.Contains(got, "## Frontend Profile") {
		t.Fatalf("expected no-profile prompt to omit frontend profile guidance, got:\n%s", got)
	}
	if !strings.Contains(got, "## Skills") || !strings.Contains(got, "## Subagent Orchestration") {
		t.Fatalf("expected standard skill and subagent suffixes to remain, got:\n%s", got)
	}
}

func TestPreparePromptWithFrontendProfileOrdersSuffixes(t *testing.T) {
	restorePromptProfile(t, promptProfileFrontend)
	restoreSingleAgent(t, false)

	got := prepareAgentPrompt("Please implement the UI.\n")

	checks := []string{
		"## Skills",
		"## Frontend Profile",
		"## Subagent Orchestration",
		"Use RLM-style context loading first",
		"Inspect existing frontend architecture",
		"product domain and audience",
		"Use familiar UI affordances",
		"unnecessary landing pages",
		"stable responsive dimensions",
		"browser or screenshot evidence",
		"render or inspect the UI",
		"text overflow",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("expected frontend prompt to contain %q", check)
		}
	}

	if count := strings.Count(got, "## Frontend Profile"); count != 1 {
		t.Fatalf("expected one frontend profile section, got %d\n%s", count, got)
	}

	skillsIndex := strings.Index(got, "## Skills")
	frontendIndex := strings.Index(got, "## Frontend Profile")
	subagentIndex := strings.Index(got, "## Subagent Orchestration")
	if !(skillsIndex < frontendIndex && frontendIndex < subagentIndex) {
		t.Fatalf("expected skills -> frontend profile -> subagents ordering, got:\n%s", got)
	}
}

func TestPreparePromptWithFrontendProfileSingleAgentOmitsSubagents(t *testing.T) {
	restorePromptProfile(t, promptProfileFrontend)
	restoreSingleAgent(t, true)

	got := prepareAgentPrompt("Please implement the UI.\n")
	if !strings.Contains(got, "## Frontend Profile") {
		t.Fatalf("expected frontend profile section, got:\n%s", got)
	}
	if strings.Contains(got, "## Subagent Orchestration") {
		t.Fatalf("expected single-agent prompt to omit subagent guidance, got:\n%s", got)
	}
}

func TestAppendPromptProfileSuffixDoesNotDuplicateFrontendSection(t *testing.T) {
	prompt := "Body\n\n## Frontend Profile\n- existing guidance"

	got := appendPromptProfileSuffix(prompt, promptProfileFrontend)
	if got != prompt {
		t.Fatalf("expected existing frontend profile section to remain unchanged, got:\n%s", got)
	}
}

func restorePromptProfile(t *testing.T, profile promptProfile) {
	t.Helper()
	previous := selectedPromptProfile
	previousExplicit := selectedPromptProfileExplicit
	selectedPromptProfile = profile
	selectedPromptProfileExplicit = profile != promptProfileNone
	t.Cleanup(func() {
		selectedPromptProfile = previous
		selectedPromptProfileExplicit = previousExplicit
	})
}

func restoreSingleAgent(t *testing.T, enabled bool) {
	t.Helper()
	previous := singleAgent
	singleAgent = enabled
	t.Cleanup(func() {
		singleAgent = previous
	})
}

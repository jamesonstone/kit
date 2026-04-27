package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunSpecFrontendProfilePersistsDependencies(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restoreSpecFlags := restoreSpecFlagState()
	defer restoreSpecFlags()
	restorePromptProfileState(t, promptProfileFrontend, true)

	cmd := newSpecProfileTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set(output-only) error = %v", err)
	}

	output := captureStdout(t, func() {
		if err := runSpec(cmd, []string{"dashboard"}); err != nil {
			t.Fatalf("runSpec() error = %v", err)
		}
	})

	specPath := filepath.Join(projectRoot, "docs", "specs", "0001-dashboard", "SPEC.md")
	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	text := string(content)
	checks := []string{
		"| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |",
		"| Design materials | design | docs/notes/0001-dashboard/design | optional frontend design input | optional |",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected SPEC.md to contain %q, got:\n%s", check, text)
		}
	}
	if !strings.Contains(output, "## Frontend Profile") {
		t.Fatalf("expected spec prompt to include frontend profile guidance, got:\n%s", output)
	}
}

func TestRunPlanFrontendProfilePersistsSpecAndPlanDependencies(t *testing.T) {
	projectRoot, _ := setupLifecycleTestProject(t)
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restorePlanFlags := restorePlanFlagState()
	defer restorePlanFlags()
	restorePromptProfileState(t, promptProfileFrontend, true)

	featurePath := filepath.Join(projectRoot, "docs", "specs", "0001-dashboard")
	writeFile(t, filepath.Join(featurePath, "SPEC.md"), documentTemplateWithSummary())

	cmd := newPlanProfileTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set(output-only) error = %v", err)
	}

	output := captureStdout(t, func() {
		if err := runPlan(cmd, []string{"dashboard"}); err != nil {
			t.Fatalf("runPlan() error = %v", err)
		}
	})

	for _, path := range []string{
		filepath.Join(featurePath, "SPEC.md"),
		filepath.Join(featurePath, "PLAN.md"),
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("os.ReadFile(%q) error = %v", path, err)
		}
		text := string(content)
		if !strings.Contains(text, "| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |") {
			t.Fatalf("expected %s to contain frontend profile dependency, got:\n%s", path, text)
		}
		if !strings.Contains(text, "| Design materials | design | docs/notes/0001-dashboard/design | optional frontend design input | optional |") {
			t.Fatalf("expected %s to contain design materials dependency, got:\n%s", path, text)
		}
	}
	if !strings.Contains(output, "## Frontend Profile") {
		t.Fatalf("expected plan prompt to include frontend profile guidance, got:\n%s", output)
	}
}

func TestRootInvalidProfileFailsBeforeBrainstormFileCreation(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restorePromptProfileState(t, promptProfileNone, false)

	previousOut := rootCmd.OutOrStdout()
	previousErr := rootCmd.ErrOrStderr()
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"--profile=backend", "brainstorm", "profile-failure", "--output-only"})
	defer func() {
		rootCmd.SetOut(previousOut)
		rootCmd.SetErr(previousErr)
		rootCmd.SetArgs(nil)
	}()

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected unsupported profile to fail")
	}
	if !strings.Contains(err.Error(), "frontend") {
		t.Fatalf("expected error to name supported frontend profile, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(projectRoot, "docs", "specs")); !os.IsNotExist(statErr) {
		t.Fatalf("expected invalid profile to avoid creating docs/specs, got %v", statErr)
	}
}

func TestRunSummarizeGenericUsesExplicitFrontendProfileOnly(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	writeFile(t, filepath.Join(projectRoot, "docs", "specs", "0001-dashboard", "SPEC.md"), dependencyDoc("| Frontend profile | profile | --profile=frontend | apply frontend-specific coding-agent instruction set | active |"))
	restore := chdirForTest(t, projectRoot)
	defer restore()
	restorePromptProfileState(t, promptProfileNone, false)
	restoreSummarize := restoreSummarizeFlagState()
	defer restoreSummarize()

	cmd := newSummarizeProfileTestCommand()
	if err := cmd.Flags().Set("output-only", "true"); err != nil {
		t.Fatalf("Flags().Set(output-only) error = %v", err)
	}

	noProfileOutput := captureStdout(t, func() {
		if err := runSummarize(cmd, nil); err != nil {
			t.Fatalf("runSummarize() without profile error = %v", err)
		}
	})
	if strings.Contains(noProfileOutput, "## Frontend Profile") {
		t.Fatalf("expected generic summarize not to infer frontend profile from feature docs, got:\n%s", noProfileOutput)
	}

	restorePromptProfileState(t, promptProfileFrontend, true)
	frontendOutput := captureStdout(t, func() {
		if err := runSummarize(cmd, nil); err != nil {
			t.Fatalf("runSummarize() with frontend profile error = %v", err)
		}
	})
	if !strings.Contains(frontendOutput, "## Frontend Profile") {
		t.Fatalf("expected generic summarize to honor explicit frontend profile, got:\n%s", frontendOutput)
	}
}

func newSpecProfileTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("template", false, "")
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().Bool("output-only", false, "")
	cmd.Flags().Bool("prompt-only", false, "")
	return cmd
}

func newPlanProfileTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("warp", false, "")
	cmd.Flags().Bool("output-only", false, "")
	cmd.Flags().Bool("prompt-only", false, "")
	return cmd
}

func newSummarizeProfileTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("output-only", false, "")
	return cmd
}

func restoreSpecFlagState() func() {
	previousCopy := specCopy
	previousEditor := specEditor
	previousInline := specInline
	previousOutputOnly := specOutputOnly
	previousUseVim := specUseVim
	specCopy = false
	specEditor = ""
	specInline = false
	specOutputOnly = false
	specUseVim = false
	return func() {
		specCopy = previousCopy
		specEditor = previousEditor
		specInline = previousInline
		specOutputOnly = previousOutputOnly
		specUseVim = previousUseVim
	}
}

func restoreSummarizeFlagState() func() {
	previousCopy := summarizeCopy
	previousOutputOnly := summarizeOutputOnly
	summarizeCopy = false
	summarizeOutputOnly = false
	return func() {
		summarizeCopy = previousCopy
		summarizeOutputOnly = previousOutputOnly
	}
}

func restorePlanFlagState() func() {
	previousCopy := planCopy
	previousOutputOnly := planOutputOnly
	planCopy = false
	planOutputOnly = false
	return func() {
		planCopy = previousCopy
		planOutputOnly = previousOutputOnly
	}
}

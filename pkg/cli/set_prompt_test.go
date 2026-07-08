package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/internal/config"
)

func TestRunSetPromptWithOptions_DefaultsToLocalInsideProject(t *testing.T) {
	projectRoot, _ := setupPromptTestProject(t)
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	editorCalls := stubSetPromptEditor(t, "local prompt body")

	output := captureStdout(t, func() {
		if err := runSetPromptWithOptions([]string{"custom", "review"}, false, false); err != nil {
			t.Fatalf("runSetPromptWithOptions() error = %v", err)
		}
	})

	if *editorCalls != 1 {
		t.Fatalf("expected one editor capture, got %d", *editorCalls)
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	got := cfg.Prompts["custom"]["review"].Content
	if got != "local prompt body" {
		t.Fatalf("local prompt content = %q, want local prompt body", got)
	}
	if !strings.Contains(output, "Saved prompt custom review to local") {
		t.Fatalf("expected save output, got %q", output)
	}
}

func TestRunSetPromptWithOptions_GlobalCreatesGlobalConfig(t *testing.T) {
	setupPromptTestEnvironment(t)
	editorCalls := stubSetPromptEditor(t, "global prompt body")

	output := captureStdout(t, func() {
		if err := runSetPromptWithOptions([]string{"custom", "review"}, false, true); err != nil {
			t.Fatalf("runSetPromptWithOptions() error = %v", err)
		}
	})

	if *editorCalls != 1 {
		t.Fatalf("expected one editor capture, got %d", *editorCalls)
	}
	cfg, found, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("config.LoadGlobal() error = %v", err)
	}
	if !found {
		t.Fatalf("expected global config to be created")
	}
	got := cfg.Prompts["custom"]["review"].Content
	if got != "global prompt body" {
		t.Fatalf("global prompt content = %q, want global prompt body", got)
	}
	if !strings.Contains(output, "Saved prompt custom review to global") {
		t.Fatalf("expected save output, got %q", output)
	}
}

func TestRunSetPromptWithOptions_DualScopeEditsOnce(t *testing.T) {
	projectRoot, _ := setupPromptTestProject(t)
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig())
	editorCalls := stubSetPromptEditor(t, "shared prompt body")

	if err := runSetPromptWithOptions([]string{"custom", "review"}, true, true); err != nil {
		t.Fatalf("runSetPromptWithOptions() error = %v", err)
	}
	if *editorCalls != 1 {
		t.Fatalf("expected one editor capture, got %d", *editorCalls)
	}

	localCfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	globalCfg, found, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("config.LoadGlobal() error = %v", err)
	}
	if !found {
		t.Fatalf("expected global config to be created")
	}
	for scope, cfg := range map[string]*config.Config{"local": localCfg, "global": globalCfg} {
		got := cfg.Prompts["custom"]["review"].Content
		if got != "shared prompt body" {
			t.Fatalf("%s prompt content = %q, want shared prompt body", scope, got)
		}
	}
}

func TestRunSetPromptWithOptions_LocalOutsideProjectFails(t *testing.T) {
	setupPromptTestEnvironment(t)
	editorCalls := stubSetPromptEditor(t, "unused")

	err := runSetPromptWithOptions([]string{"custom", "review"}, true, false)
	if err == nil {
		t.Fatalf("expected local outside project error")
	}
	if !strings.Contains(err.Error(), "--local requires a Kit project .kit.yaml") {
		t.Fatalf("unexpected error = %v", err)
	}
	if *editorCalls != 0 {
		t.Fatalf("expected no editor capture, got %d", *editorCalls)
	}
}

func TestRunSetPromptWithOptions_OutsideProjectDeclinedGlobalSaveCancels(t *testing.T) {
	setupPromptTestEnvironment(t)
	editorCalls := stubSetPromptEditor(t, "unused")

	output := withStdin(t, "n\n", func() string {
		return captureStdout(t, func() {
			if err := runSetPromptWithOptions([]string{"custom", "review"}, false, false); err != nil {
				t.Fatalf("runSetPromptWithOptions() error = %v", err)
			}
		})
	})

	if *editorCalls != 0 {
		t.Fatalf("expected no editor capture, got %d", *editorCalls)
	}
	if !strings.Contains(output, "No prompt saved.") {
		t.Fatalf("expected cancellation output, got %q", output)
	}
	if _, found, err := config.LoadGlobal(); err != nil || found {
		t.Fatalf("expected no global config, found = %v err = %v", found, err)
	}
}

func TestRunSetPromptWithOptions_OverwriteDeclineSkipsScope(t *testing.T) {
	projectRoot, _ := setupPromptTestProject(t)
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig()+`prompts:
  custom:
    review:
      content: original body
      description: original description
`)
	editorCalls := stubSetPromptEditor(t, "new body")

	output := withStdin(t, "n\n", func() string {
		return captureStdout(t, func() {
			if err := runSetPromptWithOptions([]string{"custom", "review"}, false, false); err != nil {
				t.Fatalf("runSetPromptWithOptions() error = %v", err)
			}
		})
	})

	if *editorCalls != 0 {
		t.Fatalf("expected no editor capture, got %d", *editorCalls)
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	got := cfg.Prompts["custom"]["review"].Content
	if got != "original body" {
		t.Fatalf("local prompt content = %q, want original body", got)
	}
	if !strings.Contains(output, "No prompt scopes selected. Nothing was changed.") {
		t.Fatalf("expected skipped-scope output, got %q", output)
	}
}

func TestRunSetPromptWithOptions_DualScopeOverwriteCanSkipOneScope(t *testing.T) {
	projectRoot, homeDir := setupPromptTestProject(t)
	writeFile(t, filepath.Join(projectRoot, ".kit.yaml"), defaultKitConfig()+`prompts:
  custom:
    review:
      content: original local
`)
	writeFile(t, filepath.Join(homeDir, ".config", "kit", ".kit.yaml"), `prompts:
  custom:
    review:
      content: original global
`)
	editorCalls := stubSetPromptEditor(t, "updated local")

	_ = withStdin(t, "y\nn\n", func() string {
		return captureStdout(t, func() {
			if err := runSetPromptWithOptions([]string{"custom", "review"}, true, true); err != nil {
				t.Fatalf("runSetPromptWithOptions() error = %v", err)
			}
		})
	})

	if *editorCalls != 1 {
		t.Fatalf("expected one editor capture, got %d", *editorCalls)
	}
	localCfg, err := config.Load(projectRoot)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	globalCfg, _, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("config.LoadGlobal() error = %v", err)
	}
	if got := localCfg.Prompts["custom"]["review"].Content; got != "updated local" {
		t.Fatalf("local prompt content = %q, want updated local", got)
	}
	if got := globalCfg.Prompts["custom"]["review"].Content; got != "original global" {
		t.Fatalf("global prompt content = %q, want original global", got)
	}
}

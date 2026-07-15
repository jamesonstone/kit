package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/templates"
)

func TestRunSpecCreatesV3LivingSpecAndConciseOrientation(t *testing.T) {
	projectRoot := t.TempDir()
	setWorkingDirectory(t, projectRoot)
	if err := config.Save(projectRoot, config.Default()); err != nil {
		t.Fatal(err)
	}
	restore := restoreSpecFlagState()
	defer restore()

	cmd := &cobra.Command{}
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	if err := runSpec(cmd, []string{"native-plan"}); err != nil {
		t.Fatalf("runSpec() error = %v", err)
	}

	specPath := filepath.Join(projectRoot, "docs", "specs", "0001-native-plan", "SPEC.md")
	doc, err := document.ParseFile(specPath, document.TypeSpec)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Metadata == nil || doc.Metadata.WorkflowVersion != document.WorkflowVersionV3 {
		t.Fatalf("metadata = %#v, want workflow_version 3", doc.Metadata)
	}
	for _, section := range document.SpecV3RequiredSections {
		if !doc.HasSection(section) {
			t.Fatalf("missing V3 section %q", section)
		}
	}
	if strings.Contains(out.String(), "lifecycle supervisor") || strings.Contains(out.String(), "clarification.status") {
		t.Fatalf("primary spec output contains legacy supervisor contract:\n%s", out.String())
	}
	for _, want := range []string{"native agent planning", "accepted plan", "curate repository memory"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("orientation missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunSpecPreservesExistingV2Spec(t *testing.T) {
	projectRoot := t.TempDir()
	setWorkingDirectory(t, projectRoot)
	if err := config.Save(projectRoot, config.Default()); err != nil {
		t.Fatal(err)
	}
	restore := restoreSpecFlagState()
	defer restore()

	featureDir := filepath.Join(projectRoot, "docs", "specs", "0001-existing")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(featureDir, "SPEC.md")
	original := templates.BuildSpecV2ArtifactForFeature(document.FeatureMetadataFromDir("0001-existing")) + "\n<!-- custom-v2-memory -->\n"
	writeFile(t, specPath, original)

	cmd := &cobra.Command{}
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	if err := runSpec(cmd, []string{"existing"}); err != nil {
		t.Fatalf("runSpec() error = %v", err)
	}
	if got := readFile(t, specPath); got != original {
		t.Fatalf("existing V2 spec was rewritten:\n%s", got)
	}
	if !strings.Contains(out.String(), "preserved workflow_version 2") || !strings.Contains(out.String(), "semantic curation") {
		t.Fatalf("missing V2 preservation guidance:\n%s", out.String())
	}
}

func TestLegacySupervisorRejectsV3Spec(t *testing.T) {
	projectRoot := t.TempDir()
	setWorkingDirectory(t, projectRoot)
	if err := config.Save(projectRoot, config.Default()); err != nil {
		t.Fatal(err)
	}
	featureDir := filepath.Join(projectRoot, "docs", "specs", "0001-v3")
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(featureDir, "SPEC.md"), templates.BuildSpecArtifactForFeature(document.FeatureMetadataFromDir("0001-v3")))

	restore := restoreSpecFlagState()
	defer restore()
	specLegacySupervisor = true
	cmd := newSpecProfileTestCommand()
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	err := runSpec(cmd, []string{"v3"})
	if err == nil || !strings.Contains(err.Error(), "does not support workflow_version 3") {
		t.Fatalf("runSpec() error = %v, want V3 rejection", err)
	}
	if !strings.Contains(errOut.String(), "deprecated") {
		t.Fatalf("missing deprecation warning: %s", errOut.String())
	}
}

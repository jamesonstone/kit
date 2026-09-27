package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

func contractEntryFile(block string) string {
	return "# AGENTS\n\nProject preface kept by the project.\n\n" + block + "\n## Project Notes\n\nKeep this project-owned note.\n"
}

func TestRefreshReplacesOnlyTheManagedContractBlock(t *testing.T) {
	projectRoot := t.TempDir()
	path := filepath.Join(projectRoot, "AGENTS.md")
	stale := strings.Replace(templates.UniversalContractBlock(), "Slack is read-only by default", "Slack is writable", 1)
	writeFile(t, path, contractEntryFile(stale))

	plan, err := planInstructionArtifactWrite(projectRoot, "AGENTS.md", instructionFileWriteModeAppendOnly, config.InstructionScaffoldVersionMemory)
	if err != nil {
		t.Fatal(err)
	}
	if plan.result != instructionFileUpdated || plan.legacyContract {
		t.Fatalf("plan = %#v, want updated managed block", plan)
	}
	if want := contractEntryFile(templates.UniversalContractBlock()); plan.content != want {
		t.Fatalf("refresh changed content outside the managed block:\n%s", plan.content)
	}

	writeFile(t, path, plan.content)
	again, err := planInstructionArtifactWrite(projectRoot, "AGENTS.md", instructionFileWriteModeAppendOnly, config.InstructionScaffoldVersionMemory)
	if err != nil {
		t.Fatal(err)
	}
	if again.result != instructionFileSkipped {
		t.Fatalf("second refresh result = %v, want idempotent skip", again.result)
	}
}

func TestRefreshLeavesPreContractEntryFilesUntouchedAndReportsThem(t *testing.T) {
	projectRoot, cfg := setupLifecycleTestProject(t)
	cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionMemory
	if err := config.Save(projectRoot, cfg); err != nil {
		t.Fatal(err)
	}
	legacy := "# AGENTS\n\n## Purpose\n\n- Old routing table without a managed block\n"
	path := filepath.Join(projectRoot, "AGENTS.md")
	writeFile(t, path, legacy)

	plan, err := planInstructionArtifactWrite(projectRoot, "AGENTS.md", instructionFileWriteModeAppendOnly, config.InstructionScaffoldVersionMemory)
	if err != nil {
		t.Fatal(err)
	}
	if plan.result != instructionFileSkipped || !plan.legacyContract || plan.content != "" {
		t.Fatalf("plan = %#v, want untouched legacy file", plan)
	}

	changes, notes, _, err := planRefreshInitInstructionArtifacts(projectRoot, initRefreshOptions{}, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range changes {
		if change.relativePath == "AGENTS.md" && change.after != legacy {
			t.Fatalf("refresh rewrote a pre-contract entry file:\n%s", change.after)
		}
	}
	if !containsString(notes, legacyContractNote) {
		t.Fatalf("refresh notes = %#v, want pre-contract note", notes)
	}

	var found bool
	for _, finding := range auditInstructionFiles(projectRoot, cfg) {
		if finding.FilePath == path && strings.Contains(finding.Issue, "predates the Kit-managed contract block") {
			found = true
		}
	}
	if !found {
		t.Fatal("reconcile did not report the pre-contract entry file")
	}
	if got, _ := os.ReadFile(path); string(got) != legacy {
		t.Fatal("audit mutated the entry file")
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestConstitutionBaselineFollowsContractPresence(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := config.Default()
	for _, path := range instructionFiles(cfg) {
		writeFile(t, filepath.Join(projectRoot, filepath.FromSlash(path)), "# Legacy\n\n## Purpose\n\nold\n")
	}
	if got := constitutionBaselineForProject(projectRoot, cfg, nil); got != templates.LegacyConstitutionBaselineSection {
		t.Fatal("pre-contract entry files must keep the legacy Constitution baseline")
	}
	var planned []initRefreshFileChange
	for _, path := range instructionFiles(cfg) {
		planned = append(planned, initRefreshFileChange{relativePath: path, after: templates.InstructionFile(path)})
	}
	if got := constitutionBaselineForProject(projectRoot, cfg, planned); got != templates.ConstitutionBaselineSection {
		t.Fatal("entry files carrying the contract block must use the pointer baseline")
	}
}

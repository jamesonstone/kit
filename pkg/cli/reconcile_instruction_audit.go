package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

// auditInstructionFiles reports what `kit reconcile` would change in the
// agent entry files, plus retired Kit artifacts and legacy scaffold state. It
// uses the same migration plan reconcile applies, so the two never disagree.
func auditInstructionFiles(projectRoot string, cfg *config.Config) []reconcileFinding {
	var findings []reconcileFinding
	for _, relativePath := range instructionFiles(cfg) {
		plan, err := planInstructionArtifactWrite(projectRoot, relativePath, instructionFileWriteModeConverge, false)
		if err != nil {
			return append(findings, newFinding(reconcileSeverityWarning, filepath.Join(projectRoot, relativePath), err.Error(), "inspect the file manually"))
		}
		issue := ""
		switch {
		case plan.note != "":
			issue = plan.note
		case plan.result == instructionFileCreated:
			issue = "missing Kit-managed agent entry file"
		case plan.result == instructionFileUpdated && strings.Contains(readFileOrEmpty(plan.absolutePath), templates.UniversalContractBeginMarker):
			issue = "Kit-managed contract block differs from the current universal contract"
		case plan.result == instructionFileUpdated:
			issue = "agent entry file predates the Kit-managed contract block"
		default:
			continue
		}
		findings = append(findings, newFinding(
			reconcileSeverityWarning,
			plan.absolutePath,
			issue,
			fmt.Sprintf("preview with `kit reconcile --dry-run --diff --file %s`, then run `kit reconcile`", relativePath),
		))
	}

	findings = append(findings, auditStandingAuthorityPolicy(projectRoot)...)

	artifacts, err := findRetiredArtifacts(projectRoot, cfg, nil)
	if err != nil {
		return findings
	}
	for _, artifact := range artifacts {
		issue := "retired Kit file"
		fix := "run `kit reconcile` to remove it"
		if !artifact.kitOwned {
			issue = "retired Kit file kept because it is not exactly as Kit generated it (" + artifact.reason + ")"
			fix = "move anything the project still needs elsewhere, then delete it"
		}
		finding := newFinding(
			reconcileSeverityWarning,
			filepath.Join(projectRoot, filepath.FromSlash(artifact.relativePath)),
			issue,
			fix,
		)
		finding.NonBlocking = true
		findings = append(findings, finding)
	}
	if cfg.InstructionScaffoldVersion != config.CurrentInstructionScaffoldVersion {
		finding := newFinding(
			reconcileSeverityWarning,
			filepath.Join(projectRoot, config.ConfigFileName),
			legacyScaffoldIssue(cfg.InstructionScaffoldVersion),
			"run `kit reconcile` to migrate to the current structure",
		)
		finding.NonBlocking = true
		findings = append(findings, finding)
	}
	return findings
}

func readFileOrEmpty(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func legacyScaffoldIssue(version int) string {
	if version == 0 {
		return "instruction_scaffold_version is missing, so this is a legacy Kit structure"
	}
	return fmt.Sprintf("instruction_scaffold_version %d is a legacy Kit structure", version)
}

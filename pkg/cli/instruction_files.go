package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
	"github.com/jamesonstone/kit/v3/internal/instructions"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

const (
	agentsMDPath            = instructions.AgentsMDPath
	claudeMDPath            = instructions.ClaudeMDPath
	copilotInstructionsPath = instructions.CopilotInstructionsPath
)

type instructionFileWriteResult string

type instructionFileWriteMode string

const (
	instructionFileCreated instructionFileWriteResult = "created"
	instructionFileUpdated instructionFileWriteResult = "updated"
	instructionFileMerged  instructionFileWriteResult = "merged"
	instructionFileSkipped instructionFileWriteResult = "skipped"
	// instructionFileRemoved deletes a retired Kit-owned file that Git can restore.
	instructionFileRemoved instructionFileWriteResult = "removed"

	// instructionFileWriteModeSkipExisting only creates missing files.
	instructionFileWriteModeSkipExisting instructionFileWriteMode = "skip-existing"
	// instructionFileWriteModeConverge migrates existing entry files onto the
	// managed contract block and leaves project-owned reference starters alone.
	instructionFileWriteModeConverge instructionFileWriteMode = "converge"
)

type instructionFileWritePlan struct {
	relativePath string
	absolutePath string
	content      string
	result       instructionFileWriteResult
	// converged is false when an entry file was left without a current block.
	converged bool
	note      string
}

func instructionFiles(cfg *config.Config) []string {
	return instructions.InstructionRelativePaths(cfg)
}

// instructionArtifactPaths lists every instruction artifact Kit generates: the
// agent entry files, project guidance, the testing reference, and the SOP guide.
func instructionArtifactPaths(cfg *config.Config) []string {
	return append(instructionFiles(cfg), templates.TestingReferencePath, templates.DesignPath, templates.SecurityPath, templates.SOPGuidePath)
}

func instructionArtifactContent(relativePath string) string {
	if filepath.ToSlash(relativePath) == templates.SOPGuidePath {
		return templates.SOPGuide
	}
	if content, ok := templates.ProjectGuidance(filepath.ToSlash(relativePath)); ok {
		return content
	}
	if filepath.ToSlash(relativePath) == templates.TestingReferencePath {
		return templates.TestingReference
	}
	return templates.InstructionEntryFile(relativePath)
}

func planInstructionArtifactWrite(
	projectRoot,
	relativePath string,
	mode instructionFileWriteMode,
	force bool,
	cfg *config.Config,
) (instructionFileWritePlan, error) {
	plan := instructionFileWritePlan{
		relativePath: relativePath,
		absolutePath: filepath.Join(projectRoot, filepath.FromSlash(relativePath)),
		converged:    true,
	}
	if policy := preserveSecurityPolicy(projectRoot, relativePath, force); policy != "" {
		plan.result = instructionFileSkipped
		plan.note = "preserved project-owned security policy at " + policy
		return plan, nil
	}
	if !document.Exists(plan.absolutePath) {
		plan.content = entryAwareContent(projectRoot, relativePath, instructionArtifactContent(relativePath))
		plan.result = instructionFileCreated
		return plan, nil
	}
	plan.result = instructionFileSkipped
	if mode == instructionFileWriteModeSkipExisting ||
		filepath.ToSlash(relativePath) == templates.TestingReferencePath || filepath.ToSlash(relativePath) == templates.SOPGuidePath {
		return plan, nil
	}
	data, err := os.ReadFile(plan.absolutePath)
	if err != nil {
		return instructionFileWritePlan{}, fmt.Errorf("failed to read %s: %w", relativePath, err)
	}
	if _, ok := templates.ProjectGuidance(filepath.ToSlash(relativePath)); ok {
		return planProjectGuidance(plan, string(data), force, cfg), nil
	}
	migration := migrateEntryFile(filepath.ToSlash(relativePath), string(data), force)
	plan.converged = migration.converged
	plan.note = migration.note
	base := string(data)
	if migration.content != "" && migration.content != string(data) {
		base = migration.content
		plan.content = migration.content
		plan.result = instructionFileUpdated
	}
	if plan.converged {
		next := entryAwareContent(projectRoot, relativePath, base)
		if next != string(data) {
			plan.content = next
			if plan.result == instructionFileSkipped {
				plan.result = instructionFileUpdated
			}
		}
	}
	return plan, nil
}

func entryAwareContent(projectRoot, relativePath, content string) string {
	if !isAgentEntry(relativePath) {
		return content
	}
	ledger, err := config.RepositoryWorkLedger(projectRoot)
	if err != nil {
		return content
	}
	return templates.ApplyWorkLedgerNote(content, ledger)
}

func isAgentEntry(relativePath string) bool {
	switch filepath.ToSlash(relativePath) {
	case agentsMDPath, claudeMDPath, copilotInstructionsPath:
		return true
	default:
		return false
	}
}

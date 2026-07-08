package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

func buildSpecV2SupervisorPrompt(input specV2PromptInput) string {
	cfg := input.Config
	if cfg == nil {
		cfg = config.Default()
	}

	goalPct := cfg.GoalPercentage
	constitutionPath := filepath.Join(input.ProjectRoot, "docs", "CONSTITUTION.md")
	projectProgressPath := filepath.Join(input.ProjectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")
	rlmPath := filepath.Join(input.ProjectRoot, "docs", "agents", "RLM.md")
	rulesPath := filepath.Join(input.ProjectRoot, "docs", "references", "rules")
	featureDir := filepath.Base(filepath.Dir(input.SpecPath))
	notesPath := featureNotesPath(input.ProjectRoot, featureDir)
	designPath := featureDesignMaterialsPath(input.ProjectRoot, featureDir)
	hasBrainstorm := document.Exists(input.BrainstormPath)
	useRLM := specNeedsRLM(input.FeatureSlug, input.SpecPath, input.BrainstormPath, input.Answers)
	agentTeamModeBullets := specV2AgentTeamModeBullets(input.SingleAgent)

	durableRows := [][]string{
		{"SPEC", fmt.Sprintf("%s - single durable feature artifact and workflow state", input.SpecPath)},
		{"CONSTITUTION", fmt.Sprintf("%s - durable repository facts: project-wide constraints, invariants, and development contract", constitutionPath)},
		{"PROJECT PROGRESS", fmt.Sprintf("%s - durable repository facts: highest completed artifact and prior-feature index", projectProgressPath)},
		{"KIT MANAGED RULESETS", fmt.Sprintf("%s - pointer-loaded durable repo-local rulesets managed by Kit", rulesPath)},
	}
	instructionRows := [][]string{
		{"RLM", fmt.Sprintf("%s - Kit's just-in-time context-routing pattern for progressive disclosure", rlmPath)},
	}
	instructionTargets := map[string]struct{}{
		filepath.Clean(rlmPath): {},
	}
	for _, row := range repoInstructionContextRows(input.ProjectRoot, cfg) {
		if len(row) < 2 {
			continue
		}
		cleaned := filepath.Clean(row[1])
		if _, exists := instructionTargets[cleaned]; exists {
			continue
		}
		instructionTargets[cleaned] = struct{}{}
		instructionRows = append(instructionRows, []string{
			row[0],
			fmt.Sprintf("%s - repo-local agent routing and safety guidance", row[1]),
		})
	}
	supportingRows := [][]string{
		{"FEATURE NOTES", fmt.Sprintf("%s - optional reference material supplied before or during spec work", notesPath)},
		{"DESIGN MATERIALS", fmt.Sprintf("%s - optional screenshots, references, and design inputs when relevant", designPath)},
		{"PROJECT ROOT", input.ProjectRoot},
	}
	if hasBrainstorm {
		supportingRows = append(supportingRows, []string{
			"LEGACY BRAINSTORM",
			fmt.Sprintf("%s - historical v1 research context; carry forward only still-relevant facts into SPEC.md", input.BrainstormPath),
		})
	}
	for _, row := range specSkillDiscoveryContextRows(input.ProjectRoot, cfg) {
		if len(row) < 2 {
			continue
		}
		switch row[0] {
		case "Repo Agents Entry", "Repo References Entry":
			continue
		}
		supportingRows = append(supportingRows, row)
	}

	return renderPromptDocument(func(doc *promptdoc.Document) {
		addSpecV2PromptContext(doc, input, durableRows, instructionRows, supportingRows, useRLM)
		addSpecV2PromptFoundations(doc, input, cfg, constitutionPath, projectProgressPath, rlmPath, rulesPath)
		addSpecV2PromptClarification(doc, input, goalPct)
		addSpecV2PromptAgentTeam(doc, agentTeamModeBullets, goalPct)
		addSpecV2PromptExecution(doc)
		addFinalResponseContract(doc, specV2FinalResponseContract()...)
	})
}

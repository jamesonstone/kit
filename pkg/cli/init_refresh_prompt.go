package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/promptdoc"
)

func outputInitRefreshDocumentationPrompt(projectRoot string, cfg *config.Config) error {
	prompt := buildInitRefreshDocumentationPrompt(projectRoot, cfg)
	fmt.Println()
	fmt.Println("Documentation refresh prompt:")
	if err := outputPromptWithClipboardDefault(prompt, false, false); err != nil {
		return err
	}
	printNumberedNextSteps([]string{
		"Paste the copied prompt into your agent to review semantic project documentation updates",
		"Apply only still-valid project-wide rules, structure, vocabulary, and reference changes",
		"Run `kit check --project` after documentation updates",
	})
	return nil
}

func buildInitRefreshDocumentationPrompt(projectRoot string, cfg *config.Config) string {
	constitutionPath := cfg.ConstitutionAbsPath(projectRoot)
	summaryPath := cfg.ProgressSummaryPath(projectRoot)
	specsPath := cfg.SpecsPath(projectRoot)
	agentsPath := filepath.Join(projectRoot, "docs", "agents")
	referencesPath := filepath.Join(projectRoot, "docs", "references")
	commandsPath := filepath.Join(projectRoot, "docs", "commands.md")
	readmePath := filepath.Join(projectRoot, "docs", "README.md")

	return renderPromptDocument(func(doc *promptdoc.Document) {
		doc.Heading(2, "Post Init Refresh Documentation Review")
		doc.Paragraph(fmt.Sprintf("Review semantic project documentation updates after Kit-managed project files were refreshed in %s.", projectRoot))
		doc.Paragraph("Goal:")
		doc.BulletList(
			"bring durable project documentation into alignment with refreshed Kit-managed files, registry rulesets, generated instruction docs, and command guidance",
			"promote any new global rules, source-material structure, workflow boundaries, or command behavior into canonical project docs",
			"keep the work documentation-only unless the user explicitly asks for code changes",
		)
		doc.Paragraph("Inputs to inspect:")
		doc.BulletList(
			fmt.Sprintf("Constitution: %s", constitutionPath),
			fmt.Sprintf("Project docs index: %s", readmePath),
			fmt.Sprintf("Command docs: %s", commandsPath),
			fmt.Sprintf("Project summary: %s", summaryPath),
			fmt.Sprintf("Agent routing docs: %s", agentsPath),
			fmt.Sprintf("References and rulesets: %s", referencesPath),
			fmt.Sprintf("Feature specs: %s", specsPath),
			"`AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` when present",
		)
		doc.Paragraph("Discovery commands:")
		doc.BulletList(
			"`git status --short`",
			"`git diff -- docs AGENTS.md CLAUDE.md .github/copilot-instructions.md README.md`",
			"`kit rules list`",
			"`kit capabilities reconcile --json`",
			"`kit check --project`",
		)
		doc.Paragraph("Update guidance:")
		doc.BulletList(
			"`docs/CONSTITUTION.md`: update durable project-wide principles, constraints, definitions, source-material structure, workflow rules, and codebase map entries when the refreshed Kit contract requires it",
			"`docs/agents/*`: update routing or RLM guidance only when generated docs or registry rules changed the durable agent contract",
			"`docs/references/README.md` and `docs/references/rules/*`: ensure new or changed rules are discoverable and accurately scoped",
			"`docs/commands.md`, `docs/README.md`, and command-facing docs: update when command behavior, flags, or workflow expectations changed",
			"`docs/specs/*`: update only when a current feature spec directly contradicts the refreshed global contract",
			"preserve existing project-specific wording when it remains accurate",
			"do not overwrite local custom guidance just because generated wording is newer",
		)
		doc.Paragraph("Required checks before finishing:")
		doc.BulletList(
			"`kit check --project`",
			"`kit check --all` if feature specs or repo instruction files were touched",
			"`git diff --check`",
		)
		doc.Paragraph("Final response:")
		doc.BulletList(
			"`Findings`: stale or missing project documentation, or `none`",
			"`Updates`: files changed and why, or `no documentation updates needed`",
			"`Verification`: commands run and observed results",
		)
	})
}

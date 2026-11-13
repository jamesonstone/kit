package cli

import (
	"github.com/spf13/cobra"
)

var initMint bool
var initCopy bool
var initOutputOnly bool
var initIssueTemplate bool
var initPlanDryRun bool
var initAllowOverwrite bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Kit project",
	Long: `Initialize a new Kit project in the current directory.

Creates:
  - .kit.yaml configuration file
  - .gitignore entries for Kit-local environment and generated artifacts
  - .env and .envrc local environment files
  - Makefile starter for canonical project commands
  - .coderabbit.yaml review configuration file
  - .github/pull_request_template.md pull request template
  - .github/ISSUE_TEMPLATE/issue.md only with --issue-template
  - .github/workflows/auto-assign.yml issue and pull request assignment workflow
  - README.md Kit-managed status badges and final Maintainers section
  - ~/.config/kit/.kit.yaml global configuration file
  - docs/CONSTITUTION.md
  - Repository instruction files (AGENTS.md, CLAUDE.md, .github/copilot-instructions.md)
  - DESIGN.md and SECURITY.md (project-customizable engineering guidance)
  - The core rules shipped with this Kit binary

Before it writes, init lists every file it would create, overwrite, or skip.
It warns about existing files it would change and about existing issue
templates, forms, and config.yml. An overwrite waits for confirmation. A
non-interactive run skips that file unless --force is set. --dry-run prints
the plan and writes nothing.

--issue-template opts in and records issues.issue_template: true. Reconcile
still updates a file that exactly matches a shipped Kit issue template.
Kit cannot list issue templates inherited from an organization .github
repository. When gh can identify an organization owner, init warns that a
local template overrides those inherited templates.

GitHub issue branches use GH-<number>. For Linear, look up the selected
team's key in Linear and pass it with --linear-team-key (for example ENG).
If a project has multiple teams, use the key of the team that owns the issues.

Mint Actions integration is planned automatically. Existing jobs need explicit
--mint-workflow, --mint-job and --mint-before-step placement.

Use --mint to scaffold a missing Mint v0.9.0 environment controller from an
existing team-first schema 2 .mint.yaml and project-owned build/deploy/observe
adapters. Existing workflows and activation gates are preserved.
This controller scaffold cannot be
combined with Mint workflow selection flags.

Init is for projects that are not yet Kit projects. A change to an existing
file waits for confirmation. A non-interactive run skips that change unless
--force is set. Files Kit only fills in when they are missing stay as they are.
In a project that already has .kit.yaml, run ` + "`kit reconcile`" + `
instead: it brings any existing Kit project to the current structure.

The generated Constitution starter is a valid bootstrap state. The prepared
prompt promotes only durable project-wide truth supported by repository evidence
and leaves empty-project Constitution sections unchanged.

Init copies the prepared project initialization prompt to the clipboard and
shows next steps; --output-only prints it instead.`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initMint, "mint", false, "scaffold Mint environment controls from an existing .mint.yaml and adapters")
	initCmd.Flags().BoolVar(&initCopy, "copy", false, "copy prompt to clipboard even with --output-only")
	initCmd.Flags().BoolVar(&initOutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	initCmd.Flags().BoolVar(&initIssueTemplate, "issue-template", false, "opt in to the GitHub issue template and record issues.issue_template: true")
	initCmd.Flags().BoolVar(&initPlanDryRun, "dry-run", false, "print the file plan and write nothing")
	initCmd.Flags().BoolVar(&initAllowOverwrite, "force", false, "write planned changes to existing files without confirmation")
	bindIssueProviderFlags(initCmd)
	bindMintPipelineSelection(initCmd, &initMintSelection)
	rootCmd.AddCommand(initCmd)
}

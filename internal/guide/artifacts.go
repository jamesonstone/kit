package guide

type Artifact struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Purpose      string `json:"purpose"`
	Instructions string `json:"instructions"`
}

func artifacts() []Artifact {
	return []Artifact{
		{"Agent contract", "AGENTS.md, CLAUDE.md, .github/copilot-instructions.md", "One universal contract rendered into vendor entry files.", "Read the entry file your agent supports. Kit owns its managed block; keep project additions outside it. Use kit reconcile --dry-run --diff to preview updates. These entry points invite discovery; no tool can guarantee that an agent reads them."},
		{"Constitution", "docs/CONSTITUTION.md", "Current project-wide invariants and constraints.", "Read before implementation. Preserve the Kit-managed baseline. Curate demonstrated durable truth, not feature inventories or transient plans."},
		{"Living spec", "docs/specs/<feature>/SPEC.md", "Material feature rationale, accepted plans, decisions, validation and outcomes.", "Use kit spec <feature> before implementation when rationale cannot live in code and tests. Keep the V3 sections current through delivery. Historical specs are evidence, not current instructions."},
		{"Contextual rules", "docs/references/rules/<slug>.md", "Procedures loaded when their declared triggers apply.", "Use kit rules list to list installed local rules and kit rules view <slug> to read a local or shipped rule. kit rules add <slug> installs a shipped rule; kit rules add without a slug opens an interactive selector that can install or create rules. Read triggered procedures. Preserve local overrides; rules ship with the binary."},
		{"References", "docs/references/", "Reusable project knowledge and evidence indices.", "Read relevant references before work. Record validation commands in docs/references/testing.md. Keep project facts here rather than in transcripts or generated command output."},
		{"Project SOP convention", "docs/sops/<slug>/SOP.md; docs/sops/local/<slug>/SOP.md", "Shared and Git-ignored local operating procedures.", "Use draft, active, and retired states. When present, read docs/sops/README.md and docs/references/rules/standard-operating-procedures.md. From a linked worktree, use the SOP rule to find local procedures in the primary checkout. Follow applicable active procedures under existing authorization and safety rules. This is a document convention, not a Kit creation command or execution engine; agents use their native tools."},
		{"Design and security guidance", "DESIGN.md; SECURITY.md", "Project-customizable engineering guidance.", "Read DESIGN.md for UI/UX/design. For security work, read the first existing .github/SECURITY.md, SECURITY.md, or docs/SECURITY.md. Kit preserves customized guidance during normal reconciliation."},
		{"Project configuration", ".kit.yaml", "Project settings, configured artifact paths and installed-artifact registry.", "Use kit config check to diagnose configuration and kit config set --help for supported updates. Use kit health for read-only project diagnosis and kit reconcile --dry-run --diff to preview managed-file repairs. The guide itself does not load or repair configuration."},
	}
}

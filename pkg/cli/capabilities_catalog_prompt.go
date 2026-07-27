package cli

func promptCapabilityRecords() []capabilityRecord {
	return []capabilityRecord{
		capability(
			"plan",
			"Prompt Utilities",
			"Work with plans produced by native coding agents.",
			mutationNone,
			withFileWrites("none; plan utilities may read and replace the macOS clipboard when explicitly invoked"),
			withRelated(
				related("plan challenge", "supplements a copied Codex plan for independent adversarial review"),
				related("legacy plan", "retains the deprecated staged PLAN.md generator"),
				related("spec", "persists accepted plan rationale only when repository memory is required"),
			),
			withWhenToUse("Use after a host agent such as Codex for Mac has produced a native plan."),
			withWhenNotToUse("Do not use as a plan generator; native coding agents own plan creation."),
			withExamples("kit plan challenge"),
			withCaveats("Kit never launches or calls a model through this command group."),
		),
		capability(
			"plan challenge",
			"Prompt Utilities",
			"Supplement a copied Codex plan with a paste-ready adversarial review prompt.",
			mutationNone,
			withFileWrites("none; reads the current macOS clipboard and replaces it with the generated prompt by default"),
			withFlags(
				flag("--output-only", "print the raw challenge prompt without replacing the clipboard", "read-only"),
				flag("--copy", "copy the challenge prompt even with --output-only"),
			),
			withRelated(
				related("plan", "lists native-plan utilities"),
				related("spec", "captures accepted material rationale after native planning"),
				related("legacy plan", "creates a deprecated staged PLAN.md artifact"),
			),
			withWhenToUse(
				"Use after copying the complete plan produced by Codex for Mac `/plan`.",
				"Paste the generated challenge prompt into a human-selected secondary model such as Claude.",
			),
			withWhenNotToUse(
				"Do not use when the clipboard does not contain the complete candidate plan.",
				"Do not use as model execution, desktop automation, or automatic review-result ingestion.",
			),
			withExamples("kit plan challenge", "kit plan challenge --output-only"),
			withCaveats(
				"Default execution explicitly reads and replaces the macOS clipboard; it does not watch the clipboard.",
				"The reviewer is instructed to return either IMPLEMENT THIS PLAN or paste-ready instructions for Codex's tell-what-to-do-different field.",
				"Kit does not persist the copied plan, access chat history, call the network, or launch a model.",
			),
		),
		capability(
			"instructions",
			"Prompt Utilities",
			"Print versioned provider-neutral coding-agent instructions as raw Markdown.",
			mutationNone,
			withFlags(flag("--version", "select an exact immutable vN version; defaults to the current version", "read-only")),
			withRelated(
				related("prompt", "renders reusable task-specific and compatibility prompts"),
				related("scaffold agents", "writes repository-specific agent instruction files"),
			),
			withWhenToUse(
				"Use when adding Kit's shared system-level coding-agent policy directly to Codex, Claude, or GitHub Copilot.",
				"Omit --version for the current policy or pass --version=vN to reproduce an earlier release.",
			),
			withWhenNotToUse(
				"Do not use to create or refresh repository-specific AGENTS.md, CLAUDE.md, or Copilot files; use `kit scaffold agents`.",
				"Do not use for a task-specific reusable prompt; use `kit prompt`.",
			),
			withExamples("kit instructions", "kit instructions --version=v3", "kit instructions --version=v2"),
			withCaveats(
				"The command writes only raw Markdown to stdout and never copies to the clipboard.",
				"All versions are embedded in the installed binary; selection performs no network or project configuration access.",
			),
		),
		capability("prompt", "Prompt Utilities", "Work with Kit prompt templates.", mutationNone, withFlags(flag("--output-only", "print prompt text instead of copying it with metadata"), flag("--copy", "copy prompt even with --output-only")), withRelated(related("prompt list", "lists available prompts, including the legacy V2 supervisor entry"), related("spec", "scaffolds or adopts living repository memory"), related("set prompt", "sets prompt preferences")), withWhenToUse("Use to render reusable support and compatibility prompts.", "Use this catalog entry when looking for the former kit spec prompt behavior."), withExamples("kit prompt list", "kit prompt kit spec", "kit prompt workflow spec --output-only"), withCaveats("`kit prompt kit spec` is legacy V2 compatibility; plain `kit spec <feature>` uses native-planning orientation.")),
		capability("prompt list", "Prompt Utilities", "List available Kit prompt templates, including the legacy V2 supervisor entry.", mutationNone, withRelated(related("prompt", "renders prompt templates"), related("spec", "creates or adopts SPEC.md when durable memory is needed")), withExamples("kit prompt list")),
		capability("set", "Prompt Utilities", "Update Kit configuration values.", mutationWritesFiles, withFileWrites("writes Kit local or global configuration"), withRelated(related("set prompt", "sets prompt-related configuration"))),
		capability("set prompt", "Prompt Utilities", "Set the active prompt preference.", mutationWritesFiles, withFileWrites("writes Kit prompt configuration"), withRelated(related("prompt list", "find prompt names before setting one"))),
		capability("handoff", "Prompt Utilities", "Generate handoff context for continuing work.", mutationNone, withRelated(related("summarize", "condenses project or feature context"))),
		capability("summarize", "Prompt Utilities", "Summarize Kit project or feature context.", mutationNone, withRelated(related("handoff", "packages summary for a continuation"))),
		capability("dispatch", "Prompt Utilities", "Build Agent Team Plan prompts for agents, PR review threads, and CodeRabbit prompt-prep intake.", mutationGit, withNetwork("none by default", "--pr fetches unresolved, non-outdated review threads and PR-head metadata and fetches origin; --loop can poll CodeRabbit; --resolve --yes mutates GitHub review-thread state"), withFileWrites("none for generic prompt generation", "--pr may create the canonical PR-head worktree and link the primary checkout's exact .env; editor and clipboard flags affect output outside project files"), withGitMutation("none for generic prompt generation", "--pr prompt generation may fetch origin and add or attach the exact PR-head worktree; --resolve mutates GitHub only"), withFlags(flag("--pr", "prefill from unresolved, non-outdated PR review threads and resolve the exact writable PR-head worktree", "conditional worktree creation; refuses forks and prompts before including dirty changes"), flag("--loop", "route PR review feedback through the prompt-prep review-loop workflow"), flag("--coderabbit", "with --pr, keep CodeRabbit-authored comments and extract Prompt for AI Agents blocks"), flag("--resolve", "with --pr, resolve matching unresolved review threads after fixes or no-op decisions are complete", "GitHub mutation; requires --yes"), flag("--yes", "confirm --resolve after fixes or no-op decisions are complete", "required for mutation"), flag("--watch", "with --loop, wait for current-head CodeRabbit review completion", "network read with polling"), flag("--copy", "copy generated prompt output"), flag("--output-only", "print prompt without wrapper text"), flag("--max-subagents", "bound the generated dispatch queue; default 3, hard ceiling 4")), withRelated(related("loop review", "coding-agent correctness repair loop"), related("ci", "inspects CI context"), related("code-review", "builds review-focused prompts")), withExamples("kit dispatch --file tasks.md", "kit dispatch --pr 14 --coderabbit", "kit dispatch --loop --pr 14 --watch", "kit dispatch --pr 14 --resolve --yes"), withCaveats("Generic dispatch remains prompt-only and normalizes its working directory to the current Git top-level.", "PR prompt generation reuses or creates the exact same-repository PR-head worktree, records remote/local head SHAs and the push target, and carries the user's dirty-change inclusion choice into the prompt.", "`--resolve --yes` is the explicit GitHub mutation path for already-handled review threads.", "Generated dispatch prompts require one accountable supervisor, an Agent Team Plan before subagent execution, default max 3 concurrent subagents, and hard ceiling 4.")),
		capability("code-review", "Prompt Utilities", "Generate a read-only, finding-oriented review prompt for the current branch change set.", mutationNone, withRelated(related("dispatch", "routes prompts to specialized agents"), related("legacy verify", "runs executable checks")), withCaveats("The prompt reports only evidence-backed actionable findings in severity order and does not add generic subagent implementation guidance.")),
		capability("skill", "Prompt Utilities", "Inspect or generate Kit skill prompts.", mutationNone, withRelated(related("skill mine", "mines local context for skill material"))),
		capability("skill mine", "Prompt Utilities", "Mine repository context for reusable skill material.", mutationNone, withFileWrites("none by default", "shared editor flags can copy or write generated prompt output"), withFlags(flag("--copy", "copy generated prompt output"), flag("--output-only", "print prompt without wrapper text")), withRelated(related("skill", "skill command group"))),
	}
}

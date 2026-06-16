package cli

import (
	"sort"
	"strings"
)

const capabilitiesSchemaVersion = 1

const (
	mutationNone             = "none"
	mutationWritesFiles      = "writes_files"
	mutationExecutesCommands = "executes_commands"
	mutationNetwork          = "network"
	mutationGit              = "git"
	mutationDestructive      = "destructive"
)

var capabilityCategoryOrder = map[string]int{
	"Setup":            1,
	"Workflow":         2,
	"Inspect & Repair": 3,
	"Prompt Utilities": 4,
	"Utilities":        5,
}

type capabilityBehavior struct {
	Summary       string `json:"summary"`
	FlagDependent string `json:"flag_dependent,omitempty"`
}

type capabilityFlag struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Safety  string `json:"safety,omitempty"`
}

type capabilityRelatedCommand struct {
	Command string `json:"command"`
	Note    string `json:"note,omitempty"`
}

type capabilityRecord struct {
	Command              string
	Category             string
	Summary              string
	MutationLevel        string
	NetworkUse           capabilityBehavior
	FileWrites           capabilityBehavior
	GitMutation          capabilityBehavior
	Hidden               bool
	Deprecated           bool
	DeprecationNote      string
	Aliases              []string
	ImportantFlags       []capabilityFlag
	RelatedCommands      []capabilityRelatedCommand
	WhenToUse            []string
	WhenNotToUse         []string
	Examples             []string
	Caveats              []string
	DetailedFlagBehavior []capabilityFlag
	IncludeInCompact     bool
}

type capabilityCompactRecord struct {
	Command         string                     `json:"command"`
	Category        string                     `json:"category"`
	Summary         string                     `json:"summary"`
	MutationLevel   string                     `json:"mutation_level"`
	NetworkUse      capabilityBehavior         `json:"network_use"`
	FileWrites      capabilityBehavior         `json:"file_writes"`
	GitMutation     capabilityBehavior         `json:"git_mutation"`
	Hidden          bool                       `json:"hidden"`
	Deprecated      bool                       `json:"deprecated"`
	ImportantFlags  []capabilityFlag           `json:"important_flags"`
	RelatedCommands []capabilityRelatedCommand `json:"related_commands"`
}

type capabilityDetailRecord struct {
	Command              string                     `json:"command"`
	Category             string                     `json:"category"`
	Summary              string                     `json:"summary"`
	MutationLevel        string                     `json:"mutation_level"`
	NetworkUse           capabilityBehavior         `json:"network_use"`
	FileWrites           capabilityBehavior         `json:"file_writes"`
	GitMutation          capabilityBehavior         `json:"git_mutation"`
	Hidden               bool                       `json:"hidden"`
	Deprecated           bool                       `json:"deprecated"`
	DeprecationNote      string                     `json:"deprecation_note,omitempty"`
	Aliases              []string                   `json:"aliases"`
	ImportantFlags       []capabilityFlag           `json:"important_flags"`
	RelatedCommands      []capabilityRelatedCommand `json:"related_commands"`
	WhenToUse            []string                   `json:"when_to_use"`
	WhenNotToUse         []string                   `json:"when_not_to_use"`
	Examples             []string                   `json:"examples"`
	Caveats              []string                   `json:"caveats"`
	DetailedFlagBehavior []capabilityFlag           `json:"detailed_flag_behavior"`
}

type capabilityOption func(*capabilityRecord)

func capability(command, category, summary, mutationLevel string, options ...capabilityOption) capabilityRecord {
	record := capabilityRecord{
		Command:          command,
		Category:         category,
		Summary:          summary,
		MutationLevel:    mutationLevel,
		NetworkUse:       capabilityBehavior{Summary: "none"},
		FileWrites:       capabilityBehavior{Summary: "none"},
		GitMutation:      capabilityBehavior{Summary: "none"},
		Aliases:          []string{},
		ImportantFlags:   []capabilityFlag{},
		RelatedCommands:  []capabilityRelatedCommand{},
		WhenToUse:        []string{summary},
		WhenNotToUse:     []string{"Use a narrower Kit command when one better matches the workflow step."},
		Examples:         []string{"kit " + command},
		Caveats:          []string{},
		IncludeInCompact: true,
	}
	for _, option := range options {
		option(&record)
	}
	if record.DetailedFlagBehavior == nil {
		record.DetailedFlagBehavior = append([]capabilityFlag(nil), record.ImportantFlags...)
	}
	return record
}

func withNetwork(summary string, flagDependent ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.NetworkUse = capabilityBehavior{Summary: summary}
		if len(flagDependent) > 0 {
			record.NetworkUse.FlagDependent = flagDependent[0]
		}
	}
}

func withFileWrites(summary string, flagDependent ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.FileWrites = capabilityBehavior{Summary: summary}
		if len(flagDependent) > 0 {
			record.FileWrites.FlagDependent = flagDependent[0]
		}
	}
}

func withGitMutation(summary string, flagDependent ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.GitMutation = capabilityBehavior{Summary: summary}
		if len(flagDependent) > 0 {
			record.GitMutation.FlagDependent = flagDependent[0]
		}
	}
}

func withFlags(flags ...capabilityFlag) capabilityOption {
	return func(record *capabilityRecord) {
		record.ImportantFlags = append(record.ImportantFlags, flags...)
		record.DetailedFlagBehavior = append(record.DetailedFlagBehavior, flags...)
	}
}

func withRelated(related ...capabilityRelatedCommand) capabilityOption {
	return func(record *capabilityRecord) {
		record.RelatedCommands = append(record.RelatedCommands, related...)
	}
}

func withAliases(aliases ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.Aliases = append(record.Aliases, aliases...)
	}
}

func withWhenToUse(values ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.WhenToUse = append([]string(nil), values...)
	}
}

func withWhenNotToUse(values ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.WhenNotToUse = append([]string(nil), values...)
	}
}

func withExamples(values ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.Examples = append([]string(nil), values...)
	}
}

func withCaveats(values ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.Caveats = append(record.Caveats, values...)
	}
}

func hiddenDeprecated(note string) capabilityOption {
	return func(record *capabilityRecord) {
		record.Hidden = true
		record.Deprecated = true
		record.DeprecationNote = note
		record.IncludeInCompact = false
	}
}

func flag(name, summary string, safety ...string) capabilityFlag {
	result := capabilityFlag{Name: name, Summary: summary}
	if len(safety) > 0 {
		result.Safety = safety[0]
	}
	return result
}

func related(command, note string) capabilityRelatedCommand {
	return capabilityRelatedCommand{Command: command, Note: note}
}

func capabilityCatalog() []capabilityRecord {
	records := []capabilityRecord{
		capability("init", "Setup", "Initialize Kit project scaffolding in a repository.", mutationWritesFiles, withFileWrites("writes Kit project files and docs"), withFlags(flag("--force", "replace existing generated files when supported", "review local edits first")), withRelated(related("scaffold", "generates individual workflow artifacts"))),
		capability("scaffold", "Setup", "Generate workflow artifacts for Kit features.", mutationWritesFiles, withFileWrites("writes generated docs under selected project paths"), withFlags(flag("--feature", "select the feature slug when supported")), withRelated(related("scaffold agents", "writes repo agent instructions"))),
		capability("scaffold agents", "Setup", "Generate or refresh repo-local agent instruction files.", mutationWritesFiles, withFileWrites("writes AGENTS.md and docs/agents guidance"), withFlags(flag("--force", "replace existing agent guidance files", "review local edits first")), withRelated(related("init", "creates the broader project structure"))),
		capability("brainstorm", "Workflow", "Create or inspect the brainstorm artifact for a feature.", mutationWritesFiles, withFileWrites("writes docs/specs/<feature>/BRAINSTORM.md when generating"), withRelated(related("spec", "turns brainstorm material into requirements"))),
		capability("backlog", "Workflow", "Inspect and manage feature backlog entries.", mutationWritesFiles, withFileWrites("can update backlog state depending on subcommand and flags"), withRelated(related("resume", "selects paused or pending work"))),
		capability("spec", "Workflow", "Create or update feature requirements.", mutationWritesFiles, withFileWrites("writes docs/specs/<feature>/SPEC.md"), withRelated(related("plan", "turns accepted requirements into implementation approach"))),
		capability("plan", "Workflow", "Create or update an implementation plan from a feature spec.", mutationWritesFiles, withFileWrites("writes docs/specs/<feature>/PLAN.md"), withRelated(related("tasks", "breaks the plan into executable tasks"))),
		capability("tasks", "Workflow", "Create or update the task list for a feature.", mutationWritesFiles, withFileWrites("writes docs/specs/<feature>/TASKS.md"), withRelated(related("implement", "uses tasks as the execution checklist"))),
		capability("loop", "Workflow", "Run workflow and review agent loops.", mutationExecutesCommands, withFileWrites("can write loop state under .kit/loops"), withGitMutation("none", "delegated agent commands may mutate git independently"), withFlags(flag("--dry-run", "show legacy workflow loop plan without delegated execution", "read-only"), flag("--max-iterations", "bound legacy workflow loop execution")), withRelated(related("loop workflow", "canonical feature workflow loop"), related("loop review", "changed-code correctness review loop"), related("implement", "runs outside the loop runner")), withExamples("kit loop workflow my-feature", "kit loop review", "kit loop review --pr 14")),
		capability("loop workflow", "Workflow", "Run the feature workflow through a confidence-gated local agent loop.", mutationExecutesCommands, withFileWrites("writes loop prompts, stdout, stderr, and run summaries under .kit/loops"), withGitMutation("none", "delegated agent commands may mutate files or git independently; Kit itself does not stage, commit, or push"), withFlags(flag("--dry-run", "show the next workflow stage without invoking the configured agent", "read-only"), flag("--until", "stop after a workflow stage is complete"), flag("--min-confidence", "override the required agent confidence percentage"), flag("--max-iterations", "bound loop execution"), flag("--json", "emit the loop report as JSON")), withRelated(related("loop review", "runs correctness review passes over changed code"), related("verify", "records validation evidence")), withExamples("kit loop workflow my-feature --dry-run", "kit loop workflow my-feature --until tasks")),
		capability("loop review", "Workflow", "Run a coding-agent correctness review loop over changes not in the remote mainline.", mutationExecutesCommands, withNetwork("none by default", "--pr reads PR metadata, CodeRabbit checks, and unresolved non-outdated CodeRabbit review threads through gh"), withFileWrites("writes loop prompts, stdout, stderr, and run summaries under .kit/loops; delegated agent may edit local project files"), withGitMutation("none; command prompts explicitly forbid staging, commits, pushes, PR comments, and review-thread resolution"), withFlags(flag("--base", "override the comparison base; default is origin/main, then main"), flag("--pr", "opportunistically ingest CodeRabbit feedback from a pull request", "network read"), flag("--watch", "wait for CodeRabbit completion before finalizing PR-mode review", "network read with polling"), flag("--wait-for-coderabbit", "alias for --watch", "network read with polling"), flag("--dry-run", "show the first review prompt without invoking the configured agent", "read-only"), flag("--min-confidence", "override the required correctness percentage"), flag("--max-iterations", "bound review passes; default is 10"), flag("--json", "emit the loop review report as JSON")), withRelated(related("loop workflow", "runs the feature artifact workflow loop"), related("review-loop", "legacy PR prompt-prep compatibility command"), related("dispatch --loop", "PR prompt-prep alias")), withWhenToUse("Use before publishing or updating a PR when changed code should be locally reviewed and repaired until no high, medium, or correctness-impacting issues remain.", "Use with --pr when CodeRabbit feedback should be folded into local repair passes opportunistically."), withWhenNotToUse("Do not use when you only need a prompt from PR review threads; use `dispatch --loop --pr` or legacy `review-loop`.", "Do not use as a git/GitHub delivery command; it never stages, commits, pushes, comments, or resolves threads."), withExamples("kit loop review", "kit loop review --base origin/main", "kit loop review --pr 14", "kit loop review --pr 14 --wait-for-coderabbit"), withCaveats("The correctness percentage is agent-reported and gated by the prompt plus validation evidence; it is not a mathematical proof.")),
		capability("resume", "Workflow", "Resume the next useful Kit workflow action.", mutationNone, withRelated(related("status", "shows current feature state before resuming"))),
		capability("implement", "Workflow", "Generate or print the implementation prompt for a feature.", mutationNone, withRelated(related("tasks", "defines the implementation checklist"), related("verify", "validates completed work"))),
		capability("reflect", "Workflow", "Generate or print a reflection prompt for completed feature work.", mutationNone, withRelated(related("complete", "marks finished work after validation"))),
		capability("pause", "Workflow", "Mark or package feature work as paused.", mutationWritesFiles, withFileWrites("updates Kit workflow state for the paused feature"), withRelated(related("resume", "continues paused work"))),
		capability("complete", "Workflow", "Mark a feature as complete after validation.", mutationWritesFiles, withFileWrites("updates Kit feature state and progress artifacts"), withRelated(related("verify", "runs checks before completion"), related("reflect", "captures completion notes"))),
		capability("rm", "Workflow", "Remove Kit-managed feature or state entries.", mutationDestructive, withFileWrites("removes or updates Kit-managed files depending on flags"), withAliases("remove"), withFlags(flag("--dry-run", "preview removals before changing files", "use before destructive cleanup")), withRelated(related("status", "inspect state before removal"))),
		capability("status", "Inspect & Repair", "Show current Kit project and feature status.", mutationNone, withFlags(flag("--json", "emit machine-readable status"), flag("--all", "include all known features")), withRelated(related("map", "shows feature document graph"), related("check", "validates feature artifacts"))),
		capability("map", "Inspect & Repair", "Show the artifact map for a feature.", mutationNone, withFlags(flag("--json", "emit machine-readable map output")), withRelated(related("check", "validates mapped artifacts"), related("capabilities", "explains command behavior"))),
		capability("capabilities", "Inspect & Repair", "List Kit commands and their safety behavior.", mutationNone, withFlags(flag("--json", "emit stable machine-readable capability data"), flag("--full", "include detailed hidden and deprecated records"), flag("--search", "filter visible command metadata")), withRelated(related("map", "shows feature context"), related("help", "shows command syntax")), withWhenToUse("Use before choosing a Kit command when command behavior, mutation, network, or file-write boundaries are uncertain.", "Inside the Kit source repository, use targeted lookup before implementing or extending a Kit command so the command catalog stays current."), withWhenNotToUse("Do not use as a replacement for `kit map` when the question is about feature documents or references.", "Do not use `--full` by default; prefer compact search or targeted lookup unless hidden/deprecated command metadata is needed.", "Do not maintain Kit's internal command catalog from a downstream project."), withExamples("kit capabilities --json", "kit capabilities dispatch --json", "kit capabilities --search review-loop --json"), withCaveats("Kit maintainers must keep this catalog current when command behavior changes; downstream projects should use it for discovery, not maintain it.")),
		capability("check", "Inspect & Repair", "Validate Kit feature artifacts and workflow consistency.", mutationNone, withFlags(flag("--all", "check every feature"), flag("--project", "include project-level checks")), withRelated(related("map", "finds feature artifact inputs"), related("verify", "runs implementation verification checks"))),
		capability("ci", "Inspect & Repair", "Inspect GitHub Actions CI state and prepare CI-fix prompts.", mutationNetwork, withNetwork("uses git/gh subprocesses and GitHub CLI/API to inspect PR checks and logs", "--dispatch can ask GitHub to rerun eligible workflows; --copilot can attempt Copilot-assisted diagnosis"), withFileWrites("can update .kit.yaml default branch cache", "prompt output flags can copy or write generated text"), withFlags(flag("--pr", "select a pull request to inspect"), flag("--run", "inspect one workflow run"), flag("--dispatch", "rerun eligible workflows", "network mutation"), flag("--copilot", "attempt Copilot-assisted diagnosis when supported", "optional network/tooling behavior"), flag("--no-copilot", "disable Copilot-assisted diagnosis"), flag("--json", "emit machine-readable CI summary")), withRelated(related("verify", "runs local validation"), related("dispatch", "routes review or CI prompts")), withExamples("kit ci", "kit ci --pr 14 --json", "kit ci --run 123456789 --dispatch"), withCaveats("Default diagnosis is read-oriented, but `--dispatch` can rerun eligible workflows and must be treated as a network mutation.")),
		capability("verify", "Inspect & Repair", "Run local implementation verification for a feature.", mutationExecutesCommands, withFileWrites("writes verification run artifacts by default", "--dry-run and --no-write avoid run artifact writes"), withFlags(flag("--task", "limit verification to one task"), flag("--dry-run", "plan verification without executing commands", "read-only"), flag("--no-write", "avoid writing verification artifacts", "read-only"), flag("--allow-shell", "allow shell checks to execute", "executes local commands")), withRelated(related("check", "validates docs before implementation verification"), related("complete", "uses verification as a completion gate"))),
		capability("trace", "Inspect & Repair", "Trace Kit artifact relationships for a feature.", mutationNone, withRelated(related("map", "shows the feature artifact graph"))),
		capability("replay", "Inspect & Repair", "Replay recorded Kit verification or workflow context.", mutationExecutesCommands, withFileWrites("can write replay output or run artifacts"), withFlags(flag("--dry-run", "preview replay plan when supported", "prefer before executing")), withRelated(related("verify", "creates verification run context"))),
		capability("state", "Inspect & Repair", "Inspect or refresh Kit project state.", mutationWritesFiles, withFileWrites("default inspection is read-only", "the refresh action updates Kit state files"), withFlags(flag("--json", "emit machine-readable state output")), withRelated(related("status", "renders state for humans"))),
		capability("eval", "Inspect & Repair", "Evaluate Kit project quality signals.", mutationNone, withFlags(flag("--json", "emit machine-readable evaluation output")), withRelated(related("check", "validates feature artifacts"), related("verify", "runs task verification"))),
		capability("rules", "Inspect & Repair", "Manage Kit rule bindings and inspect available rules.", mutationWritesFiles, withFileWrites("root is mostly dispatch; add and link write rule files or feature links"), withAliases("rule"), withRelated(related("rules list", "lists available rules"), related("rules add", "adds rules to the project"))),
		capability("rules add", "Inspect & Repair", "Add a rule to the project rule set.", mutationWritesFiles, withFileWrites("writes rule files or rule references under docs/references/rules"), withFlags(flag("--force", "replace existing rule content when supported", "review current rule first")), withRelated(related("rules list", "find rules that can be added"), related("rules view", "inspect a rule before adding"))),
		capability("rules list", "Inspect & Repair", "List available or installed Kit rules.", mutationNone, withRelated(related("rules add", "adds one listed rule"))),
		capability("rules view", "Inspect & Repair", "Show the content or metadata for a Kit rule.", mutationNone, withRelated(related("rules add", "adds a viewed rule"))),
		capability("rules link", "Inspect & Repair", "Link a rule to a feature or artifact.", mutationWritesFiles, withFileWrites("updates feature docs or rule references to record the link"), withRelated(related("rules view", "review the rule first"), related("map", "shows feature artifact context"))),
		capability("reconcile", "Inspect & Repair", "Compare Kit docs and state for drift.", mutationNone, withRelated(related("check", "validates known invariants"), related("state", "refreshes state when needed"))),
		capability("prompt", "Prompt Utilities", "Work with Kit prompt templates.", mutationNone, withRelated(related("prompt list", "lists available prompts"), related("set prompt", "sets prompt preferences"))),
		capability("prompt list", "Prompt Utilities", "List available Kit prompt templates.", mutationNone, withRelated(related("prompt", "renders prompt templates"))),
		capability("set", "Prompt Utilities", "Update Kit configuration values.", mutationWritesFiles, withFileWrites("writes Kit local or global configuration"), withRelated(related("set prompt", "sets prompt-related configuration"))),
		capability("set prompt", "Prompt Utilities", "Set the active prompt preference.", mutationWritesFiles, withFileWrites("writes Kit prompt configuration"), withRelated(related("prompt list", "find prompt names before setting one"))),
		capability("handoff", "Prompt Utilities", "Generate handoff context for continuing work.", mutationNone, withRelated(related("summarize", "condenses project or feature context"))),
		capability("summarize", "Prompt Utilities", "Summarize Kit project or feature context.", mutationNone, withRelated(related("handoff", "packages summary for a continuation"))),
		capability("review-loop", "Prompt Utilities", "Legacy PR prompt-prep compatibility command for current unresolved review feedback.", mutationNone, hiddenDeprecated("use `kit dispatch --loop --pr <target>` for prompt prep or `kit loop review --pr <target>` for the coding-agent repair loop"), withNetwork("uses gh to read PR metadata, review threads, and optional CodeRabbit check status"), withFileWrites("none by default", "editor and clipboard output use temporary/editor state outside project files"), withGitMutation("none"), withFlags(flag("--pr", "select the pull request from a URL, Markdown link, owner/repo#number, or current-repo number", "network read"), flag("--coderabbit", "include only CodeRabbit-authored review comments and Prompt for AI Agents blocks"), flag("--watch", "wait for current-head CodeRabbit review completion before collecting feedback", "network read with polling"), flag("--copy", "copy generated dispatch prompt output"), flag("--output-only", "print prompt without wrapper text"), flag("--max-subagents", "bound the generated dispatch queue")), withRelated(related("loop review", "canonical coding-agent repair loop"), related("dispatch --loop", "alias path for prompt preparation")), withExamples("kit review-loop --pr 14 --coderabbit", "kit review-loop --pr owner/repo#14 --watch"), withCaveats("This compatibility command prepares prompts; it does not run the repair loop.")),
		capability("dispatch", "Prompt Utilities", "Build dispatch prompts for agents, PR review threads, and CodeRabbit prompt-prep intake.", mutationNetwork, withNetwork("none by default", "--pr fetches unresolved, non-outdated GitHub PR review threads through gh api graphql; --loop also reads PR metadata and optional CodeRabbit check status; --resolve --yes mutates GitHub review-thread state"), withFileWrites("none by default", "editor, --copy, and output flags can copy or write generated prompt output outside project files"), withGitMutation("none"), withFlags(flag("--pr", "prefill the dispatch editor from unresolved, non-outdated PR review threads", "network read"), flag("--loop", "route PR review feedback through the legacy prompt-prep review-loop workflow"), flag("--coderabbit", "with --pr, keep CodeRabbit-authored comments and extract Prompt for AI Agents blocks"), flag("--resolve", "with --pr, resolve matching unresolved review threads after fixes or no-op decisions are complete", "GitHub mutation; requires --yes"), flag("--yes", "confirm --resolve after fixes or no-op decisions are complete", "required for mutation"), flag("--watch", "with --loop, wait for current-head CodeRabbit review completion", "network read with polling"), flag("--copy", "copy generated prompt output"), flag("--output-only", "print prompt without wrapper text"), flag("--max-subagents", "bound the generated dispatch queue")), withRelated(related("loop review", "coding-agent correctness repair loop"), related("review-loop", "legacy prompt-prep command"), related("ci", "inspects CI context"), related("code-review", "builds review-focused prompts")), withExamples("kit dispatch --file tasks.md", "kit dispatch --pr 14 --coderabbit", "kit dispatch --loop --pr 14 --watch", "kit dispatch --pr 14 --resolve --yes"), withCaveats("Default dispatch generates prompts only; `--resolve --yes` is the explicit GitHub mutation path for already-handled review threads.")),
		capability("code-review", "Prompt Utilities", "Generate a code-review prompt from current context.", mutationNone, withRelated(related("dispatch", "routes prompts to specialized agents"), related("verify", "runs executable checks"))),
		capability("skill", "Prompt Utilities", "Inspect or generate Kit skill prompts.", mutationNone, withRelated(related("skill mine", "mines local context for skill material"))),
		capability("skill mine", "Prompt Utilities", "Mine repository context for reusable skill material.", mutationNone, withFileWrites("none by default", "shared editor flags can copy or write generated prompt output"), withFlags(flag("--copy", "copy generated prompt output"), flag("--output-only", "print prompt without wrapper text")), withRelated(related("skill", "skill command group"))),
		capability("upgrade", "Utilities", "Upgrade the Kit CLI installation.", mutationNetwork, withNetwork("downloads release metadata or binaries"), withFileWrites("writes the installed Kit binary or related install files"), withFlags(flag("--check", "check for an upgrade without installing when supported", "prefer for read-only inspection")), withRelated(related("version", "shows current installed version"))),
		capability("version", "Utilities", "Print the Kit CLI version.", mutationNone, withRelated(related("upgrade", "updates the installed version"))),
		capability("completion", "Utilities", "Generate shell completion scripts.", mutationNone, withFileWrites("none by default", "the shell may redirect output to a completion file outside Kit"), withRelated(related("help", "shows command syntax"))),
		capability("help", "Utilities", "Show command help and flag syntax.", mutationNone, withRelated(related("capabilities", "adds behavior and safety metadata"))),
		capability("update", "Utilities", "Deprecated alias for upgrading Kit.", mutationNetwork, hiddenDeprecated("Use `kit upgrade` instead."), withNetwork("downloads release metadata or binaries"), withFileWrites("writes the installed Kit binary or related install files"), withRelated(related("upgrade", "canonical replacement"))),
		capability("skills", "Prompt Utilities", "Deprecated alias for skill helpers.", mutationNone, hiddenDeprecated("Use `kit skill` instead."), withRelated(related("skill", "canonical replacement"))),
		capability("skills mine", "Prompt Utilities", "Deprecated alias for skill mine.", mutationNone, hiddenDeprecated("Use `kit skill mine` instead."), withRelated(related("skill mine", "canonical replacement"))),
		capability("catchup", "Workflow", "Deprecated compatibility command for resuming context.", mutationNone, hiddenDeprecated("Use `kit resume` or `kit summarize` depending on the workflow need."), withRelated(related("resume", "canonical workflow continuation"), related("summarize", "context summary replacement"))),
		capability("rollup", "Inspect & Repair", "Hidden compatibility command for progress rollups.", mutationWritesFiles, hiddenDeprecated("Use the current completion and progress-summary workflow instead."), withFileWrites("can update progress summary artifacts"), withRelated(related("complete", "current feature completion flow"))),
	}

	sort.SliceStable(records, func(i, j int) bool {
		return lessCapabilityRecord(records[i], records[j])
	})
	return records
}

func compactCapabilityRecords(records []capabilityRecord) []capabilityCompactRecord {
	compact := make([]capabilityCompactRecord, 0, len(records))
	for _, record := range records {
		if !record.IncludeInCompact || record.Hidden || record.Deprecated {
			continue
		}
		compact = append(compact, record.compact())
	}
	return compact
}

func detailCapabilityRecords(records []capabilityRecord) []capabilityDetailRecord {
	detail := make([]capabilityDetailRecord, 0, len(records))
	for _, record := range records {
		detail = append(detail, record.detail())
	}
	return detail
}

func capabilityByCommandPath(commandPath string) (capabilityRecord, bool) {
	normalized := normalizeCapabilityQuery(commandPath)
	for _, record := range capabilityCatalog() {
		if record.matchesCommandPath(normalized) {
			return record, true
		}
	}
	return capabilityRecord{}, false
}

func searchCapabilityRecords(query string) []capabilityRecord {
	normalized := normalizeCapabilityQuery(query)
	if normalized == "" {
		return visibleCapabilityRecords()
	}

	var matches []capabilityRecord
	for _, record := range visibleCapabilityRecords() {
		if record.matchesSearch(normalized) {
			matches = append(matches, record)
		}
	}
	return matches
}

func visibleCapabilityRecords() []capabilityRecord {
	records := capabilityCatalog()
	visible := make([]capabilityRecord, 0, len(records))
	for _, record := range records {
		if record.IncludeInCompact && !record.Hidden && !record.Deprecated {
			visible = append(visible, record)
		}
	}
	return visible
}

func suggestCapabilityCommands(commandPath string) []string {
	normalized := normalizeCapabilityQuery(commandPath)
	if normalized == "" {
		return nil
	}

	suggestions := make([]string, 0, 3)
	for _, record := range capabilityCatalog() {
		candidates := append([]string{record.Command}, record.Aliases...)
		for _, candidate := range candidates {
			candidate = normalizeCapabilityQuery(candidate)
			if strings.HasPrefix(candidate, normalized) || strings.HasPrefix(normalized, candidate) || strings.Contains(candidate, normalized) {
				suggestions = appendUniqueSuggestion(suggestions, record.Command)
			}
			if len(suggestions) >= 3 {
				return suggestions
			}
		}
	}
	return suggestions
}

func appendUniqueSuggestion(suggestions []string, command string) []string {
	for _, existing := range suggestions {
		if existing == command {
			return suggestions
		}
	}
	return append(suggestions, command)
}

func (record capabilityRecord) compact() capabilityCompactRecord {
	return capabilityCompactRecord{
		Command:         record.Command,
		Category:        record.Category,
		Summary:         record.Summary,
		MutationLevel:   record.MutationLevel,
		NetworkUse:      record.NetworkUse,
		FileWrites:      record.FileWrites,
		GitMutation:     record.GitMutation,
		Hidden:          record.Hidden,
		Deprecated:      record.Deprecated,
		ImportantFlags:  cloneCapabilityFlags(record.ImportantFlags),
		RelatedCommands: cloneRelatedCommands(record.RelatedCommands),
	}
}

func (record capabilityRecord) detail() capabilityDetailRecord {
	return capabilityDetailRecord{
		Command:              record.Command,
		Category:             record.Category,
		Summary:              record.Summary,
		MutationLevel:        record.MutationLevel,
		NetworkUse:           record.NetworkUse,
		FileWrites:           record.FileWrites,
		GitMutation:          record.GitMutation,
		Hidden:               record.Hidden,
		Deprecated:           record.Deprecated,
		DeprecationNote:      record.DeprecationNote,
		Aliases:              cloneStrings(record.Aliases),
		ImportantFlags:       cloneCapabilityFlags(record.ImportantFlags),
		RelatedCommands:      cloneRelatedCommands(record.RelatedCommands),
		WhenToUse:            cloneStrings(record.WhenToUse),
		WhenNotToUse:         cloneStrings(record.WhenNotToUse),
		Examples:             cloneStrings(record.Examples),
		Caveats:              cloneStrings(record.Caveats),
		DetailedFlagBehavior: cloneCapabilityFlags(record.DetailedFlagBehavior),
	}
}

func lessCapabilityRecord(left, right capabilityRecord) bool {
	leftCategory := capabilityCategoryOrder[left.Category]
	rightCategory := capabilityCategoryOrder[right.Category]
	if leftCategory != rightCategory {
		return leftCategory < rightCategory
	}

	leftOrder := capabilityCommandOrder(left.Command)
	rightOrder := capabilityCommandOrder(right.Command)
	if leftOrder != rightOrder {
		return leftOrder < rightOrder
	}
	return left.Command < right.Command
}

func capabilityCommandOrder(commandPath string) int {
	rootName := commandPath
	if fields := strings.Fields(commandPath); len(fields) > 0 {
		rootName = fields[0]
	}
	if order, ok := commandOrder[rootName]; ok {
		return order
	}
	return 1000
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func cloneCapabilityFlags(values []capabilityFlag) []capabilityFlag {
	if len(values) == 0 {
		return []capabilityFlag{}
	}
	return append([]capabilityFlag(nil), values...)
}

func cloneRelatedCommands(values []capabilityRelatedCommand) []capabilityRelatedCommand {
	if len(values) == 0 {
		return []capabilityRelatedCommand{}
	}
	return append([]capabilityRelatedCommand(nil), values...)
}

func (record capabilityRecord) matchesCommandPath(commandPath string) bool {
	if normalizeCapabilityQuery(record.Command) == commandPath {
		return true
	}
	for _, alias := range record.Aliases {
		if normalizeCapabilityQuery(alias) == commandPath {
			return true
		}
	}
	return false
}

func (record capabilityRecord) matchesSearch(query string) bool {
	searchable := []string{
		record.Command,
		record.Category,
		record.Summary,
		record.MutationLevel,
		record.NetworkUse.Summary,
		record.NetworkUse.FlagDependent,
		record.FileWrites.Summary,
		record.FileWrites.FlagDependent,
		record.GitMutation.Summary,
		record.GitMutation.FlagDependent,
	}
	searchable = append(searchable, record.Aliases...)
	searchable = append(searchable, record.WhenToUse...)
	searchable = append(searchable, record.WhenNotToUse...)
	searchable = append(searchable, record.Examples...)
	searchable = append(searchable, record.Caveats...)
	for _, flag := range record.ImportantFlags {
		searchable = append(searchable, flag.Name, flag.Summary, flag.Safety)
	}
	for _, related := range record.RelatedCommands {
		searchable = append(searchable, related.Command, related.Note)
	}

	for _, value := range searchable {
		if strings.Contains(normalizeCapabilityQuery(value), query) {
			return true
		}
	}
	return false
}

func normalizeCapabilityQuery(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

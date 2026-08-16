package releaseprompt

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/promptdoc"
	"gopkg.in/yaml.v3"
)

func Render(config Config) (string, error) {
	if err := Validate(config); err != nil {
		return "", err
	}
	document := promptdoc.New()
	addObjectiveAndScope(document, config)
	addContextContract(document, config)
	addHardRules(document, config)
	addDiscoveryAndGraph(document, config)
	addPreparationAndCompatibility(document, config)
	addReleaseAndInfrastructure(document, config)
	addVerificationAndFailure(document, config)
	addCompletionAndReport(document, config)
	prompt := document.String() + "\n"
	if strings.Contains(prompt, "{{") || strings.Contains(prompt, "}}") {
		return "", fmt.Errorf("rendered prompt contains an unresolved placeholder")
	}
	return prompt, nil
}

func RenderDryRun(config Config, prompt string) (string, error) {
	resolved, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("encode resolved configuration: %w", err)
	}
	document := promptdoc.New()
	document.Heading(1, "Release Orchestration Dry Run")
	document.Paragraph("Read-only configuration resolution and prompt rendering. No release, repository, GitHub, deployment, infrastructure, agent, or clipboard mutation was performed by this generator.")
	document.Heading(2, "Resolved Configuration")
	document.CodeBlock("yaml", safeCode(string(resolved)))
	document.Heading(2, "Generated Prompt")
	document.Raw(strings.TrimSpace(prompt))
	return document.String() + "\n", nil
}

func addObjectiveAndScope(document *promptdoc.Document, config Config) {
	document.Heading(1, "Coding-Agent Release Orchestration")
	document.Heading(2, "Objective")
	document.Paragraph(fmt.Sprintf("Safely deliver all relevant open pull requests across `%s` through **discovery → dependency analysis → Global Release Graph → remediation → validation → merge → deployment → production verification → integrated-system verification** until every intended safe change reaches its required environment and the integrated system is verified healthy.", inline(config.Project)))
	document.Paragraph("Optimize in this order: correctness, zero downtime, data integrity, security, backward and forward compatibility, production stability, dependency-safe ordering, failure isolation, efficient tool and API usage, then throughput.")
	document.Paragraph("A PR is not complete when merged. It is complete only when its intended behavior is deployed and verified.")

	document.Heading(2, "Scope")
	document.Paragraph("These repositories define the minimum required release scope:")
	document.CodeBlock("text", repositoryInventory(config.Repositories))
	document.Paragraph("Scope expansion policy: `" + inline(config.ScopeExpansion) + "`")
	document.Paragraph(scopePolicy(config.ScopeExpansion))
	document.Paragraph("Organization context:")
	document.CodeBlock("text", safeCode(config.Organization))
	document.Paragraph("Feature or release context:")
	document.CodeBlock("text", safeCode(config.FeatureContext))
	document.Paragraph("Use local state, issues, PR descriptions and discussions, code and history, runtime relationships, and deployment relationships to find permitted related work. Never include an unrelated PR merely because it is open.")
}

func addContextContract(document *promptdoc.Document, config Config) {
	document.Heading(2, "Repository-Local Evidence Contract")
	document.Paragraph("Before mutation in each Kit-managed repository:")
	document.OrderedList(1,
		"Run `kit capabilities context resolve --json` and inspect capabilities for every other Kit command before relying on it.",
		"When `docs/references/workflows/release-orchestration.md` is present, run `kit context resolve --workflow release-orchestration --json` with relevant feature and path hints.",
		"If that workflow is absent, run `kit context resolve --workflow implementation-delivery --json` and also resolve `cross-repository-program-coordination` when its documented trigger applies.",
		"Before any pull-request merge mutation, resolve `kit context resolve --workflow pull-request-merge --json` and load the selected merge-authority, identity, readiness, repository-policy, and post-merge evidence.",
		"Load every required selected artifact in order. Treat blocked resolution as an evidence gap and do not guess.",
		"Report the missing release workflow as scaffold drift. Do not refresh or rewrite repository guidance during the release unless that mutation is separately authorized and placed in the graph.",
	)
	document.Paragraph("Merge authority comes only from a direct merge request or an accepted bounded merge plan. PR-delivery consent, successful checks, subagent assignment, and a program ledger do not create it. Preserve compatible authority across revalidation and routine retries; obtain follow-up authorization only for material scope expansion or changed effects.")
	document.Paragraph("Cross-repository coordination is conditional: create or adopt one canonical `docs/programs/<program>/PROGRAM.md` only when the work spans multiple repositories and also has dependent deliverables, staged deployment or activation, or expected agent/session handoff. The ledger records and reconciles the authorized PR set and frontier but never creates authority. Otherwise keep the Global Release Graph in the task's working report without inventing a program ledger.")
	if additionalRules := strings.TrimSpace(config.AdditionalHardRules); additionalRules != "" {
		document.Paragraph(inline(additionalRules))
	}
}

func addHardRules(document *promptdoc.Document, config Config) {
	document.Heading(2, "Hard Rules")
	document.OrderedList(1,
		"Construct the initial Global Release Graph before any source, GitHub, deployment, or infrastructure mutation.",
		"Identify the direct request or accepted bounded plan that authorizes the exact merge set before treating any node as authorized.",
		"The graph, not PR age, number, repository order, or apparent urgency, determines release ordering.",
		"Treat merge order and deployment order as separate decisions.",
		"Prefer local Git and repository evidence; use `"+inline(config.SourceControl.CLI)+"` conservatively with batched JSON, cached results, and bounded targeted refreshes. If it is unavailable or unauthenticated, stop and establish the repository-approved authenticated client rather than guessing remote state.",
		"Address all actionable human and automated review feedback and resolve merge conflicts by intended behavior before merge.",
		"Green CI alone is insufficient; review the integrated diff, contracts, migrations, runtime prerequisites, and mixed-version behavior.",
		"Assume rolling mixed-version operation unless the deployment system proves an atomic stronger guarantee.",
		"Use additive contracts and expand-and-contract schema evolution unless stronger repository-specific guarantees are proven.",
		"Parallelize independent discovery, remediation, and testing; serialize overlapping files, shared mutation surfaces, and potentially interacting production changes.",
		"Keep one accountable supervisor for scope, graph state, integration, validation, delivery gates, and reporting.",
		"Verify production from actual runtime identity and state, not from merge or deployment command success.",
		"Only an authorized node with exact current evidence in state `MERGE_READY` may enter the reconciled merge frontier; `BLOCKED` and `UNKNOWN` never do.",
		"Convert failures requiring source changes into normal corrective PRs and first-class graph nodes; obtain merge authority for corrective or newly discovered PRs not already inside the accepted bounded plan.",
		"Run the configured final integration suite after all intended units pass individual environment verification.",
		"Never trade correctness, availability, security, compatibility, or data integrity for completion.",
		"Preserve every repository's local issue, branch, worktree, delivery, testing, infrastructure, and approval rules.",
		"Do not merge, deploy, mutate infrastructure, or resolve review state when authority, target identity, ownership, evidence, repository policy, or approval is ambiguous.",
	)
}

func repositoryInventory(repositories []Repository) string {
	var lines []string
	for _, repository := range repositories {
		identity := repository.GitHub
		if identity == "" {
			identity = "infer GitHub identity from repository evidence"
		}
		branch := repository.DefaultBranch
		if branch == "" {
			branch = "infer default branch before mutation"
		}
		workflow := "not Kit-managed"
		if repository.KitManaged && repository.ReleaseWorkflowPresent {
			workflow = "Kit-managed; release-orchestration workflow present"
		} else if repository.KitManaged {
			workflow = "Kit-managed; release-orchestration workflow missing (use fallback)"
		}
		verification := repository.VerificationHint
		if verification == "" {
			verification = "discover"
		}
		integration := repository.IntegrationSuiteHint
		if integration == "" {
			integration = "discover"
		}
		lines = append(lines, fmt.Sprintf("- %s | %s | default=%s | %s | verify=%s | integration=%s", repository.Path, identity, branch, workflow, verification, integration))
	}
	return safeCode(strings.Join(lines, "\n"))
}

func scopePolicy(scope string) string {
	switch scope {
	case "strict":
		return "Inspect only the listed repositories. Record an external dependency as a blocker rather than expanding scope without authorization."
	case "organization":
		return "Expand within the resolved organization only when repository or PR evidence shows a material release relationship; organization membership alone is insufficient."
	default:
		return "Expand only to repositories and PRs materially related by feature, contract, runtime, infrastructure, deployment, or validation evidence."
	}
}

func inline(value string) string {
	value = strings.ReplaceAll(value, "`", "'")
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " ")
}

func safeCode(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "```", "`` `")
}

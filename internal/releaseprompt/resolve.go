package releaseprompt

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

func Resolve(ctx context.Context, input Input, runner Runner) (Config, error) {
	if runner == nil {
		return Config{}, fmt.Errorf("command runner is required")
	}
	repositories, err := discoverRepositories(ctx, input, runner)
	if err != nil {
		return Config{}, err
	}
	config := Config{Repositories: repositories}
	var source ValueSource
	config.Project, source = resolveProject(input, repositories)
	config.record("project", config.Project, source)
	config.Organization, source = resolveOrganization(input.Organization, repositories)
	config.record("organization", config.Organization, source)
	config.ScopeExpansion, source = choose(input.ScopeExpansion, "", "related")
	config.record("scope_expansion", config.ScopeExpansion, source)
	config.FeatureContext, source = choose(input.FeatureContext, "", "Infer the feature and release intent from scoped PRs, linked issues, code, and repository memory.")
	config.record("feature_context", config.FeatureContext, source)

	config.SourceControl.Provider = "github"
	providerSource := SourceDefaulted
	if everyRepositoryHasGitHubIdentity(repositories) {
		providerSource = SourceDiscovered
	}
	config.record("source_control.provider", config.SourceControl.Provider, providerSource)
	config.SourceControl.CLI = "gh"
	cliSource := SourceDefaulted
	if _, err := runner.LookPath("gh"); err == nil {
		cliSource = SourceDiscovered
	}
	config.record("source_control.cli", config.SourceControl.CLI, cliSource)

	if err := resolveInfrastructure(&config, input, discoverInfrastructure(repositories)); err != nil {
		return Config{}, err
	}
	config.Production.Environment, source = choose(input.Environment, "", "production")
	config.record("production.environment", config.Production.Environment, source)
	verificationHint := singleRepositoryHint(repositories, true)
	config.Production.Verification, source = choose(input.ProductionVerification, verificationHint, "auto")
	config.record("production.verification", config.Production.Verification, source)
	integrationHint := singleRepositoryHint(repositories, false)
	config.IntegrationSuite, source = choose(input.IntegrationSuite, integrationHint, "auto")
	config.record("integration_suite", config.IntegrationSuite, source)

	config.DeploymentContext = "Infer deployment ownership, target identity, release tooling, and rollback procedure from each repository before mutation."
	config.ReviewSystems = "Discover human and automated review systems from current pull requests; classify findings before remediation."
	config.RequiredChecks = "Discover repository-required checks and run all applicable code-level and release-specific validation; green CI alone is insufficient."
	config.DatabaseMigrationPolicy = "Use expand-and-contract by default; prove any stronger repository-specific compatibility guarantee before relying on it."
	config.AdditionalHardRules = "Resolve repository-local Kit evidence before mutation, preserve each repository's delivery lane, and apply program coordination only when its documented trigger is satisfied."
	config.FinalReportRequirements = "Separate verified, inferred, not-applicable, unresolved, local, hosted, deployment, and production evidence; never convert an unobserved state into a pass."
	if err := Validate(config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func choose(explicit, discovered, fallback string) (string, ValueSource) {
	if value := strings.TrimSpace(explicit); value != "" {
		return value, SourceExplicit
	}
	if value := strings.TrimSpace(discovered); value != "" {
		return value, SourceDiscovered
	}
	return fallback, SourceDefaulted
}

func resolveProject(input Input, repositories []Repository) (string, ValueSource) {
	if value := strings.TrimSpace(input.Project); value != "" {
		return value, SourceExplicit
	}
	if len(repositories) == 1 {
		return repositories[0].Name, SourceDiscovered
	}
	if root := strings.TrimSpace(input.Root); root != "" {
		if absolute, err := filepath.Abs(root); err == nil && filepath.Base(absolute) != "." {
			return filepath.Base(absolute), SourceDiscovered
		}
	}
	if organization, source := resolveOrganization("", repositories); source == SourceDiscovered {
		return organization, SourceDiscovered
	}
	return "multi-repository-release", SourceDefaulted
}

func resolveOrganization(explicit string, repositories []Repository) (string, ValueSource) {
	if value := strings.TrimSpace(explicit); value != "" {
		return value, SourceExplicit
	}
	var owner string
	for _, repository := range repositories {
		parts := strings.Split(repository.GitHub, "/")
		if len(parts) != 2 {
			return "Infer organization from repository remotes and related release evidence.", SourceDefaulted
		}
		if owner == "" {
			owner = parts[0]
		} else if owner != parts[0] {
			return "Infer organization boundaries from repository remotes and related release evidence.", SourceDefaulted
		}
	}
	if owner != "" {
		return owner, SourceDiscovered
	}
	return "Infer organization from repository remotes and related release evidence.", SourceDefaulted
}

func everyRepositoryHasGitHubIdentity(repositories []Repository) bool {
	for _, repository := range repositories {
		if repository.GitHub == "" {
			return false
		}
	}
	return len(repositories) > 0
}

func singleRepositoryHint(repositories []Repository, verification bool) string {
	if len(repositories) != 1 {
		return ""
	}
	hint := repositories[0].IntegrationSuiteHint
	if verification {
		hint = repositories[0].VerificationHint
	}
	if hint == "" {
		return ""
	}
	return "instruction:run " + hint + " from repository " + repositories[0].Name
}

func resolveInfrastructure(config *Config, input Input, hint infrastructureHint) error {
	var source ValueSource
	config.Infrastructure.Mode, source = choose(input.InfrastructureMode, hint.Mode, "auto")
	config.record("infrastructure.mode", config.Infrastructure.Mode, source)
	if config.Infrastructure.Mode == "none" {
		if strings.TrimSpace(input.InfrastructureProvider) != "" || strings.TrimSpace(input.InfrastructureCLI) != "" {
			return fmt.Errorf("infrastructure mode none cannot be combined with --infra-provider or --infra-cli")
		}
		config.Infrastructure.Provider = "not applicable"
		config.Infrastructure.CLI = "not applicable"
		config.record("infrastructure.provider", config.Infrastructure.Provider, SourceDefaulted)
		config.record("infrastructure.cli", config.Infrastructure.CLI, SourceDefaulted)
	} else {
		config.Infrastructure.Provider, source = choose(input.InfrastructureProvider, hint.Provider, "infer from repository and provider state")
		config.record("infrastructure.provider", config.Infrastructure.Provider, source)
		cliHint := hint.CLI
		if cliHint == "" {
			cliHint = providerCLI(config.Infrastructure.Provider)
		}
		config.Infrastructure.CLI, source = choose(input.InfrastructureCLI, cliHint, "infer from repository tooling and provider source of truth")
		config.record("infrastructure.cli", config.Infrastructure.CLI, source)
		if strings.TrimSpace(input.InfrastructureMode) != "" && requiresConcreteInfrastructure(config.Infrastructure.Mode) {
			if strings.HasPrefix(config.Infrastructure.Provider, "infer from repository") {
				return fmt.Errorf("infrastructure mode %s requires --infra-provider or a discoverable provider", config.Infrastructure.Mode)
			}
			if strings.HasPrefix(config.Infrastructure.CLI, "infer from repository") {
				return fmt.Errorf("infrastructure mode %s requires --infra-cli or a discoverable CLI", config.Infrastructure.Mode)
			}
		}
	}
	config.Infrastructure.IdentityCheck = identityCheck(config.Infrastructure.Provider)
	config.Infrastructure.Policy = "IAM, network, KMS, secrets, database schema or data-loss, infrastructure create/replace/delete, destructive, and nonstandard deployment changes require their own complete approval boundary and post-change proof. Standing authority covers only recorded standard deployment workflows on already-provisioned targets."
	return nil
}

func providerCLI(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "aws":
		return "aws"
	case "azure":
		return "az"
	case "gcp", "google cloud":
		return "gcloud"
	default:
		return ""
	}
}

func identityCheck(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "not applicable":
		return "not applicable"
	case "aws":
		return "When .kit.yaml enables AWS, run kit aws verify; otherwise run aws sts get-caller-identity and verify account, ARN, region, and environment."
	case "azure":
		return "Run az account show and verify tenant, subscription, region, and environment."
	case "gcp", "google cloud":
		return "Run gcloud auth list and gcloud config list, then verify account, project, region, and environment."
	default:
		return "Use the repository's strongest provider-specific identity command and verify account, project or subscription, region, cluster when applicable, and environment."
	}
}

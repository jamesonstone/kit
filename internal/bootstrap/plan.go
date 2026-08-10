package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jamesonstone/kit/internal/registry"
)

type starterFile struct {
	path     string
	strategy string
	content  string
}

var createIfMissingFiles = []starterFile{
	{"Makefile", "create-if-missing", makefileStarter},
	{".coderabbit.yaml", "create-if-missing", codeRabbitStarter},
	{".github/pull_request_template.md", "create-if-missing", pullRequestStarter},
	{".github/workflows/auto-assign.yml", "create-if-missing", autoAssignStarter},
	{"docs/PROJECT_PROGRESS_SUMMARY.md", "create-if-missing", progressSummaryStarter},
	{"docs/agents/WORKFLOWS.md", "create-if-missing", workflowsStarter},
	{"docs/agents/RLM.md", "create-if-missing", rlmStarter},
	{"docs/agents/TOOLING.md", "create-if-missing", toolingStarter},
	{"docs/agents/GUARDRAILS.md", "create-if-missing", guardrailsStarter},
	{"docs/references/README.md", "create-if-missing", referencesReadmeStarter},
	{"docs/references/testing.md", "create-if-missing", testingStarter},
	{"docs/references/tooling.md", "create-if-missing", projectToolingStarter},
	{"docs/references/external-systems.md", "create-if-missing", externalSystemsStarter},
	{"docs/references/worktrees.md", "create-if-missing", worktreesStarter},
}

var routingStarters = map[string]string{
	"AGENTS.md":                       providerRouter,
	"CLAUDE.md":                       providerRouter,
	".github/copilot-instructions.md": providerRouter,
	"docs/agents/README.md":           agentsReadmeStarter,
}

func BuildPlan(
	ctx context.Context,
	root string,
	source registry.Source,
	sourceConfig registry.SourceConfig,
	userDefaults UserConfig,
	userConfig UserConfigDisposition,
) (Plan, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Plan{}, err
	}
	registryPlan, fresh, err := planRegistry(ctx, root, source, sourceConfig)
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{
		SchemaVersion: PlanSchemaVersion, State: "current", Fresh: fresh,
		Registry: registryPlan, UserConfig: userConfig, root: root,
		userDefaults: userDefaults,
		Prompt:       repositoryBootstrapPrompt(),
		NextActions: []string{
			"review the complete bootstrap file dispositions",
			"give the prompt to a coding agent to populate only repository-verified context",
		},
	}
	if err := planProjectFiles(&plan); err != nil {
		return Plan{}, err
	}
	if err := validatePlan(plan); err != nil {
		return Plan{}, err
	}
	if hasPlannedChanges(plan) {
		plan.State = "planned"
	}
	if !hasRecord(plan.Registry.Config.Registry.Artifacts, registry.KindWorkflow, "repository-bootstrap") {
		plan.Diagnostics = append(plan.Diagnostics,
			"workflow/repository-bootstrap is not installed; preview registry drift with `kit reconcile`")
	}
	return plan, nil
}

func planRegistry(
	ctx context.Context,
	root string,
	source registry.Source,
	sourceConfig registry.SourceConfig,
) (registry.Plan, bool, error) {
	if _, err := os.Stat(filepath.Join(root, registry.ProjectFile)); os.IsNotExist(err) {
		if source == nil {
			return registry.Plan{}, false, fmt.Errorf("registry source is required for fresh initialization")
		}
		plan, buildErr := registry.BuildInitPlan(ctx, root, source, sourceConfig)
		return plan, true, buildErr
	} else if err != nil {
		return registry.Plan{}, false, err
	}
	config, migration, err := registry.LoadProject(root)
	if err != nil {
		return registry.Plan{}, false, err
	}
	if migration {
		return registry.Plan{}, false, fmt.Errorf("schema-v1 project requires `kit reconcile --json --diff` before initialization")
	}
	plan := registry.Plan{State: "current", Revision: config.Registry.Source.Revision, Config: config}
	for _, record := range config.Registry.Artifacts {
		plan.Artifacts = append(plan.Artifacts, registry.ArtifactPlan{
			Kind: record.Kind, Slug: record.Slug, Path: record.Path,
			State: record.State, Action: "none",
		})
	}
	before, _, err := registry.ReadOptional(root, registry.ProjectFile)
	if err != nil {
		return registry.Plan{}, false, err
	}
	after, err := registry.MarshalProject(config)
	if err != nil {
		return registry.Plan{}, false, err
	}
	if before != string(after) {
		plan.Changes = append(plan.Changes, registry.Change{
			Path: registry.ProjectFile, Action: "update", Before: before, After: string(after),
		})
	}
	return plan, false, nil
}

func planProjectFiles(plan *Plan) error {
	projectDisposition := FileDisposition{
		Path: registry.ProjectFile, Strategy: "schema-v2-provenance",
		State: "preserved", Action: "none",
	}
	if change := changeForPath(plan.Registry.Changes, registry.ProjectFile); change != nil {
		projectDisposition.State, projectDisposition.Action = "planned", change.Action
	}
	plan.Files = append(plan.Files, projectDisposition)
	if err := planEnvironment(plan); err != nil {
		return err
	}
	if err := planDirectory(plan, "docs/specs"); err != nil {
		return err
	}
	if err := planGitignore(plan); err != nil {
		return err
	}
	for _, starter := range createIfMissingFiles {
		if starter.path == ".github/workflows/auto-assign.yml" {
			starter.content = buildAutoAssign(plan.userDefaults.GitHub.DefaultAssignees)
		}
		if err := planCreateIfMissing(plan, starter); err != nil {
			return err
		}
	}
	if err := planManagedFile(plan, "README.md", "bounded-badges-and-maintainers", readmeContent); err != nil {
		return err
	}
	if err := planManagedFile(plan, "docs/CONSTITUTION.md", "bounded-baseline", constitutionContent); err != nil {
		return err
	}
	return planRoutingFiles(plan)
}

func planCreateIfMissing(plan *Plan, starter starterFile) error {
	before, exists, err := registry.ReadOptional(plan.root, starter.path)
	if err != nil {
		return err
	}
	disposition := FileDisposition{Path: starter.path, Strategy: starter.strategy, State: "preserved", Action: "none"}
	if !exists {
		disposition.State, disposition.Action = "planned", "create"
		plan.Registry.Changes = append(plan.Registry.Changes, registry.Change{
			Path: starter.path, Action: "create", Before: before, After: starter.content,
		})
	}
	plan.Files = append(plan.Files, disposition)
	return nil
}

func hasRecord(records []registry.ArtifactRecord, kind, slug string) bool {
	_, found := registry.RecordByKey(records, kind, slug)
	return found
}

func hasPlannedChanges(plan Plan) bool {
	return len(plan.Registry.Changes) > 0 || len(plan.exclusive) > 0 ||
		(plan.UserConfig.Action != "" && plan.UserConfig.Action != "none") ||
		(len(plan.Directories) > 0 && plan.Directories[0].Action != "none")
}

func validatePlan(plan Plan) error {
	seen := map[string]bool{}
	for _, change := range plan.Registry.Changes {
		if seen[change.Path] {
			return fmt.Errorf("bootstrap plans duplicate change for %s", change.Path)
		}
		seen[change.Path] = true
	}
	sort.Slice(plan.Files, func(i, j int) bool { return plan.Files[i].Path < plan.Files[j].Path })
	sort.Slice(plan.Registry.Changes, func(i, j int) bool {
		return plan.Registry.Changes[i].Path < plan.Registry.Changes[j].Path
	})
	return nil
}

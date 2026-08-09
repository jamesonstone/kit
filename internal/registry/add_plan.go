package registry

import (
	"context"
	"errors"
	"fmt"
)

var ErrArtifactNotFound = errors.New("registry artifact not found")

func BuildAddPlan(ctx context.Context, root string, source Source, kind, slug string) (Plan, error) {
	cfg, migration, err := LoadProject(root)
	if err != nil {
		return Plan{}, err
	}
	if migration {
		return Plan{}, fmt.Errorf("migrate %s with `kit reconcile --apply` before adding artifacts", ProjectFile)
	}
	loaded, revision, err := loadVisibleArtifacts(ctx, source, cfg.Registry.Source)
	if err != nil {
		return Plan{}, err
	}
	for _, item := range loaded {
		if item.Catalog.Kind != kind || item.Catalog.Slug != slug {
			continue
		}
		record, tracked := RecordByKey(cfg.Registry.Artifacts, kind, slug)
		artifactPlan, changes, updated := planArtifact(root, cfg.Registry.Source, revision, item, record, tracked, false)
		cfg.Registry.Artifacts = UpsertRecord(cfg.Registry.Artifacts, updated)
		plan := Plan{State: "planned", Revision: revision, Artifacts: []ArtifactPlan{artifactPlan}, Changes: changes, Config: cfg}
		if artifactPlan.State == StateConflict {
			plan.State = "attention-needed"
		}
		return finishAddPlan(root, plan)
	}
	return Plan{}, fmt.Errorf("%w: %s", ErrArtifactNotFound, ArtifactKey(kind, slug))
}

func BuildLocalRulesetPlan(root, slug string) (Plan, error) {
	if !ValidSlug(slug) {
		return Plan{}, fmt.Errorf("ruleset slug %q is invalid", slug)
	}
	cfg, migration, err := LoadProject(root)
	if err != nil {
		return Plan{}, err
	}
	if migration {
		return Plan{}, fmt.Errorf("migrate %s with `kit reconcile --apply` before adding artifacts", ProjectFile)
	}
	path := "docs/references/rules/" + slug + ".md"
	if _, exists, err := ReadOptional(root, path); err != nil {
		return Plan{}, err
	} else if exists {
		return Plan{}, fmt.Errorf("ruleset %q already exists", slug)
	}
	content := localRulesetTemplate(slug)
	record := ArtifactRecord{
		Kind: KindRuleset, Slug: slug, Description: "Project-local " + slug + " rules",
		Path: path, Version: 1, ReadPolicy: "conditional", State: StateLocalCustom,
		ContentHash: HashContent(content),
	}
	cfg.Registry.Artifacts = UpsertRecord(cfg.Registry.Artifacts, record)
	plan := Plan{
		State: "planned", Config: cfg,
		Artifacts: []ArtifactPlan{{Kind: KindRuleset, Slug: slug, Path: path, State: StateLocalCustom, Action: "create"}},
		Changes:   []Change{{Path: path, Action: "create", After: content}},
	}
	return finishAddPlan(root, plan)
}

func finishAddPlan(root string, plan Plan) (Plan, error) {
	routing, err := PlanRouting(root, plan.Config.Registry.Artifacts)
	if err != nil {
		return Plan{}, err
	}
	plan.Changes = append(plan.Changes, routing...)
	before, _, err := ReadOptional(root, ProjectFile)
	if err != nil {
		return Plan{}, err
	}
	after, err := MarshalProject(plan.Config)
	if err != nil {
		return Plan{}, err
	}
	if before != string(after) {
		plan.Changes = append(plan.Changes, Change{Path: ProjectFile, Action: "update", Before: before, After: string(after)})
	}
	return plan, nil
}

func localRulesetTemplate(slug string) string {
	return "---\n" +
		"kind: ruleset\n" +
		"slug: " + slug + "\n" +
		"description: Project-local " + slug + " rules\n" +
		"status: active\n" +
		"registry_scope: local\n" +
		"applies_to:\n  - coding-agent\n" +
		"read_policy_default: conditional\n" +
		"---\n\n# Ruleset: " + slug + "\n\n" +
		"## Purpose\n\nDescribe the local rule purpose.\n\n" +
		"## Applies When\n\nDescribe when agents load this rule.\n\n" +
		"## Rules\n\n- Replace this starter with explicit project rules.\n\n" +
		"## Anti-Patterns\n\n- Do not apply the rule outside its stated scope.\n\n" +
		"## Verification\n\n- Confirm the rule was followed when applicable.\n\n" +
		"## Examples\n\nNo examples recorded yet.\n"
}

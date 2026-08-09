package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func BuildInitPlan(ctx context.Context, root string, source Source, sourceCfg SourceConfig) (Plan, error) {
	if _, err := os.Stat(filepath.Join(root, ProjectFile)); err == nil {
		return Plan{}, fmt.Errorf("%s already exists; use `kit reconcile`", ProjectFile)
	} else if !os.IsNotExist(err) {
		return Plan{}, err
	}
	loaded, revision, err := loadVisibleArtifacts(ctx, source, sourceCfg)
	if err != nil {
		return Plan{}, err
	}
	cfg := NewProjectConfig(sourceCfg)
	cfg.Registry.Source.Revision = revision
	plan := Plan{State: "planned", Revision: revision, Config: cfg}
	for _, item := range loaded {
		before, exists, err := ReadOptional(root, item.Catalog.TargetPath)
		if err != nil {
			return Plan{}, err
		}
		state, action, finalContent := StateManaged, "none", item.Content
		if exists && HashContent(before) != item.Catalog.Digest {
			state, finalContent = StateLocalCustom, before
			plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("%s exists with local custom content", item.Catalog.TargetPath))
		} else if !exists {
			action = "create"
			plan.Changes = append(plan.Changes, Change{Path: item.Catalog.TargetPath, Action: action, After: item.Content})
		}
		record := recordFromRemote(cfg.Registry.Source, revision, item, finalContent, state)
		cfg.Registry.Artifacts = UpsertRecord(cfg.Registry.Artifacts, record)
		plan.Artifacts = append(plan.Artifacts, ArtifactPlan{
			Kind: item.Catalog.Kind, Slug: item.Catalog.Slug, Path: item.Catalog.TargetPath, State: state, Action: action,
		})
	}
	routingChanges, err := PlanRouting(root, cfg.Registry.Artifacts)
	if err != nil {
		return Plan{}, err
	}
	plan.Changes = append(plan.Changes, routingChanges...)
	configContent, err := MarshalProject(cfg)
	if err != nil {
		return Plan{}, err
	}
	plan.Changes = append(plan.Changes, Change{Path: ProjectFile, Action: "create", After: string(configContent)})
	plan.Config = cfg
	return plan, nil
}

package registry

import (
	"context"
	"fmt"
	"sort"
)

func BuildReconcilePlan(ctx context.Context, root string, source Source, accepts map[string]bool) (Plan, error) {
	cfg, migration, err := LoadProject(root)
	if err != nil {
		return Plan{}, err
	}
	loaded, revision, err := loadVisibleArtifacts(ctx, source, cfg.Registry.Source)
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{State: "current", Migration: migration, Revision: revision, Config: cfg}
	remoteKeys := map[string]bool{}
	for _, item := range loaded {
		key := ArtifactKey(item.Catalog.Kind, item.Catalog.Slug)
		remoteKeys[key] = true
		record, tracked := RecordByKey(cfg.Registry.Artifacts, item.Catalog.Kind, item.Catalog.Slug)
		if tracked && record.Path != "" && record.Path != item.Catalog.TargetPath {
			moveChanges, diagnostic, blocked := planMovedPath(root, record, item.Catalog.TargetPath)
			plan.Changes = append(plan.Changes, moveChanges...)
			if diagnostic != "" {
				plan.Diagnostics = append(plan.Diagnostics, diagnostic)
			}
			if blocked {
				plan.State = "attention-needed"
			}
		}
		artifactPlan, changes, updated := planArtifact(root, cfg.Registry.Source, revision, item, record, tracked, accepts[key])
		plan.Artifacts = append(plan.Artifacts, artifactPlan)
		plan.Changes = append(plan.Changes, changes...)
		cfg.Registry.Artifacts = UpsertRecord(cfg.Registry.Artifacts, updated)
		if artifactPlan.State == StateConflict {
			plan.State = "attention-needed"
			plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("%s has conflicting sections: %v", key, artifactPlan.Conflicts))
		}
	}
	installed := append([]ArtifactRecord(nil), cfg.Registry.Artifacts...)
	for _, record := range installed {
		key := ArtifactKey(record.Kind, record.Slug)
		if remoteKeys[key] {
			continue
		}
		local, exists, readErr := ReadOptional(root, record.Path)
		if readErr != nil {
			plan.State = "attention-needed"
			plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("inspect retired artifact %s: %v", key, readErr))
			continue
		}
		expectedLocalHash := record.ContentHash
		if expectedLocalHash == "" {
			expectedLocalHash = record.InstalledHash
		}
		if !exists || (record.State == StateManaged && HashContent(local) == expectedLocalHash) {
			action := "none"
			if exists {
				action = "delete"
				plan.Changes = append(plan.Changes, Change{Path: record.Path, Action: action, Before: local})
			}
			plan.Artifacts = append(plan.Artifacts, ArtifactPlan{
				Kind: record.Kind, Slug: record.Slug, Path: record.Path, State: StateMissing, Action: action,
			})
			cfg.Registry.Artifacts = RemoveRecord(cfg.Registry.Artifacts, record.Kind, record.Slug)
			continue
		}
		record.State = StateLocalCustom
		record.ContentHash = HashContent(local)
		cfg.Registry.Artifacts = UpsertRecord(cfg.Registry.Artifacts, record)
		plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("retired artifact %s is preserved as local-custom", key))
	}
	cfg.SchemaVersion = ProjectSchemaVersion
	cfg.Registry.SchemaVersion = CatalogSchemaVersion
	cfg.Registry.Source.Revision = revision
	routingChanges, err := PlanRouting(root, cfg.Registry.Artifacts)
	if err != nil {
		return Plan{}, err
	}
	plan.Changes = append(plan.Changes, routingChanges...)
	beforeConfig, _, err := ReadOptional(root, ProjectFile)
	if err != nil {
		return Plan{}, err
	}
	afterConfig, err := MarshalProject(cfg)
	if err != nil {
		return Plan{}, err
	}
	if beforeConfig != string(afterConfig) {
		plan.Changes = append(plan.Changes, Change{Path: ProjectFile, Action: "update", Before: beforeConfig, After: string(afterConfig)})
	}
	if plan.State == "current" && (len(plan.Changes) > 0 || migration) {
		plan.State = "changes-available"
	}
	switch plan.State {
	case "attention-needed":
		plan.NextActions = []string{
			"review `kit reconcile --json --diff` diagnostics",
			"resolve conflicts manually or accept exact artifacts with `--accept-registry`, then apply explicitly",
		}
	case "changes-available":
		plan.NextActions = []string{
			"review `kit reconcile --json --diff`",
			"apply conflict-free changes with `kit reconcile --apply`",
		}
	}
	plan.Config = cfg
	sort.Slice(plan.Artifacts, func(i, j int) bool {
		return ArtifactKey(plan.Artifacts[i].Kind, plan.Artifacts[i].Slug) < ArtifactKey(plan.Artifacts[j].Kind, plan.Artifacts[j].Slug)
	})
	return plan, nil
}

func planMovedPath(root string, record ArtifactRecord, target string) ([]Change, string, bool) {
	local, exists, err := ReadOptional(root, record.Path)
	key := ArtifactKey(record.Kind, record.Slug)
	if err != nil {
		return nil, fmt.Sprintf("inspect previous path for moved artifact %s: %v", key, err), true
	}
	if !exists {
		return nil, "", false
	}
	expected := record.ContentHash
	if expected == "" {
		expected = record.InstalledHash
	}
	if record.State == StateManaged && HashContent(local) == expected {
		return []Change{{Path: record.Path, Action: "delete", Before: local}}, "", false
	}
	return nil, fmt.Sprintf("moved artifact %s preserves customized previous path %s while installing %s", key, record.Path, target), false
}

func planArtifact(
	root string,
	source SourceConfig,
	revision string,
	remote loadedArtifact,
	record ArtifactRecord,
	tracked bool,
	accept bool,
) (ArtifactPlan, []Change, ArtifactRecord) {
	artifact := remote.Catalog
	local, exists, readErr := ReadOptional(root, artifact.TargetPath)
	result := ArtifactPlan{Kind: artifact.Kind, Slug: artifact.Slug, Path: artifact.TargetPath, State: StateManaged, Action: "none"}
	if readErr != nil {
		result.State, result.Action, result.Conflicts = StateConflict, "blocked", []string{readErr.Error()}
		if tracked {
			record.State = StateConflict
			return result, nil, record
		}
		return result, nil, recordFromRemote(source, revision, remote, remote.Content, StateConflict)
	}
	if accept || !exists {
		action := "update"
		if !exists {
			action = "create"
		}
		result.Action = action
		updated := recordFromRemote(source, revision, remote, remote.Content, StateManaged)
		return result, []Change{{Path: artifact.TargetPath, Action: action, Before: local, After: remote.Content}}, updated
	}
	if HashContent(local) == artifact.Digest {
		return result, nil, recordFromRemote(source, revision, remote, local, StateManaged)
	}
	if !tracked {
		result.State = StateLocalCustom
		return result, nil, recordFromRemote(source, revision, remote, local, StateLocalCustom)
	}
	return reconcileTrackedArtifact(source, revision, remote, record, local, result)
}

func reconcileTrackedArtifact(
	source SourceConfig,
	revision string,
	remote loadedArtifact,
	record ArtifactRecord,
	local string,
	result ArtifactPlan,
) (ArtifactPlan, []Change, ArtifactRecord) {
	localHash := HashContent(local)
	localChanged := localHash != record.InstalledHash
	remoteChanged := record.InstalledHash != remote.Catalog.Digest
	switch {
	case !localChanged && !remoteChanged:
		return result, nil, recordFromRemote(source, revision, remote, local, StateManaged)
	case !localChanged && remoteChanged:
		result.Action = "update"
		updated := recordFromRemote(source, revision, remote, remote.Content, StateManaged)
		return result, []Change{{Path: remote.Catalog.TargetPath, Action: "update", Before: local, After: remote.Content}}, updated
	case localChanged && !remoteChanged:
		result.State = StateLocalCustom
		return result, nil, recordFromRemote(source, revision, remote, local, StateLocalCustom)
	case len(record.Sections) == 0:
		result.State, result.Action = StateConflict, "blocked"
		result.Conflicts = []string{"base section hashes are unavailable"}
		record.State = StateConflict
		record.ContentHash = localHash
		return result, nil, record
	}
	merged, conflicts, err := MergeSections(record.Sections, local, remote.Content)
	if err != nil {
		conflicts = []string{err.Error()}
	}
	if len(conflicts) > 0 {
		result.State, result.Action, result.Conflicts = StateConflict, "blocked", conflicts
		record.State = StateConflict
		record.ContentHash = localHash
		return result, nil, record
	}
	result.Action = "update"
	result.State = StateLocalCustom
	if HashContent(merged) == remote.Catalog.Digest {
		result.State = StateManaged
	}
	updated := recordFromRemote(source, revision, remote, merged, result.State)
	return result, []Change{{Path: remote.Catalog.TargetPath, Action: "update", Before: local, After: merged}}, updated
}

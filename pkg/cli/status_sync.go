package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

const (
	statusKitManagedStateSynced   = "synced"
	statusKitManagedStateUnsynced = "unsynced"
	statusKitManagedStateStale    = "stale"
	statusKitManagedStateUnknown  = "unknown"
)

type statusKitManagedSummary struct {
	State        string                    `json:"state"`
	SyncChecked  bool                      `json:"sync_checked"`
	ManagedFiles statusManagedFilesSummary `json:"managed_files"`
	Registry     statusRegistrySummary     `json:"registry"`
	Items        []statusKitManagedItem    `json:"items,omitempty"`
	NextActions  []string                  `json:"next_actions,omitempty"`
}

type statusManagedFilesSummary struct {
	Created  int `json:"created"`
	Updated  int `json:"updated"`
	Merged   int `json:"merged"`
	Skipped  int `json:"skipped"`
	Unsynced int `json:"unsynced"`
}

type statusRegistrySummary struct {
	Checked         bool   `json:"checked"`
	SourceRepo      string `json:"source_repo,omitempty"`
	SourceBranch    string `json:"source_branch,omitempty"`
	SourceCommit    string `json:"source_commit,omitempty"`
	Total           int    `json:"total"`
	Managed         int    `json:"managed"`
	Missing         int    `json:"missing"`
	UpdateAvailable int    `json:"update_available"`
	LocalCustom     int    `json:"local_custom"`
	Conflicts       int    `json:"conflicts"`
	Unknown         int    `json:"unknown"`
}

type statusKitManagedItem struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	State  string `json:"state"`
	Detail string `json:"detail,omitempty"`
}

func buildStatusKitManagedSummary(
	ctx context.Context,
	projectRoot string,
	cfg *config.Config,
	syncChecked bool,
) (*statusKitManagedSummary, error) {
	summary := &statusKitManagedSummary{
		State:       statusKitManagedStateUnknown,
		SyncChecked: syncChecked,
		Registry: statusRegistrySummary{
			SourceRepo:   cfg.Registry.Source.Repo,
			SourceBranch: cfg.Registry.Source.Branch,
		},
	}

	changes, err := planStatusManagedFileChanges(projectRoot)
	if err != nil {
		return nil, err
	}
	recordStatusManagedFileChanges(summary, changes)

	if syncChecked {
		if err := recordStatusRemoteRegistry(ctx, projectRoot, summary); err != nil {
			return nil, err
		}
	} else {
		recordStatusLocalRegistry(projectRoot, cfg, summary)
	}

	summary.State = determineStatusKitManagedState(summary)
	summary.NextActions = statusKitManagedNextActions(summary)
	return summary, nil
}

func planStatusManagedFileChanges(projectRoot string) ([]initRefreshFileChange, error) {
	targets := map[string]bool{}
	opts := initRefreshOptions{dryRun: true, outputOnly: true}

	cfg, configChange, err := initRefreshConfig(projectRoot, opts, targets)
	if err != nil {
		return nil, err
	}

	var changes []initRefreshFileChange
	if configChange != nil {
		changes = append(changes, *configChange)
	}
	scaffoldChanges, err := planRefreshInitScaffoldFiles(projectRoot, opts, cfg, targets)
	if err != nil {
		return nil, err
	}
	changes = append(changes, scaffoldChanges...)

	constitutionChange, err := planRefreshInitConstitution(projectRoot, cfg, targets)
	if err != nil {
		return nil, err
	}
	if constitutionChange != nil {
		changes = append(changes, *constitutionChange)
	}

	instructionChanges, err := planRefreshInitInstructionArtifacts(projectRoot, opts, cfg, targets)
	if err != nil {
		return nil, err
	}
	changes = append(changes, instructionChanges...)
	return changes, nil
}

func recordStatusManagedFileChanges(summary *statusKitManagedSummary, changes []initRefreshFileChange) {
	for _, change := range changes {
		switch change.result {
		case instructionFileCreated:
			summary.ManagedFiles.Created++
		case instructionFileUpdated:
			summary.ManagedFiles.Updated++
		case instructionFileMerged:
			summary.ManagedFiles.Merged++
		case instructionFileSkipped:
			summary.ManagedFiles.Skipped++
		}
		if change.result == instructionFileSkipped {
			continue
		}
		summary.ManagedFiles.Unsynced++
		summary.Items = append(summary.Items, statusKitManagedItem{
			Path:  change.relativePath,
			Kind:  "managed-file",
			State: string(change.result),
		})
	}
}

func recordStatusLocalRegistry(projectRoot string, cfg *config.Config, summary *statusKitManagedSummary) {
	for _, artifact := range cfg.Registry.Artifacts {
		if artifact.Kind != rulesetKind {
			continue
		}
		summary.Registry.Total++
		path := artifact.Path
		if path == "" {
			path = rulesetTarget(artifact.Slug)
		}
		if !document.Exists(filepath.Join(projectRoot, filepath.FromSlash(path))) {
			summary.Registry.Missing++
			summary.Items = append(summary.Items, statusKitManagedItem{
				Path:   path,
				Kind:   "registry-ruleset",
				State:  "missing",
				Detail: "tracked in .kit.yaml but missing locally",
			})
			continue
		}
		switch artifact.State {
		case registryArtifactStateManaged:
			summary.Registry.Managed++
		case registryArtifactStateLocalCustom:
			summary.Registry.LocalCustom++
			summary.Items = append(summary.Items, statusKitManagedItem{
				Path:   path,
				Kind:   "registry-ruleset",
				State:  registryArtifactStateLocalCustom,
				Detail: "local custom content is not registry-managed",
			})
		case registryArtifactStateConflict:
			summary.Registry.Conflicts++
			summary.Items = append(summary.Items, statusKitManagedItem{
				Path:   path,
				Kind:   "registry-ruleset",
				State:  registryArtifactStateConflict,
				Detail: "registry refresh previously detected a conflict",
			})
		default:
			summary.Registry.Unknown++
		}
	}
}

func recordStatusRemoteRegistry(ctx context.Context, projectRoot string, summary *statusKitManagedSummary) error {
	registry, err := rulesetRegistryFetcher(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh Kit ruleset registry for status: %w", err)
	}
	registry = projectRulesetRegistry(registry)
	summary.Registry.Checked = true
	summary.Registry.SourceRepo = rulesetRegistryRepoFullName()
	summary.Registry.SourceBranch = rulesetRegistryBranch
	for _, item := range registry {
		if item.SourceCommit != "" {
			summary.Registry.SourceCommit = item.SourceCommit
			break
		}
	}

	entries, err := buildRegistrySelectorEntries(projectRoot, registry)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := rulesetTarget(entry.Registry.Slug)
		summary.Registry.Total++
		if !entry.Installed {
			summary.Registry.Missing++
			summary.Items = append(summary.Items, statusKitManagedItem{
				Path:   path,
				Kind:   "registry-ruleset",
				State:  "missing",
				Detail: "available in registry but not installed locally",
			})
			continue
		}
		recordStatusRemoteRegistryEntry(summary, path, entry.RegistryState)
	}
	return nil
}

func recordStatusRemoteRegistryEntry(summary *statusKitManagedSummary, path string, state string) {
	switch state {
	case registryArtifactStateManaged:
		summary.Registry.Managed++
	case "update-available":
		summary.Registry.UpdateAvailable++
		summary.Items = append(summary.Items, statusKitManagedItem{
			Path:   path,
			Kind:   "registry-ruleset",
			State:  state,
			Detail: "registry has newer managed content",
		})
	case registryArtifactStateLocalCustom:
		summary.Registry.LocalCustom++
		summary.Items = append(summary.Items, statusKitManagedItem{
			Path:   path,
			Kind:   "registry-ruleset",
			State:  state,
			Detail: "local custom content is not registry-managed",
		})
	case registryArtifactStateConflict:
		summary.Registry.Conflicts++
		summary.Items = append(summary.Items, statusKitManagedItem{
			Path:   path,
			Kind:   "registry-ruleset",
			State:  state,
			Detail: "registry refresh detected a conflict",
		})
	default:
		summary.Registry.Unknown++
	}
}

func determineStatusKitManagedState(summary *statusKitManagedSummary) string {
	if summary.Registry.Conflicts+summary.Registry.LocalCustom+summary.Registry.UpdateAvailable > 0 {
		return statusKitManagedStateStale
	}
	if summary.ManagedFiles.Unsynced+summary.Registry.Missing > 0 {
		return statusKitManagedStateUnsynced
	}
	if !summary.SyncChecked {
		return statusKitManagedStateUnknown
	}
	return statusKitManagedStateSynced
}

func statusKitManagedNextActions(summary *statusKitManagedSummary) []string {
	var actions []string
	if summary.ManagedFiles.Unsynced+summary.Registry.Missing > 0 {
		actions = append(actions, "run `kit init --refresh --dry-run --diff` to preview managed-file updates")
	}
	if summary.Registry.Conflicts+summary.Registry.LocalCustom > 0 {
		actions = append(actions, "review registry ruleset notes, then use `kit init --refresh --force` only when accepting registry content is intended")
	}
	if !summary.SyncChecked {
		actions = append(actions, "run `kit status --sync` to fetch the registry and check for remote rule updates")
	}
	if len(actions) == 0 {
		actions = append(actions, "no Kit-managed refresh action needed")
	}
	return actions
}

package cli

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

const (
	statusKitManagedStateCurrent          = "current"
	statusKitManagedStateRefreshAvailable = "refresh_available"
	statusKitManagedStateAttentionNeeded  = "attention_needed"
	statusKitManagedStateUnknown          = "unknown"
	statusKitManagedRefreshTimeout        = 30 * time.Second
)

type statusKitManagedSummary struct {
	State        string                    `json:"state"`
	ManagedFiles statusManagedFilesSummary `json:"managed_files"`
	Registry     statusRegistrySummary     `json:"registry"`
	Items        []statusKitManagedItem    `json:"items,omitempty"`
	NextActions  []string                  `json:"next_actions,omitempty"`
}

type statusManagedFilesSummary struct {
	Created    int    `json:"created"`
	Updated    int    `json:"updated"`
	Merged     int    `json:"merged"`
	Skipped    int    `json:"skipped"`
	Planned    int    `json:"planned_changes"`
	Unchecked  bool   `json:"unchecked,omitempty"`
	CheckError string `json:"check_error,omitempty"`
}

type statusRegistrySummary struct {
	SourceRepo   string `json:"source_repo,omitempty"`
	SourceBranch string `json:"source_branch,omitempty"`
	Total        int    `json:"total"`
	Managed      int    `json:"managed"`
	Missing      int    `json:"missing"`
	LocalCustom  int    `json:"local_custom"`
	Conflicts    int    `json:"conflicts"`
	Unknown      int    `json:"unknown"`
}

type statusKitManagedItem struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	State  string `json:"state"`
	Detail string `json:"detail,omitempty"`
}

type statusManagedFileChangePlan struct {
	changes    []initRefreshFileChange
	unchecked  bool
	checkError string
}

func buildStatusKitManagedSummary(
	projectRoot string,
	cfg *config.Config,
) (*statusKitManagedSummary, error) {
	summary := &statusKitManagedSummary{
		State: statusKitManagedStateCurrent,
		Registry: statusRegistrySummary{
			SourceRepo:   cfg.Registry.Source.Repo,
			SourceBranch: cfg.Registry.Source.Branch,
		},
	}

	filePlan, err := planStatusManagedFileChanges(projectRoot)
	if err != nil {
		return nil, err
	}
	if filePlan.unchecked {
		summary.ManagedFiles.Unchecked = true
		summary.ManagedFiles.CheckError = filePlan.checkError
	} else {
		recordStatusManagedFileChanges(summary, filePlan.changes)
	}

	recordStatusLocalRegistry(projectRoot, cfg, summary)

	summary.State = determineStatusKitManagedState(summary)
	summary.NextActions = statusKitManagedNextActions(summary)
	return summary, nil
}

func planStatusManagedFileChanges(projectRoot string) (statusManagedFileChangePlan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), statusKitManagedRefreshTimeout)
	defer cancel()

	plan, err := buildInitRefreshPlan(ctx, projectRoot, initRefreshOptions{dryRun: true, outputOnly: true})
	if err != nil {
		var registryErr *initRefreshRegistryError
		if errors.As(err, &registryErr) || errors.Is(err, context.DeadlineExceeded) {
			checkError := err.Error()
			if registryErr != nil {
				checkError = registryErr.Error()
			}
			return statusManagedFileChangePlan{unchecked: true, checkError: checkError}, nil
		}
		return statusManagedFileChangePlan{}, err
	}
	return statusManagedFileChangePlan{changes: plan.changes}, nil
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
		summary.ManagedFiles.Planned++
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

func determineStatusKitManagedState(summary *statusKitManagedSummary) string {
	if summary.Registry.Conflicts+summary.Registry.LocalCustom+summary.Registry.Unknown > 0 {
		return statusKitManagedStateAttentionNeeded
	}
	if summary.ManagedFiles.Unchecked {
		return statusKitManagedStateUnknown
	}
	if summary.ManagedFiles.Planned+summary.Registry.Missing > 0 {
		return statusKitManagedStateRefreshAvailable
	}
	return statusKitManagedStateCurrent
}

func statusKitManagedNextActions(summary *statusKitManagedSummary) []string {
	var actions []string
	attentionNeeded := summary.Registry.Conflicts+summary.Registry.LocalCustom+summary.Registry.Unknown > 0
	refreshAvailable := summary.ManagedFiles.Planned+summary.Registry.Missing > 0
	if summary.ManagedFiles.Unchecked {
		actions = append(actions, "managed-file freshness was not checked because the registry was unavailable; rerun `kit status` when registry access is restored")
	}
	if attentionNeeded {
		actions = append(actions, "run `kit reconcile --output-only` to audit local custom, conflicted, or unknown Kit-managed files")
		actions = append(actions, "run `kit reconcile --include-files --force` only when accepting registry content is intended")
	}
	if refreshAvailable {
		actions = append(actions, "run `kit reconcile --include-files --dry-run --diff` to preview managed-file updates")
		actions = append(actions, "run `kit reconcile --include-files` to apply reviewed managed-file updates")
	}
	if len(actions) == 0 {
		actions = append(actions, "no Kit-managed refresh action needed")
	}
	return actions
}

// package state generates non-authoritative agent-readable Kit state.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
	"github.com/jamesonstone/kit/internal/runstore"
	"github.com/jamesonstone/kit/internal/verify"
)

const StatePath = ".kit/state.json"

type SourceFingerprint struct {
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

type FeatureState struct {
	Status             *feature.FeatureStatus `json:"status"`
	NextCommand        string                 `json:"next_command,omitempty"`
	TaskBundles        []verify.TaskBundle    `json:"task_bundles,omitempty"`
	LatestVerification *runstore.IndexEntry   `json:"latest_verification,omitempty"`
	ParseErrors        []string               `json:"parse_errors,omitempty"`
}

type State struct {
	SchemaVersion int                 `json:"schema_version"`
	GeneratedAt   time.Time           `json:"generated_at"`
	Authoritative bool                `json:"authoritative"`
	Sources       []SourceFingerprint `json:"sources"`
	ActiveFeature *FeatureState       `json:"active_feature,omitempty"`
	Features      []FeatureState      `json:"features"`
}

func Generate(projectRoot string, cfg *config.Config) (State, error) {
	specsDir := cfg.SpecsPath(projectRoot)
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return State{}, err
	}
	latestRuns, err := latestRunsByFeature(projectRoot)
	if err != nil {
		return State{}, err
	}
	active, err := feature.FindActiveFeatureWithState(specsDir, cfg)
	if err != nil {
		return State{}, err
	}

	generated := State{
		SchemaVersion: verify.SchemaVersion,
		GeneratedAt:   time.Now().UTC(),
		Authoritative: false,
	}

	for i := range features {
		entry, err := buildFeatureState(projectRoot, &features[i], latestRuns)
		if err != nil {
			return State{}, err
		}
		generated.Features = append(generated.Features, entry)
		if active != nil && active.DirName == features[i].DirName {
			copy := entry
			generated.ActiveFeature = &copy
		}
		generated.Sources = append(generated.Sources, featureSources(projectRoot, features[i].Path)...)
	}
	generated.Sources = append(generated.Sources, sourceFingerprint(projectRoot, filepath.Join(projectRoot, "docs", "PROJECT_PROGRESS_SUMMARY.md")))
	return generated, nil
}

func Write(projectRoot string, generated State) error {
	path := filepath.Join(projectRoot, filepath.FromSlash(StatePath))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(generated, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func buildFeatureState(
	projectRoot string,
	feat *feature.Feature,
	latestRuns map[string]runstore.IndexEntry,
) (FeatureState, error) {
	status, err := feature.GetFeatureStatus(feat)
	if err != nil {
		return FeatureState{}, err
	}
	entry := FeatureState{
		Status:      status,
		NextCommand: nextCommand(status),
	}
	if latest, ok := latestRuns[feat.DirName]; ok {
		copy := latest
		entry.LatestVerification = &copy
	}
	tasksPath := filepath.Join(feat.Path, "TASKS.md")
	if _, err := os.Stat(tasksPath); err == nil {
		bundles, err := verify.LoadTaskBundles(tasksPath, verify.FeatureRefFromDir(feat.Path), false)
		if err != nil {
			entry.ParseErrors = append(entry.ParseErrors, err.Error())
		} else {
			entry.TaskBundles = bundles
		}
	}
	return entry, nil
}

func latestRunsByFeature(projectRoot string) (map[string]runstore.IndexEntry, error) {
	entries, err := runstore.List(projectRoot)
	if err != nil {
		return nil, err
	}
	latest := make(map[string]runstore.IndexEntry)
	for _, entry := range entries {
		latest[entry.FeatureDir] = entry
	}
	return latest, nil
}

func nextCommand(status *feature.FeatureStatus) string {
	if status == nil {
		return ""
	}
	switch status.Phase {
	case feature.PhaseSpec:
		return fmt.Sprintf("kit legacy plan %s", status.Name)
	case feature.PhasePlan:
		return fmt.Sprintf("kit legacy tasks %s", status.Name)
	case feature.PhaseTasks:
		return fmt.Sprintf("kit legacy implement %s", status.Name)
	case feature.PhaseImplement:
		return fmt.Sprintf("kit legacy implement %s", status.Name)
	case feature.PhaseReflect:
		return fmt.Sprintf("kit legacy reflect %s", status.Name)
	case feature.PhaseComplete:
		return ""
	default:
		return fmt.Sprintf("kit spec %s", status.Name)
	}
}

func featureSources(projectRoot, featurePath string) []SourceFingerprint {
	var sources []SourceFingerprint
	for _, name := range []string{"BRAINSTORM.md", "SPEC.md", "PLAN.md", "TASKS.md"} {
		path := filepath.Join(featurePath, name)
		if _, err := os.Stat(path); err == nil {
			sources = append(sources, sourceFingerprint(projectRoot, path))
		}
	}
	return sources
}

func sourceFingerprint(projectRoot, path string) SourceFingerprint {
	info, err := os.Stat(path)
	rel := path
	if next, relErr := filepath.Rel(projectRoot, path); relErr == nil {
		rel = filepath.ToSlash(next)
	}
	if err != nil {
		return SourceFingerprint{Path: rel}
	}
	return SourceFingerprint{
		Path:    rel,
		Size:    info.Size(),
		ModTime: info.ModTime().UTC(),
	}
}

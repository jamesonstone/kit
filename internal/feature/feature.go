package feature

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/document"
)

// Feature represents a feature directory and its metadata.
type Feature struct {
	Number    int       `json:"number"`
	Slug      string    `json:"slug"`
	DirName   string    `json:"dir_name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"-"`
	Phase     Phase     `json:"phase"`
	Paused    bool      `json:"paused"`
}

// Phase represents the current phase of a feature.
type Phase string

const (
	PhaseClarify   Phase = "clarify"
	PhaseReady     Phase = "ready"
	PhaseImplement Phase = "implement"
	PhaseValidate  Phase = "validate"
	PhaseReflect   Phase = "reflect"
	PhaseDeliver   Phase = "deliver"
	PhaseComplete  Phase = "complete"
	PhaseBlocked   Phase = "blocked"
	PhaseRemoved   Phase = "removed"

	// Legacy staged phases are retained only for existing v1 artifacts and
	// explicit `kit legacy` commands.
	PhaseBrainstorm Phase = "brainstorm"
	PhaseSpec       Phase = "spec"
	PhasePlan       Phase = "plan"
	PhaseTasks      Phase = "tasks"
)

// ReflectionCompleteMarker is the marker that indicates reflection is complete.
const ReflectionCompleteMarker = "<!-- REFLECTION_COMPLETE -->"

var (
	// featureDirPattern matches feature directories like "0001-feat-name"
	featureDirPattern = regexp.MustCompile(`^(\d+)-(.+)$`)
	// slugPattern validates slugs: lowercase, kebab-case
	slugPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
)

// ValidateSlug checks if a slug meets requirements:
// - lowercase only
// - kebab-case
// - max 5 words
func ValidateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}

	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("slug must be lowercase kebab-case (e.g., 'my-feature-name')")
	}

	words := strings.Split(slug, "-")
	if len(words) > 5 {
		return fmt.Errorf("slug cannot exceed 5 words (got %d)", len(words))
	}

	return nil
}

// NormalizeSlug converts a string to a valid slug.
func NormalizeSlug(input string) string {

	slug := strings.ToLower(input)

	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// remove any characters that aren't alphanumeric or hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()

	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	slug = strings.Trim(slug, "-")
	return slug
}

// ListFeatures returns all features in the specs directory, sorted by number.
func ListFeatures(specsDir string) ([]Feature, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	var features []Feature
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		matches := featureDirPattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		num, _ := strconv.Atoi(matches[1])
		feat := Feature{
			Number:  num,
			Slug:    matches[2],
			DirName: entry.Name(),
			Path:    filepath.Join(specsDir, entry.Name()),
		}

		feat.Phase = DeterminePhase(feat.Path)

		info, err := entry.Info()
		if err == nil {
			feat.CreatedAt = info.ModTime()
		}

		features = append(features, feat)
	}

	sort.Slice(features, func(i, j int) bool {
		if features[i].Number != features[j].Number {
			return features[i].Number < features[j].Number
		}
		return features[i].DirName < features[j].DirName
	})

	return features, nil
}

// DeterminePhase checks feature documents and returns the current phase.
// v2 SPEC.md front matter is authoritative when present. Legacy staged artifact
// inference is retained only as fallback for historical v1 feature directories.
func DeterminePhase(featurePath string) Phase {
	tasksPath := filepath.Join(featurePath, "TASKS.md")
	planPath := filepath.Join(featurePath, "PLAN.md")
	specPath := filepath.Join(featurePath, "SPEC.md")
	brainstormPath := filepath.Join(featurePath, "BRAINSTORM.md")

	if phase, ok := determineV2SpecPhase(specPath); ok {
		return phase
	}

	if _, err := os.Stat(tasksPath); err == nil {
		return DeterminePhaseFromTasks(tasksPath)
	}
	if _, err := os.Stat(planPath); err == nil {
		return PhasePlan
	}
	if _, err := os.Stat(specPath); err == nil {
		return PhaseSpec
	}
	if _, err := os.Stat(brainstormPath); err == nil {
		return PhaseBrainstorm
	}
	return PhaseBrainstorm
}

func determineV2SpecPhase(specPath string) (Phase, bool) {
	if _, err := os.Stat(specPath); err != nil {
		return "", false
	}
	doc, err := document.ParseFile(specPath, document.TypeSpec)
	if err != nil || doc.Metadata == nil || doc.Metadata.WorkflowVersion != 2 {
		return "", false
	}
	if phase, ok := V2PhaseFromString(doc.Metadata.Phase); ok {
		return phase, true
	}
	return PhaseClarify, true
}

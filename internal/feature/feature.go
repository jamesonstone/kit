// package feature handles feature numbering, slug validation, and directory management.
package feature

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
)

// Feature represents a feature directory and its metadata.
type Feature struct {
	Number    int
	Slug      string
	DirName   string
	Path      string
	CreatedAt time.Time
	Phase     Phase
}

// Phase represents the current phase of a feature in the artifact pipeline.
type Phase string

const (
	PhaseSpec      Phase = "spec"
	PhasePlan      Phase = "plan"
	PhaseTasks     Phase = "tasks"
	PhaseImplement Phase = "implement"
	PhaseReflect   Phase = "reflect"
	PhaseComplete  Phase = "complete"
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
	// lowercase
	slug := strings.ToLower(input)
	// replace spaces and underscores with hyphens
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
	// collapse multiple hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	// trim leading/trailing hyphens
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

		// determine phase
		feat.Phase = DeterminePhase(feat.Path)

		// get creation time from directory
		info, err := entry.Info()
		if err == nil {
			feat.CreatedAt = info.ModTime()
		}

		features = append(features, feat)
	}

	// sort by number ascending
	sort.Slice(features, func(i, j int) bool {
		return features[i].Number < features[j].Number
	})

	return features, nil
}

// DeterminePhase checks which documents exist and returns the current phase.
// phase progression: spec → plan → tasks → implement → reflect
func DeterminePhase(featurePath string) Phase {
	tasksPath := filepath.Join(featurePath, "TASKS.md")
	planPath := filepath.Join(featurePath, "PLAN.md")
	specPath := filepath.Join(featurePath, "SPEC.md")

	// if tasks file exists, check task completion for implement vs reflect
	if _, err := os.Stat(tasksPath); err == nil {
		return DeterminePhaseFromTasks(tasksPath)
	}
	if _, err := os.Stat(planPath); err == nil {
		return PhasePlan
	}
	if _, err := os.Stat(specPath); err == nil {
		return PhaseSpec
	}
	return PhaseSpec
}

// DeterminePhaseFromTasks determines phase based on task progress.
// - no tasks defined: PhaseTasks (needs task definition)
// - all tasks complete + reflection marker: PhaseComplete
// - all tasks complete: PhaseReflect
// - some tasks incomplete: PhaseImplement
func DeterminePhaseFromTasks(tasksPath string) Phase {
	progress, hasReflectionMarker, err := parseTaskProgressFromPath(tasksPath)
	if err != nil || progress.Total == 0 {
		return PhaseTasks
	}
	if progress.Complete == progress.Total {
		if hasReflectionMarker {
			return PhaseComplete
		}
		return PhaseReflect
	}
	return PhaseImplement
}

// parseTaskProgressFromPath is a lightweight task counter used by DeterminePhase.
// returns: progress counts, whether reflection marker is present, error
func parseTaskProgressFromPath(tasksPath string) (struct{ Total, Complete int }, bool, error) {
	progress := struct{ Total, Complete int }{}
	hasReflectionMarker := false

	file, err := os.Open(tasksPath)
	if err != nil {
		return progress, false, err
	}
	defer file.Close()

	incompletePattern := regexp.MustCompile(`^\s*-\s*\[\s*\]`)
	completePattern := regexp.MustCompile(`^\s*-\s*\[[xX]\]`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if incompletePattern.MatchString(line) {
			progress.Total++
		} else if completePattern.MatchString(line) {
			progress.Total++
			progress.Complete++
		}
		if strings.Contains(line, ReflectionCompleteMarker) {
			hasReflectionMarker = true
		}
	}

	return progress, hasReflectionMarker, scanner.Err()
}

// NextNumber returns the next available feature number.
func NextNumber(specsDir string) (int, error) {
	features, err := ListFeatures(specsDir)
	if err != nil {
		return 0, err
	}

	if len(features) == 0 {
		return 1, nil
	}

	return features[len(features)-1].Number + 1, nil
}

// FormatDirName formats a feature directory name from number and slug.
func FormatDirName(cfg *config.Config, number int, slug string) string {
	format := fmt.Sprintf("%%0%dd%s%%s", cfg.FeatureNaming.NumericWidth, cfg.FeatureNaming.Separator)
	return fmt.Sprintf(format, number, slug)
}

// ParseDirName extracts number and slug from a feature directory name.
func ParseDirName(dirName string) (number int, slug string, ok bool) {
	matches := featureDirPattern.FindStringSubmatch(dirName)
	if matches == nil {
		return 0, "", false
	}
	num, _ := strconv.Atoi(matches[1])
	return num, matches[2], true
}

// FindBySlug finds a feature by its slug (case-insensitive partial match).
func FindBySlug(specsDir string, slug string) (*Feature, error) {
	features, err := ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	slug = strings.ToLower(slug)
	for _, f := range features {
		if strings.ToLower(f.Slug) == slug {
			return &f, nil
		}
	}

	return nil, fmt.Errorf("feature '%s' not found. Run 'kit spec %s' to create it", slug, slug)
}

// FindByDirName finds a feature by its full directory name.
func FindByDirName(specsDir string, dirName string) (*Feature, error) {
	features, err := ListFeatures(specsDir)
	if err != nil {
		return nil, err
	}

	for _, f := range features {
		if f.DirName == dirName {
			return &f, nil
		}
	}

	return nil, fmt.Errorf("feature directory '%s' not found", dirName)
}

// Resolve resolves a feature reference (either slug or full dir name).
func Resolve(specsDir string, ref string) (*Feature, error) {
	// first try exact directory match
	feat, err := FindByDirName(specsDir, ref)
	if err == nil {
		return feat, nil
	}

	// then try slug match
	return FindBySlug(specsDir, ref)
}

// Create creates a new feature directory with the given slug.
func Create(cfg *config.Config, specsDir string, slug string) (*Feature, error) {
	// validate slug
	if err := ValidateSlug(slug); err != nil {
		return nil, err
	}

	// check if slug already exists
	existing, _ := FindBySlug(specsDir, slug)
	if existing != nil {
		return nil, fmt.Errorf("feature '%s' already exists at %s", slug, existing.Path)
	}

	// get next number
	num, err := NextNumber(specsDir)
	if err != nil {
		return nil, err
	}

	// format directory name
	dirName := FormatDirName(cfg, num, slug)
	path := filepath.Join(specsDir, dirName)

	// create directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create feature directory: %w", err)
	}

	return &Feature{
		Number:    num,
		Slug:      slug,
		DirName:   dirName,
		Path:      path,
		CreatedAt: time.Now(),
		Phase:     PhaseSpec,
	}, nil
}

// EnsureExists ensures a feature exists, creating it if necessary.
func EnsureExists(cfg *config.Config, specsDir string, ref string) (*Feature, bool, error) {
	// try to resolve existing
	feat, err := Resolve(specsDir, ref)
	if err == nil {
		return feat, false, nil
	}

	// normalize and create
	slug := NormalizeSlug(ref)
	if err := ValidateSlug(slug); err != nil {
		return nil, false, err
	}

	feat, err = Create(cfg, specsDir, slug)
	if err != nil {
		return nil, false, err
	}

	return feat, true, nil
}

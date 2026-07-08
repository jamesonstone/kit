package feature

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
)

func V2PhaseFromString(value string) (Phase, bool) {
	switch Phase(strings.TrimSpace(value)) {
	case PhaseClarify:
		return PhaseClarify, true
	case PhaseReady:
		return PhaseReady, true
	case PhaseImplement:
		return PhaseImplement, true
	case PhaseValidate:
		return PhaseValidate, true
	case PhaseReflect:
		return PhaseReflect, true
	case PhaseDeliver:
		return PhaseDeliver, true
	case PhaseComplete:
		return PhaseComplete, true
	case PhaseBlocked:
		return PhaseBlocked, true
	default:
		return "", false
	}
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

	if err := scanner.Err(); err != nil {
		_ = file.Close()
		return progress, hasReflectionMarker, err
	}
	if err := file.Close(); err != nil {
		return progress, hasReflectionMarker, err
	}
	return progress, hasReflectionMarker, nil
}

// NextNumber returns the next available feature number, coordinating across
// worktrees from the same clone when a shared Git common dir is available.
func NextNumber(projectRoot, specsDir string) (int, error) {
	features, err := ListFeatures(specsDir)
	if err != nil {
		return 0, err
	}

	return reserveNextFeatureNumber(projectRoot, highestFeatureNumber(features))
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

	feat, err := FindByDirName(specsDir, ref)
	if err == nil {
		return feat, nil
	}

	return FindBySlug(specsDir, ref)
}

// Create creates a new feature directory with the given slug.
func Create(cfg *config.Config, projectRoot, specsDir string, slug string) (*Feature, error) {

	if err := ValidateSlug(slug); err != nil {
		return nil, err
	}

	existing, _ := FindBySlug(specsDir, slug)
	if existing != nil {
		return nil, fmt.Errorf("feature '%s' already exists at %s", slug, existing.Path)
	}

	num, err := NextNumber(projectRoot, specsDir)
	if err != nil {
		return nil, err
	}

	dirName := FormatDirName(cfg, num, slug)
	path := filepath.Join(specsDir, dirName)

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create feature directory: %w", err)
	}

	return &Feature{
		Number:    num,
		Slug:      slug,
		DirName:   dirName,
		Path:      path,
		CreatedAt: time.Now(),
		Phase:     PhaseBrainstorm,
	}, nil
}

// EnsureExists ensures a feature exists, creating it if necessary.
func EnsureExists(cfg *config.Config, projectRoot, specsDir string, ref string) (*Feature, bool, error) {

	feat, err := Resolve(specsDir, ref)
	if err == nil {
		return feat, false, nil
	}

	slug := NormalizeSlug(ref)
	if err := ValidateSlug(slug); err != nil {
		return nil, false, err
	}

	feat, err = Create(cfg, projectRoot, specsDir, slug)
	if err != nil {
		return nil, false, err
	}

	return feat, true, nil
}

func highestFeatureNumber(features []Feature) int {
	if len(features) == 0 {
		return 0
	}

	return features[len(features)-1].Number
}

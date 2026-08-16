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

	"github.com/jamesonstone/kit/v3/internal/config"
)

// parseTaskProgressFromPath counts task completion and reports whether the
// task file contains a reflection marker.
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
	// first try exact directory match
	feat, err := FindByDirName(specsDir, ref)
	if err == nil {
		return feat, nil
	}

	// then try slug match
	return FindBySlug(specsDir, ref)
}

// Create creates a new feature directory with the given slug.
func Create(cfg *config.Config, projectRoot, specsDir string, slug string) (*Feature, error) {
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
	num, err := NextNumber(projectRoot, specsDir)
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
		Phase:     PhaseBrainstorm,
	}, nil
}

// EnsureExists ensures a feature exists, creating it if necessary.
func EnsureExists(cfg *config.Config, projectRoot, specsDir string, ref string) (*Feature, bool, error) {
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

func DuplicateNumberGroups(features []Feature) map[int][]Feature {
	if len(features) == 0 {
		return nil
	}

	groups := make(map[int][]Feature)
	for _, feat := range features {
		groups[feat.Number] = append(groups[feat.Number], feat)
	}

	duplicates := make(map[int][]Feature)
	for number, group := range groups {
		if len(group) < 2 {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			return group[i].DirName < group[j].DirName
		})
		duplicates[number] = group
	}

	if len(duplicates) == 0 {
		return nil
	}

	return duplicates
}

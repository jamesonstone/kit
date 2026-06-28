package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

type projectRefreshStatus struct {
	Due                       bool
	Reasons                   []string
	CompletedFeatures         int
	LastCompletedFeatureCount int
	FeaturesSinceLastReview   int
	FeatureInterval           int
	MaxAgeDays                int
	LastReviewedAt            string
	LastReviewedAgeDays       int
	LastReviewedKnown         bool
}

func calculateProjectRefreshStatus(projectRoot string, cfg *config.Config, now time.Time) (projectRefreshStatus, error) {
	if cfg == nil {
		cfg = config.Default()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()

	completed, err := countCompletedFeatures(projectRoot, cfg)
	if err != nil {
		return projectRefreshStatus{}, err
	}

	constitution := cfg.ProjectRefresh.Constitution
	featureInterval := constitution.FeatureInterval
	if featureInterval <= 0 {
		featureInterval = config.DefaultProjectRefreshFeatureInterval
	}
	maxAgeDays := constitution.MaxAgeDays
	if maxAgeDays <= 0 {
		maxAgeDays = config.DefaultProjectRefreshMaxAgeDays
	}

	featuresSinceLastReview := completed - constitution.LastCompletedFeatureCount
	if featuresSinceLastReview < 0 {
		featuresSinceLastReview = 0
	}

	status := projectRefreshStatus{
		CompletedFeatures:         completed,
		LastCompletedFeatureCount: constitution.LastCompletedFeatureCount,
		FeaturesSinceLastReview:   featuresSinceLastReview,
		FeatureInterval:           featureInterval,
		MaxAgeDays:                maxAgeDays,
		LastReviewedAt:            strings.TrimSpace(constitution.LastReviewedAt),
	}

	if featureInterval > 0 && featuresSinceLastReview >= featureInterval {
		status.Due = true
		status.Reasons = append(status.Reasons, fmt.Sprintf(
			"%d completed feature(s) since the last Constitution refresh review (threshold %d)",
			featuresSinceLastReview,
			featureInterval,
		))
	}

	if status.LastReviewedAt == "" {
		return status, nil
	}

	lastReviewedAt, ok := parseProjectRefreshReviewTime(status.LastReviewedAt)
	if !ok {
		status.Due = true
		status.Reasons = append(status.Reasons, fmt.Sprintf(
			"last Constitution refresh review timestamp %q is not parseable",
			status.LastReviewedAt,
		))
		return status, nil
	}

	status.LastReviewedKnown = true
	if lastReviewedAt.After(now) {
		return status, nil
	}

	status.LastReviewedAgeDays = int(now.Sub(lastReviewedAt).Hours() / 24)
	if maxAgeDays > 0 && status.LastReviewedAgeDays >= maxAgeDays {
		status.Due = true
		status.Reasons = append(status.Reasons, fmt.Sprintf(
			"last Constitution refresh review was %d day(s) ago (threshold %d)",
			status.LastReviewedAgeDays,
			maxAgeDays,
		))
	}

	return status, nil
}

func countCompletedFeatures(projectRoot string, cfg *config.Config) (int, error) {
	features, err := feature.ListFeaturesWithState(cfg.SpecsPath(projectRoot), cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to list project features: %w", err)
	}

	completed := 0
	for _, feat := range features {
		if feat.Phase == feature.PhaseComplete {
			completed++
		}
	}
	return completed, nil
}

func parseProjectRefreshReviewTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func recordProjectRefreshReview(projectRoot string, cfg *config.Config, now time.Time) (projectRefreshStatus, error) {
	if cfg == nil {
		cfg = config.Default()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	completed, err := countCompletedFeatures(projectRoot, cfg)
	if err != nil {
		return projectRefreshStatus{}, err
	}

	if cfg.ProjectRefresh.Constitution.FeatureInterval <= 0 {
		cfg.ProjectRefresh.Constitution.FeatureInterval = config.DefaultProjectRefreshFeatureInterval
	}
	if cfg.ProjectRefresh.Constitution.MaxAgeDays <= 0 {
		cfg.ProjectRefresh.Constitution.MaxAgeDays = config.DefaultProjectRefreshMaxAgeDays
	}
	cfg.ProjectRefresh.Constitution.LastReviewedAt = now.UTC().Format(time.RFC3339)
	cfg.ProjectRefresh.Constitution.LastCompletedFeatureCount = completed
	if err := config.Save(projectRoot, cfg); err != nil {
		return projectRefreshStatus{}, err
	}

	return calculateProjectRefreshStatus(projectRoot, cfg, now)
}

func formatProjectRefreshDueSummary(status projectRefreshStatus) string {
	if status.Due {
		return "due: " + strings.Join(status.Reasons, "; ")
	}
	return fmt.Sprintf(
		"not due: %d/%d completed feature(s) since last review; max age %d day(s)",
		status.FeaturesSinceLastReview,
		status.FeatureInterval,
		status.MaxAgeDays,
	)
}

func projectRefreshStatusBullets(status projectRefreshStatus) []string {
	lastReviewed := "never"
	if status.LastReviewedAt != "" {
		lastReviewed = status.LastReviewedAt
		if status.LastReviewedKnown {
			lastReviewed = fmt.Sprintf("%s (%d day(s) ago)", status.LastReviewedAt, status.LastReviewedAgeDays)
		}
	}

	items := []string{
		fmt.Sprintf("Due state: %s", formatProjectRefreshDueSummary(status)),
		fmt.Sprintf("Completed features: %d", status.CompletedFeatures),
		fmt.Sprintf("Completed features since last review: %d (threshold %d)", status.FeaturesSinceLastReview, status.FeatureInterval),
		fmt.Sprintf("Last reviewed: %s", lastReviewed),
		fmt.Sprintf("Max review age: %d day(s)", status.MaxAgeDays),
	}
	if status.Due {
		items = append(items, "When due, run `kit project refresh` and make reviewed semantic updates before final handoff.")
	}
	return items
}

func printProjectRefreshStatusSummary(out io.Writer, status projectRefreshStatus) error {
	if status.Due {
		_, err := fmt.Fprintf(out, "  ⚠ Project refresh due: %s. Run `kit project refresh`.\n", strings.Join(status.Reasons, "; "))
		return err
	}
	_, err := fmt.Fprintf(out, "  ℹ Project refresh %s. Run `kit project refresh` anytime for an ad hoc semantic refresh.\n", formatProjectRefreshDueSummary(status))
	return err
}

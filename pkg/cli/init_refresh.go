package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

const constitutionBaselineHeading = templates.ConstitutionBaselineHeading

type initRefreshOptions struct {
	force      bool
	dryRun     bool
	diff       bool
	files      []string
	outputOnly bool
}

type initRefreshStats struct {
	created int
	updated int
	merged  int
	removed int
	skipped int
}

func (s initRefreshStats) changed() int {
	return s.created + s.updated + s.merged + s.removed
}

type initRefreshPlan struct {
	cfg     *config.Config
	targets map[string]bool
	changes []initRefreshFileChange
	notes   []string
	stats   initRefreshStats
}

type initRefreshRegistryError struct {
	err error
}

func (e *initRefreshRegistryError) Error() string {
	return fmt.Sprintf("failed to refresh Kit ruleset registry: %v", e.err)
}

func (e *initRefreshRegistryError) Unwrap() error {
	return e.err
}

func runInitRefreshWithSnapshot(
	projectRoot string,
	opts initRefreshOptions,
) ([]managedFileDeliverySnapshot, error) {
	if !opts.outputOnly && !opts.dryRun {
		fmt.Println("🔄 Refreshing Kit-managed project files...")
	}

	ctx := context.Background()
	plan, err := buildInitRefreshPlan(ctx, projectRoot, opts)
	if err != nil {
		return nil, err
	}
	deliverySnapshot := managedFileDeliverySnapshotFromInitRefresh(projectRoot, plan.changes)

	if opts.dryRun {
		printInitRefreshDryRun(plan.changes, plan.stats, opts)
		printInitRefreshNotes(plan.notes, opts)
		return deliverySnapshot, nil
	}

	if err := applyInitRefreshFileChangesAtomically(plan.changes); err != nil {
		return nil, err
	}

	if !opts.outputOnly {
		fmt.Println("\n✅ Kit managed refresh complete!")
		if plan.stats.changed() == 0 {
			fmt.Println("   No Kit-managed project changes needed.")
		}
		fmt.Printf(
			"   Created: %d, Updated: %d, Merged: %d, Removed: %d, Skipped: %d\n",
			plan.stats.created,
			plan.stats.updated,
			plan.stats.merged,
			plan.stats.removed,
			plan.stats.skipped,
		)
		printInitRefreshNotes(plan.notes, opts)
	}
	return deliverySnapshot, nil
}

func buildInitRefreshPlan(ctx context.Context, projectRoot string, opts initRefreshOptions) (*initRefreshPlan, error) {
	targets, err := normalizeInitRefreshTargets(opts.files)
	if err != nil {
		return nil, err
	}

	needsRegistry := len(targets) == 0
	for target := range targets {
		if strings.HasPrefix(target, rulesetDirRelPath+"/") {
			needsRegistry = true
			break
		}
	}

	var registry []registryRuleset
	if needsRegistry {
		registry, err = rulesetRegistryFetcher(ctx)
		if err != nil {
			return nil, &initRefreshRegistryError{err: err}
		}
	}

	cfg, configChange, err := initRefreshConfig(projectRoot, opts, targets)
	if err != nil {
		return nil, err
	}

	knownTargets := initRefreshKnownTargets(cfg, registry)
	if err := validateInitRefreshTargets(targets, knownTargets); err != nil {
		return nil, err
	}

	var changes []initRefreshFileChange
	var notes []string

	var stats initRefreshStats
	scaffoldChanges, err := planRefreshInitScaffoldFiles(projectRoot, opts, cfg, targets)
	if err != nil {
		return nil, err
	}
	changes = append(changes, scaffoldChanges...)
	readmeChange, err := planRefreshReadmeFile(projectRoot, cfg, targets)
	if err != nil {
		return nil, err
	}
	if readmeChange != nil {
		changes = append(changes, *readmeChange)
	}
	// Plan entry files first: the Constitution baseline moves to the contract
	// pointer only when every entry file converges, in the same pass.
	instructionChanges, instructionNotes, entriesConverged, err := planRefreshInitInstructionArtifacts(projectRoot, opts, cfg, targets)
	if err != nil {
		return nil, err
	}
	constitutionChange, err := planRefreshInitConstitution(projectRoot, cfg, targets, entriesConverged)
	if err != nil {
		return nil, err
	}
	if constitutionChange != nil {
		changes = append(changes, *constitutionChange)
	}
	changes = append(changes, instructionChanges...)
	notes = append(notes, instructionNotes...)
	retiredChanges, retiredNotes, retiredRegistryChanged, err := planRetiredArtifacts(projectRoot, cfg, targets)
	if err != nil {
		return nil, err
	}
	changes = append(changes, retiredChanges...)
	rulesetChanges, rulesetNotes, registryChanged, err := planRefreshInitRulesets(ctx, projectRoot, opts, cfg, targets, registry)
	if err != nil {
		return nil, err
	}
	notes = append(notes, rulesetNotes...)
	notes = append(notes, retiredNotes...)
	changes = append(changes, rulesetChanges...)
	versionChanged := false
	if entriesConverged && requiredStructurePresent(projectRoot, cfg, registry, changes) &&
		initRefreshTargetMatches(targets, config.ConfigFileName) && cfg.InstructionScaffoldVersion != config.CurrentInstructionScaffoldVersion {
		cfg.InstructionScaffoldVersion = config.CurrentInstructionScaffoldVersion
		versionChanged = true
	}
	if configChange != nil || registryChanged || retiredRegistryChanged || versionChanged || initRefreshTargetMatches(targets, config.ConfigFileName) {
		configChange, err = finalizeInitRefreshConfigChange(projectRoot, cfg, configChange)
		if err != nil {
			return nil, err
		}
		if configChange != nil {
			changes = append([]initRefreshFileChange{*configChange}, changes...)
			if note := droppedConfigNote(configChange.before, configChange.after); note != "" {
				notes = append(notes, note)
			}
		}
	}

	for _, change := range changes {
		stats.recordFileChange(change)
	}

	return &initRefreshPlan{
		cfg:     cfg,
		targets: targets,
		changes: changes,
		notes:   notes,
		stats:   stats,
	}, nil
}

func initRefreshKnownTargets(cfg *config.Config, registry []registryRuleset) map[string]bool {
	known := map[string]bool{
		config.ConfigFileName:                  true,
		gitignorePath:                          true,
		envPath:                                true,
		envrcPath:                              true,
		makefilePath:                           true,
		codeRabbitConfigPath:                   true,
		pullRequestTemplatePath:                true,
		autoAssignWorkflowPath:                 true,
		readmePath:                             true,
		cfg.ConstitutionPath:                   true,
		filepath.ToSlash(cfg.ConstitutionPath): true,
	}
	for _, relativePath := range instructionArtifactPaths(cfg) {
		known[filepath.ToSlash(relativePath)] = true
	}
	for _, item := range registry {
		known[rulesetTarget(item.Slug)] = true
	}
	return known
}

func normalizeInitRefreshTargets(files []string) (map[string]bool, error) {
	targets := make(map[string]bool, len(files))
	for _, file := range files {
		target := strings.TrimSpace(file)
		if target == "" {
			return nil, fmt.Errorf("--file target cannot be blank")
		}
		if filepath.IsAbs(target) {
			return nil, fmt.Errorf("--file target %q must be relative to the project root", file)
		}
		target = filepath.ToSlash(filepath.Clean(target))
		target = strings.TrimPrefix(target, "./")
		if target == "." || strings.HasPrefix(target, "../") {
			return nil, fmt.Errorf("--file target %q must stay inside the project root", file)
		}
		targets[target] = true
	}
	return targets, nil
}

func validateInitRefreshTargets(targets, known map[string]bool) error {
	if len(targets) == 0 {
		return nil
	}
	var unknown []string
	for target := range targets {
		if !known[target] {
			unknown = append(unknown, target)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return fmt.Errorf("%s is not a Kit-managed refresh target", strings.Join(unknown, ", "))
}

func printInitRefreshNotes(notes []string, opts initRefreshOptions) {
	if opts.outputOnly || len(notes) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Notes:")
	for _, note := range notes {
		fmt.Printf("   - %s\n", note)
	}
}

package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/internal/config"
	"gopkg.in/yaml.v3"
)

const constitutionBaselineHeading = "Kit-Managed Baseline Rules"

const constitutionBaselineSection = `### ` + constitutionBaselineHeading + `

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat ` + "`docs/CONSTITUTION.md`" + ` as the canonical project contract.
- Keep ` + "`AGENTS.md`" + `, ` + "`CLAUDE.md`" + `, and ` + "`.github/copilot-instructions.md`" + ` aligned with the repo-local docs tree.
- Prefer implementation/source code files around 300 lines or less when splitting improves clarity and ownership.
- Do not apply the code-file size guideline to documentation files, all ` + "`docs/**`" + `, all ` + "`.kit/**`" + `, or ` + "`.kit.yaml`" + `.
- Do not split or rewrite docs, generated state, or Kit config artifacts solely because they exceed 300 lines.
<!-- END KIT-MANAGED BASELINE RULES -->`

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
	skipped int
}

func runInitRefresh(projectRoot string, opts initRefreshOptions) error {
	ctx := context.Background()
	targets, err := normalizeInitRefreshTargets(opts.files)
	if err != nil {
		return err
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
			return fmt.Errorf("failed to refresh Kit ruleset registry: %w", err)
		}
	}

	cfg, configChange, err := initRefreshConfig(projectRoot, opts, targets)
	if err != nil {
		return err
	}

	knownTargets := initRefreshKnownTargets(cfg, registry)
	if err := validateInitRefreshTargets(targets, knownTargets); err != nil {
		return err
	}

	if !opts.outputOnly && !opts.dryRun {
		fmt.Println("🔄 Refreshing Kit-managed project files...")
	}

	var changes []initRefreshFileChange
	var notes []string

	var stats initRefreshStats
	scaffoldChanges, err := planRefreshInitScaffoldFiles(projectRoot, opts, targets)
	if err != nil {
		return err
	}
	changes = append(changes, scaffoldChanges...)
	constitutionChange, err := planRefreshInitConstitution(projectRoot, cfg, targets)
	if err != nil {
		return err
	}
	if constitutionChange != nil {
		changes = append(changes, *constitutionChange)
	}
	instructionChanges, err := planRefreshInitInstructionArtifacts(projectRoot, opts, cfg, targets)
	if err != nil {
		return err
	}
	changes = append(changes, instructionChanges...)
	rulesetChanges, rulesetNotes, registryChanged, err := planRefreshInitRulesets(ctx, projectRoot, opts, cfg, targets, registry)
	if err != nil {
		return err
	}
	notes = append(notes, rulesetNotes...)
	changes = append(changes, rulesetChanges...)
	if configChange != nil || registryChanged {
		configChange, err = finalizeInitRefreshConfigChange(projectRoot, cfg, configChange)
		if err != nil {
			return err
		}
		if configChange != nil {
			changes = append([]initRefreshFileChange{*configChange}, changes...)
		}
	}

	if opts.dryRun {
		for _, change := range changes {
			stats.recordFileChange(change)
		}
		printInitRefreshDryRun(changes, stats, opts)
		printInitRefreshNotes(notes, opts)
		return nil
	}

	for _, change := range changes {
		if err := applyInitRefreshFileChange(change); err != nil {
			return err
		}
		stats.recordFileChange(change)
	}

	if !opts.outputOnly {
		fmt.Println("\n✅ Kit project refresh complete!")
		fmt.Printf(
			"   Created: %d, Updated: %d, Merged: %d, Skipped: %d\n",
			stats.created,
			stats.updated,
			stats.merged,
			stats.skipped,
		)
		printInitRefreshNotes(notes, opts)
	}
	return nil
}

func initRefreshConfig(
	projectRoot string,
	opts initRefreshOptions,
	targets map[string]bool,
) (*config.Config, *initRefreshFileChange, error) {
	cfg := defaultInitConfig()
	configSelected := initRefreshTargetMatches(targets, config.ConfigFileName)
	shouldTouchConfig := len(targets) == 0 || configSelected
	path := filepath.Join(projectRoot, config.ConfigFileName)
	exists := config.Exists(projectRoot)
	var before string

	if exists {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read %s: %w", config.ConfigFileName, err)
		}
		before = string(data)
		existing, err := config.Load(projectRoot)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load %s: %w", config.ConfigFileName, err)
		}
		cfg = existing
	}

	if configSelected && opts.force {
		cfg = defaultInitConfig()
		cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
		after, err := marshalInitRefreshConfig(cfg)
		if err != nil {
			return nil, nil, err
		}
		result := instructionFileCreated
		if exists {
			result = instructionFileUpdated
		}
		return cfg, newInitRefreshFileChange(projectRoot, config.ConfigFileName, before, after, result), nil
	}

	if cfg.InstructionScaffoldVersion != config.InstructionScaffoldVersionTOC {
		cfg.InstructionScaffoldVersion = config.InstructionScaffoldVersionTOC
		if shouldTouchConfig {
			after, err := marshalInitRefreshConfig(cfg)
			if err != nil {
				return nil, nil, err
			}
			return cfg, newInitRefreshFileChange(projectRoot, config.ConfigFileName, before, after, instructionFileUpdated), nil
		}
	}

	if !exists && shouldTouchConfig {
		after, err := marshalInitRefreshConfig(cfg)
		if err != nil {
			return nil, nil, err
		}
		return cfg, newInitRefreshFileChange(projectRoot, config.ConfigFileName, before, after, instructionFileCreated), nil
	}

	return cfg, nil, nil
}

func initRefreshKnownTargets(cfg *config.Config, registry []registryRuleset) map[string]bool {
	known := map[string]bool{
		config.ConfigFileName:                  true,
		gitignorePath:                          true,
		envPath:                                true,
		envrcPath:                              true,
		codeRabbitConfigPath:                   true,
		pullRequestTemplatePath:                true,
		cfg.ConstitutionPath:                   true,
		filepath.ToSlash(cfg.ConstitutionPath): true,
	}
	for _, relativePath := range instructionArtifactPaths(
		cfg,
		instructionFileSelection{},
		config.InstructionScaffoldVersionTOC,
		true,
	) {
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

func finalizeInitRefreshConfigChange(projectRoot string, cfg *config.Config, planned *initRefreshFileChange) (*initRefreshFileChange, error) {
	before := ""
	result := instructionFileCreated
	if planned != nil {
		before = planned.before
		result = planned.result
	} else if config.Exists(projectRoot) {
		data, err := os.ReadFile(filepath.Join(projectRoot, config.ConfigFileName))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", config.ConfigFileName, err)
		}
		before = string(data)
		result = instructionFileUpdated
	}

	after, err := marshalInitRefreshConfig(cfg)
	if err != nil {
		return nil, err
	}
	if before == after {
		return nil, nil
	}
	return newInitRefreshFileChange(projectRoot, config.ConfigFileName, before, after, result), nil
}

func printInitRefreshNotes(notes []string, opts initRefreshOptions) {
	if opts.outputOnly || len(notes) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("Ruleset registry notes:")
	for _, note := range notes {
		fmt.Printf("   - %s\n", note)
	}
}

func initRefreshTargetMatches(targets map[string]bool, relativePath string) bool {
	if len(targets) == 0 {
		return true
	}
	_, ok := targets[filepath.ToSlash(relativePath)]
	return ok
}

func (s *initRefreshStats) recordFileChange(change initRefreshFileChange) {
	s.recordResult(change.result)
}

func (s *initRefreshStats) recordResult(result instructionFileWriteResult) {
	switch result {
	case instructionFileCreated:
		s.created++
	case instructionFileUpdated:
		s.updated++
	case instructionFileMerged:
		s.merged++
	case instructionFileSkipped:
		s.skipped++
	}
}

func marshalInitRefreshConfig(cfg *config.Config) (string, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return string(data), nil
}

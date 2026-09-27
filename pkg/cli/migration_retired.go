package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/legacy"
	"github.com/jamesonstone/kit/v3/internal/templates"
)

// retiredArtifact is a file an earlier Kit generated that the current
// structure no longer has.
type retiredArtifact struct {
	relativePath string
	content      string
	kitOwned     bool   // exactly what Kit generated
	reason       string // why a Kit-owned file is kept, or what was edited
}

// retiredDirectories are removed once migration empties them.
var retiredDirectories = []string{"docs/agents", "docs/references/workflows"}

// planRetiredArtifacts removes retired Kit-generated files that are exactly as
// Kit wrote them and that Git can restore, prunes registry entries for retired
// rules, and reports every retired file it keeps.
func planRetiredArtifacts(
	projectRoot string,
	cfg *config.Config,
	targets map[string]bool,
) ([]initRefreshFileChange, []string, bool, error) {
	artifacts, err := findRetiredArtifacts(projectRoot, cfg, targets)
	if err != nil {
		return nil, nil, false, err
	}
	registryChanged := pruneRetiredRegistryEntries(cfg, targets)

	candidates := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.kitOwned {
			candidates = append(candidates, artifact.relativePath)
		}
	}
	restorable, err := gitRestorablePaths(projectRoot, candidates)
	if err != nil {
		return nil, nil, false, err
	}

	var changes []initRefreshFileChange
	var kept []string
	for _, artifact := range artifacts {
		switch {
		case artifact.kitOwned && restorable[artifact.relativePath]:
			changes = append(changes, *newInitRefreshFileChange(projectRoot, artifact.relativePath, artifact.content, "", instructionFileRemoved))
		case artifact.kitOwned:
			kept = append(kept, fmt.Sprintf("%s (unmodified, but not committed to Git, so removal could not be undone)", artifact.relativePath))
		default:
			kept = append(kept, fmt.Sprintf("%s (%s)", artifact.relativePath, artifact.reason))
		}
	}
	var notes []string
	if len(kept) > 0 {
		notes = append(notes, "kept retired Kit files that are not exactly as Kit generated them; review and delete them yourself when they hold nothing the project needs: "+strings.Join(kept, "; "))
	}
	return changes, notes, registryChanged, nil
}

func findRetiredArtifacts(projectRoot string, cfg *config.Config, targets map[string]bool) ([]retiredArtifact, error) {
	current := map[string]bool{
		legacyConstitutionPath:                  true,
		filepath.ToSlash(cfg.ConstitutionPath):  true,
		templates.TestingReferencePath:          true,
		filepath.ToSlash(config.ConfigFileName): true,
	}
	for _, path := range instructionFiles(cfg) {
		current[filepath.ToSlash(path)] = true
	}

	var artifacts []retiredArtifact
	add := func(relativePath string, classify func(string) (bool, string)) error {
		if current[relativePath] || !initRefreshTargetMatches(targets, relativePath) {
			return nil
		}
		path := filepath.Join(projectRoot, filepath.FromSlash(relativePath))
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", relativePath, err)
		}
		owned, reason := classify(string(data))
		artifacts = append(artifacts, retiredArtifact{relativePath: relativePath, content: string(data), kitOwned: owned, reason: reason})
		return nil
	}

	for _, relativePath := range legacy.GeneratedPaths() {
		relativePath := relativePath
		if err := add(relativePath, func(content string) (bool, string) {
			doc := legacy.Classify(relativePath, content)
			if doc.Unmodified() {
				return true, ""
			}
			return false, describeEditedDocument(doc)
		}); err != nil {
			return nil, err
		}
	}
	if err := add(legacy.ProgressSummaryPath, func(content string) (bool, string) {
		if legacy.ProgressSummaryGenerated(content) {
			return true, ""
		}
		return false, "sections Kit's progress rollup never wrote"
	}); err != nil {
		return nil, err
	}
	slugs := make([]string, 0, len(retiredRulesets))
	for slug := range retiredRulesets {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	for _, slug := range slugs {
		slug := slug
		state, _ := rulesetRegistryState(cfg, slug)
		if err := add(rulesetTarget(slug), func(content string) (bool, string) {
			if legacy.RuleKnown(slug, content) || installedUnmodified(state, content) {
				return true, ""
			}
			return false, "retired rule with local edits; it is no longer registered and is now a project rule"
		}); err != nil {
			return nil, err
		}
	}
	return artifacts, nil
}

func installedUnmodified(state config.RegistryArtifact, content string) bool {
	parsed := parseRuleset(content, "")
	hash, err := normalizedRulesetContentHash(content, parsed.Metadata.Status)
	return err == nil && kitWroteUnmodified(state, hash)
}

func describeEditedDocument(doc legacy.Document) string {
	var parts []string
	if len(doc.Modified) > 0 {
		parts = append(parts, "edited sections: "+strings.Join(doc.ModifiedKeys(), ", "))
	}
	if len(doc.Project) > 0 {
		parts = append(parts, "project content Kit never wrote")
	}
	if len(parts) == 0 {
		return "no recognizable Kit content"
	}
	return strings.Join(parts, "; ")
}

// pruneRetiredRegistryEntries drops `.kit.yaml` registrations of rules Kit no
// longer ships; a kept file becomes an ordinary project rule.
func pruneRetiredRegistryEntries(cfg *config.Config, targets map[string]bool) bool {
	if len(targets) > 0 && !targets[config.ConfigFileName] {
		return false
	}
	kept := cfg.Registry.Artifacts[:0]
	changed := false
	for _, artifact := range cfg.Registry.Artifacts {
		if artifact.Kind == rulesetKind && isRetiredRuleset(artifact.Slug) {
			changed = true
			continue
		}
		kept = append(kept, artifact)
	}
	cfg.Registry.Artifacts = kept
	return changed
}

// removeEmptyRetiredDirectories deletes retired Kit directories left empty.
func removeEmptyRetiredDirectories(removedFile string) {
	for dir := filepath.Dir(removedFile); ; dir = filepath.Dir(dir) {
		retired := false
		for _, candidate := range retiredDirectories {
			if strings.HasSuffix(filepath.ToSlash(dir), "/"+candidate) {
				retired = true
			}
		}
		if !retired || !dirIsEmpty(dir) || os.Remove(dir) != nil {
			return
		}
	}
}

func dirIsEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) == 0
}

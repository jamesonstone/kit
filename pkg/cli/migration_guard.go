package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
)

// guardLinkedTargets turns every planned write or removal whose path is a
// symlink, or resolves outside the project, into a skip with a note. Kit never
// writes through links: they can point at local-only files such as the
// primary checkout's .env or at content shared outside the repository.
func guardLinkedTargets(projectRoot string, changes []initRefreshFileChange) ([]initRefreshFileChange, map[string]bool, []string) {
	blocked := map[string]bool{}
	var notes []string
	root := resolvedOrClean(projectRoot)
	for i, change := range changes {
		if change.result == instructionFileSkipped {
			continue
		}
		reason := ""
		if info, err := os.Lstat(change.absolutePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
			reason = "is a symbolic link"
		} else if parent := resolvedOrClean(filepath.Dir(change.absolutePath)); parent != root && !strings.HasPrefix(parent, root+string(filepath.Separator)) {
			reason = "resolves outside the project"
		}
		if reason == "" {
			continue
		}
		blocked[change.relativePath] = true
		notes = append(notes, fmt.Sprintf("%s %s, so Kit left it unchanged; Kit never writes or removes through links", change.relativePath, reason))
		changes[i].after = change.before
		changes[i].result = instructionFileSkipped
	}
	return changes, blocked, notes
}

// resolvedOrClean resolves symlinks in the longest existing prefix of path.
func resolvedOrClean(path string) string {
	path = filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	parent := filepath.Dir(path)
	if parent == path {
		return path
	}
	return filepath.Join(resolvedOrClean(parent), filepath.Base(path))
}

// recordBlockedRulesAsLocal keeps the registry honest when the guard skipped a
// rule: the file on disk was not replaced, so it is recorded as local-custom
// with its actual hash.
func recordBlockedRulesAsLocal(projectRoot string, cfg *config.Config, registry []registryRuleset, blocked map[string]bool) {
	for _, item := range registry {
		if !blocked[rulesetTarget(item.Slug)] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(rulesetTarget(item.Slug))))
		if err != nil {
			continue
		}
		hash, err := normalizedRulesetContentHash(string(data), item.Metadata.Status)
		if err != nil {
			hash = ""
		}
		recordRulesetRegistryState(cfg, item, registryArtifactStateLocalCustom, hash)
	}
}

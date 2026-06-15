package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

func applyRegistryRulesetSelection(projectRoot string, entries []registrySelectorEntry) (registrySelectorSummary, error) {
	var summary registrySelectorSummary
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return summary, fmt.Errorf("failed to load config: %w", err)
	}
	configChanged := false
	for _, entry := range entries {
		if !entry.Installed && !entry.DesiredActive {
			summary.Unchanged++
			continue
		}
		if entry.Installed && entry.DesiredActive == entry.CurrentActive {
			summary.Unchanged++
			continue
		}

		targetStatus := inactiveRulesetStatus
		if entry.DesiredActive {
			targetStatus = document.ReferenceStatusActive
		}

		content := entry.Registry.Content
		if entry.Installed {
			content = entry.LocalContent
		}
		updated, err := setRulesetStatus(content, targetStatus)
		if err != nil {
			return summary, fmt.Errorf("failed to update ruleset %s status: %w", entry.Registry.Slug, err)
		}
		path := rulesetPath(projectRoot, entry.Registry.Slug)
		if err := document.Write(path, updated); err != nil {
			return summary, fmt.Errorf("failed to write %s: %w", rulesetTarget(entry.Registry.Slug), err)
		}
		normalizedHash, err := normalizedRulesetContentHash(updated, entry.Registry.Metadata.Status)
		if err != nil {
			return summary, fmt.Errorf("failed to hash ruleset %s: %w", entry.Registry.Slug, err)
		}
		state := registryArtifactStateManaged
		hash := entry.Registry.NormalizedHash
		if normalizedHash != entry.Registry.NormalizedHash {
			state = registryArtifactStateLocalCustom
			hash = normalizedHash
		}
		recordRulesetRegistryState(cfg, entry.Registry, state, hash, updated)
		configChanged = true

		switch {
		case !entry.Installed && entry.DesiredActive:
			summary.Imported++
		case entry.DesiredActive:
			summary.Activated++
		default:
			summary.Deactivated++
		}
	}
	if configChanged {
		if err := config.Save(projectRoot, cfg); err != nil {
			return summary, fmt.Errorf("failed to save registry state: %w", err)
		}
	}
	return summary, nil
}

func printRegistryRulesetSummary(out io.Writer, summary registrySelectorSummary) error {
	if summary.Imported == 0 && summary.Activated == 0 && summary.Deactivated == 0 {
		_, err := fmt.Fprintln(out, "No ruleset changes selected.")
		return err
	}
	_, err := fmt.Fprintf(
		out,
		"Rulesets updated. Imported: %d, Activated: %d, Deactivated: %d, Unchanged: %d\n",
		summary.Imported,
		summary.Activated,
		summary.Deactivated,
		summary.Unchanged,
	)
	return err
}

func setRulesetStatus(content, status string) (string, error) {
	if !validRulesetStatus(status) {
		return "", fmt.Errorf("invalid ruleset status %q", status)
	}
	raw, body, err := splitRulesetFrontMatter(content)
	if err != nil {
		return "", err
	}
	lines := strings.Split(raw, "\n")
	changed := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "status:") {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			lines[i] = indent + "status: " + status
			changed = true
			break
		}
	}
	if !changed {
		return "", fmt.Errorf("front matter status is missing")
	}

	frontMatter := strings.TrimRight(strings.Join(lines, "\n"), "\n")
	return "---\n" + frontMatter + "\n---\n\n" + strings.TrimLeft(body, "\n"), nil
}

func loadRulesetViewContent(ctx context.Context, projectRoot, slug string) (string, string, error) {
	localPath := rulesetPath(projectRoot, slug)
	if document.Exists(localPath) {
		content, err := os.ReadFile(localPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to read %s: %w", rulesetTarget(slug), err)
		}
		parsed := parseRuleset(string(content), localPath)
		if issues := validateRulesetDocument(parsed, slug); len(issues) > 0 {
			return "", "", fmt.Errorf("local ruleset %s is invalid: %s", rulesetTarget(slug), strings.Join(issues, "; "))
		}
		return string(content), rulesetTarget(slug), nil
	}

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	registry, err := rulesetRegistryFetcher(ctx)
	if err != nil {
		return "", "", err
	}
	for _, item := range projectRulesetRegistry(registry) {
		if item.Slug == slug {
			return item.Content, rulesetRegistryRulesetURL(slug), nil
		}
	}
	return "", "", fmt.Errorf("ruleset %q was not found locally or in the Kit registry", slug)
}

package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
	"gopkg.in/yaml.v3"
)

func validateRulesetSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("ruleset slug cannot be empty")
	}
	if !rulesetSlugPattern.MatchString(slug) {
		return fmt.Errorf("ruleset slug must be lowercase kebab-case")
	}
	return nil
}

func rulesetPath(projectRoot, slug string) string {
	return filepath.Join(projectRoot, filepath.FromSlash(rulesetTarget(slug)))
}

func rulesetTarget(slug string) string {
	return filepath.ToSlash(filepath.Join(rulesetDirRelPath, slug+".md"))
}

func defaultRulesetAppliesTo(slug string) []string {
	first, _, ok := strings.Cut(slug, "-")
	if ok && first != "" {
		return []string{first}
	}
	return []string{slug}
}

func listRulesets(projectRoot string) ([]rulesetDocument, error) {
	dir := filepath.Join(projectRoot, filepath.FromSlash(rulesetDirRelPath))
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read rulesets directory: %w", err)
	}

	var rulesets []rulesetDocument
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		ruleset, err := parseRulesetFile(path)
		if err != nil {
			return nil, err
		}
		if issues := validateRulesetDocument(ruleset, strings.TrimSuffix(entry.Name(), ".md")); len(issues) > 0 {
			return nil, fmt.Errorf("invalid ruleset %s: %s", filepath.ToSlash(path), strings.Join(issues, "; "))
		}
		rulesets = append(rulesets, ruleset)
	}

	sort.SliceStable(rulesets, func(i, j int) bool {
		return rulesets[i].Metadata.Slug < rulesets[j].Metadata.Slug
	})
	return rulesets, nil
}

func printRulesetList(w io.Writer, projectRoot string, cfg *config.Config, rulesets []rulesetDocument) error {
	writer := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "SLUG\tPATH\tSTATUS\tREGISTRY\tAPPLIES_TO"); err != nil {
		return err
	}
	for _, ruleset := range rulesets {
		relPath, err := filepath.Rel(projectRoot, ruleset.Path)
		if err != nil {
			relPath = ruleset.Path
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\n",
			ruleset.Metadata.Slug,
			filepath.ToSlash(relPath),
			ruleset.Metadata.Status,
			rulesetListRegistryState(cfg, ruleset.Metadata.Slug),
			strings.Join(ruleset.Metadata.AppliesTo, ","),
		); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to render ruleset list: %w", err)
	}
	return nil
}

func rulesetListRegistryState(cfg *config.Config, slug string) string {
	artifact, ok := rulesetRegistryState(cfg, slug)
	if !ok || strings.TrimSpace(artifact.State) == "" {
		return "untracked"
	}
	return artifact.State
}

func loadRuleset(projectRoot, slug string) (rulesetDocument, error) {
	path := rulesetPath(projectRoot, slug)
	if !document.Exists(path) {
		return rulesetDocument{}, fmt.Errorf("ruleset %q not found at %s", slug, rulesetTarget(slug))
	}
	return parseRulesetFile(path)
}

func parseRulesetFile(path string) (rulesetDocument, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return rulesetDocument{}, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return parseRuleset(string(content), path), nil
}

func parseRuleset(content, path string) rulesetDocument {
	raw, body, err := splitRulesetFrontMatter(content)
	ruleset := rulesetDocument{
		Path:     path,
		Body:     body,
		Sections: rulesetSections(body),
		ParseErr: err,
	}
	if err != nil {
		return ruleset
	}
	if strings.TrimSpace(raw) == "" {
		ruleset.ParseErr = fmt.Errorf("front matter is empty")
		return ruleset
	}
	if err := yaml.Unmarshal([]byte(raw), &ruleset.Metadata); err != nil {
		ruleset.ParseErr = fmt.Errorf("failed to parse front matter: %w", err)
	}
	return ruleset
}

func splitRulesetFrontMatter(content string) (string, string, error) {
	lines := strings.SplitAfter(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", content, fmt.Errorf("missing YAML front matter")
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[1:i], ""), strings.Join(lines[i+1:], ""), nil
		}
	}
	return "", content, fmt.Errorf("missing closing front matter delimiter")
}

func rulesetSections(body string) map[string]string {
	sections := make(map[string]string)
	matches := rulesetSectionRe.FindAllStringSubmatchIndex(body, -1)
	for i, match := range matches {
		name := strings.ToUpper(strings.TrimSpace(body[match[2]:match[3]]))
		start := match[1]
		end := len(body)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		sections[name] = strings.TrimSpace(body[start:end])
	}
	return sections
}

func validateRulesetDocument(ruleset rulesetDocument, expectedSlug string) []string {
	var issues []string
	if ruleset.ParseErr != nil {
		issues = append(issues, ruleset.ParseErr.Error())
		return issues
	}
	if ruleset.Metadata.Kind != rulesetKind {
		issues = append(issues, "front matter kind must be ruleset")
	}
	if ruleset.Metadata.Slug == "" {
		issues = append(issues, "front matter slug cannot be empty")
	} else if err := validateRulesetSlug(ruleset.Metadata.Slug); err != nil {
		issues = append(issues, err.Error())
	} else if expectedSlug != "" && ruleset.Metadata.Slug != expectedSlug {
		issues = append(issues, fmt.Sprintf("front matter slug %q does not match file slug %q", ruleset.Metadata.Slug, expectedSlug))
	}
	if ruleset.Metadata.Status == "" || !validRulesetStatus(ruleset.Metadata.Status) {
		issues = append(issues, "front matter status must be active, optional, or stale")
	}
	if len(ruleset.Metadata.AppliesTo) == 0 {
		issues = append(issues, "front matter applies_to must contain at least one entry")
	}
	for _, appliesTo := range ruleset.Metadata.AppliesTo {
		if err := validateRulesetSlug(appliesTo); err != nil {
			issues = append(issues, fmt.Sprintf("front matter applies_to entry %q is invalid", appliesTo))
		}
	}
	if ruleset.Metadata.ReadPolicyDefault == "" || !validRulesetReadPolicy(ruleset.Metadata.ReadPolicyDefault) {
		issues = append(issues, "front matter read_policy_default must be must, conditional, evidence, or skip")
	}
	if !validRulesetRegistryScope(ruleset.Metadata.RegistryScope) {
		issues = append(issues, "front matter registry_scope must be downstream or kit-maintainer when set")
	}
	for _, section := range requiredRulesetSections() {
		content, ok := ruleset.Sections[strings.ToUpper(section)]
		if !ok {
			issues = append(issues, fmt.Sprintf("missing required section ## %s", section))
			continue
		}
		if !meaningfulSectionContent(content) {
			issues = append(issues, fmt.Sprintf("required section ## %s is empty or placeholder-only", section))
		}
	}
	return issues
}

func validRulesetRegistryScope(scope string) bool {
	switch strings.TrimSpace(scope) {
	case "", rulesetRegistryScopeDownstream, rulesetRegistryScopeKitMaintainer:
		return true
	default:
		return false
	}
}

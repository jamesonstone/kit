package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/document"
)

func parseRulesetAppliesTo(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	var appliesTo []string
	for _, part := range parts {
		normalized := normalizeRulesetSlug(part)
		if normalized == "" {
			continue
		}
		if err := validateRulesetSlug(normalized); err != nil {
			return nil, fmt.Errorf("invalid applies_to entry %q: %w", strings.TrimSpace(part), err)
		}
		appliesTo = append(appliesTo, normalized)
	}
	if len(appliesTo) == 0 {
		return nil, fmt.Errorf("applies_to must contain at least one entry")
	}
	return appliesTo, nil
}

func normalizeRulesetSlug(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	var builder strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	slug = builder.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}

func buildRulesetOptimizationPrompt(projectRoot, path string, input rulesetAddInput) string {
	relPath, err := filepath.Rel(projectRoot, path)
	if err != nil {
		relPath = path
	}
	relPath = filepath.ToSlash(relPath)
	referencesReadme := filepath.Join(projectRoot, "docs", "references", "README.md")
	rlmPath := filepath.Join(projectRoot, "docs", "agents", "RLM.md")
	initContractPath := filepath.Join(projectRoot, "docs", "specs", "0000_INIT_PROJECT.md")

	var sb strings.Builder
	sb.WriteString("Optimize this Kit durable ruleset for correctness, semantic clarity, and RLM just-in-time loading.\n\n")
	sb.WriteString("Ruleset file:\n")
	sb.WriteString("- " + filepath.Join(projectRoot, filepath.FromSlash(relPath)) + "\n\n")
	sb.WriteString("Creation context:\n")
	sb.WriteString("- name: " + input.Name + "\n")
	sb.WriteString("- slug: " + input.Slug + "\n")
	sb.WriteString("- applies_to: " + strings.Join(input.AppliesTo, ", ") + "\n")
	sb.WriteString("- read_policy_default: " + input.ReadPolicyDefault + "\n\n")
	sb.WriteString("Task:\n")
	sb.WriteString("1. Read the ruleset file and treat the captured context as the human source of truth.\n")
	sb.WriteString("2. Load only the Kit contract sections needed for this decision, starting with:\n")
	sb.WriteString("   - " + referencesReadme + "\n")
	sb.WriteString("   - " + rlmPath + "\n")
	sb.WriteString("   - " + initContractPath + "\n")
	sb.WriteString("3. Rewrite or reorganize the ruleset so it is durable, concise, scan-friendly, and directly useful to coding agents.\n")
	sb.WriteString("4. Preserve valid YAML front matter with `kind: ruleset`, `slug`, `status`, `applies_to`, and `read_policy_default`.\n")
	sb.WriteString("5. Preserve these required sections exactly: `Purpose`, `Applies When`, `Rules`, `Anti-Patterns`, `Verification`, and `Examples`.\n")
	sb.WriteString("6. Make rules specific and testable; move vague advice into concrete acceptance or verification guidance.\n")
	sb.WriteString("7. Keep the artifact pointer-loaded: do not inline it into AGENTS.md, CLAUDE.md, copilot instructions, or generated prompt bodies by default.\n")
	sb.WriteString("8. Do not create a broad policy engine or unrelated docs churn.\n")
	sb.WriteString("9. Run `kit check --project` and `kit rules list` after editing.\n\n")
	sb.WriteString("Output expectation:\n")
	sb.WriteString("- Edit only the ruleset unless the validation evidence requires a small contract/doc fix.\n")
	sb.WriteString("- Summarize changed files and verification results.\n")
	return sb.String()
}

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

package cli

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	kit "github.com/jamesonstone/kit/v3"
)

// embeddedRulesetRegistry returns the rules shipped inside this Kit binary.
func embeddedRulesetRegistry(context.Context) ([]registryRuleset, error) {
	entries, err := fs.ReadDir(kit.Rules, kit.RulesDir)
	if err != nil {
		return nil, fmt.Errorf("read embedded rules: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	var rulesets []registryRuleset
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		content, err := fs.ReadFile(kit.Rules, path.Join(kit.RulesDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read embedded rule %s: %w", entry.Name(), err)
		}
		slug := strings.TrimSuffix(entry.Name(), ".md")
		parsed := parseRuleset(string(content), entry.Name())
		if issues := validateRulesetDocument(parsed, slug); len(issues) > 0 {
			return nil, fmt.Errorf("embedded rule %s is invalid: %s", entry.Name(), strings.Join(issues, "; "))
		}
		hash, err := normalizedRulesetContentHash(string(content), parsed.Metadata.Status)
		if err != nil {
			return nil, fmt.Errorf("hash embedded rule %s: %w", entry.Name(), err)
		}
		rulesets = append(rulesets, registryRuleset{
			Slug:           parsed.Metadata.Slug,
			Content:        string(content),
			Metadata:       parsed.Metadata,
			NormalizedHash: hash,
		})
	}
	return rulesets, nil
}

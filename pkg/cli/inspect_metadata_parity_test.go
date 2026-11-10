package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/guidance"
)

func TestInspectRulesetMetadataValidatorParity(t *testing.T) {
	type metadataCase struct{ scope, status, policy, slug, applies string }
	base := metadataCase{"downstream", "active", "conditional", "sample", "testing"}
	cases := []metadataCase{base}
	for _, scope := range []string{"", "optional", "kit-maintainer", "unknown", " optional "} {
		c := base
		c.scope = scope
		cases = append(cases, c)
	}
	for _, status := range []string{"optional", "stale", "draft", "retired", "unknown", ""} {
		c := base
		c.status = status
		cases = append(cases, c)
	}
	for _, policy := range []string{"must", "evidence", "skip", "unknown", ""} {
		c := base
		c.policy = policy
		cases = append(cases, c)
	}
	for _, slug := range []string{"a-rule", "1-rule", "Bad", "", "wrong_slug"} {
		c := base
		c.slug = slug
		cases = append(cases, c)
	}
	for _, applies := range []string{"a-trigger", "1-trigger", "Bad", "", "wrong_trigger"} {
		c := base
		c.applies = applies
		cases = append(cases, c)
	}
	for i, c := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, ".kit.yaml"), []byte("schema_version: 2\n"), 0644); err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(root, "docs/references/rules")
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			name := c.slug
			if name == "" {
				name = "empty"
			}
			data := fmt.Sprintf("---\nkind: ruleset\nslug: %q\nstatus: %q\nregistry_scope: %q\nread_policy_default: %q\napplies_to: [%q]\n---\n", c.slug, c.status, c.scope, c.policy, c.applies)
			if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(data), 0644); err != nil {
				t.Fatal(err)
			}
			want := validRulesetRegistryScope(c.scope) && validRulesetStatus(c.status) && validRulesetReadPolicy(c.policy) && validateRulesetSlug(c.slug) == nil && validateRulesetSlug(c.applies) == nil
			report, err := guidance.Inspect(root)
			if (err == nil) != want {
				t.Fatalf("metadata %+v: inspect error %v, canonical valid=%v; %v", c, err, want, report.Errors)
			}
		})
	}
}

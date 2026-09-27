package cli

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func TestSyncRulesetUsesEmbeddedVersionUnlessLocallyEdited(t *testing.T) {
	old := registryRulesetForTest("delivery", []string{"git"})
	embedded := registryRulesetForTest("delivery", []string{"git"})
	embedded.Content = strings.Replace(embedded.Content, "## Rules", "## Rules\n\n- New embedded guidance.", 1)
	embedded.NormalizedHash = ""
	installed := config.RegistryArtifact{Kind: rulesetKind, Slug: "delivery", InstalledHash: old.NormalizedHash, State: registryArtifactStateManaged}

	t.Run("unmodified install takes the embedded version", func(t *testing.T) {
		result, err := syncRulesetRegistryContent(embedded, installed, old.Content, false)
		if err != nil || result.state != registryArtifactStateManaged || !strings.Contains(result.content, "New embedded guidance") {
			t.Fatalf("result = %#v, err = %v", result, err)
		}
	})
	t.Run("local edit is preserved and reported", func(t *testing.T) {
		edited := strings.Replace(old.Content, "## Rules", "## Rules\n\n- Project-specific guidance.", 1)
		result, err := syncRulesetRegistryContent(embedded, installed, edited, false)
		if err != nil || result.state != registryArtifactStateLocalCustom || result.content != edited || len(result.conflicts) == 0 {
			t.Fatalf("result = %#v, err = %v", result, err)
		}
		forced, err := syncRulesetRegistryContent(embedded, installed, edited, true)
		if err != nil || !strings.Contains(forced.content, "New embedded guidance") {
			t.Fatalf("forced = %#v, err = %v", forced, err)
		}
	})
	t.Run("legacy merged or conflict entries keep their local edits", func(t *testing.T) {
		edited := strings.Replace(old.Content, "## Rules", "## Rules\n\n- Merged project guidance.", 1)
		editedHash, err := normalizedRulesetContentHash(edited, old.Metadata.Status)
		if err != nil {
			t.Fatal(err)
		}
		legacy := []config.RegistryArtifact{
			{Kind: rulesetKind, Slug: "delivery", InstalledHash: editedHash, State: registryArtifactStateManaged,
				Sections: []config.RegistryArtifactSection{{Key: "RULES", InstalledHash: "sha256:base"}}},
			{Kind: rulesetKind, Slug: "delivery", InstalledHash: editedHash, State: registryArtifactStateConflict},
		}
		for _, state := range legacy {
			result, err := syncRulesetRegistryContent(embedded, state, edited, false)
			if err != nil || result.state != registryArtifactStateLocalCustom || result.content != edited {
				t.Fatalf("state %#v: result = %#v, err = %v", state, result, err)
			}
		}
	})
	t.Run("missing installed hash never counts as unmodified", func(t *testing.T) {
		result, err := syncRulesetRegistryContent(embedded, config.RegistryArtifact{State: registryArtifactStateManaged}, old.Content, false)
		if err != nil || result.state != registryArtifactStateLocalCustom || result.content != old.Content {
			t.Fatalf("result = %#v, err = %v", result, err)
		}
	})
	t.Run("identical content is managed without change", func(t *testing.T) {
		result, err := syncRulesetRegistryContent(embedded, config.RegistryArtifact{}, embedded.Content, false)
		if err != nil || result.state != registryArtifactStateManaged || result.content != embedded.Content {
			t.Fatalf("result = %#v, err = %v", result, err)
		}
	})
}

func TestEmbeddedRegistryShipsValidRulesWithoutNetwork(t *testing.T) {
	registry, err := embeddedRulesetRegistry(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(registry) < 10 {
		t.Fatalf("embedded registry has %d rules", len(registry))
	}
	for _, item := range registry {
		if item.NormalizedHash == "" || item.Metadata.Slug != item.Slug {
			t.Errorf("embedded rule %s is incomplete", item.Slug)
		}
	}
}

func TestRetiredRulesetsAreNotShipped(t *testing.T) {
	registry, err := embeddedRulesetRegistry(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range registry {
		if isRetiredRuleset(item.Slug) {
			t.Errorf("retired rule %s is still shipped", item.Slug)
		}
	}
}

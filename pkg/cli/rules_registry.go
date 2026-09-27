package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jamesonstone/kit/v3/internal/document"
)

const (
	rulesetRegistryScopeDownstream   = "downstream"
	rulesetRegistryScopeOptional     = "optional"
	inactiveRulesetStatus            = document.ReferenceStatusOptional
	registryArtifactStateManaged     = "managed"
	registryArtifactStateLocalCustom = "local-custom"
	registryArtifactStateConflict    = "conflict"

	registrySelectorDefaultTableWidth = 118
	registrySelectorMinimumTableWidth = 88
	registrySelectorMinimumSlugWidth  = 18
	registrySelectorMaximumSlugWidth  = 30
	registrySelectorMinimumDescWidth  = 22
)

type rulesetRegistryFetchFunc func(context.Context) ([]registryRuleset, error)

var rulesetRegistryFetcher rulesetRegistryFetchFunc = embeddedRulesetRegistry

type registryRuleset struct {
	Slug           string
	Content        string
	Metadata       rulesetMetadata
	NormalizedHash string
}

type registrySelectorEntry struct {
	Registry      registryRuleset
	Local         *rulesetDocument
	LocalContent  string
	Installed     bool
	Modified      bool
	CurrentActive bool
	DesiredActive bool
	Touched       bool
	RegistryState string
}

type registrySelectorSummary struct {
	Imported    int
	Activated   int
	Deactivated int
	Unchanged   int
}

func runRulesAddRegistrySelector(cmd interface {
	InOrStdin() io.Reader
	OutOrStdout() io.Writer
}, projectRoot string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	registry, err := rulesetRegistryFetcher(ctx)
	if err != nil {
		return err
	}
	if len(registry) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "No registry rulesets found.")
		return err
	}

	entries, err := buildRegistrySelectorEntries(projectRoot, registry)
	if err != nil {
		return err
	}
	if err := selectRegistryRulesets(cmd.InOrStdin(), cmd.OutOrStdout(), entries); err != nil {
		return err
	}

	summary, err := applyRegistryRulesetSelection(projectRoot, entries)
	if err != nil {
		return err
	}
	return printRegistryRulesetSummary(cmd.OutOrStdout(), summary)
}

func rulesetRegistrySourceDescription() string {
	return "rules embedded in Kit " + Version
}

func rulesetRegistryRulesetURL(slug string) string {
	return "Kit " + Version + " embedded rule " + rulesetTarget(slug)
}

func ensureTrailingNewline(content string) string {
	if strings.HasSuffix(content, "\n") {
		return content
	}
	return content + "\n"
}

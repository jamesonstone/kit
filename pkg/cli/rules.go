package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/document"
)

const (
	rulesetKind                    = "ruleset"
	rulesetDirRelPath              = "docs/references/rules"
	rulesetReferenceType           = "ruleset"
	rulesetReferenceIDPrefix       = "ruleset-"
	defaultRulesetReadPolicy       = document.ReferenceReadPolicyConditional
	rulesetReferenceStatus         = document.ReferenceStatusActive
	frontendRulesetAppliesTo       = "frontend"
	frontendProfileReferenceMarker = "frontend-profile"
)

var (
	rulesAddCopy        bool
	rulesAddEditor      string
	rulesAddEvidence    bool
	rulesAddForce       bool
	rulesAddInline      bool
	rulesAddMust        bool
	rulesAddOutputOnly  bool
	rulesAddSkip        bool
	rulesAddCustom      bool
	rulesAddConditional bool
	rulesAddUseVim      bool
	rulesLinkReadPolicy string
	rulesetSlugPattern  = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	rulesetSectionRe    = regexp.MustCompile(`(?m)^##\s+(.+)$`)
)

var rulesCmd = &cobra.Command{
	Use:     "rules",
	Aliases: []string{"rule"},
	Short:   "Manage durable repo-local rulesets",
	Long: `Import, preview, create, list, and link durable repo-local rulesets.

Rulesets live under docs/references/rules/ and are loaded through feature
front matter references. They are not inlined into always-loaded instruction
files or prompt bodies by default.`,
}

var rulesAddCmd = &cobra.Command{
	Use:   "add [slug]",
	Short: "Import or create a durable repo-local ruleset",
	Long: `Import or create a durable repo-local ruleset.

Without a slug, opens the registry selector so users can import available
rulesets from the Kit GitHub registry or toggle existing registry rules active
and inactive.

With a slug argument, creates a custom ruleset non-interactively for scripts.
Use --custom without a slug for the interactive custom ruleset builder; it asks
for the ruleset name, loading policy, applicability, and rule context, then
opens $EDITOR for the context by default, falls back to a vim-compatible editor
when $EDITOR is unset, and copies an agent optimization prompt after the
ruleset is saved.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRulesAdd,
}

var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List durable repo-local rulesets",
	Args:  cobra.NoArgs,
	RunE:  runRulesList,
}

var rulesViewCmd = &cobra.Command{
	Use:   "view <slug>",
	Short: "View a local or registry ruleset before adding it",
	Args:  cobra.ExactArgs(1),
	RunE:  runRulesView,
}

var rulesLinkCmd = &cobra.Command{
	Use:   "link <feature> <slug>",
	Short: "Link a ruleset to a feature through canonical references",
	Args:  cobra.ExactArgs(2),
	RunE:  runRulesLink,
}

type rulesetMetadata struct {
	Kind              string   `yaml:"kind"`
	Slug              string   `yaml:"slug"`
	Description       string   `yaml:"description"`
	Status            string   `yaml:"status"`
	AppliesTo         []string `yaml:"applies_to"`
	ReadPolicyDefault string   `yaml:"read_policy_default"`
	RegistryScope     string   `yaml:"registry_scope"`
}

type rulesetDocument struct {
	Path      string
	Body      string
	Metadata  rulesetMetadata
	Sections  map[string]string
	ParseErr  error
	LineHints []string
}

type rulesetAddInput struct {
	Name              string
	Slug              string
	AppliesTo         []string
	ReadPolicyDefault string
	Context           string
}

func init() {
	addFreeTextInputFlags(rulesAddCmd, &rulesAddUseVim, &rulesAddEditor)
	addInlineTextInputFlag(rulesAddCmd, &rulesAddInline)
	rulesAddCmd.Flags().BoolVar(&rulesAddCopy, "copy", false, "copy optimization prompt to clipboard even with --output-only")
	rulesAddCmd.Flags().BoolVar(&rulesAddEvidence, "evidence", false, "set read_policy_default to evidence")
	rulesAddCmd.Flags().BoolVar(&rulesAddForce, "force", false, "overwrite an existing ruleset file")
	rulesAddCmd.Flags().BoolVar(&rulesAddMust, "must", false, "set read_policy_default to must")
	rulesAddCmd.Flags().BoolVar(&rulesAddOutputOnly, "output-only", false, "output optimization prompt text instead of copying it to the clipboard")
	rulesAddCmd.Flags().BoolVar(&rulesAddSkip, "skip", false, "set read_policy_default to skip")
	rulesAddCmd.Flags().BoolVar(&rulesAddCustom, "custom", false, "open the interactive custom ruleset builder instead of the registry selector")
	rulesAddCmd.Flags().BoolVar(&rulesAddConditional, "conditional", false, "set read_policy_default to conditional")
	rulesLinkCmd.Flags().StringVar(&rulesLinkReadPolicy, "read-policy", defaultRulesetReadPolicy, "ruleset read policy for this feature reference (must or conditional)")

	rulesCmd.AddCommand(rulesAddCmd)
	rulesCmd.AddCommand(rulesListCmd)
	rulesCmd.AddCommand(rulesViewCmd)
	rulesCmd.AddCommand(rulesLinkCmd)
	rootCmd.AddCommand(rulesCmd)
}

func runRulesAdd(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	readPolicyDefault, policyExplicit, err := selectedRulesetReadPolicy()
	if err != nil {
		return err
	}
	if rulesAddInline && (rulesAddUseVim || rulesAddEditor != "") {
		return fmt.Errorf("--inline cannot be used with --vim or --editor")
	}

	if len(args) == 0 {
		if !rulesAddCustom && !rulesAddCopy && !rulesAddOutputOnly && !rulesAddUseVim && rulesAddEditor == "" && !rulesAddInline && !policyExplicit {
			return runRulesAddRegistrySelector(cmd, projectRoot)
		}
		return runRulesAddInteractive(cmd, projectRoot, readPolicyDefault, policyExplicit)
	}

	if rulesAddCustom {
		return fmt.Errorf("--custom requires interactive `kit rules add` with no slug")
	}
	if rulesAddCopy || rulesAddOutputOnly || rulesAddUseVim || rulesAddEditor != "" || rulesAddInline {
		return fmt.Errorf("--copy, --output-only, --vim, --editor, and --inline require interactive `kit rules add` with no slug")
	}

	slug := strings.TrimSpace(args[0])
	if err := validateRulesetSlug(slug); err != nil {
		return err
	}

	input := rulesetAddInput{
		Name:              slug,
		Slug:              slug,
		AppliesTo:         defaultRulesetAppliesTo(slug),
		ReadPolicyDefault: readPolicyDefault,
	}
	_, err = createRuleset(projectRoot, input)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created ruleset %s at %s\n", slug, rulesetTarget(slug))
	return err
}

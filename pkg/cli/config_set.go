package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/spf13/cobra"
)

type configSetReport struct {
	Target                 string              `json:"target"`
	TargetNotice           string              `json:"target_notice,omitempty"`
	Worktree               bool                `json:"worktree"`
	DryRun                 bool                `json:"dry_run"`
	Changed                bool                `json:"changed"`
	Issues                 config.IssueConfig  `json:"issues"`
	BranchPrefix           string              `json:"branch_prefix,omitempty"`
	ReconciliationRequired bool                `json:"reconciliation_required"`
	Next                   []string            `json:"next_actions"`
	Plugin                 string              `json:"plugin,omitempty"`
	PluginEnabled          *bool               `json:"plugin_enabled,omitempty"`
	Plugins                config.PluginConfig `json:"plugins,omitempty"`
}

var resolveConfigSetTarget = resolveReconcileTarget

func init() { configCmd.AddCommand(newConfigSetCommand()) }

func newConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "set", Short: "Set the issue provider or one plugin and report canonical reconcile steps",
		Long: `Set repository issue-provider configuration, or enable or disable one plugin, preserving unrelated YAML.
Issue mode requires --issue-provider github or linear and never prompts or stores credentials.
Linear needs a verified workspace slug and team UUID, with optional project UUID.
Use --linear-team-key for the verified team's branch prefix (for example ENG).
Existing target fields are retained unless explicitly supplied; pass an empty
--linear-project-id or --linear-team-key to clear that field. Changing the
workspace or team clears the stored key unless a new key is supplied.
GitHub remains the omitted default.

Primary checkouts stay read-only: writes use the same kit-reconcile target as
kit reconcile. Preview with --dry-run first; previews describe the invoking
checkout, while a primary-checkout write uses the canonical base/worktree.

This command changes configuration only. The result explicitly reports the target
and next actions: run kit reconcile --dry-run --diff, then kit reconcile there.
Reconcile alone owns managed rule/template migration and preserves custom files.
In issue mode, --json reports issue settings and next actions, not other config values.
Plugin mode adds the plugin id and the enabled list.

To enable or disable one plugin, pass --plugin <id> with --enable or --disable
and without issue-provider flags. The plugin must already be at
.kit/plugins/<id>/. --disable removes that id and drops plugins when the list
is empty. That mode also changes configuration only and reports the same
reconcile steps. Kit does not read the plugin config file.`,
		Args: cobra.NoArgs, SilenceUsage: true, RunE: runConfigSet,
	}
	bindIssueProviderFlags(cmd)
	cmd.Flags().String("plugin", "", "plugin id to enable or disable")
	cmd.Flags().Bool("enable", false, "enable the plugin named by --plugin")
	cmd.Flags().Bool("disable", false, "disable the plugin named by --plugin")
	cmd.Flags().Bool("dry-run", false, "preview issue settings without writing or preparing a worktree")
	cmd.Flags().Bool("json", false, "emit structured settings, target, and reconcile next actions")
	return cmd
}

func desiredIssueConfig(cmd *cobra.Command, old config.IssueConfig) (config.IssueConfig, error) {
	if !cmd.Flags().Changed("issue-provider") {
		return old, fmt.Errorf("--issue-provider github|linear is required")
	}
	old.Provider, _ = cmd.Flags().GetString("issue-provider")
	if old.Provider != "github" && old.Provider != "linear" {
		return old, fmt.Errorf("--issue-provider must be github or linear")
	}
	previous := old.Linear
	for _, field := range []struct {
		name  string
		value *string
	}{
		{"linear-workspace", &old.Linear.Workspace},
		{"linear-team-id", &old.Linear.TeamID},
		{"linear-project-id", &old.Linear.ProjectID},
		{"linear-team-key", &old.Linear.TeamKey},
	} {
		if cmd.Flags().Changed(field.name) {
			if old.Provider != "linear" {
				return old, fmt.Errorf("linear target flags require --issue-provider linear")
			}
			*field.value, _ = cmd.Flags().GetString(field.name)
		}
	}
	if !cmd.Flags().Changed("linear-team-key") && !previous.SameTeam(old.Linear) {
		old.Linear.TeamKey = ""
	}
	return old, old.Validate()
}

func loadIssueSetting(root string) (*config.Config, error) {
	cfg, inspection, err := config.LoadWithInspection(root)
	if err != nil {
		return nil, err
	}
	if inspection.SchemaState == config.SchemaStateNewer {
		return nil, fmt.Errorf("upgrade Kit before editing a newer config schema")
	}
	return cfg, nil // permits repairing invalid issue settings without defaulting
}

type issueSettingPlan struct {
	issues        config.IssueConfig
	before, after []byte
}

func planIssueSetting(root string, cmd *cobra.Command) (issueSettingPlan, error) {
	cfg, err := loadIssueSetting(root)
	if err != nil {
		return issueSettingPlan{}, err
	}
	desired, err := desiredIssueConfig(cmd, cfg.Issues)
	if err != nil {
		return issueSettingPlan{}, err
	}
	before, err := os.ReadFile(filepath.Join(root, config.ConfigFileName))
	if err != nil {
		return issueSettingPlan{}, err
	}
	after, err := config.PlanIssueConfig(before, desired)
	return issueSettingPlan{issues: desired, before: before, after: after}, err
}

func runConfigSet(cmd *cobra.Command, _ []string) error {
	pluginMode, err := pluginConfigSetMode(cmd)
	if err != nil {
		return err
	}
	if pluginMode {
		if issueConfigSetRequested(cmd) {
			return fmt.Errorf("set a plugin separately from the issue provider")
		}
		return runPluginConfigSet(cmd)
	}
	root, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	plan, err := planIssueSetting(root, cmd)
	if err != nil {
		return err
	}
	dry, _ := cmd.Flags().GetBool("dry-run")
	// Preserve target guidance without adding prose outside the JSON report.
	var targetNotice bytes.Buffer
	target, err := resolveConfigSetTarget(&targetNotice, root, dry, bytes.Equal(plan.before, plan.after))
	if err != nil {
		return err
	}
	if target.projectRoot != root {
		plan, err = planIssueSetting(target.projectRoot, cmd)
		if err != nil {
			return err
		}
	}

	changed := !bytes.Equal(plan.before, plan.after)
	if !dry {
		if err := config.ApplyIssueConfig(filepath.Join(target.projectRoot, config.ConfigFileName), plan.before, plan.after); err != nil {
			return err
		}
	}
	report := configSetReport{Target: target.projectRoot, TargetNotice: strings.TrimSpace(targetNotice.String()), Worktree: target.worktree, DryRun: dry, Changed: changed, Issues: plan.issues, BranchPrefix: plan.issues.BranchPrefix(), ReconciliationRequired: true,
		Next: []string{"cd " + shellQuoteArgument(target.projectRoot) + " && kit reconcile --dry-run --diff", "cd " + shellQuoteArgument(target.projectRoot) + " && kit reconcile"}}
	structured, _ := cmd.Flags().GetBool("json")
	if structured {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
	}
	if report.TargetNotice != "" {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), report.TargetNotice); err != nil {
			return err
		}
	}
	verb := "Configured"
	if dry {
		verb = "Would configure"
	}
	if !changed {
		verb = "Already configured"
	}
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "%s issue provider %s in %s\n", verb, plan.issues.EffectiveProvider(), report.Target); err != nil {
		return err
	}
	if err := printIssueBranchPrefix(out, report.BranchPrefix); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "Configuration only; canonical reconciliation is required."); err != nil {
		return err
	}
	for _, next := range report.Next {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), next); err != nil {
			return err
		}
	}
	return nil
}

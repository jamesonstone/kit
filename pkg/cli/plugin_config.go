package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/config"
	"github.com/jamesonstone/kit/v3/internal/plugin"
	"github.com/spf13/cobra"
)

func pluginConfigSetMode(cmd *cobra.Command) (bool, error) {
	enable, _ := cmd.Flags().GetBool("enable")
	disable, _ := cmd.Flags().GetBool("disable")
	if !cmd.Flags().Changed("plugin") && !enable && !disable {
		return false, nil
	}
	if enable && disable {
		return false, fmt.Errorf("pass only one of --enable or --disable")
	}
	if !enable && !disable {
		return false, fmt.Errorf("--plugin requires --enable or --disable")
	}
	id, _ := cmd.Flags().GetString("plugin")
	if strings.TrimSpace(id) == "" {
		return false, fmt.Errorf("--enable and --disable require --plugin")
	}
	return true, nil
}

func issueConfigSetRequested(cmd *cobra.Command) bool {
	for _, name := range []string{"issue-provider", "linear-workspace", "linear-team-id", "linear-project-id", "linear-team-key"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

type pluginSettingPlan struct {
	id            string
	enable        bool
	plugins       config.PluginConfig
	issues        config.IssueConfig
	before, after []byte
}

func planPluginSetting(root string, cmd *cobra.Command) (pluginSettingPlan, error) {
	before, err := os.ReadFile(filepath.Join(root, config.ConfigFileName))
	if err != nil {
		return pluginSettingPlan{}, err
	}
	cfg, inspection, err := config.ParseWithInspection(before)
	if err != nil {
		return pluginSettingPlan{}, err
	}
	if inspection.SchemaState == config.SchemaStateNewer {
		return pluginSettingPlan{}, fmt.Errorf("upgrade Kit before editing a newer config schema")
	}
	id, _ := cmd.Flags().GetString("plugin")
	enable, _ := cmd.Flags().GetBool("enable")
	next, err := togglePluginID(cfg.Plugins.Enabled, id, enable)
	if err != nil {
		return pluginSettingPlan{}, err
	}
	if _, err := plugin.LoadEnabled(root, next); err != nil {
		return pluginSettingPlan{}, err
	}
	after, err := config.PlanPluginEnabled(before, next)
	if err != nil {
		return pluginSettingPlan{}, err
	}
	return pluginSettingPlan{id: id, enable: enable, plugins: config.PluginConfig{Enabled: next}, issues: cfg.Issues, before: before, after: after}, nil
}

func togglePluginID(current []string, id string, enable bool) ([]string, error) {
	if err := plugin.ValidateID(id); err != nil {
		return nil, err
	}
	if enable {
		for _, existing := range current {
			if existing == id {
				return current, nil
			}
		}
		return append(append([]string{}, current...), id), nil
	}
	next := make([]string, 0, len(current))
	for _, existing := range current {
		if existing != id {
			next = append(next, existing)
		}
	}
	return next, nil
}

func runPluginConfigSet(cmd *cobra.Command) error {
	root, err := config.FindProjectRoot()
	if err != nil {
		return err
	}
	plan, err := planPluginSetting(root, cmd)
	if err != nil {
		return err
	}
	dry, _ := cmd.Flags().GetBool("dry-run")
	var targetNotice bytes.Buffer
	target, err := resolveConfigSetTarget(&targetNotice, root, dry, bytes.Equal(plan.before, plan.after))
	if err != nil {
		return err
	}
	if target.projectRoot != root {
		plan, err = planPluginSetting(target.projectRoot, cmd)
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
	enabled := plan.enable
	report := configSetReport{
		Target: target.projectRoot, TargetNotice: strings.TrimSpace(targetNotice.String()), Worktree: target.worktree,
		DryRun: dry, Changed: changed, Issues: plan.issues, BranchPrefix: plan.issues.BranchPrefix(),
		ReconciliationRequired: true, Plugin: plan.id, PluginEnabled: &enabled, Plugins: &plan.plugins,
		Next: []string{
			"cd " + shellQuoteArgument(target.projectRoot) + " && kit reconcile --dry-run --diff",
			"cd " + shellQuoteArgument(target.projectRoot) + " && kit reconcile",
		},
	}
	structured, _ := cmd.Flags().GetBool("json")
	if structured {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
	}
	return printPluginConfigSet(cmd, report)
}

func printPluginConfigSet(cmd *cobra.Command, report configSetReport) error {
	if report.TargetNotice != "" {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), report.TargetNotice); err != nil {
			return err
		}
	}
	verb := "Enabled"
	if report.PluginEnabled != nil && !*report.PluginEnabled {
		verb = "Disabled"
	}
	if !report.Changed {
		verb = "Already enabled"
		if report.PluginEnabled != nil && !*report.PluginEnabled {
			verb = "Already disabled"
		}
	} else if report.DryRun {
		verb = "Would enable"
		if report.PluginEnabled != nil && !*report.PluginEnabled {
			verb = "Would disable"
		}
	}
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "%s plugin %s in %s\n", verb, report.Plugin, report.Target); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "Configuration only; canonical reconciliation is required."); err != nil {
		return err
	}
	for _, next := range report.Next {
		if _, err := fmt.Fprintln(out, next); err != nil {
			return err
		}
	}
	return nil
}

func selectInitPlugins(cmd *cobra.Command, projectRoot string, cfg *config.Config) error {
	if cmd.Flags().Lookup("plugin") == nil {
		return nil
	}
	ids, err := cmd.Flags().GetStringArray("plugin")
	if err != nil || len(ids) == 0 {
		return err
	}
	if _, err := plugin.LoadEnabled(projectRoot, ids); err != nil {
		return err
	}
	cfg.Plugins.Enabled = ids
	return nil
}

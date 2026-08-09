package agentcli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

type registryListItem struct {
	Kind        string `json:"kind"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Version     int    `json:"version"`
	Installed   bool   `json:"installed"`
	State       string `json:"state,omitempty"`
	Path        string `json:"path"`
	Revision    string `json:"revision"`
}

func newRegistryCommand() *cobra.Command {
	command := &cobra.Command{Use: "registry", Short: "Inspect and administer typed contract artifacts"}
	command.AddCommand(newRegistryListCommand(""))
	command.AddCommand(newRegistryViewCommand(""))
	command.AddCommand(newRegistryAddCommand(""))
	command.AddCommand(newRegistryStatusCommand())
	return command
}

func newRegistryListCommand(kind string) *cobra.Command {
	jsonOutput := false
	command := &cobra.Command{
		Use:   "list",
		Short: "List registry artifacts and installed state",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runRegistryList(command, kind, jsonOutput)
		},
	}
	command.Flags().BoolVar(&jsonOutput, "json", false, "emit artifact state as JSON")
	return command
}

func runRegistryList(command *cobra.Command, kind string, jsonOutput bool) error {
	root, cfg, source, err := projectRegistryContext()
	if err != nil {
		return err
	}
	_ = root
	catalog, revision, err := source.LoadCatalog(command.Context(), cfg.Registry.Source)
	if err != nil {
		return err
	}
	var items []registryListItem
	for _, artifact := range registry.VisibleArtifacts(catalog) {
		if kind != "" && artifact.Kind != kind {
			continue
		}
		record, installed := registry.RecordByKey(cfg.Registry.Artifacts, artifact.Kind, artifact.Slug)
		items = append(items, registryListItem{
			Kind: artifact.Kind, Slug: artifact.Slug, Description: artifact.Description,
			Version: artifact.Version, Installed: installed, State: record.State,
			Path: artifact.TargetPath, Revision: revision,
		})
	}
	for _, record := range sortedRecords(cfg.Registry.Artifacts, kind) {
		if _, found := registry.FindArtifact(catalog, record.Kind, record.Slug); found {
			continue
		}
		items = append(items, registryListItem{
			Kind: record.Kind, Slug: record.Slug, Description: record.Description,
			Version: record.Version, Installed: true, State: record.State, Path: record.Path,
		})
	}
	if jsonOutput {
		return writeJSON(command.OutOrStdout(), items)
	}
	for _, item := range items {
		state := "available"
		if item.Installed {
			state = item.State
		}
		if _, err := fmt.Fprintf(command.OutOrStdout(), "%-10s %-34s %-16s %s\n", item.Kind, item.Slug, state, item.Description); err != nil {
			return err
		}
	}
	return nil
}

func newRegistryViewCommand(kind string) *cobra.Command {
	sourceOnly, localOnly, diff, provenance := false, false, false, false
	use := "view <kind>/<slug>"
	if kind != "" {
		use = "view <slug>"
	}
	command := &cobra.Command{
		Use:   use,
		Short: "Inspect local and registry artifact content",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			key := args[0]
			if kind != "" {
				key = registry.ArtifactKey(kind, key)
			}
			return runRegistryView(command, key, sourceOnly, localOnly, diff, provenance)
		},
	}
	command.Flags().BoolVar(&sourceOnly, "source", false, "show catalog source content")
	command.Flags().BoolVar(&localOnly, "local", false, "show materialized local content")
	command.Flags().BoolVar(&diff, "diff", false, "show local-to-source unified diff")
	command.Flags().BoolVar(&provenance, "provenance", false, "show installed provenance and state as JSON")
	return command
}

func runRegistryView(command *cobra.Command, key string, sourceOnly, localOnly, diff, provenance bool) error {
	kind, slug, err := parseArtifactKey(key)
	if err != nil {
		return err
	}
	root, cfg, source, err := projectRegistryContext()
	if err != nil {
		return err
	}
	catalog, revision, err := source.LoadCatalog(command.Context(), cfg.Registry.Source)
	if err != nil {
		return err
	}
	artifact, found := registry.FindArtifact(catalog, kind, slug)
	var sourceContent string
	if found {
		sourceContent, err = source.LoadArtifact(command.Context(), cfg.Registry.Source, artifact, revision)
		if err != nil {
			return err
		}
	}
	record, installed := registry.RecordByKey(cfg.Registry.Artifacts, kind, slug)
	var localContent string
	if installed {
		localContent, _, err = registry.ReadOptional(root, record.Path)
		if err != nil {
			return err
		}
	}
	switch {
	case provenance:
		if !installed {
			return fmt.Errorf("artifact %s is not installed", key)
		}
		return writeJSON(command.OutOrStdout(), record)
	case diff:
		if !found || !installed {
			return fmt.Errorf("--diff requires both local and registry content")
		}
		if _, err := fmt.Fprint(command.OutOrStdout(), registry.RenderDiff(registry.Plan{Changes: []registry.Change{{Path: record.Path, Before: localContent, After: sourceContent}}})); err != nil {
			return err
		}
	case sourceOnly:
		if !found {
			return fmt.Errorf("registry artifact %s was not found", key)
		}
		_, err = fmt.Fprint(command.OutOrStdout(), sourceContent)
		return err
	case localOnly || installed:
		if !installed {
			return fmt.Errorf("artifact %s is not installed", key)
		}
		_, err = fmt.Fprint(command.OutOrStdout(), localContent)
		return err
	case found:
		_, err = fmt.Fprint(command.OutOrStdout(), sourceContent)
		return err
	default:
		return fmt.Errorf("artifact %s was not found", key)
	}
	return nil
}

func newRegistryAddCommand(kind string) *cobra.Command {
	use := "add <kind>/<slug>"
	if kind != "" {
		use = "add <slug>"
	}
	return &cobra.Command{
		Use: use, Short: "Install one registry artifact", Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			key := args[0]
			if kind != "" {
				key = registry.ArtifactKey(kind, key)
			}
			return runRegistryAdd(command, key, kind == registry.KindRuleset)
		},
	}
}

func runRegistryAdd(command *cobra.Command, key string, allowLocal bool) error {
	kind, slug, err := parseArtifactKey(key)
	if err != nil {
		return err
	}
	root, cfg, source, err := projectRegistryContext()
	if err != nil {
		return err
	}
	plan, err := registry.BuildAddPlan(command.Context(), root, source, kind, slug)
	if errors.Is(err, registry.ErrArtifactNotFound) && allowLocal {
		plan, err = registry.BuildLocalRulesetPlan(root, slug)
	}
	if err != nil {
		return err
	}
	if plan.State == "attention-needed" {
		return &exitError{code: 2, err: fmt.Errorf("artifact requires reconciliation before it can be installed")}
	}
	if err := registry.ApplyPlan(root, plan); err != nil {
		return err
	}
	_ = cfg
	return writePlanHuman(command.OutOrStdout(), plan, true)
}

func parseArtifactKey(key string) (string, string, error) {
	parts := strings.Split(key, "/")
	if len(parts) != 2 || (parts[0] != registry.KindRuleset && parts[0] != registry.KindWorkflow) || !registry.ValidSlug(parts[1]) {
		return "", "", fmt.Errorf("artifact %q must be ruleset/<slug> or workflow/<slug>", key)
	}
	return parts[0], parts[1], nil
}

func projectRegistryContext() (string, registry.ProjectConfig, registry.Source, error) {
	root, err := registry.FindProjectRoot(mustGetwd())
	if err != nil {
		return "", registry.ProjectConfig{}, nil, err
	}
	cfg, _, err := registry.LoadProject(root)
	if err != nil {
		return "", registry.ProjectConfig{}, nil, err
	}
	return root, cfg, sourceFor(root, cfg.Registry.Source), nil
}

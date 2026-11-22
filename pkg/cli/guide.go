package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/guide"
	"github.com/spf13/cobra"
)

func newGuideCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guide [command path]",
		Short: "Find and retrieve installed command guidance for coding agents",
		Long: `Use this guide to choose a Kit command. The index maps installed command paths to their purposes and lists repository artifact locations. Give a command path to retrieve its usage, flags, and examples. Use --json for detailed structured output, with or without a command path.

The guide derives command guidance from the installed binary's help. It works outside a project without network access, configuration changes, or usage recording.`,
		Example:           "  kit guide\n  kit guide rules view\n  kit guide reconcile --json",
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
		RunE:              runGuide,
	}
	cmd.Flags().Bool("json", false, "Output detailed guidance as deterministic JSON")
	return cmd
}

func runGuide(cmd *cobra.Command, args []string) error {
	root := cmd.Root()
	d := guide.Build(root)
	out := cmd.OutOrStdout()
	asJSON, _ := cmd.Flags().GetBool("json")
	if len(args) == 0 {
		if asJSON {
			return guide.WriteJSON(out, d)
		}
		return guide.WriteIndex(out, d)
	}
	path := strings.Join(args, " ")
	if path != root.Name() && !strings.HasPrefix(path, root.Name()+" ") {
		path = root.Name() + " " + path
	}
	index := slices.IndexFunc(d.Commands, func(c guide.Command) bool { return c.Path == path })
	if index == -1 {
		return fmt.Errorf("unknown installed command %q. Run %s guide to list command paths", path, root.Name())
	}
	if asJSON {
		d.Instructions = "Read your repository's agent instructions. This entry describes an installed command."
		d.Artifacts = []guide.Artifact{}
		d.Commands = d.Commands[index : index+1]
		return guide.WriteJSON(out, d)
	}
	return guide.WriteCommand(out, d.Commands[index])
}

func init() { rootCmd.AddCommand(newGuideCommand()) }

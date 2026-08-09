package agentcli

import (
	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

func newRulesCommand() *cobra.Command {
	command := &cobra.Command{Use: "rules", Short: "Manage repository-local coding-agent rules"}
	command.AddCommand(newRegistryAddCommand(registry.KindRuleset))
	command.AddCommand(newRegistryListCommand(registry.KindRuleset))
	command.AddCommand(newRegistryViewCommand(registry.KindRuleset))
	return command
}

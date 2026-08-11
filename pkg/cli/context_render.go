package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	contextcontract "github.com/jamesonstone/kit/internal/context"
)

func renderContextContract(cmd *cobra.Command, contract contextcontract.Contract) error {
	out := cmd.OutOrStdout()
	state := "ready"
	if contract.Blocked {
		state = "blocked"
	}
	if _, err := fmt.Fprintf(out, "Context %s: workflow=%s evidence=%d\n", state, contract.Request.Workflow, len(contract.Evidence)); err != nil {
		return err
	}
	for _, item := range contract.Evidence {
		requirement := "optional"
		if item.Required {
			requirement = "required"
		}
		if _, err := fmt.Fprintf(out, "  %-12s %-8s %-7s %s\n", item.Kind, requirement, item.State, item.Path); err != nil {
			return err
		}
	}
	for _, diagnostic := range contract.Diagnostics {
		path := ""
		if diagnostic.Path != "" {
			path = " [" + diagnostic.Path + "]"
		}
		if _, err := fmt.Fprintf(out, "  %s %s%s: %s\n", strings.ToUpper(diagnostic.Level), diagnostic.Code, path, diagnostic.Message); err != nil {
			return err
		}
	}
	return nil
}

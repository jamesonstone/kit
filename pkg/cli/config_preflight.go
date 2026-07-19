package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jamesonstone/kit/internal/config"
)

func runAutomaticConfigCheck(cmd *cobra.Command, args []string) error {
	if skipAutomaticConfigCheck(cmd) {
		return nil
	}
	projectRoot, found, err := config.FindProjectRootOptional()
	if err != nil || !found {
		return err
	}
	cfg, inspection, err := config.LoadWithInspection(projectRoot)
	if err != nil {
		return err
	}
	if inspection.SchemaState == config.SchemaStateNewer {
		return fmt.Errorf("%s", inspection.Findings[0].Message)
	}
	for _, finding := range inspection.Findings {
		if finding.Severity == config.FindingError && !strings.HasPrefix(finding.Field, "aws.") {
			return fmt.Errorf("invalid .kit.yaml field %s: %s", finding.Field, finding.Message)
		}
	}
	if !automaticConfigPromptAllowed(cmd) {
		return nil
	}
	_, err = remediateProjectConfig(projectRoot, cfg, inspection, configRemediationOptions{
		Interactive: true,
		Input:       cmd.InOrStdin(),
		Output:      cmd.ErrOrStderr(),
	})
	return err
}

func skipAutomaticConfigCheck(cmd *cobra.Command) bool {
	parts := strings.Fields(cmd.CommandPath())
	if len(parts) < 2 {
		return true
	}
	switch parts[1] {
	case "init", "capabilities", "config", "aws", "health", "registry", "instructions", "upgrade", "version", "completion", "help":
		return true
	default:
		return false
	}
}

func automaticConfigPromptAllowed(cmd *cobra.Command) bool {
	for _, name := range []string{"json", "output-only", "dry-run"} {
		if flag := cmd.Flags().Lookup(name); flag != nil && flag.Value.String() == "true" {
			return false
		}
	}
	in, ok := cmd.InOrStdin().(*os.File)
	return ok && term.IsTerminal(int(in.Fd())) && terminalWriterCheck(cmd.ErrOrStderr())
}

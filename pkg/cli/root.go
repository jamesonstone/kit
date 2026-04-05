package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "kit",
	Short: "🧰 Kit is a general-purpose harness for thought work",
	Long: banner() + `
Kit is a general-purpose harness for disciplined thought work.
Its strongest engine is a document-first, spec-driven workflow, but the
harness also supports ad hoc execution, catch-up, handoff, summarization,
review, and orchestration.

The current command surface is packaged around repository and software
workflows, but the underlying harness patterns generalize to research,
strategy, operations, writing, policy, and other structured fields.

` + flowDiagram(),
	Version: Version,
}

func init() {
	rootCmd.SetVersionTemplate("kit version {{.Version}}\n")
	configureRootHelp()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

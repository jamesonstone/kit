package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

const (
	scaffoldAgentsVersionColumnWidth = 7
	scaffoldAgentsModelColumnWidth   = 12
	scaffoldAgentsNotesColumnWidth   = 56
)

func init() {
	scaffoldAgentsCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_ = renderScaffoldAgentsHelp(cmd)
	})
}

func renderScaffoldAgentsHelp(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	style := styleForWriter(out)

	if _, err := fmt.Fprintln(out, cmd.Long); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, style.title("🚀", "Usage")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  %s\n", cmd.UseLine()); err != nil {
		return err
	}
	if len(cmd.Aliases) > 0 {
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, style.title("🏷️", "Aliases")); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "  %s\n", cmd.NameAndAliases()); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, style.title("🧭", "Version Models")); err != nil {
		return err
	}
	printScaffoldAgentsVersionTable(out, style)

	if flags := cmd.LocalFlags().FlagUsages(); flags != "" {
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, style.title("⚙️", "Flags")); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, flags); err != nil {
			return err
		}
	}
	if inherited := cmd.InheritedFlags().FlagUsages(); inherited != "" {
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, style.title("🌐", "Global Flags")); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(out, inherited); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(out, "\n%s \"%s --version 2 --force\" to migrate a legacy repo to the thin ToC model.\n",
		rootHelpMoreInfoLabel(style.enabled), cmd.CommandPath())
	return err
}

func printScaffoldAgentsVersionTable(w io.Writer, style humanOutputStyle) {
	printScaffoldAgentsTableBorder(w, style, "┌", "┬", "┐")
	printScaffoldAgentsTableRow(w, style, "Version", "Model", "Notes")
	printScaffoldAgentsTableBorder(w, style, "├", "┼", "┤")
	printScaffoldAgentsTableRow(w, style, "1", "verbose", "legacy AGENTS/CLAUDE model with dense top-level instruction files")
	printScaffoldAgentsTableBorder(w, style, "├", "┼", "┤")
	printScaffoldAgentsTableRow(w, style, "2", "toc/rlm", "recommended thin entrypoints plus docs/agents, docs/references, and repo-local RLM routing")
	printScaffoldAgentsTableBorder(w, style, "└", "┴", "┘")
}

func printScaffoldAgentsTableBorder(w io.Writer, style humanOutputStyle, left, middle, right string) {
	line := left +
		strings.Repeat("─", scaffoldAgentsVersionColumnWidth+2) +
		middle +
		strings.Repeat("─", scaffoldAgentsModelColumnWidth+2) +
		middle +
		strings.Repeat("─", scaffoldAgentsNotesColumnWidth+2) +
		right
	if style.enabled {
		line = dim + line + reset
	}
	fmt.Fprintln(w, line)
}

func printScaffoldAgentsTableRow(w io.Writer, style humanOutputStyle, version, model, notes string) {
	versionText := version
	modelText := model
	if style.enabled && version == "2" {
		versionText = whiteBold + version + reset
		modelText = whiteBold + model + reset
	}

	fmt.Fprintf(
		w,
		"│ %-*s │ %-*s │ %-*s │\n",
		scaffoldAgentsVersionColumnWidth,
		versionText,
		scaffoldAgentsModelColumnWidth,
		modelText,
		scaffoldAgentsNotesColumnWidth,
		notes,
	)
}

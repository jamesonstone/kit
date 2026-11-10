package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var commandOrder = map[string]int{
	"guide": 0, "init": 1, "spec": 2,
	"inspect": 9, "status": 10, "registry": 11, "health": 12,
	"config": 14, "aws": 15, "check": 16,
	"rules": 19, "reconcile": 20, "usage": 21,
	"upgrade": 31, "version": 32, "completion": 33, "help": 34,
}

type commandSection struct {
	title    string
	commands []string
}

var rootCommandSections = []commandSection{
	{title: "Agent Workflow", commands: []string{"guide", "init", "spec"}},
	{title: "Inspect & Repair", commands: []string{"inspect", "status", "registry", "health", "config", "aws", "check", "rules", "reconcile", "usage"}},
	{title: "Utilities", commands: []string{"upgrade", "version", "completion", "help"}},
}

func configureRootHelp() {
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		sortSubcommands(cmd)
		if cmd == rootCmd {
			_ = renderRootHelp(cmd)
			return
		}
		cmd.SetHelpTemplate(helpTemplate(terminalWriterCheck(cmd.OutOrStdout())))
		defaultHelp(cmd, args)
	})

	defaultUsage := rootCmd.UsageFunc()
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		sortSubcommands(cmd)
		cmd.SetUsageTemplate(usageTemplate(terminalWriterCheck(cmd.OutOrStdout())))
		return defaultUsage(cmd)
	})
}

func sortSubcommands(cmd *cobra.Command) {
	sort.SliceStable(cmd.Commands(), func(i, j int) bool {
		iOrder, iOk := commandOrder[cmd.Commands()[i].Name()]
		jOrder, jOk := commandOrder[cmd.Commands()[j].Name()]
		if !iOk {
			iOrder = 50
		}
		if !jOk {
			jOrder = 50
		}
		return iOrder < jOrder
	})
}

func renderRootHelp(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	style := styleForWriter(out)

	if _, err := fmt.Fprintln(out, strings.TrimRight(rootLong(style), "\n")); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, style.title("🚀", "Usage")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  %s [command]\n", cmd.CommandPath()); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "\n  Coding agents: run `%s guide` first. It is derived from each command's help.\n  `%s <command> --help` shows that command. `%s guide --json` is the same guide as data.\n", cmd.CommandPath(), cmd.CommandPath(), cmd.CommandPath()); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, style.title("🧰", "Available Commands")); err != nil {
		return err
	}

	namePadding := rootHelpNamePadding(cmd)
	for _, section := range rootCommandSections {
		rendered := false
		for _, name := range section.commands {
			command := findVisibleSubcommand(cmd, name)
			if command == nil {
				continue
			}
			if !rendered {
				if _, err := fmt.Fprintln(out); err != nil {
					return err
				}
				if _, err := fmt.Fprintln(out, style.label(section.title)); err != nil {
					return err
				}
				rendered = true
			}
			if _, err := fmt.Fprintf(out, "  %s %s\n", padRight(command.Name(), namePadding), command.Short); err != nil {
				return err
			}
		}
	}

	flags := strings.TrimSpace(cmd.Flags().FlagUsages())
	if flags != "" {
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
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

	_, err := fmt.Fprintf(out, "\n%s \"%s [command] --help\" for more information about a command.\n",
		rootHelpMoreInfoLabel(style.enabled), cmd.CommandPath())
	return err
}

func rootHelpNamePadding(cmd *cobra.Command) int {
	maxWidth := 0
	for _, section := range rootCommandSections {
		for _, name := range section.commands {
			command := findVisibleSubcommand(cmd, name)
			if command == nil {
				continue
			}
			if width := len(command.Name()); width > maxWidth {
				maxWidth = width
			}
		}
	}
	if maxWidth == 0 {
		return 12
	}
	return maxWidth + 2
}

func findVisibleSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subcommand := range cmd.Commands() {
		if subcommand.Name() != name {
			continue
		}
		if !subcommand.IsAvailableCommand() && subcommand.Name() != "help" {
			return nil
		}
		return subcommand
	}
	return nil
}

func rootHelpMoreInfoLabel(human bool) string {
	if human {
		return "🔎 Use"
	}
	return "Use"
}

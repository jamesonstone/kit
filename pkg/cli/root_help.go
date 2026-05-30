package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var commandOrder = map[string]int{
	"init":        1,
	"agents":      7,
	"scaffold":    8,
	"brainstorm":  9,
	"backlog":     10,
	"spec":        11,
	"plan":        12,
	"tasks":       13,
	"loop":        14,
	"resume":      15,
	"implement":   16,
	"reflect":     17,
	"pause":       18,
	"complete":    19,
	"status":      20,
	"map":         21,
	"rm":          22,
	"remove":      22,
	"check":       23,
	"verify":      24,
	"trace":       25,
	"replay":      26,
	"state":       27,
	"eval":        28,
	"rules":       29,
	"rollup":      30,
	"code-review": 31,
	"reconcile":   32,
	"handoff":     33,
	"summarize":   34,
	"catchup":     35,
	"dispatch":    36,
	"prompt":      37,
	"set":         38,
	"skill":       40,
	"skills":      40,
	"upgrade":     88,
	"update":      88,
	"version":     89,
	"completion":  91,
	"help":        92,
}

type commandSection struct {
	title    string
	commands []string
}

var rootCommandSections = []commandSection{
	{title: "Setup", commands: []string{"init", "scaffold"}},
	{
		title: "Workflow",
		commands: []string{
			"brainstorm",
			"backlog",
			"spec",
			"plan",
			"tasks",
			"loop",
			"resume",
			"implement",
			"reflect",
			"pause",
			"complete",
			"rm",
		},
	},
	{title: "Inspect & Repair", commands: []string{"status", "map", "check", "verify", "trace", "replay", "state", "eval", "rules", "reconcile"}},
	{title: "Prompt Utilities", commands: []string{"prompt", "set", "handoff", "summarize", "dispatch", "code-review", "skill"}},
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
		if cmd == scaffoldAgentsCmd {
			_ = renderScaffoldAgentsHelp(cmd)
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

	if _, err := fmt.Fprintln(out, strings.TrimRight(cmd.Long, "\n")); err != nil {
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

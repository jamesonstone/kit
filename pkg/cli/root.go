// package cli implements the Kit command-line interface.
package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

// ANSI color codes for consistent theming.
const (
	reset        = "\033[0m"
	dim          = "\033[38;5;245m"
	whiteBold    = "\033[1;37m"
	gray         = "\033[38;5;240m"
	constitution = "\033[38;5;220m" // gold/yellow
	spec         = "\033[38;5;39m"  // bright cyan
	plan         = "\033[38;5;82m"  // bright green
	tasks        = "\033[38;5;213m" // bright pink
	implement    = "\033[38;5;208m" // orange
	reflect      = "\033[38;5;141m" // light purple
)

// banner returns the Kit ASCII art banner with pink-to-black gradient.
func banner() string {
	// ANSI 256-color codes for pink-to-black gradient
	colors := []string{
		"\033[38;5;213m", // bright pink
		"\033[38;5;177m", // pink
		"\033[38;5;134m", // dark pink/magenta
		"\033[38;5;97m",  // dark magenta
		"\033[38;5;60m",  // very dark purple
		"\033[38;5;238m", // near black
	}

	lines := []string{
		"██╗  ██╗██╗████████╗",
		"██║ ██╔╝██║╚══██╔══╝",
		"█████╔╝ ██║   ██║   ",
		"██╔═██╗ ██║   ██║   ",
		"██║  ██╗██║   ██║   ",
		"╚═╝  ╚═╝╚═╝   ╚═╝   ",
	}

	var result string
	for i, line := range lines {
		result += "                                        " + colors[i] + line + reset + "\n"
	}
	result += "\n"
	result += "                                   " + dim + "Spec-Driven Development Toolkit" + reset + "\n"
	return result
}

// flowDiagram returns the colorized artifact pipeline flow diagram.
func flowDiagram() string {
	return whiteBold + "Project Initialization" + reset + dim + " (run once, update as needed):" + reset + `
` + gray + `┌──────────────┐` + reset + `
` + gray + `│ ` + constitution + `Constitution` + reset + gray + ` │  ← ` + reset + dim + `global constraints, principles, priors` + reset + `
` + gray + `└──────────────┘` + reset + `

` + whiteBold + `Core Development Loop:` + reset + `
` + gray + `┌───────────────┐    ┌──────┐    ┌───────┐    ┌────────────────┐    ┌────────────┐` + reset + `
` + gray + `│ ` + spec + `Specification` + reset + gray + ` │ ─▶ │ ` + plan + `Plan` + reset + gray + ` │ ─▶ │ ` + tasks + `Tasks` + reset + gray + ` │ ─▶ │ ` + implement + `Implementation` + reset + gray + ` │ ─▶ │ ` + reflect + `Reflection` + reset + gray + ` │ ─┐` + reset + `
` + gray + `└───────────────┘    └──────┘    └───────┘    └────────────────┘    └────────────┘  │` + reset + `
` + gray + `       ▲                                                                            │` + reset + `
` + gray + `       └────────────────────────────────────────────────────────────────────────────┘` + reset + `

` + whiteBold + `Artifact Pipeline:` + reset + `
  1. ` + constitution + `Constitution` + reset + dim + `   — strategy, patterns, long-term vision (kept updated)` + reset + `
  2. ` + spec + `Specification` + reset + dim + `  — what is being built and why` + reset + `
  3. ` + plan + `Plan` + reset + dim + `           — how it will be built` + reset + `
  4. ` + tasks + `Tasks` + reset + dim + `          — executable work units` + reset + `
  5. ` + implement + `Implementation` + reset + dim + ` — execution outside Kit's core scope` + reset + `
  6. ` + reflect + `Reflection` + reset + dim + `     — verify correctness, refine understanding` + reset
}

var rootCmd = &cobra.Command{
	Use:   "kit",
	Short: "Kit is a document-centered CLI for spec-driven development",
	Long: banner() + `
Kit helps teams reach a high-confidence understanding of a problem
and its solution before implementation, using open standards and
universally portable documents.

` + flowDiagram(),
	Version: Version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// commandOrder defines the display order of commands in help.
var commandOrder = map[string]int{
	// project initialization
	"init": 1,
	// core development loop
	"spec":      10,
	"plan":      11,
	"tasks":     12,
	"implement": 13,
	"status":    14,
	// verification and state
	"check":  20,
	"rollup": 21,
	// context management
	"handoff":   30,
	"summarize": 31,
	"reflect":   32,
	// utility
	"scaffold-agents": 90,
	"completion":      91,
	"help":            92,
}

func init() {
	rootCmd.SetVersionTemplate("kit version {{.Version}}\n")

	// custom help to order commands
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// sort subcommands by our custom order
		sort.SliceStable(cmd.Commands(), func(i, j int) bool {
			iOrder, iOk := commandOrder[cmd.Commands()[i].Name()]
			jOrder, jOk := commandOrder[cmd.Commands()[j].Name()]
			if !iOk {
				iOrder = 50 // default middle
			}
			if !jOk {
				jOrder = 50
			}
			return iOrder < jOrder
		})
		defaultHelp(cmd, args)
	})

	// custom usage to order commands
	defaultUsage := rootCmd.UsageFunc()
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
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
		return defaultUsage(cmd)
	})
}

// unused but required to avoid "imported and not used" error for strings
var _ = strings.TrimSpace

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
var clipboardCopyFunc = copyToClipboard

// ANSI color codes for consistent theming.
const (
	reset        = "\033[0m"
	dim          = "\033[38;5;245m"
	whiteBold    = "\033[1;37m"
	gray         = "\033[38;5;240m"
	constitution = "\033[38;5;220m" // gold/yellow
	brainstorm   = "\033[38;5;117m" // light blue
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
	lines := []string{
		whiteBold + "Project Initialization" + reset + dim + " (run once, update as needed):" + reset,
		gray + "┌──────────────┐" + reset,
		gray + "│ " + constitution + "Constitution" + reset + gray + " │  ← " + reset + dim + "global constraints, principles, priors" + reset,
		gray + "└──────────────┘" + reset,
		"",
		whiteBold + "Optional Research Step:" + reset,
		gray + "  ┌────────────┐" + reset,
		gray + "  │ " + brainstorm + "Brainstorm" + reset + gray + " │  ← " + reset + dim + "codebase research, framing, options, affected files" + reset,
		gray + "  └─────┬──────┘" + reset,
		gray + "        │" + reset,
		gray + "        ▼" + reset,
		"",
		whiteBold + "Core Development Loop:" + reset,
		gray + "┌───────────────┐    ┌──────┐    ┌───────┐    ┌────────────────┐    ┌────────────┐" + reset,
		gray + "│ " + spec + "Specification" + reset + gray + " │ ─▶ │ " + plan + "Plan" + reset + gray + " │ ─▶ │ " + tasks + "Tasks" + reset + gray + " │ ─▶ │ " + implement + "Implementation" + reset + gray + " │ ─▶ │ " + reflect + "Reflection" + reset + gray + " │ ─┐" + reset,
		gray + "└───────────────┘    └──────┘    └───────┘    └────────────────┘    └────────────┘  │" + reset,
		gray + "       ▲                                                                            │" + reset,
		gray + "       └────────────────────────────────────────────────────────────────────────────┘" + reset,
		"",
		whiteBold + "Artifact Pipeline:" + reset,
		"  1. " + constitution + "Constitution" + reset + dim + "   — strategy, patterns, long-term vision (kept updated)" + reset,
		"  2. " + brainstorm + "Brainstorm" + reset + dim + "     — optional research and framing before the spec" + reset,
		"  3. " + spec + "Specification" + reset + dim + "  — what is being built and why" + reset,
		"  4. " + plan + "Plan" + reset + dim + "           — how it will be built" + reset,
		"  5. " + tasks + "Tasks" + reset + dim + "          — executable work units" + reset,
		"  6. " + implement + "Implementation" + reset + dim + " — execution outside Kit's core scope" + reset,
		"  7. " + reflect + "Reflection" + reset + dim + "     — verify correctness, refine understanding" + reset,
	}

	return strings.Join(lines, "\n")
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
	"scaffold":   8,
	"brainstorm": 9,
	"spec":       10,
	"plan":       11,
	"tasks":      12,
	"implement":  13,
	"reflect":    14,
	"complete":   15,
	"status":     16,
	// verification and state
	"check":       20,
	"rollup":      21,
	"code-review": 22,
	// context management
	"handoff":   30,
	"summarize": 31,
	"catchup":   32,
	"dispatch":  33,
	"skill":     40,
	"skills":    40,
	// utility
	"upgrade":         88,
	"update":          88,
	"version":         89,
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

func formatAgentInstructionBlock(prompt string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(prompt)
	if !strings.HasSuffix(prompt, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("---\n")
	return sb.String()
}

// outputPrompt handles consistent output behavior for --output-only and --copy flags.
// if copy=true, copies the raw prompt to the clipboard
// if outputOnly=true, outputs the raw prompt without status text or dividers
// otherwise, outputs the prompt inside a standardized markdown copy block
func outputPrompt(prompt string, outputOnly, copy bool) error {
	return writePrompt(prepareAgentPrompt(prompt), outputOnly, copy)
}

func outputPromptWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	return writePromptWithClipboardDefault(prepareAgentPrompt(prompt), outputOnly, copy)
}

func outputPromptWithoutSubagentsWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	return writePromptWithClipboardDefault(preparePromptWithoutSubagents(prompt), outputOnly, copy)
}

func outputPromptWithoutSubagents(prompt string, outputOnly, copy bool) error {
	return writePrompt(preparePromptWithoutSubagents(prompt), outputOnly, copy)
}

func writePrompt(prompt string, outputOnly, copy bool) error {
	if copy {
		if err := clipboardCopyFunc(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		if outputOnly {
			fmt.Print(prompt)
			return nil
		}
		fmt.Println("Copied agent instructions to clipboard.")
		return nil
	}
	if outputOnly {
		fmt.Print(prompt)
		return nil
	}

	fmt.Println("Copy this section to the Agent:")
	fmt.Print(formatAgentInstructionBlock(prompt))
	return nil
}

func writePromptWithClipboardDefault(prompt string, outputOnly, copy bool) error {
	shouldCopy := !outputOnly || copy
	if shouldCopy {
		if err := clipboardCopyFunc(prompt); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
	}

	if outputOnly {
		fmt.Print(prompt)
		return nil
	}

	fmt.Println("Copied agent instructions to the clipboard.")
	return nil
}

func printWorkflowInstructions(currentStep string, nextSteps []string) {
	fmt.Println(whiteBold + "Workflow" + reset)
	fmt.Println(dim + "Pipeline: [optional brainstorm] -> spec -> plan -> tasks -> implement -> reflect" + reset)
	fmt.Printf("Current step: %s\n", currentStep)
	if len(nextSteps) > 0 {
		fmt.Println("Next steps:")
		for _, step := range nextSteps {
			fmt.Printf("  - %s\n", step)
		}
	}
	fmt.Println()
}

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
	result += "                                   " + dim + "General-Purpose Thought-Work Harness" + reset + "\n"
	return result
}

// flowDiagram returns the colorized artifact pipeline flow diagram.
func flowDiagram() string {
	lines := []string{
		whiteBold + "🧱 Project Initialization" + reset + dim + " (run once, update as needed):" + reset,
		gray + "┌──────────────┐" + reset,
		gray + "│ " + constitution + "Constitution" + reset + gray + " │  ← " + reset + dim + "global constraints, principles, priors" + reset,
		gray + "└──────────────┘" + reset,
		"",
		whiteBold + "🧠 Optional Research Step" + reset + dim + ":" + reset,
		gray + "  ┌────────────┐" + reset,
		gray + "  │ " + brainstorm + "Brainstorm" + reset + gray + " │  ← " + reset + dim + "research, framing, options, affected artifacts" + reset,
		gray + "  └─────┬──────┘" + reset,
		gray + "        │" + reset,
		gray + "        ▼" + reset,
		"",
		whiteBold + "🔁 Core Development Loop" + reset + dim + ":" + reset,
		gray + "┌───────────────┐    ┌──────┐    ┌───────┐    ┌────────────────┐    ┌────────────┐" + reset,
		gray + "│ " + spec + "Specification" + reset + gray + " │ ─▶ │ " + plan + "Plan" + reset + gray + " │ ─▶ │ " + tasks + "Tasks" + reset + gray + " │ ─▶ │ " + implement + "Implementation" + reset + gray + " │ ─▶ │ " + reflect + "Reflection" + reset + gray + " │ ─┐" + reset,
		gray + "└───────────────┘    └──────┘    └───────┘    └────────────────┘    └────────────┘  │" + reset,
		gray + "       ▲                                                                            │" + reset,
		gray + "       └────────────────────────────────────────────────────────────────────────────┘" + reset,
		"",
		whiteBold + "🗂️ Structured Engine: Artifact Pipeline" + reset,
		"  1. " + constitution + "Constitution" + reset + dim + "   — strategy, patterns, long-term vision (kept updated)" + reset,
		"  2. " + brainstorm + "Brainstorm" + reset + dim + "     — optional research and framing before the spec" + reset,
		"  3. " + spec + "Specification" + reset + dim + "  — what is being built and why" + reset,
		"  4. " + plan + "Plan" + reset + dim + "           — how it will be built" + reset,
		"  5. " + tasks + "Tasks" + reset + dim + "          — executable work units" + reset,
		"  6. " + implement + "Implementation" + reset + dim + " — execution begins after the readiness gate" + reset,
		"  7. " + reflect + "Reflection" + reset + dim + "     — verify correctness, refine understanding" + reset,
	}

	return strings.Join(lines, "\n")
}

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
	"pause":      15,
	"complete":   16,
	"status":     17,
	"remove":     18,
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
		cmd.SetHelpTemplate(helpTemplate(terminalWriterCheck(cmd.OutOrStdout())))
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
		cmd.SetUsageTemplate(usageTemplate(terminalWriterCheck(cmd.OutOrStdout())))
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

	fmt.Println(styleForStdout().clipboardAcknowledgement())
	return nil
}

func printWorkflowInstructions(currentStep string, nextSteps []string) {
	style := styleForStdout()

	fmt.Println(style.title("🧭", "Workflow"))
	if divider := style.sectionDivider(); divider != "" {
		fmt.Println(divider)
	}
	fmt.Println(style.muted("Pipeline: [optional brainstorm] -> spec -> plan -> tasks -> implement -> reflect"))
	fmt.Println()
	fmt.Println(style.currentStepLine(currentStep))
	if len(nextSteps) > 0 {
		fmt.Println()
		fmt.Println(style.nextStepsTitle())
		for _, step := range nextSteps {
			fmt.Printf("  %s\n", style.bullet(step))
		}
	}
	fmt.Println()
}

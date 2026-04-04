package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var terminalWriterCheck = isTerminalWriter

type humanOutputStyle struct {
	enabled bool
}

func styleForWriter(w io.Writer) humanOutputStyle {
	return humanOutputStyle{enabled: terminalWriterCheck(w)}
}

func styleForStdout() humanOutputStyle {
	return styleForWriter(os.Stdout)
}

func isTerminalWriter(w io.Writer) bool {
	fileLike, ok := w.(interface{ Fd() uintptr })
	if !ok {
		return false
	}

	return term.IsTerminal(int(fileLike.Fd()))
}

func (s humanOutputStyle) title(emoji, text string) string {
	if !s.enabled {
		return text
	}

	return whiteBold + emoji + " " + text + reset
}

func (s humanOutputStyle) label(text string) string {
	if !s.enabled {
		return text
	}

	return whiteBold + text + reset
}

func (s humanOutputStyle) muted(text string) string {
	if !s.enabled {
		return text
	}

	return dim + text + reset
}

func (s humanOutputStyle) bullet(text string) string {
	prefix := "- "
	if s.enabled {
		prefix = "• "
	}

	return prefix + text
}

func (s humanOutputStyle) clipboardAcknowledgement() string {
	if !s.enabled {
		return "Copied the prepared text to the clipboard."
	}

	return s.title("📋", "Copied the prepared text to the clipboard.")
}

func (s humanOutputStyle) selectionTitle(text string) string {
	return s.title("🧭", text)
}

func (s humanOutputStyle) selectionPrompt() string {
	if !s.enabled {
		return "Enter number: "
	}

	return whiteBold + "👉 Enter number: " + reset
}

func (s humanOutputStyle) nextStepsTitle() string {
	return s.title("🪜", "Next steps")
}

func (s humanOutputStyle) currentStepLine(step string) string {
	label := "Current step:"
	if s.enabled {
		label = "📍 Current step:"
	}

	return fmt.Sprintf("%s %s", s.label(label), step)
}

func (s humanOutputStyle) sectionDivider() string {
	if !s.enabled {
		return ""
	}

	return s.muted(strings.Repeat("─", 72))
}

func helpTemplate(enabled bool) string {
	usageHeader := "Usage:"
	aliasesHeader := "Aliases:"
	examplesHeader := "Examples:"
	commandsHeader := "Available Commands:"
	flagsHeader := "Flags:"
	globalFlagsHeader := "Global Flags:"
	additionalHelpHeader := "Additional Help Topics:"
	moreInfoLabel := "Use"

	if enabled {
		usageHeader = "🚀 Usage"
		aliasesHeader = "🏷️ Aliases"
		examplesHeader = "🧪 Examples"
		commandsHeader = "🧰 Available Commands"
		flagsHeader = "⚙️ Flags"
		globalFlagsHeader = "🌐 Global Flags"
		additionalHelpHeader = "📚 Additional Help Topics"
		moreInfoLabel = "🔎 Use"
	}

	return fmt.Sprintf(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}

%s
  {{if .Runnable}}{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

%s
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}  {{rpad .CommandPath .CommandPathPadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

%s "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}`,
		usageHeader,
		aliasesHeader,
		examplesHeader,
		commandsHeader,
		flagsHeader,
		globalFlagsHeader,
		additionalHelpHeader,
		moreInfoLabel,
	)
}

func usageTemplate(enabled bool) string {
	header := "Usage:"
	commandsHeader := "Available Commands:"
	flagsHeader := "Flags:"
	globalFlagsHeader := "Global Flags:"

	if enabled {
		header = "🚀 Usage"
		commandsHeader = "🧰 Available Commands"
		flagsHeader = "⚙️ Flags"
		globalFlagsHeader = "🌐 Global Flags"
	}

	return fmt.Sprintf(`%s
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

%s
{{range .Commands}}{{if .IsAvailableCommand}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`, header, commandsHeader, flagsHeader, globalFlagsHeader)
}

func printSelectionHeader(title string) {
	printSelectionHeaderTo(os.Stdout, title)
}

func printSelectionHeaderTo(w io.Writer, title string) {
	style := styleForWriter(w)
	fmt.Fprintln(w)
	fmt.Fprintln(w, style.selectionTitle(title))
	fmt.Fprintln(w)
}

func selectionPrompt(w io.Writer) string {
	return styleForWriter(w).selectionPrompt()
}

func printNumberedNextSteps(steps []string) {
	style := styleForStdout()

	fmt.Println()
	fmt.Println(style.nextStepsTitle())
	for i, step := range steps {
		fmt.Printf("  %d. %s\n", i+1, step)
	}
}

func printSectionBanner(emoji, title string) {
	style := styleForStdout()

	fmt.Println()
	if divider := style.sectionDivider(); divider != "" {
		fmt.Println(divider)
	}
	fmt.Println(style.title(emoji, title))
	if divider := style.sectionDivider(); divider != "" {
		fmt.Println(divider)
	}
	fmt.Println()
}

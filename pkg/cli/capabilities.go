package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

type capabilitiesOptions struct {
	json   bool
	full   bool
	search string
}

type capabilitiesIndexPayload struct {
	SchemaVersion int                       `json:"schema_version"`
	Kind          string                    `json:"kind"`
	GeneratedBy   string                    `json:"generated_by"`
	Commands      []capabilityCompactRecord `json:"commands"`
}

type capabilityDetailPayload struct {
	SchemaVersion int                    `json:"schema_version"`
	Kind          string                 `json:"kind"`
	GeneratedBy   string                 `json:"generated_by"`
	Command       capabilityDetailRecord `json:"command"`
}

type capabilitiesFullPayload struct {
	SchemaVersion int                      `json:"schema_version"`
	Kind          string                   `json:"kind"`
	GeneratedBy   string                   `json:"generated_by"`
	Commands      []capabilityDetailRecord `json:"commands"`
}

type capabilitiesSearchPayload struct {
	SchemaVersion int                       `json:"schema_version"`
	Kind          string                    `json:"kind"`
	GeneratedBy   string                    `json:"generated_by"`
	Query         string                    `json:"query"`
	Commands      []capabilityCompactRecord `json:"commands"`
}

var capabilitiesCmd = newCapabilitiesCommand()

func init() {
	rootCmd.AddCommand(capabilitiesCmd)
}

func newCapabilitiesCommand() *cobra.Command {
	options := &capabilitiesOptions{}
	cmd := &cobra.Command{
		Use:   "capabilities [command]",
		Short: "List Kit commands and their safety behavior",
		Long: strings.TrimSpace(`
List Kit command capabilities, mutation behavior, network use, and important
flags. The command is read-only and does not require a Kit project root.

Use a command path such as "verify" or "rules add" for a detailed record.
`),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCapabilities(cmd, args, *options)
		},
	}
	cmd.Flags().BoolVar(&options.json, "json", false, "emit machine-readable JSON")
	cmd.Flags().BoolVar(&options.full, "full", false, "include detailed hidden and deprecated command records")
	cmd.Flags().StringVar(&options.search, "search", "", "search visible command capability metadata")
	return cmd
}

func runCapabilities(cmd *cobra.Command, args []string, options capabilitiesOptions) error {
	commandPath := normalizeCapabilityQuery(strings.Join(args, " "))
	search := strings.TrimSpace(options.search)
	if search != "" && commandPath != "" {
		return fmt.Errorf("--search cannot be combined with a command path")
	}
	if options.full && commandPath != "" {
		return fmt.Errorf("--full cannot be combined with a command path")
	}

	if commandPath != "" {
		record, ok := capabilityByCommandPath(commandPath)
		if !ok {
			return unknownCapabilityCommandError(commandPath)
		}
		payload := capabilityDetailPayload{
			SchemaVersion: capabilitiesSchemaVersion,
			Kind:          "capability_detail",
			GeneratedBy:   "kit capabilities",
			Command:       record.detail(),
		}
		if options.json {
			return outputJSON(cmd.OutOrStdout(), payload)
		}
		return renderCapabilityDetail(cmd.OutOrStdout(), payload.Command)
	}

	if search != "" {
		records := searchCapabilityRecords(search)
		payload := capabilitiesSearchPayload{
			SchemaVersion: capabilitiesSchemaVersion,
			Kind:          "capabilities_search",
			GeneratedBy:   "kit capabilities",
			Query:         search,
			Commands:      compactCapabilityRecords(records),
		}
		if options.json {
			return outputJSON(cmd.OutOrStdout(), payload)
		}
		return renderCapabilitiesIndex(cmd.OutOrStdout(), payload.Commands, fmt.Sprintf("Search results for %q", search))
	}

	if options.full {
		payload := capabilitiesFullPayload{
			SchemaVersion: capabilitiesSchemaVersion,
			Kind:          "capabilities_full",
			GeneratedBy:   "kit capabilities",
			Commands:      detailCapabilityRecords(capabilityCatalog()),
		}
		if options.json {
			return outputJSON(cmd.OutOrStdout(), payload)
		}
		return renderCapabilitiesFull(cmd.OutOrStdout(), payload.Commands)
	}

	payload := capabilitiesIndexPayload{
		SchemaVersion: capabilitiesSchemaVersion,
		Kind:          "capabilities_index",
		GeneratedBy:   "kit capabilities",
		Commands:      compactCapabilityRecords(capabilityCatalog()),
	}
	if options.json {
		return outputJSON(cmd.OutOrStdout(), payload)
	}
	return renderCapabilitiesIndex(cmd.OutOrStdout(), payload.Commands, "Kit capabilities")
}

func unknownCapabilityCommandError(commandPath string) error {
	suggestions := suggestCapabilityCommands(commandPath)
	if len(suggestions) > 0 {
		return fmt.Errorf("unknown Kit command path %q. Did you mean %s? Run `kit capabilities --json` to list commands", commandPath, strings.Join(suggestions, ", "))
	}
	return fmt.Errorf("unknown Kit command path %q. Run `kit capabilities --json` to list commands", commandPath)
}

func renderCapabilitiesIndex(out io.Writer, records []capabilityCompactRecord, title string) error {
	if _, err := fmt.Fprintf(out, "%s\n", title); err != nil {
		return err
	}
	if len(records) == 0 {
		_, err := fmt.Fprintln(out, "No matching commands.")
		return err
	}

	currentCategory := ""
	for _, record := range records {
		if record.Category != currentCategory {
			currentCategory = record.Category
			if _, err := fmt.Fprintf(out, "\n%s\n", currentCategory); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(out, "  %-18s %s [%s]\n", record.Command, record.Summary, record.MutationLevel); err != nil {
			return err
		}
	}
	return nil
}

func renderCapabilitiesFull(out io.Writer, records []capabilityDetailRecord) error {
	if _, err := fmt.Fprintln(out, "Kit capabilities"); err != nil {
		return err
	}
	for _, record := range records {
		if err := renderCapabilityDetail(out, record); err != nil {
			return err
		}
	}
	return nil
}

func renderCapabilityDetail(out io.Writer, record capabilityDetailRecord) error {
	labels := make([]string, 0, 2)
	if record.Hidden {
		labels = append(labels, "hidden")
	}
	if record.Deprecated {
		labels = append(labels, "deprecated")
	}

	labelText := ""
	if len(labels) > 0 {
		labelText = " (" + strings.Join(labels, ", ") + ")"
	}
	if _, err := fmt.Fprintf(out, "\n%s%s\n", record.Command, labelText); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  Category: %s\n", record.Category); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  Summary: %s\n", record.Summary); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  Mutation: %s\n", record.MutationLevel); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  Network: %s\n", behaviorText(record.NetworkUse)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  File writes: %s\n", behaviorText(record.FileWrites)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  Git mutation: %s\n", behaviorText(record.GitMutation)); err != nil {
		return err
	}
	if record.DeprecationNote != "" {
		if _, err := fmt.Fprintf(out, "  Deprecation: %s\n", record.DeprecationNote); err != nil {
			return err
		}
	}
	if len(record.ImportantFlags) > 0 {
		if _, err := fmt.Fprintln(out, "  Important flags:"); err != nil {
			return err
		}
		for _, flag := range record.ImportantFlags {
			if _, err := fmt.Fprintf(out, "    %s: %s\n", flag.Name, flag.Summary); err != nil {
				return err
			}
		}
	}
	if len(record.RelatedCommands) > 0 {
		if _, err := fmt.Fprintln(out, "  Related commands:"); err != nil {
			return err
		}
		for _, related := range record.RelatedCommands {
			if _, err := fmt.Fprintf(out, "    %s", related.Command); err != nil {
				return err
			}
			if related.Note != "" {
				if _, err := fmt.Fprintf(out, ": %s", related.Note); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(out); err != nil {
				return err
			}
		}
	}
	return nil
}

func behaviorText(behavior capabilityBehavior) string {
	if behavior.FlagDependent == "" {
		return behavior.Summary
	}
	return behavior.Summary + " (" + behavior.FlagDependent + ")"
}

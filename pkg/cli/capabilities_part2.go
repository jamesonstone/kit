package cli

import (
	"fmt"
	"io"
	"strings"
)

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
	if len(record.Aliases) > 0 {
		if _, err := fmt.Fprintf(out, "  Aliases: %s\n", strings.Join(record.Aliases, ", ")); err != nil {
			return err
		}
	}
	if err := renderCapabilityStringList(out, "When to use", record.WhenToUse); err != nil {
		return err
	}
	if err := renderCapabilityStringList(out, "When not to use", record.WhenNotToUse); err != nil {
		return err
	}
	if err := renderCapabilityStringList(out, "Examples", record.Examples); err != nil {
		return err
	}
	if err := renderCapabilityStringList(out, "Caveats", record.Caveats); err != nil {
		return err
	}
	if len(record.ImportantFlags) > 0 {
		if _, err := fmt.Fprintln(out, "  Important flags:"); err != nil {
			return err
		}
		for _, flag := range record.ImportantFlags {
			if _, err := fmt.Fprintf(out, "    %s: %s", flag.Name, flag.Summary); err != nil {
				return err
			}
			if flag.Safety != "" {
				if _, err := fmt.Fprintf(out, " [%s]", flag.Safety); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(out); err != nil {
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

func renderCapabilityStringList(out io.Writer, title string, values []string) error {
	if len(values) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(out, "  %s:\n", title); err != nil {
		return err
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, err := fmt.Fprintf(out, "    - %s\n", value); err != nil {
			return err
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

package cli

import (
	"fmt"
	"io"
	"strings"
)

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

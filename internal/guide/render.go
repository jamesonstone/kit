package guide

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func WriteJSON(out io.Writer, d Document) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(d)
}

func WriteText(out io.Writer, d Document) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Kit functionality guide\n\n%s\n\nArtifact types\n", d.Instructions)
	for _, a := range d.Artifacts {
		fmt.Fprintf(&b, "\n%s (%s)\n%s\n%s\n", a.Name, a.Path, a.Purpose, a.Instructions)
	}
	b.WriteString("\nInstalled commands\n")
	for _, c := range d.Commands {
		fmt.Fprintf(&b, "\n%s\n  Purpose: %s\n  Usage: %s\n", c.Path, c.Purpose, c.Usage)
		if c.Instructions != "" {
			fmt.Fprintf(&b, "%s\n", c.Instructions)
		}
		if len(c.Aliases) > 0 {
			fmt.Fprintf(&b, "  Aliases: %s\n", strings.Join(c.Aliases, ", "))
		}
		if c.Example != "" {
			fmt.Fprintf(&b, "Examples:\n%s\n", c.Example)
		}
		for _, f := range c.Flags {
			fmt.Fprintf(&b, "  --%s", f.Name)
			if f.Shorthand != "" {
				fmt.Fprintf(&b, ", -%s", f.Shorthand)
			}
			fmt.Fprintf(&b, " (%s, default %q", f.Type, f.Default)
			if f.Inherited {
				b.WriteString(", inherited")
			}
			if f.Required {
				b.WriteString(", required")
			}
			fmt.Fprintf(&b, "): %s\n", f.Usage)
		}
	}
	_, err := io.WriteString(out, b.String())
	return err
}

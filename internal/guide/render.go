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

// WriteIndex lists the available commands and artifact locations for quick retrieval.
func WriteIndex(out io.Writer, d Document) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Kit guide for coding agents\n\n%s\n\n## Commands\n", d.Instructions)
	for _, c := range d.Commands {
		fmt.Fprintf(&b, "- `%s` — %s\n", c.Path, c.Purpose)
	}
	b.WriteString("\n## Repository artifacts\n")
	for _, a := range d.Artifacts {
		fmt.Fprintf(&b, "- %s (%s) — %s\n", a.Name, a.Path, a.Purpose)
	}
	_, err := io.WriteString(out, b.String())
	return err
}

// WriteCommand renders the detailed help for one command.
func WriteCommand(out io.Writer, c Command) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n%s\n\nUsage: %s\n", c.Path, c.Purpose, c.Usage)
	if c.Instructions != "" {
		fmt.Fprintf(&b, "\n%s\n", c.Instructions)
	}
	if len(c.Aliases) > 0 {
		fmt.Fprintf(&b, "\nAliases: %s\n", strings.Join(c.Aliases, ", "))
	}
	if c.Example != "" {
		fmt.Fprintf(&b, "\n## Examples\n%s\n", c.Example)
	}
	if len(c.Flags) > 0 {
		b.WriteString("\n## Flags\n")
	}
	for _, f := range c.Flags {
		fmt.Fprintf(&b, "- --%s", f.Name)
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
	_, err := io.WriteString(out, b.String())
	return err
}

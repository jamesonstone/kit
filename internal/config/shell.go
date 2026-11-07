package config

import (
	"fmt"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// ShellConfig describes an agent's shell invocation, not Kit's own process.
// Empty Command uses the current user's initialized shell on the execution host.
// A nonempty Command is an executable and argument prefix; the host appends the
// task command as one argument. Kit never evaluates or runs these strings.
type ShellConfig struct {
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
}

// UnmarshalYAML rejects misspelled fields instead of silently using the default.
func (c *ShellConfig) UnmarshalYAML(node *yaml.Node) error {
	type plain ShellConfig
	var value plain
	if err := node.Decode(&value); err != nil {
		return fmt.Errorf("shell must be a mapping with a command argument list")
	}
	node = shellValueNode(node)
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value != "command" {
			return fmt.Errorf("shell supports only the command field")
		}
		command := shellValueNode(node.Content[i+1])
		for _, argument := range command.Content {
			argument = shellValueNode(argument)
			if argument.Kind != yaml.ScalarNode || argument.Tag != "!!str" {
				return fmt.Errorf("shell.command arguments must be strings")
			}
		}
	}
	*c = ShellConfig(value)
	return nil
}

// Source reports whether the executor resolves the user's shell or uses the
// configured argument prefix. It does not inspect or capture the host's state.
func (c ShellConfig) Source() string {
	if len(c.Command) == 0 {
		return "user"
	}
	return "command"
}

// Validate checks local structure only, without resolving executables, expanding
// variables, reading profiles, or checking credentials and network access.
func (c ShellConfig) Validate() error {
	for _, arg := range c.Command {
		if strings.TrimSpace(arg) == "" || strings.IndexFunc(arg, unicode.IsControl) >= 0 {
			return fmt.Errorf("shell.command arguments must be nonempty strings without control characters")
		}
	}
	return nil
}

// Decode has already rejected cyclic aliases before this helper is used.
func shellValueNode(node *yaml.Node) *yaml.Node {
	for node.Kind == yaml.AliasNode && node.Alias != nil {
		node = node.Alias
	}
	return node
}

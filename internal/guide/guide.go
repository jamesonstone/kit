// Package guide derives an agent-readable view from authoritative command help.
package guide

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Document struct {
	SchemaVersion int        `json:"schema_version"`
	Instructions  string     `json:"instructions"`
	Artifacts     []Artifact `json:"artifacts"`
	Commands      []Command  `json:"commands"`
}

type Command struct {
	Path         string   `json:"path"`
	Usage        string   `json:"usage"`
	Purpose      string   `json:"purpose"`
	Instructions string   `json:"instructions,omitempty"`
	Example      string   `json:"example,omitempty"`
	Aliases      []string `json:"aliases,omitempty"`
	Flags        []Flag   `json:"flags"`
}

type Flag struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default"`
	Usage     string `json:"usage"`
	Inherited bool   `json:"inherited"`
	Required  bool   `json:"required"`
}

// Build snapshots the current tree, including commands registered after startup.
// It does not execute commands, inspect projects, or load configuration.
func Build(root *cobra.Command) Document {
	d := Document{SchemaVersion: 1, Instructions: "Read the repository's agent contract first. Use this guide to select installed commands and understand artifact types. Run kit <command> --help for focused help. Command descriptions come from this binary's live command tree; examples are instructions, not executed actions. Project authorization and contextual rules still apply. Paths below are defaults; project configuration can override supported paths. kit guide --json returns the same information as structured data.", Artifacts: artifacts(), Commands: []Command{}}
	var visit func(*cobra.Command)
	visit = func(cmd *cobra.Command) {
		if cmd != root && (!cmd.IsAvailableCommand() && cmd.Name() != "help" || cmd.Hidden) {
			return
		}
		cmd.InitDefaultHelpFlag()
		cmd.InitDefaultVersionFlag()
		aliases := append([]string(nil), cmd.Aliases...)
		sort.Strings(aliases)
		c := Command{Path: cmd.CommandPath(), Usage: cmd.UseLine(), Purpose: cmd.Short, Instructions: strings.TrimSpace(cmd.Long), Example: strings.TrimSpace(cmd.Example), Aliases: aliases, Flags: []Flag{}}
		add := func(fs *pflag.FlagSet, inherited bool) {
			fs.VisitAll(func(f *pflag.Flag) {
				if f.Hidden || f.Deprecated != "" {
					return
				}
				c.Flags = append(c.Flags, Flag{Name: f.Name, Shorthand: f.Shorthand, Type: f.Value.Type(), Default: f.DefValue, Usage: f.Usage, Inherited: inherited, Required: len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0})
			})
		}
		add(cmd.LocalFlags(), false)
		add(cmd.InheritedFlags(), true)
		sort.Slice(c.Flags, func(i, j int) bool { return c.Flags[i].Name < c.Flags[j].Name })
		d.Commands = append(d.Commands, c)
		for _, child := range cmd.Commands() {
			visit(child)
		}
	}
	visit(root)
	sort.Slice(d.Commands, func(i, j int) bool { return d.Commands[i].Path < d.Commands[j].Path })
	return d
}

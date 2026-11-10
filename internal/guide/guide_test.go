package guide

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestBuildTracksLiveCommandsAndFlags(t *testing.T) {
	root := &cobra.Command{Use: "kit", Short: "Root", Version: "1.2.3"}
	root.PersistentFlags().String("profile", "local", "Select profile")
	group := &cobra.Command{Use: "future", Short: "New group"}
	leaf := &cobra.Command{Use: "run <target>", Short: "Run purpose", Long: "Detailed instructions", Example: "kit future run sample", Aliases: []string{"z", "a"}, Run: func(*cobra.Command, []string) {}}
	leaf.Flags().StringP("value", "v", "safe", "Choose value")
	if err := leaf.MarkFlagRequired("value"); err != nil {
		t.Fatal(err)
	}
	leaf.Flags().Bool("secret", false, "Hidden")
	if err := leaf.Flags().MarkHidden("secret"); err != nil {
		t.Fatal(err)
	}
	root.AddCommand(group)
	group.AddCommand(leaf)
	root.AddCommand(&cobra.Command{Use: "hidden", Hidden: true, Run: func(*cobra.Command, []string) {}})
	d := Build(root)
	if len(d.Commands) != 3 {
		t.Fatalf("unexpected commands: %#v", d.Commands)
	}
	versionFound := false
	for _, f := range d.Commands[0].Flags {
		if f.Name == "version" {
			versionFound = true
		}
	}
	if !versionFound {
		t.Fatal("root version flag missing")
	}
	got := d.Commands[2]
	if got.Path != "kit future run" || got.Purpose != leaf.Short || got.Instructions != leaf.Long || got.Example != leaf.Example || !strings.Contains(got.Usage, "<target>") {
		t.Fatalf("help mismatch: %#v", got)
	}
	if !reflect.DeepEqual(got.Aliases, []string{"a", "z"}) {
		t.Fatal(got.Aliases)
	}
	flags := map[string]Flag{}
	for _, f := range got.Flags {
		flags[f.Name] = f
	}
	if !flags["profile"].Inherited || flags["profile"].Default != "local" || !flags["value"].Required || flags["value"].Shorthand != "v" || flags["value"].Type != "string" {
		t.Fatalf("flags: %#v", flags)
	}
	if _, ok := flags["secret"]; ok {
		t.Fatal("hidden flag included")
	}
	var first, second bytes.Buffer
	if err := WriteJSON(&first, d); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(&second, Build(root)); err != nil {
		t.Fatal(err)
	}
	if first.String() != second.String() {
		t.Fatal("JSON changed between builds")
	}
	leaf.Short = "Updated purpose"
	if Build(root).Commands[2].Purpose != "Updated purpose" {
		t.Fatal("help was cached")
	}
}

func TestArtifactsAndText(t *testing.T) {
	d := Build(&cobra.Command{Use: "kit"})
	var b bytes.Buffer
	if err := WriteText(&b, d); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"docs/CONSTITUTION.md", "docs/specs/<feature>/SPEC.md", "docs/references/rules/<slug>.md", "docs/references/testing.md", "docs/sops/<slug>/SOP.md", "When present", "docs/sops/README.md", "docs/references/rules/standard-operating-procedures.md", "draft, active, and retired", "not a Kit creation command or execution engine", "existing authorization", "kit spec <feature>", "kit rules view <slug>"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("missing %q", want)
		}
	}
}

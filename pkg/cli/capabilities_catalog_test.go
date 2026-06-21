package cli

import (
	"strings"
	"testing"
)

func TestCapabilityCatalogCoversVisibleRootCommands(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Hidden || cmd.Deprecated != "" {
			continue
		}
		if cmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		if _, ok := capabilityByCommandPath(cmd.Name()); !ok {
			t.Fatalf("visible root command %q is missing from capability catalog", cmd.Name())
		}
	}
}

func TestCapabilityCatalogRecordsIncludeAgentGuidance(t *testing.T) {
	for _, record := range capabilityCatalog() {
		if record.Hidden || record.Deprecated {
			continue
		}
		detail := record.detail()
		if strings.TrimSpace(detail.Summary) == "" {
			t.Fatalf("%s missing summary", detail.Command)
		}
		if strings.TrimSpace(detail.MutationLevel) == "" {
			t.Fatalf("%s missing mutation level", detail.Command)
		}
		if strings.TrimSpace(detail.NetworkUse.Summary) == "" {
			t.Fatalf("%s missing network behavior", detail.Command)
		}
		if strings.TrimSpace(detail.FileWrites.Summary) == "" {
			t.Fatalf("%s missing file-write behavior", detail.Command)
		}
		if strings.TrimSpace(detail.GitMutation.Summary) == "" {
			t.Fatalf("%s missing git mutation behavior", detail.Command)
		}
		if len(detail.WhenToUse) == 0 {
			t.Fatalf("%s missing when_to_use guidance", detail.Command)
		}
		if len(detail.WhenNotToUse) == 0 {
			t.Fatalf("%s missing when_not_to_use guidance", detail.Command)
		}
		if len(detail.Examples) == 0 {
			t.Fatalf("%s missing examples", detail.Command)
		}
	}
}

func TestCapabilityCatalogNestedCommandsAreRegistered(t *testing.T) {
	for _, commandPath := range [][]string{
		{"scaffold", "agents"},
		{"prompt", "list"},
		{"set", "prompt"},
		{"skill", "mine"},
		{"loop", "workflow"},
		{"loop", "review"},
		{"rules", "add"},
		{"rules", "list"},
		{"rules", "view"},
		{"rules", "link"},
	} {
		commandName := strings.Join(commandPath, " ")
		cmd, _, err := rootCmd.Find(commandPath)
		if err != nil {
			t.Fatalf("rootCmd.Find(%s) error = %v", commandName, err)
		}
		if cmd == nil {
			t.Fatalf("expected %s to be registered", commandName)
		}
		if _, ok := capabilityByCommandPath(commandName); !ok {
			t.Fatalf("registered nested command %q is missing from capability catalog", commandName)
		}
	}
}

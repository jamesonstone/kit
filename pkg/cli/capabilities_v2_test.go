package cli

import (
	"bytes"
	"encoding/json"
	stdreflect "reflect"
	"sort"
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/commandset"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestCapabilityCatalogMatchesSupportedKitCommandPaths(t *testing.T) {
	seen := map[string]bool{}
	for _, path := range commandset.ProtectedPaths() {
		fields := strings.Fields(path)
		for index := range fields {
			seen[strings.Join(fields[:index+1], " ")] = true
		}
	}
	want := make([]string, 0, len(seen))
	for path := range seen {
		want = append(want, path)
	}
	sort.Strings(want)
	var got []string
	for _, record := range capabilityCatalog() {
		if strings.HasPrefix(record.Command, "git wt") || record.Command == "help" {
			continue
		}
		got = append(got, record.Command)
	}
	sort.Strings(got)
	if !stdreflect.DeepEqual(got, want) {
		t.Fatalf("capability paths mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCapabilitiesIndexContainsNoRemovedCommands(t *testing.T) {
	cmd := newCapabilitiesCommand()
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{"--full", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var payload capabilitiesFullPayload
	if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	for _, removed := range []string{"notes", "project refresh", "loop", "ci", "scaffold", "prompt", "legacy", "skill"} {
		for _, record := range payload.Commands {
			if record.Command == removed || strings.HasPrefix(record.Command, removed+" ") {
				t.Fatalf("removed capability %q remains", record.Command)
			}
		}
	}
}

func TestContextAndUsageCapabilitiesExposeSafetyBoundaries(t *testing.T) {
	contextRecord, ok := capabilityByCommandPath("context resolve")
	if !ok || contextRecord.MutationLevel != mutationNone || contextRecord.NetworkUse.Summary != "none" {
		t.Fatalf("context capability = %#v", contextRecord)
	}
	usageRecord, ok := capabilityByCommandPath("usage")
	if !ok || usageRecord.NetworkUse.Summary != "none" || !strings.Contains(strings.Join(usageRecord.Caveats, " "), "secrets") {
		t.Fatalf("usage capability = %#v", usageRecord)
	}
}

func TestCapabilitiesExposeRootPersistentFlagsOnce(t *testing.T) {
	want := flagSetNames(rootCmd.PersistentFlags())
	for name, args := range map[string][]string{
		"index":  {"--json"},
		"detail": {"reconcile", "--json"},
		"full":   {"--full", "--json"},
		"search": {"--search", "reconcile", "--json"},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := newCapabilitiesCommand()
			output := &bytes.Buffer{}
			cmd.SetOut(output)
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			var payload struct {
				GlobalFlags []capabilityFlag `json:"global_flags"`
			}
			if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if got := capabilityFlagNames(payload.GlobalFlags); !stdreflect.DeepEqual(got, want) {
				t.Fatalf("global flags = %#v, want %#v", got, want)
			}
		})
	}

	global := map[string]bool{}
	for _, name := range want {
		global[name] = true
	}
	for _, record := range capabilityCatalog() {
		for _, item := range record.ImportantFlags {
			if global[item.Name] {
				t.Fatalf("%s command record duplicates global flag %s", record.Command, item.Name)
			}
		}
	}
}

func TestReconcileCapabilityExposesMaintenanceContracts(t *testing.T) {
	cmd := newCapabilitiesCommand()
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{"reconcile", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var payload capabilityDetailPayload
	if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	whenToUse := strings.Join(payload.Command.WhenToUse, " ")
	if !strings.Contains(whenToUse, "exact 300-physical-line limit") {
		t.Fatalf("reconcile when_to_use omits source-size limit: %#v", payload.Command.WhenToUse)
	}
	caveats := strings.Join(payload.Command.Caveats, " ")
	for _, contract := range []string{
		"ordered Codex pre-response thread-title and thread-pin gate",
		"fail-visible first-commentary semantics",
		"source-file-size audit: complete",
		"candidate, eligible-file, and violation counts",
	} {
		if !strings.Contains(caveats, contract) {
			t.Errorf("reconcile caveats omit %q: %#v", contract, payload.Command.Caveats)
		}
	}
}

func TestCapabilityCatalogFlagsMatchCommandTree(t *testing.T) {
	rootCmd.InitDefaultCompletionCmd()
	globalFlags := stringSetFromFlagSet(rootCmd.PersistentFlags())
	globalFlags["help"] = true

	for _, record := range capabilityCatalog() {
		if record.Command == "help" {
			continue
		}
		command, _, err := rootCmd.Find(strings.Fields(record.Command))
		if err != nil || command == nil || command.CommandPath() != "kit "+record.Command {
			t.Fatalf("resolve command %q: command=%#v err=%v", record.Command, command, err)
		}

		want := commandSpecificFlagNames(command, globalFlags)
		got := capabilityFlagNames(record.ImportantFlags)
		if !stdreflect.DeepEqual(got, want) {
			t.Errorf("%s capability flags mismatch\ngot:  %#v\nwant: %#v", record.Command, got, want)
		}
	}
}

func TestMutatingLeafCapabilitiesDescribeSideEffects(t *testing.T) {
	records := capabilityCatalog()
	for _, record := range records {
		if !isLeafCapability(record.Command, records) {
			continue
		}
		switch record.MutationLevel {
		case mutationWritesFiles, mutationDestructive:
			if record.FileWrites.Summary == "none" {
				t.Errorf("%s mutation %s omits file-write behavior", record.Command, record.MutationLevel)
			}
		case mutationNetwork:
			if record.NetworkUse.Summary == "none" {
				t.Errorf("%s network mutation omits network behavior", record.Command)
			}
		case mutationGit:
			if record.GitMutation.Summary == "none" {
				t.Errorf("%s Git mutation omits Git behavior", record.Command)
			}
		}
	}
}

func TestCapabilityCommandGroupsDescribeExactReadOnlyInvocation(t *testing.T) {
	for _, command := range []string{"context", "registry", "usage", "config", "aws", "pr", "improve", "rules"} {
		record, ok := capabilityByCommandPath(command)
		if !ok {
			t.Errorf("capability group %s is missing", command)
			continue
		}
		if record.MutationLevel != mutationNone || record.NetworkUse.Summary != "none" || record.FileWrites.Summary != "none" || record.GitMutation.Summary != "none" {
			t.Errorf("capability group %s describes child side effects instead of its exact invocation: %#v", command, record)
		}
	}
}

func commandSpecificFlagNames(command *cobra.Command, excluded map[string]bool) []string {
	names := []string{}
	seen := map[string]bool{}
	appendFlag := func(item *pflag.Flag) {
		if !excluded[item.Name] && !seen[item.Name] {
			names = append(names, "--"+item.Name)
			seen[item.Name] = true
		}
	}
	command.Flags().VisitAll(appendFlag)
	command.PersistentFlags().VisitAll(appendFlag)
	for parent := command.Parent(); parent != nil && parent != rootCmd; parent = parent.Parent() {
		parent.PersistentFlags().VisitAll(appendFlag)
	}
	sort.Strings(names)
	return names
}

func capabilityFlagNames(flags []capabilityFlag) []string {
	names := make([]string, 0, len(flags))
	for _, item := range flags {
		names = append(names, item.Name)
	}
	sort.Strings(names)
	return names
}

func stringSetFromFlagSet(flags *pflag.FlagSet) map[string]bool {
	result := map[string]bool{}
	flags.VisitAll(func(item *pflag.Flag) {
		result[item.Name] = true
	})
	return result
}

func flagSetNames(flags *pflag.FlagSet) []string {
	names := []string{}
	flags.VisitAll(func(item *pflag.Flag) {
		names = append(names, "--"+item.Name)
	})
	sort.Strings(names)
	return names
}

func isLeafCapability(command string, records []capabilityRecord) bool {
	prefix := command + " "
	for _, candidate := range records {
		if strings.HasPrefix(candidate.Command, prefix) {
			return false
		}
	}
	return true
}

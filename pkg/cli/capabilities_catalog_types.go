package cli

const (
	mutationNone             = "none"
	mutationWritesFiles      = "writes_files"
	mutationExecutesCommands = "executes_commands"
	mutationNetwork          = "network"
	mutationGit              = "git"
	mutationDestructive      = "destructive"
)

var capabilityCategoryOrder = map[string]int{
	"Setup":            1,
	"Workflow":         2,
	"Legacy":           3,
	"Inspect & Repair": 4,
	"Prompt Utilities": 5,
	"Utilities":        6,
}

type capabilityBehavior struct {
	Summary       string `json:"summary"`
	FlagDependent string `json:"flag_dependent,omitempty"`
}

type capabilityFlag struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Safety  string `json:"safety,omitempty"`
}

type capabilityRelatedCommand struct {
	Command string `json:"command"`
	Note    string `json:"note,omitempty"`
}

type capabilityRecord struct {
	Command              string
	Category             string
	Summary              string
	MutationLevel        string
	NetworkUse           capabilityBehavior
	FileWrites           capabilityBehavior
	GitMutation          capabilityBehavior
	Hidden               bool
	Deprecated           bool
	DeprecationNote      string
	Aliases              []string
	ImportantFlags       []capabilityFlag
	RelatedCommands      []capabilityRelatedCommand
	WhenToUse            []string
	WhenNotToUse         []string
	Examples             []string
	Caveats              []string
	DetailedFlagBehavior []capabilityFlag
	IncludeInCompact     bool
}

type capabilityCompactRecord struct {
	Command         string                     `json:"command"`
	Category        string                     `json:"category"`
	Summary         string                     `json:"summary"`
	MutationLevel   string                     `json:"mutation_level"`
	NetworkUse      capabilityBehavior         `json:"network_use"`
	FileWrites      capabilityBehavior         `json:"file_writes"`
	GitMutation     capabilityBehavior         `json:"git_mutation"`
	Hidden          bool                       `json:"hidden"`
	Deprecated      bool                       `json:"deprecated"`
	ImportantFlags  []capabilityFlag           `json:"important_flags"`
	RelatedCommands []capabilityRelatedCommand `json:"related_commands"`
}

type capabilityDetailRecord struct {
	Command              string                     `json:"command"`
	Category             string                     `json:"category"`
	Summary              string                     `json:"summary"`
	MutationLevel        string                     `json:"mutation_level"`
	NetworkUse           capabilityBehavior         `json:"network_use"`
	FileWrites           capabilityBehavior         `json:"file_writes"`
	GitMutation          capabilityBehavior         `json:"git_mutation"`
	Hidden               bool                       `json:"hidden"`
	Deprecated           bool                       `json:"deprecated"`
	DeprecationNote      string                     `json:"deprecation_note,omitempty"`
	Aliases              []string                   `json:"aliases"`
	ImportantFlags       []capabilityFlag           `json:"important_flags"`
	RelatedCommands      []capabilityRelatedCommand `json:"related_commands"`
	WhenToUse            []string                   `json:"when_to_use"`
	WhenNotToUse         []string                   `json:"when_not_to_use"`
	Examples             []string                   `json:"examples"`
	Caveats              []string                   `json:"caveats"`
	DetailedFlagBehavior []capabilityFlag           `json:"detailed_flag_behavior"`
}

type capabilityOption func(*capabilityRecord)

func capability(command, category, summary, mutationLevel string, options ...capabilityOption) capabilityRecord {
	record := capabilityRecord{
		Command:          command,
		Category:         category,
		Summary:          summary,
		MutationLevel:    mutationLevel,
		NetworkUse:       capabilityBehavior{Summary: "none"},
		FileWrites:       capabilityBehavior{Summary: "none"},
		GitMutation:      capabilityBehavior{Summary: "none"},
		Aliases:          []string{},
		ImportantFlags:   []capabilityFlag{},
		RelatedCommands:  []capabilityRelatedCommand{},
		WhenToUse:        []string{summary},
		WhenNotToUse:     []string{"Use a narrower Kit command when one better matches the workflow step."},
		Examples:         []string{"kit " + command},
		Caveats:          []string{},
		IncludeInCompact: true,
	}
	for _, option := range options {
		option(&record)
	}
	if record.DetailedFlagBehavior == nil {
		record.DetailedFlagBehavior = append([]capabilityFlag(nil), record.ImportantFlags...)
	}
	return record
}

func withNetwork(summary string, flagDependent ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.NetworkUse = capabilityBehavior{Summary: summary}
		if len(flagDependent) > 0 {
			record.NetworkUse.FlagDependent = flagDependent[0]
		}
	}
}

func withFileWrites(summary string, flagDependent ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.FileWrites = capabilityBehavior{Summary: summary}
		if len(flagDependent) > 0 {
			record.FileWrites.FlagDependent = flagDependent[0]
		}
	}
}

func withGitMutation(summary string, flagDependent ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.GitMutation = capabilityBehavior{Summary: summary}
		if len(flagDependent) > 0 {
			record.GitMutation.FlagDependent = flagDependent[0]
		}
	}
}

func withFlags(flags ...capabilityFlag) capabilityOption {
	return func(record *capabilityRecord) {
		record.ImportantFlags = append(record.ImportantFlags, flags...)
		record.DetailedFlagBehavior = append(record.DetailedFlagBehavior, flags...)
	}
}

func withRelated(related ...capabilityRelatedCommand) capabilityOption {
	return func(record *capabilityRecord) {
		record.RelatedCommands = append(record.RelatedCommands, related...)
	}
}

func withAliases(aliases ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.Aliases = append(record.Aliases, aliases...)
	}
}

func withWhenToUse(values ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.WhenToUse = append([]string(nil), values...)
	}
}

func withWhenNotToUse(values ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.WhenNotToUse = append([]string(nil), values...)
	}
}

func withExamples(values ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.Examples = append([]string(nil), values...)
	}
}

func withCaveats(values ...string) capabilityOption {
	return func(record *capabilityRecord) {
		record.Caveats = append(record.Caveats, values...)
	}
}

func hiddenDeprecated(note string) capabilityOption {
	return func(record *capabilityRecord) {
		record.Hidden = true
		record.Deprecated = true
		record.DeprecationNote = note
		record.IncludeInCompact = false
	}
}

func deprecated(note string) capabilityOption {
	return func(record *capabilityRecord) {
		record.Deprecated = true
		record.DeprecationNote = note
		record.IncludeInCompact = false
	}
}

func flag(name, summary string, safety ...string) capabilityFlag {
	result := capabilityFlag{Name: name, Summary: summary}
	if len(safety) > 0 {
		result.Safety = safety[0]
	}
	return result
}

func related(command, note string) capabilityRelatedCommand {
	return capabilityRelatedCommand{Command: command, Note: note}
}

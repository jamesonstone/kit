package usage

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func BuildReport(filter Filter) (Report, error) {
	dir, err := Directory()
	if err != nil {
		return Report{}, err
	}
	report := Report{SchemaVersion: SchemaVersion, GeneratedAt: time.Now().UTC(), Since: filter.Since, Diagnostics: []Diagnostic{}}
	commandCounts := map[string]*CommandSummary{}
	projectCounts := map[string]int{}
	shards, err := listShards(dir)
	if err != nil {
		return report, err
	}
	for _, item := range shards {
		readErr := readEvents(item.path, func(event Event) error {
			if event.Timestamp.Before(filter.Since) || !matchesFilter(event, filter) {
				return nil
			}
			includeEvent(&report, commandCounts, projectCounts, event)
			return nil
		})
		if readErr != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: "error", Message: readErr.Error()})
		}
	}
	for _, summary := range commandCounts {
		report.Commands = append(report.Commands, *summary)
	}
	sort.Slice(report.Commands, func(i, j int) bool {
		if report.Commands[i].Calls != report.Commands[j].Calls {
			return report.Commands[i].Calls > report.Commands[j].Calls
		}
		return report.Commands[i].Command < report.Commands[j].Command
	})
	for projectID, calls := range projectCounts {
		report.Projects = append(report.Projects, ProjectSummary{ProjectID: projectID, Calls: calls})
	}
	sort.Slice(report.Projects, func(i, j int) bool {
		if report.Projects[i].Calls != report.Projects[j].Calls {
			return report.Projects[i].Calls > report.Projects[j].Calls
		}
		return report.Projects[i].ProjectID < report.Projects[j].ProjectID
	})
	return report, nil
}

func matchesFilter(event Event, filter Filter) bool {
	if filter.Command != "" && event.Command != filter.Command {
		return false
	}
	return filter.ProjectID == "" || event.ProjectID == filter.ProjectID
}

func includeEvent(report *Report, commands map[string]*CommandSummary, projects map[string]int, event Event) {
	report.TotalCalls++
	if event.Success {
		report.Successes++
	} else {
		report.Failures++
	}
	if event.Interactive {
		report.Interactive++
	} else {
		report.NonInteractive++
	}
	if report.CoverageStart == nil || event.Timestamp.Before(*report.CoverageStart) {
		value := event.Timestamp
		report.CoverageStart = &value
	}
	if report.CoverageEnd == nil || event.Timestamp.After(*report.CoverageEnd) {
		value := event.Timestamp
		report.CoverageEnd = &value
	}
	summary := commands[event.Command]
	if summary == nil {
		summary = &CommandSummary{Command: event.Command}
		commands[event.Command] = summary
	}
	summary.Calls++
	if event.Success {
		summary.Successes++
	} else {
		summary.Failures++
	}
	if event.Interactive {
		summary.Interactive++
	} else {
		summary.NonInteractive++
	}
	if event.ProjectID != "" {
		projects[event.ProjectID]++
	}
}

func Status(projectRoot string) (StorageStatus, error) {
	settings, settingsErr := EffectiveSettings(projectRoot)
	dir, dirErr := Directory()
	if dirErr != nil {
		return StorageStatus{}, dirErr
	}
	status := StorageStatus{
		SchemaVersion: SchemaVersion,
		Enabled:       settings.Enabled, GlobalEnabled: settings.GlobalEnabled,
		ProjectState: settings.ProjectState, Directory: dir,
		RetentionDays: int(DefaultRetention.Hours() / 24),
		MaxTotalBytes: MaxTotalBytes, MaxShardBytes: MaxShardBytes,
		Diagnostics: []Diagnostic{},
	}
	if settingsErr != nil {
		status.Diagnostics = append(status.Diagnostics, Diagnostic{Level: "error", Message: settingsErr.Error()})
	}
	shards, err := listShards(dir)
	if err != nil {
		return status, err
	}
	status.ShardCount = len(shards)
	for _, item := range shards {
		status.TotalBytes += item.size
		if item.size > MaxShardBytes {
			status.Diagnostics = append(status.Diagnostics, Diagnostic{Level: "warning", Message: fmt.Sprintf("%s exceeds the shard bound", item.name)})
		}
		if err := readEvents(item.path, func(event Event) error {
			if status.CoverageStart == nil || event.Timestamp.Before(*status.CoverageStart) {
				value := event.Timestamp
				status.CoverageStart = &value
			}
			if status.CoverageEnd == nil || event.Timestamp.After(*status.CoverageEnd) {
				value := event.Timestamp
				status.CoverageEnd = &value
			}
			return nil
		}); err != nil {
			status.Diagnostics = append(status.Diagnostics, Diagnostic{Level: "error", Message: err.Error()})
		}
	}
	if status.TotalBytes > MaxTotalBytes {
		status.Diagnostics = append(status.Diagnostics, Diagnostic{Level: "warning", Message: "usage storage exceeds the total bound; run kit usage refresh"})
	}
	return status, nil
}

func CurrentProjectID(projectRoot string) (string, error) {
	dir, err := Directory()
	if err != nil {
		return "", err
	}
	return projectIdentifier(dir, strings.TrimSpace(projectRoot), false)
}

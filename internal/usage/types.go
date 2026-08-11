package usage

import "time"

const (
	SchemaVersion    = "kit.usage/v1"
	DefaultRetention = 365 * 24 * time.Hour
	MaxTotalBytes    = int64(16 * 1024 * 1024)
	MaxShardBytes    = int64(2 * 1024 * 1024)
)

type Event struct {
	SchemaVersion string    `json:"schema_version"`
	Timestamp     time.Time `json:"timestamp"`
	Command       string    `json:"command"`
	Version       string    `json:"version"`
	ExitCode      int       `json:"exit_code"`
	Success       bool      `json:"success"`
	ElapsedMS     int64     `json:"elapsed_ms"`
	ProjectID     string    `json:"project_id,omitempty"`
	Interactive   bool      `json:"interactive"`
}

type RecordInput struct {
	Command     string
	Version     string
	ExitCode    int
	Elapsed     time.Duration
	ProjectRoot string
	Interactive bool
}

type Settings struct {
	Enabled       bool   `json:"enabled"`
	GlobalEnabled bool   `json:"global_enabled"`
	ProjectState  string `json:"project_state"`
	ProjectRoot   string `json:"-"`
}

type Diagnostic struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type StorageStatus struct {
	SchemaVersion string       `json:"schema_version"`
	Enabled       bool         `json:"enabled"`
	GlobalEnabled bool         `json:"global_enabled"`
	ProjectState  string       `json:"project_state"`
	Directory     string       `json:"directory"`
	ShardCount    int          `json:"shard_count"`
	TotalBytes    int64        `json:"total_bytes"`
	RetentionDays int          `json:"retention_days"`
	MaxTotalBytes int64        `json:"max_total_bytes"`
	MaxShardBytes int64        `json:"max_shard_bytes"`
	CoverageStart *time.Time   `json:"coverage_start,omitempty"`
	CoverageEnd   *time.Time   `json:"coverage_end,omitempty"`
	PrunedShards  int          `json:"pruned_shards,omitempty"`
	PrunedEvents  int          `json:"pruned_events,omitempty"`
	Diagnostics   []Diagnostic `json:"diagnostics"`
}

type CommandSummary struct {
	Command        string `json:"command"`
	Calls          int    `json:"calls"`
	Successes      int    `json:"successes"`
	Failures       int    `json:"failures"`
	Interactive    int    `json:"interactive"`
	NonInteractive int    `json:"non_interactive"`
}

type ProjectSummary struct {
	ProjectID string `json:"project_id"`
	Calls     int    `json:"calls"`
}

type Report struct {
	SchemaVersion   string           `json:"schema_version"`
	GeneratedAt     time.Time        `json:"generated_at"`
	Since           time.Time        `json:"since"`
	CoverageStart   *time.Time       `json:"coverage_start,omitempty"`
	CoverageEnd     *time.Time       `json:"coverage_end,omitempty"`
	TotalCalls      int              `json:"total_calls"`
	Successes       int              `json:"successes"`
	Failures        int              `json:"failures"`
	Interactive     int              `json:"interactive"`
	NonInteractive  int              `json:"non_interactive"`
	Commands        []CommandSummary `json:"commands"`
	Projects        []ProjectSummary `json:"projects"`
	ZeroUseCommands []string         `json:"zero_use_commands"`
	Diagnostics     []Diagnostic     `json:"diagnostics"`
}

type Filter struct {
	Since     time.Time
	Command   string
	ProjectID string
}

type RefreshResult struct {
	Status  StorageStatus `json:"status"`
	DryRun  bool          `json:"dry_run"`
	Changed bool          `json:"changed"`
}

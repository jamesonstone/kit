// Package improve implements the deterministic core for kit improve.
package improve

import "time"

const SchemaVersion = 1

type Suite struct {
	SchemaVersion  int          `json:"schema_version" yaml:"schema_version"`
	ID             string       `json:"id" yaml:"id"`
	Title          string       `json:"title" yaml:"title"`
	HeldIn         TaskSelector `json:"held_in" yaml:"held_in"`
	HeldOut        TaskSelector `json:"held_out" yaml:"held_out"`
	Repeat         int          `json:"repeat" yaml:"repeat"`
	MinimumTasks   int          `json:"minimum_tasks" yaml:"minimum_tasks"`
	SelectionRules []string     `json:"selection_rules" yaml:"selection_rules"`
}

type TaskSelector struct {
	IncludeTags        []string `json:"include_tags" yaml:"include_tags"`
	HiddenFromProposer bool     `json:"hidden_from_proposer,omitempty" yaml:"hidden_from_proposer,omitempty"`
}

type Task struct {
	SchemaVersion     int         `json:"schema_version" yaml:"schema_version"`
	ID                string      `json:"id" yaml:"id"`
	Title             string      `json:"title" yaml:"title"`
	Category          string      `json:"category" yaml:"category"`
	Fixture           string      `json:"fixture" yaml:"fixture"`
	Persona           string      `json:"persona,omitempty" yaml:"persona,omitempty"`
	TimeoutSeconds    int         `json:"timeout_seconds,omitempty" yaml:"timeout_seconds,omitempty"`
	InputPrompt       string      `json:"input_prompt,omitempty" yaml:"input_prompt,omitempty"`
	ExpectedBehavior  string      `json:"expected_behavior" yaml:"expected_behavior"`
	Oracle            string      `json:"oracle" yaml:"oracle"`
	MutationPolicy    string      `json:"mutation_policy" yaml:"mutation_policy"`
	AllowedSurfaces   []string    `json:"allowed_surfaces" yaml:"allowed_surfaces"`
	Commands          []string    `json:"commands" yaml:"commands"`
	Assertions        []Assertion `json:"assertions" yaml:"assertions"`
	RegressionTags    []string    `json:"regression_tags" yaml:"regression_tags"`
	HeldOutEligible   bool        `json:"held_out_eligible" yaml:"held_out_eligible"`
	KnownFailureModes []string    `json:"known_failure_modes,omitempty" yaml:"known_failure_modes,omitempty"`
}

type Assertion struct {
	Type         string `json:"type" yaml:"type"`
	CommandIndex int    `json:"command_index,omitempty" yaml:"command_index,omitempty"`
	Value        string `json:"value,omitempty" yaml:"value,omitempty"`
	Max          int    `json:"max,omitempty" yaml:"max,omitempty"`
}

type RunManifest struct {
	SchemaVersion int                 `json:"schema_version"`
	Kind          string              `json:"kind"`
	RunID         string              `json:"run_id"`
	Suite         string              `json:"suite"`
	StartedAt     time.Time           `json:"started_at"`
	EndedAt       time.Time           `json:"ended_at"`
	Status        string              `json:"status"`
	RunDir        string              `json:"run_dir"`
	Provenance    BenchmarkProvenance `json:"provenance"`
	Metrics       RunMetrics          `json:"metrics"`
	Traces        []Trace             `json:"traces"`
}

type BenchmarkProvenance struct {
	SuiteDefinitionSHA256 string `json:"suite_definition_sha256"`
	RunnerBinaryPath      string `json:"runner_binary_path"`
	RunnerBinarySHA256    string `json:"runner_binary_sha256"`
	KitBinaryPath         string `json:"kit_binary_path"`
	KitBinarySHA256       string `json:"kit_binary_sha256"`
	KitVersion            string `json:"kit_version"`
	HarnessGitCommit      string `json:"harness_git_commit"`
}

type RunMetrics struct {
	TaskRuns            int         `json:"task_runs"`
	PassedTaskRuns      int         `json:"passed_task_runs"`
	FailedTaskRuns      int         `json:"failed_task_runs"`
	Assertions          int         `json:"assertions"`
	PassedAssertions    int         `json:"passed_assertions"`
	FailedAssertions    int         `json:"failed_assertions"`
	TaskSuccessRate     float64     `json:"task_success_rate"`
	OutputCompleteness  float64     `json:"output_completeness"`
	CommandDurationMS   int64       `json:"command_duration_ms"`
	RepeatedTasks       int         `json:"repeated_tasks"`
	StableRepeatedTasks int         `json:"stable_repeated_tasks"`
	DeterminismRate     float64     `json:"determinism_rate"`
	Stdout              TextMetrics `json:"stdout"`
}

type TextMetrics struct {
	Lines           int `json:"lines"`
	Words           int `json:"words"`
	Bytes           int `json:"bytes"`
	EstimatedTokens int `json:"estimated_tokens"`
}

type Trace struct {
	SchemaVersion            int               `json:"schema_version"`
	TaskID                   string            `json:"task_id"`
	Suite                    string            `json:"suite"`
	KitVersion               string            `json:"kit_version"`
	GitCommit                string            `json:"git_commit"`
	StartedAt                time.Time         `json:"started_at"`
	DurationMS               int64             `json:"duration_ms"`
	Status                   string            `json:"status"`
	WorkspacePath            string            `json:"workspace_path"`
	BaselineTraceID          string            `json:"baseline_trace_id,omitempty"`
	RepeatIndex              int               `json:"repeat_index"`
	Seed                     string            `json:"seed"`
	Commands                 []CommandTrace    `json:"commands"`
	Assertions               []AssertionResult `json:"assertions"`
	ChangedFiles             []string          `json:"changed_files"`
	AllowedSurfaceViolations []string          `json:"allowed_surface_violations"`
	OracleResults            []OracleResult    `json:"oracle_results"`
	FailureSignature         string            `json:"failure_signature,omitempty"`
}

type CommandTrace struct {
	Argv         []string    `json:"argv"`
	ExitCode     int         `json:"exit_code"`
	Status       string      `json:"status"`
	DurationMS   int64       `json:"duration_ms"`
	Error        string      `json:"error,omitempty"`
	TimedOut     bool        `json:"timed_out,omitempty"`
	Stdout       TextMetrics `json:"stdout"`
	StdoutSHA256 string      `json:"stdout_sha256"`
	StdoutPath   string      `json:"stdout_path,omitempty"`
	StderrPath   string      `json:"stderr_path,omitempty"`
}

type AssertionResult struct {
	Type         string `json:"type"`
	CommandIndex int    `json:"command_index,omitempty"`
	Status       string `json:"status"`
	Message      string `json:"message,omitempty"`
}

type OracleResult struct {
	Oracle  string `json:"oracle"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type WeaknessReport struct {
	SchemaVersion int               `json:"schema_version"`
	Kind          string            `json:"kind"`
	SourceDir     string            `json:"source_dir"`
	Clusters      []WeaknessCluster `json:"clusters"`
}

type WeaknessCluster struct {
	Signature            string   `json:"signature"`
	AffectedTasks        []string `json:"affected_tasks"`
	RepresentativeTraces []string `json:"representative_traces"`
	ObservedFailureMode  string   `json:"observed_failure_mode"`
	LikelyHarnessSurface string   `json:"likely_harness_surface"`
	Actionability        string   `json:"actionability"`
	Confidence           string   `json:"confidence"`
	ReproducibilityCount int      `json:"reproducibility_count"`
	FlakeRate            float64  `json:"flake_rate"`
	ProposedEvalTasks    []string `json:"proposed_eval_tasks"`
}

type Candidate struct {
	SchemaVersion    int      `json:"schema_version"`
	ID               string   `json:"id"`
	TargetCluster    string   `json:"target_cluster"`
	EditableSurfaces []string `json:"editable_surfaces"`
	PatchPath        string   `json:"patch_path,omitempty"`
	PromptPath       string   `json:"prompt_path,omitempty"`
	Summary          string   `json:"summary"`
	ExpectedEffect   string   `json:"expected_effect"`
	Rationale        string   `json:"rationale"`
	NegativeControls []string `json:"negative_controls,omitempty"`
	RegressionRisks  []string `json:"regression_risks"`
	Rollback         string   `json:"rollback"`
	Status           string   `json:"status"`
}

type Scorecard struct {
	SchemaVersion      int      `json:"schema_version"`
	CandidateID        string   `json:"candidate_id"`
	Score              int      `json:"score"`
	Acceptance         string   `json:"acceptance"`
	Reasons            []string `json:"reasons"`
	ValidationCommands []string `json:"validation_commands"`
}

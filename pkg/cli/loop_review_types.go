package cli

import (
	"io"
	"regexp"
	"time"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

const loopReviewSchemaVersion = 1
const loopReviewProgressEvery = 30 * time.Second

var loopReviewCorrectnessPattern = regexp.MustCompile(`(?mi)^\s*Correctness:\s*(\d{1,3})\s*%`)

type loopReviewOptions struct {
	ProjectRoot       string
	Config            *config.Config
	Feature           *feature.Feature
	BaseRef           string
	PRRef             string
	WaitForCodeRabbit bool
	MinConfidence     int
	MaxIterations     int
	DryRun            bool
	JSON              bool
	UseSubagents      bool
	Agent             config.LoopAgentConfig
	Progress          io.Writer
}

type loopReviewReport struct {
	SchemaVersion int                   `json:"schema_version"`
	RunID         string                `json:"run_id,omitempty"`
	Status        string                `json:"status"`
	StopReason    string                `json:"stop_reason,omitempty"`
	Feature       string                `json:"feature,omitempty"`
	BaseRef       string                `json:"base_ref,omitempty"`
	PRRef         string                `json:"pr_ref,omitempty"`
	PRStatus      string                `json:"pr_status,omitempty"`
	Correctness   int                   `json:"correctness,omitempty"`
	MinConfidence int                   `json:"min_confidence"`
	MaxIterations int                   `json:"max_iterations"`
	ArtifactDir   string                `json:"artifact_dir,omitempty"`
	StartedAt     time.Time             `json:"started_at"`
	EndedAt       time.Time             `json:"ended_at"`
	Iterations    []loopReviewIteration `json:"iterations"`
}

type loopReviewIteration struct {
	Index      int                    `json:"index"`
	PromptPath string                 `json:"prompt_path,omitempty"`
	StdoutPath string                 `json:"stdout_path,omitempty"`
	StderrPath string                 `json:"stderr_path,omitempty"`
	Result     *loopReviewAgentResult `json:"result,omitempty"`
	ExitCode   int                    `json:"exit_code,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StartedAt  time.Time              `json:"started_at"`
	EndedAt    time.Time              `json:"ended_at"`
	DurationMS int64                  `json:"duration_ms"`
	DryRun     bool                   `json:"dry_run,omitempty"`
}

type loopReviewTarget struct {
	BaseRef        string
	ChangedFiles   []string
	DiffStat       string
	WorkingStat    string
	StagedStat     string
	NoLocalChanges bool
}

type loopReviewAgentResult struct {
	Done        bool     `json:"done"`
	Correctness int      `json:"correctness"`
	Bullets     []string `json:"bullets,omitempty"`
	RawSummary  string   `json:"raw_summary,omitempty"`
}

type loopReviewPRFeedback struct {
	Status            reviewLoopCheckStatus
	StatusLabel       string
	Found             bool
	Fingerprint       string
	RenderedTasks     string
	CommonInstruction string
	Pending           bool
}

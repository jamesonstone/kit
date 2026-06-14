package cli

import "time"

type ciOptions struct {
	PRRef       string
	RunID       string
	JobRef      string
	WorkflowRef string
	RepoPath    string
	JSON        bool
	Dispatch    bool
	UseCopilot  bool
	NoCopilot   bool
	LogLines    int
	InputConfig freeTextInputConfig
}

type ciRepoTarget struct {
	Owner    string `json:"owner"`
	Repo     string `json:"repo"`
	FullName string `json:"full_name"`
}

type ciRepoContext struct {
	Directory        string
	ProjectRoot      string
	ConfigEligible   bool
	Target           ciRepoTarget
	DefaultBranch    string
	DefaultBranchSrc string
}

type ciTarget struct {
	Kind       string `json:"kind"`
	Repository string `json:"repository"`
	Branch     string `json:"branch,omitempty"`
	PRNumber   int    `json:"pr_number,omitempty"`
	PRURL      string `json:"pr_url,omitempty"`
	RunID      int64  `json:"run_id,omitempty"`
	Workflow   string `json:"workflow,omitempty"`
	Job        string `json:"job,omitempty"`
	HeadSHA    string `json:"head_sha,omitempty"`
}

type ciDiagnosis struct {
	Target         ciTarget       `json:"target"`
	FailureFound   bool           `json:"failure_found"`
	FailingChecks  []ciCheck      `json:"failing_checks,omitempty"`
	ExternalChecks []ciCheck      `json:"external_checks,omitempty"`
	Runs           []ciRunFailure `json:"runs,omitempty"`
	RootCause      string         `json:"root_cause"`
	Evidence       []string       `json:"evidence"`
	Recommendation string         `json:"recommendation"`
	AgentPrompt    string         `json:"agent_prompt"`
	Copilot        ciCopilotInfo  `json:"copilot"`
}

type ciCopilotInfo struct {
	Requested bool   `json:"requested"`
	Used      bool   `json:"used"`
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

type ciCheck struct {
	Name        string `json:"name"`
	Workflow    string `json:"workflow,omitempty"`
	State       string `json:"state,omitempty"`
	Bucket      string `json:"bucket,omitempty"`
	Link        string `json:"link,omitempty"`
	Description string `json:"description,omitempty"`
}

type ciRunFailure struct {
	RunID        int64          `json:"run_id"`
	Name         string         `json:"name,omitempty"`
	Workflow     string         `json:"workflow,omitempty"`
	Conclusion   string         `json:"conclusion,omitempty"`
	Status       string         `json:"status,omitempty"`
	HeadBranch   string         `json:"head_branch,omitempty"`
	HeadSHA      string         `json:"head_sha,omitempty"`
	URL          string         `json:"url,omitempty"`
	FailedJobs   []ciJobFailure `json:"failed_jobs,omitempty"`
	LogTruncated bool           `json:"log_truncated,omitempty"`
}

type ciJobFailure struct {
	JobID       int64    `json:"job_id,omitempty"`
	Name        string   `json:"name"`
	Conclusion  string   `json:"conclusion,omitempty"`
	Status      string   `json:"status,omitempty"`
	URL         string   `json:"url,omitempty"`
	FailedSteps []string `json:"failed_steps,omitempty"`
	LogExcerpt  []string `json:"log_excerpt,omitempty"`
}

type ciRun struct {
	Attempt            int       `json:"attempt"`
	Conclusion         string    `json:"conclusion"`
	CreatedAt          time.Time `json:"createdAt"`
	DatabaseID         int64     `json:"databaseId"`
	DisplayTitle       string    `json:"displayTitle"`
	Event              string    `json:"event"`
	HeadBranch         string    `json:"headBranch"`
	HeadSHA            string    `json:"headSha"`
	Name               string    `json:"name"`
	Number             int       `json:"number"`
	StartedAt          time.Time `json:"startedAt"`
	Status             string    `json:"status"`
	UpdatedAt          time.Time `json:"updatedAt"`
	URL                string    `json:"url"`
	WorkflowDatabaseID int64     `json:"workflowDatabaseId"`
	WorkflowName       string    `json:"workflowName"`
	Jobs               []ciJob   `json:"jobs"`
}

type ciJob struct {
	CompletedAt time.Time `json:"completedAt"`
	Conclusion  string    `json:"conclusion"`
	DatabaseID  int64     `json:"databaseId"`
	Name        string    `json:"name"`
	StartedAt   time.Time `json:"startedAt"`
	Status      string    `json:"status"`
	Steps       []ciStep  `json:"steps"`
	URL         string    `json:"url"`
}

type ciStep struct {
	Conclusion string `json:"conclusion"`
	Name       string `json:"name"`
	Number     int    `json:"number"`
	Status     string `json:"status"`
}

type ciWorkflow struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Path  string `json:"path"`
	State string `json:"state"`
}

type ciPR struct {
	Number      int    `json:"number"`
	HeadRefName string `json:"headRefName"`
	HeadRefOID  string `json:"headRefOid"`
	Title       string `json:"title"`
	URL         string `json:"url"`
}

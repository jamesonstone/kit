package prfix

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/jamesonstone/kit/internal/registry"
)

const (
	DefaultMaxSubagents = 3
	HardMaxSubagents    = 4
)

type Repository struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (repository Repository) Slug() string { return repository.Owner + "/" + repository.Name }

type Target struct {
	Repository
	Number int `json:"number"`
}

func (target Target) URL() string {
	return fmt.Sprintf("https://github.com/%s/pull/%d", target.Slug(), target.Number)
}

type OpenPullRequest struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	HeadRefName    string `json:"headRefName"`
	BaseRefName    string `json:"baseRefName"`
	IsDraft        bool   `json:"isDraft"`
	ReviewDecision string `json:"reviewDecision"`
}

type PullRequest struct {
	Number            int    `json:"number"`
	Title             string `json:"title"`
	URL               string `json:"url"`
	State             string `json:"state"`
	HeadRefName       string `json:"headRefName"`
	HeadRefOID        string `json:"headRefOid"`
	BaseRefName       string `json:"baseRefName"`
	IsCrossRepository bool   `json:"isCrossRepository"`
}

type Feedback struct {
	Kind        string `json:"kind"`
	NodeID      string `json:"node_id"`
	ThreadID    string `json:"thread_id,omitempty"`
	Path        string `json:"path,omitempty"`
	Line        int    `json:"line,omitempty"`
	Author      string `json:"author"`
	URL         string `json:"url"`
	Body        string `json:"body"`
	Task        string `json:"task"`
	Fingerprint string `json:"fingerprint"`
}

type CollectionOptions struct {
	CodeRabbitOnly      bool
	TrustedCommentUsers []string
}

type Collection struct {
	Items        []Feedback `json:"items"`
	ActiveCount  int        `json:"active_count"`
	RequestCount int        `json:"request_count"`
}

type Lane struct {
	Repository     string   `json:"repository"`
	PRNumber       int      `json:"pull_request"`
	PRURL          string   `json:"pull_request_url"`
	HeadBranch     string   `json:"head_branch"`
	ExpectedHead   string   `json:"expected_remote_head"`
	LocalHead      string   `json:"local_head"`
	WorktreePath   string   `json:"worktree_path"`
	PushTarget     string   `json:"push_target"`
	Created        bool     `json:"worktree_created"`
	DirtyStatus    string   `json:"dirty_status,omitempty"`
	DirtyPaths     []string `json:"dirty_paths,omitempty"`
	DirtyOwnership string   `json:"dirty_ownership"`
}

type GitHub interface {
	CurrentRepository(context.Context, string) (Repository, error)
	ListOpenPullRequests(context.Context, string, Repository) ([]OpenPullRequest, error)
	PullRequest(context.Context, string, Target) (PullRequest, error)
	Collect(context.Context, Target, registry.PRFeedbackContract, CollectionOptions, *Budget) (Collection, error)
	Status(context.Context, Target, string) (registry.PRFeedbackObservation, error)
	ResolveThread(context.Context, string) error
}

type Sleeper func(context.Context, time.Duration) error

type Output struct {
	Stdout io.Writer
	Stderr io.Writer
}

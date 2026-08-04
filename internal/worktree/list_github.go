package worktree

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultListPRTimeout = 2 * time.Second

	listPRNoneMarker      = "-"
	listPRMissingGHMarker = "NG"
	listPRRateLimitMarker = "RL"
	listPRTimeoutMarker   = "TO"
	listPRUnknownMarker   = "??"
)

type listPullRequest struct {
	Number            int    `json:"number"`
	HeadRefName       string `json:"headRefName"`
	IsCrossRepository bool   `json:"isCrossRepository"`
	Title             string `json:"title"`
}

type listPRAnnotation struct {
	numbers string
	titles  string
}

type listPRLookup struct {
	byBranch map[string]listPRAnnotation
	fallback string
}

type listPRResolverFunc func(context.Context, string) listPRLookup

func successfulListPRLookup(byBranch map[string]listPRAnnotation) listPRLookup {
	return listPRLookup{byBranch: byBranch, fallback: listPRNoneMarker}
}

func failedListPRLookup(marker string) listPRLookup {
	return listPRLookup{fallback: marker}
}

func (a *App) populateListPullRequests(
	ctx context.Context,
	cwd string,
	entries []worktreeEntry,
) {
	lookup := a.resolveListPRs(ctx, cwd)
	if lookup.fallback == "" {
		lookup.fallback = listPRUnknownMarker
	}
	for i := range entries {
		entries[i].prText = lookup.fallback
		entries[i].prTitle = lookup.fallback
		if annotation, ok := lookup.byBranch[entries[i].branch]; ok {
			entries[i].prText = annotation.numbers
			entries[i].prTitle = annotation.titles
		}
	}
}

func (a *App) resolveListPullRequests(ctx context.Context, cwd string) listPRLookup {
	if _, err := a.lookPath("gh"); err != nil {
		return failedListPRLookup(listPRMissingGHMarker)
	}

	lookupCtx, cancel := context.WithTimeout(ctx, a.listPRTimeout)
	defer cancel()
	output, err := a.run(
		lookupCtx,
		cwd,
		"gh",
		"pr",
		"list",
		"--state",
		"open",
		"--limit",
		"1000",
		"--json",
		"number,headRefName,isCrossRepository,title",
	)
	if err != nil {
		return failedListPRLookup(listPRFailureMarker(lookupCtx, output, err))
	}

	var prs []listPullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return failedListPRLookup(listPRUnknownMarker)
	}
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].Number < prs[j].Number
	})
	byBranch := make(map[string]listPRAnnotation, len(prs))
	for _, pr := range prs {
		if pr.IsCrossRepository || pr.Number <= 0 || pr.HeadRefName == "" {
			continue
		}
		title := pr.Title
		if strings.TrimSpace(title) == "" {
			title = listPRUnknownMarker
		}
		number := strconv.Itoa(pr.Number)
		annotation := byBranch[pr.HeadRefName]
		if annotation.numbers != "" {
			annotation.numbers += ","
			annotation.titles += " | "
		}
		annotation.numbers += number
		annotation.titles += title
		byBranch[pr.HeadRefName] = annotation
	}
	return successfulListPRLookup(byBranch)
}

func listPRFailureMarker(
	ctx context.Context,
	output []byte,
	err error,
) string {
	if ctx.Err() == context.DeadlineExceeded {
		return listPRTimeoutMarker
	}
	detail := strings.ToLower(string(output) + "\n" + err.Error())
	switch {
	case strings.Contains(detail, "rate limit"),
		strings.Contains(detail, "ratelimit"),
		strings.Contains(detail, "too many requests"),
		strings.Contains(detail, "http 429"),
		strings.Contains(detail, "status 429"):
		return listPRRateLimitMarker
	case strings.Contains(detail, "executable file not found"),
		strings.Contains(detail, "command not found"):
		return listPRMissingGHMarker
	default:
		return listPRUnknownMarker
	}
}

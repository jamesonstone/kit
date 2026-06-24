package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	prFixLoopReviewRunner = runLoopReviewCommand
	prFixOpenPRLister     = listPRFixOpenPullRequests
)

type prFixOptions struct {
	PRRef             string
	WaitForCodeRabbit bool
	MinConfidence     int
	MaxIterations     int
	DryRun            bool
	JSON              bool
	UseSubagents      bool
}

type prFixOpenPullRequest struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	HeadRefName    string `json:"headRefName"`
	BaseRefName    string `json:"baseRefName"`
	IsDraft        bool   `json:"isDraft"`
	ReviewDecision string `json:"reviewDecision"`
}

func init() {
	rootCmd.AddCommand(newPRCommand())
}

func newPRCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pr",
		Short:         "Run pull-request repair workflows",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Run pull-request repair workflows.

Use kit pr fix to select or target a pull request, ingest current CodeRabbit
review feedback, and run the local review repair loop. Kit itself does not
stage, commit, push, post PR comments, or resolve review threads from this
command; those remain behind the repo-local delivery gate.`,
	}
	cmd.AddCommand(newPRFixCommand())
	return cmd
}

func newPRFixCommand() *cobra.Command {
	opts := prFixOptions{}
	cmd := &cobra.Command{
		Use:           "fix [feature]",
		Short:         "Repair a pull request from CodeRabbit feedback",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Repair a pull request from current CodeRabbit feedback.

With --pr, the target can be a GitHub PR URL, a Markdown PR link,
owner/repo#number, or a pull request number in the current repository.

Without --pr, Kit lists open pull requests in the current repository and asks
which one to repair. The selected PR is passed to the same correctness loop as
kit loop review --pr, so local changes can be made by the configured agent but
GitHub delivery remains a separate, explicit step.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPRFixCommand(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.PRRef, "pr", "", "pull request URL, Markdown link, owner/repo#number, or current-repo number")
	cmd.Flags().BoolVar(&opts.WaitForCodeRabbit, "watch", false, "wait for CodeRabbit completion before finalizing PR-mode review")
	cmd.Flags().BoolVar(&opts.WaitForCodeRabbit, "wait-for-coderabbit", false, "wait for CodeRabbit completion before finalizing PR-mode review")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "show the first review prompt without running the configured agent")
	cmd.Flags().IntVar(&opts.MinConfidence, "min-confidence", 0, "minimum correctness percentage required to stop (0 uses loop config, goal_percentage, then 95)")
	cmd.Flags().IntVar(&opts.MaxIterations, "max-iterations", 0, "maximum review iterations (0 uses loop config, then 10)")
	cmd.Flags().BoolVar(&opts.UseSubagents, "subagents", false, "allow the review agent to pre-analyze and choose subagents")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output loop review report as JSON")
	return cmd
}

func runPRFixCommand(cmd *cobra.Command, args []string, opts prFixOptions) error {
	prRef := strings.TrimSpace(opts.PRRef)
	if prRef == "" {
		if opts.JSON {
			return fmt.Errorf("--json requires --pr because interactive PR selection writes human-readable output")
		}
		selected, err := selectPRFixOpenPullRequest(cmd.InOrStdin(), cmd.OutOrStdout())
		if err != nil {
			return err
		}
		prRef = selected
	}

	return prFixLoopReviewRunner(cmd, args, loopReviewOptions{
		PRRef:             prRef,
		WaitForCodeRabbit: opts.WaitForCodeRabbit,
		MinConfidence:     opts.MinConfidence,
		MaxIterations:     opts.MaxIterations,
		DryRun:            opts.DryRun,
		JSON:              opts.JSON,
		UseSubagents:      opts.UseSubagents,
	})
}

func listPRFixOpenPullRequests() ([]prFixOpenPullRequest, error) {
	output, err := commandOutput(
		"gh",
		"pr",
		"list",
		"--state",
		"open",
		"--limit",
		"50",
		"--json",
		"number,title,url,headRefName,baseRefName,isDraft,reviewDecision",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list open pull requests: %w", err)
	}

	var prs []prFixOpenPullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse open pull request list: %w", err)
	}
	return prs, nil
}

func selectPRFixOpenPullRequest(input io.Reader, output io.Writer) (string, error) {
	prs, err := prFixOpenPRLister()
	if err != nil {
		return "", err
	}
	if len(prs) == 0 {
		return "", fmt.Errorf("no open pull requests found in the current repository; pass --pr <url|owner/repo#number|number> to target another PR")
	}

	if _, err := fmt.Fprintln(output, "Open pull requests:"); err != nil {
		return "", err
	}
	for index, pr := range prs {
		if _, err := fmt.Fprintf(output, "  %d. #%d %s [%s -> %s] %s\n",
			index+1,
			pr.Number,
			strings.TrimSpace(pr.Title),
			strings.TrimSpace(pr.HeadRefName),
			strings.TrimSpace(pr.BaseRefName),
			prFixPRStateLabel(pr),
		); err != nil {
			return "", err
		}
	}
	if _, err := fmt.Fprintf(output, "Select PR (1-%d or PR number): ", len(prs)); err != nil {
		return "", err
	}

	reader := bufio.NewReader(input)
	line, err := reader.ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return "", fmt.Errorf("failed to read PR selection: %w", err)
	}
	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return "", fmt.Errorf("invalid PR selection %q", strings.TrimSpace(line))
	}

	for _, pr := range prs {
		if pr.Number == choice {
			return prFixPRRef(pr), nil
		}
	}
	if choice >= 1 && choice <= len(prs) {
		return prFixPRRef(prs[choice-1]), nil
	}
	return "", fmt.Errorf("PR selection %d is not in the listed PRs", choice)
}

func prFixPRRef(pr prFixOpenPullRequest) string {
	if strings.TrimSpace(pr.URL) != "" {
		return strings.TrimSpace(pr.URL)
	}
	return strconv.Itoa(pr.Number)
}

func prFixPRStateLabel(pr prFixOpenPullRequest) string {
	var parts []string
	if pr.IsDraft {
		parts = append(parts, "draft")
	} else {
		parts = append(parts, "ready")
	}
	if strings.TrimSpace(pr.ReviewDecision) != "" {
		parts = append(parts, strings.ToLower(strings.TrimSpace(pr.ReviewDecision)))
	}
	return strings.Join(parts, ", ")
}

package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	prFixDispatchInputLoader = loadDispatchPRInput
	prFixDispatchRunner      = runPRFixDispatchPrompt
	prFixDispatchTasksLoader = loadDispatchPRTasks
	prFixOpenPRLister        = listPRFixOpenPullRequests
)

type prFixOptions struct {
	PRRef          string
	CodeRabbitOnly bool
	Copy           bool
	Edit           bool
	Editor         string
	MaxSubagents   int
	OutputOnly     bool
	UseVim         bool
}

type prFixDispatchOptions struct {
	PRRef          string
	CodeRabbitOnly bool
	Copy           bool
	Edit           bool
	Editor         string
	MaxSubagents   int
	OutputOnly     bool
	UseVim         bool
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

Use kit pr fix to select or target a pull request, ingest current PR review
feedback, and copy the resulting agent prompt. Editing the review tasks is
opt-in with --edit, --vim, or --editor. Kit itself does not edit files, stage,
commit, push, post PR comments, or resolve review threads from this path. After
fixes or no-op decisions are complete, pushed, and reflected against the PR
head, resolve handled review threads explicitly with
kit dispatch --pr <target> --resolve --yes.`,
	}
	cmd.AddCommand(newPRFixCommand())
	return cmd
}

func newPRFixCommand() *cobra.Command {
	opts := prFixOptions{}
	cmd := &cobra.Command{
		Use:           "fix",
		Short:         "Repair a pull request from review feedback",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Repair a pull request from current review feedback.

With --pr, the target can be a GitHub PR URL, a Markdown PR link,
owner/repo#number, or a pull request number in the current repository.

Without --pr, Kit lists open pull requests in the current repository and asks
which one to repair. The selected PR uses the same prompt-producing flow as
kit dispatch --pr: only active (unresolved, non-outdated) review threads
become a dispatch prompt, and the prompt is copied for pasting to a coding
agent. Pass --edit to review and change the task list in the default editor
before it is copied; --vim and --editor also opt into editing. GitHub delivery
remains a separate, explicit step. The generated prompt requires post-push
reflection before resolving verified addressed review conversations.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPRFixCommand(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.PRRef, "pr", "", "pull request URL, Markdown link, owner/repo#number, or current-repo number")
	cmd.Flags().BoolVar(&opts.CodeRabbitOnly, "coderabbit", false, "include only CodeRabbit-authored review comments")
	cmd.Flags().BoolVar(&opts.Copy, "copy", false, "copy prompt to clipboard even with --output-only")
	cmd.Flags().BoolVar(&opts.Edit, "edit", false, "open review tasks in the default editor before generating the prompt")
	cmd.Flags().StringVar(&opts.Editor, "editor", "", "open review tasks in a specific editor command before generating the prompt")
	cmd.Flags().BoolVar(&opts.OutputOnly, "output-only", false, "output prompt text to stdout instead of copying it to the clipboard")
	cmd.Flags().BoolVar(&opts.UseVim, "vim", false, "open review tasks in a vim-compatible editor before generating the prompt")
	cmd.Flags().IntVar(&opts.MaxSubagents, "max-subagents", defaultDispatchMaxSubagents, "maximum concurrent subagents allowed in the generated prompt; default 3, hard ceiling 4")
	return cmd
}

func runPRFixCommand(cmd *cobra.Command, _ []string, opts prFixOptions) error {
	prRef := strings.TrimSpace(opts.PRRef)
	if prRef == "" {
		selected, err := selectPRFixOpenPullRequest(cmd.InOrStdin(), cmd.OutOrStdout())
		if err != nil {
			return err
		}
		prRef = selected
	}

	return prFixDispatchRunner(cmd, prFixDispatchOptions{
		PRRef:          prRef,
		CodeRabbitOnly: opts.CodeRabbitOnly,
		Copy:           opts.Copy,
		Edit:           opts.Edit,
		Editor:         opts.Editor,
		MaxSubagents:   opts.MaxSubagents,
		OutputOnly:     opts.OutputOnly,
		UseVim:         opts.UseVim,
	})
}

func runPRFixDispatchPrompt(cmd *cobra.Command, opts prFixDispatchOptions) error {
	if err := validateDispatchMaxSubagents(opts.MaxSubagents); err != nil {
		return err
	}

	prInput, found, err := loadPRFixDispatchInput(opts)
	if err != nil {
		return err
	}
	if !found {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "No actionable PR review comments found.")
		return err
	}

	tasks, err := normalizeDispatchTasks(prInput.RawTasks)
	if err != nil {
		return err
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	prompt := buildDispatchPrompt(
		tasks,
		opts.MaxSubagents,
		workingDirectory,
		dispatchInputSourcePR,
		dispatchPromptOptions{
			CodeRabbitOnly:          opts.CodeRabbitOnly,
			CommonReviewInstruction: prInput.CommonReviewInstruction,
			PRTarget:                opts.PRRef,
		},
	)
	if err := outputPromptWithoutSubagentsWithClipboardDefault(prompt, opts.OutputOnly, opts.Copy); err != nil {
		return err
	}

	if !opts.OutputOnly {
		printWorkflowInstructions("pr fix (dispatch prompt)", []string{
			"paste the copied prompt into your coding agent",
			"verify each finding against current code before changing files",
			"after push-to-PR and reflection, resolve verified handled threads explicitly with kit dispatch --pr <target> --resolve --yes",
		})
	}

	return nil
}

func loadPRFixDispatchInput(opts prFixDispatchOptions) (dispatchPRInput, bool, error) {
	if !shouldEditPRFixTasks(opts) {
		return prFixDispatchTasksLoader(opts.PRRef, opts.CodeRabbitOnly)
	}

	inputCfg := newFreeTextInputConfig(opts.UseVim, opts.Editor, false, opts.Edit)
	return prFixDispatchInputLoader(opts.PRRef, opts.CodeRabbitOnly, inputCfg)
}

func shouldEditPRFixTasks(opts prFixDispatchOptions) bool {
	return opts.Edit || opts.UseVim || strings.TrimSpace(opts.Editor) != ""
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

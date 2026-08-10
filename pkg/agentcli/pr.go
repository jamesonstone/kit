package agentcli

import (
	"time"

	"github.com/jamesonstone/kit/internal/prfix"
	"github.com/spf13/cobra"
)

type prFixOptions struct {
	PRRef                 string
	CodeRabbitOnly        bool
	Copy                  bool
	Edit                  bool
	Editor                string
	OutputOnly            bool
	UseVim                bool
	Wait                  bool
	Timeout               time.Duration
	MaxSubagents          int
	IncludeDirty          bool
	ExcludeDirty          bool
	TrustedCommentAuthors []string
	Resolve               bool
	Threads               []string
	Head                  string
	Yes                   bool
}

func newPRCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "pr",
		Short: "Use the protected pull-request feedback fallback",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(newPRFixCommand())
	return command
}

func newPRFixCommand() *cobra.Command {
	options := &prFixOptions{}
	command := &cobra.Command{
		Use:   "fix",
		Short: "Generate a coding-agent prompt from active PR feedback",
		Long: `Generate a coding-agent repair prompt from one exact pull request.

Without --pr, Kit lists open pull requests in the current repository. With
--pr, pass a GitHub URL, Markdown link, owner/repo#number, or current-repository
number. One-shot collection includes active human feedback by default;
--coderabbit filters to CodeRabbit authors. Kit prepares the exact writable
same-repository PR-head lane and never repairs from detached PR-N.

Kit does not launch agents, edit source, stage, commit, push, post comments,
resolve threads implicitly, or merge. --wait uses the bounded local workflow
monitor. --resolve requires an exact pushed head, explicit active thread IDs,
and --yes.`,
		Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runPRFix(command, *options)
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.PRRef, "pr", "", "PR URL, Markdown link, owner/repo#number, or current-repository number")
	flags.BoolVar(&options.CodeRabbitOnly, "coderabbit", false, "include only CodeRabbit-authored feedback")
	flags.BoolVar(&options.Copy, "copy", false, "copy the prompt, including with --output-only")
	flags.BoolVar(&options.Edit, "edit", false, "edit collected feedback in the default editor before prompt generation")
	flags.StringVar(&options.Editor, "editor", "", "edit collected feedback with this editor command")
	flags.BoolVar(&options.OutputOnly, "output-only", false, "write only the generated prompt to stdout")
	flags.BoolVar(&options.UseVim, "vim", false, "edit collected feedback with nvim or vim")
	flags.BoolVar(&options.Wait, "wait", false, "boundedly await confirmed CodeRabbit completion before collection")
	flags.DurationVar(&options.Timeout, "timeout", 25*time.Minute, "await timeout, bounded by the local workflow ceiling")
	flags.IntVar(&options.MaxSubagents, "max-subagents", prfix.DefaultMaxSubagents, "maximum concurrent repair lanes; default 3, hard ceiling 4")
	flags.BoolVar(&options.IncludeDirty, "include-dirty", false, "explicitly include existing repair-lane changes")
	flags.BoolVar(&options.ExcludeDirty, "exclude-dirty", false, "explicitly exclude existing repair-lane changes")
	flags.StringSliceVar(&options.TrustedCommentAuthors, "trusted-comment-author", nil, "include marked top-level feedback from this trusted author; repeat as needed")
	flags.BoolVar(&options.Resolve, "resolve", false, "resolve only explicitly named verified threads")
	flags.StringArrayVar(&options.Threads, "thread", nil, "exact review-thread ID verified addressed at --head; repeat as needed")
	flags.StringVar(&options.Head, "head", "", "exact pushed PR-head SHA required by --resolve")
	flags.BoolVar(&options.Yes, "yes", false, "confirm explicit GitHub thread resolution")
	return command
}

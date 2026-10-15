package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/cleanup"
	"github.com/spf13/cobra"
)

func newCleanupCommand() *cobra.Command {
	var record, apply, writersStopped, disposable bool
	var path, repo, kind, owner, receiptPath, head, stop, remote, ref string
	var pr int
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Reclaim proven-recoverable task artifacts without routine prompts",
		Long:  "Record creation ownership, then prove completion, recovery and exclusive handoff immediately before bounded cleanup. Unknown or unique state is retained. Read docs/references/rules/work-cleanup.md before use.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), cleanup.DefaultTimeout)
			defer cancel()
			service := cleanup.New()
			if record {
				if (kind == "scratch" || kind == "cache") && !disposable {
					return fmt.Errorf("generated output requires --disposable: the task produced it and it never became authoritative")
				}
				if apply || receiptPath != "" || head != "" || pr != 0 || stop != "" || writersStopped {
					return fmt.Errorf("record cannot apply or assert completion")
				}
				receipt, err := service.Record(ctx, kind, path, repo, owner, ref)
				if err != nil {
					return err
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(receipt)
			}
			if receiptPath == "" {
				return fmt.Errorf("--receipt is required; use --record at task artifact creation")
			}
			if path != "" || kind != "" || ref != "" || disposable || cmd.Flags().Changed("repository") {
				return fmt.Errorf("apply targets come only from the recorded receipt")
			}
			file, err := os.Open(receiptPath)
			if err != nil {
				return err
			}
			defer func() { _ = file.Close() }()
			receipt, err := decodeCleanupReceipt(file)
			if err != nil {
				return err
			}
			absolute, err := filepath.Abs(receiptPath)
			if err != nil {
				return err
			}
			absolute, err = filepath.EvalSymlinks(absolute)
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(receipt.Path, absolute)
			if err != nil || relative == "." || (!filepath.IsAbs(relative) && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
				return fmt.Errorf("ownership receipt must remain outside its target")
			}
			result := service.Execute(ctx, receipt, cleanup.Options{Owner: owner, FinishedHead: head, StopEvidence: stop, Remote: remote, PR: pr, WritersStopped: writersStopped, Apply: apply})
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		},
	}
	flags := cmd.Flags()
	flags.BoolVar(&disposable, "disposable", false, "record only task-produced, never-authoritative scratch/cache output")
	flags.BoolVar(&record, "record", false, "record a newly task-created artifact; primary records bounded sync scope")
	flags.StringVar(&kind, "kind", "", "record worktree, branch, ref, scratch, cache, or primary")
	flags.StringVar(&path, "path", "", "exact physical artifact directory, for record only")
	flags.StringVar(&repo, "repository", ".", "repository for Git evidence, for record only")
	flags.StringVar(&owner, "owner", "", "stable current task/session identifier")
	flags.StringVar(&ref, "ref", "", "exact branch/tracking ref or optional primary default ref, for record only")
	flags.StringVar(&receiptPath, "receipt", "", "creation receipt retained outside the target")
	flags.StringVar(&head, "head", "", "exact completed task source head")
	flags.IntVar(&pr, "pr", 0, "pull request whose current completion matches the head")
	flags.StringVar(&stop, "stop-evidence", "", "current human stop/supersession instruction reference, instead of a PR")
	flags.StringVar(&remote, "remote", "origin", "selected recovery remote")
	flags.BoolVar(&writersStopped, "writers-stopped", false, "current owner exclusively hands off after all writers stop; never infer")
	flags.BoolVar(&apply, "apply", false, "re-prove evidence under cleanup lease, then act; otherwise preview")
	return cmd
}

func init() { rootCmd.AddCommand(newCleanupCommand()) }

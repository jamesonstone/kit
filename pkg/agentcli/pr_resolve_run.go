package agentcli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/prfix"
	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

func runPRFixResolution(
	command *cobra.Command,
	runtime prFixRuntime,
	cwd string,
	target prfix.Target,
	contract registry.PRFeedbackContract,
	options prFixOptions,
) error {
	pullRequest, err := runtime.github.PullRequest(command.Context(), cwd, target)
	if err != nil {
		return err
	}
	if options.Head != pullRequest.HeadRefOID {
		return fmt.Errorf("--head %s does not match current PR head %s", options.Head, pullRequest.HeadRefOID)
	}
	lane, err := runtime.lane.Resolve(command.Context(), cwd, target, pullRequest)
	if err != nil {
		return err
	}
	if len(lane.DirtyPaths) > 0 {
		return fmt.Errorf("repair lane has uncommitted changes; refuse thread resolution until pushed-head verification is clean")
	}
	budget := &prfix.Budget{Limit: contract.RequestBudgetPerHead}
	collection, err := runtime.github.Collect(command.Context(), target, contract, prfix.CollectionOptions{
		CodeRabbitOnly: options.CodeRabbitOnly, TrustedCommentUsers: options.TrustedCommentAuthors,
	}, budget)
	if err != nil {
		return err
	}
	refreshed, err := runtime.github.PullRequest(command.Context(), cwd, target)
	if err != nil {
		return err
	}
	if err := verifyHeadUnchanged(options.Head, refreshed); err != nil {
		return err
	}
	threadIDs, err := prfix.ValidateResolution(options.Head, refreshed, options.Threads, collection.Items)
	if err != nil {
		return err
	}
	for index, threadID := range threadIDs {
		if err := runtime.github.ResolveThread(command.Context(), threadID); err != nil {
			return fmt.Errorf("resolved %d/%d threads before failure: %w", index, len(threadIDs), err)
		}
	}
	if _, err := fmt.Fprintf(command.OutOrStdout(), "Resolved %d verified review thread(s) at head %s:\n", len(threadIDs), options.Head); err != nil {
		return err
	}
	for _, threadID := range threadIDs {
		if _, err := fmt.Fprintln(command.OutOrStdout(), "-", threadID); err != nil {
			return err
		}
	}
	return nil
}

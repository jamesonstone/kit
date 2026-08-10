package agentcli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/prfix"
	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

func runPRFixPrompt(
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
	budget := &prfix.Budget{Limit: contract.RequestBudgetPerHead}
	var lane prfix.Lane
	await := prfix.AwaitResult{
		SchemaVersion: 1, Mode: "collect", State: registry.PRFeedbackCompleted,
		Repository: target.Slug(), PullRequest: target.Number, ExpectedHead: pullRequest.HeadRefOID,
		ObservedHead: pullRequest.HeadRefOID, Reason: "explicit one-shot collection",
	}
	if options.Wait {
		lane, err = runtime.lane.Resolve(command.Context(), cwd, target, pullRequest)
		if err != nil {
			return err
		}
		lane, err = resolveDirtyOwnership(command, lane, options, nil)
		if err != nil {
			return err
		}
		release, lockErr := runtime.state.Acquire(target, pullRequest.HeadRefOID)
		if lockErr != nil {
			return lockErr
		}
		defer release()
		await = runtime.monitor.Await(command.Context(), runtime.github, target,
			pullRequest.HeadRefOID, contract, options.Timeout, budget)
		if await.State != registry.PRFeedbackCompleted {
			if saveErr := runtime.state.Save(target, pullRequest.HeadRefOID, await, nil); saveErr != nil {
				return fmt.Errorf("await ended in %s and state persistence failed: %v", await.State, saveErr)
			}
			_ = writeJSON(command.OutOrStdout(), await)
			return &exitError{code: 2, err: fmt.Errorf("PR feedback await ended in %s: %s", await.State, await.Reason)}
		}
	}
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
	if err := verifyHeadUnchanged(pullRequest.HeadRefOID, refreshed); err != nil {
		return err
	}
	if len(collection.Items) == 0 {
		if err := runtime.state.Save(target, pullRequest.HeadRefOID, await, nil); err != nil {
			return err
		}
		output := command.OutOrStdout()
		if options.OutputOnly {
			output = command.ErrOrStderr()
		}
		_, err := fmt.Fprintln(output, "No actionable unresolved, non-outdated PR feedback found.")
		return err
	}
	if !options.Wait {
		lane, err = runtime.lane.Resolve(command.Context(), cwd, target, pullRequest)
		if err != nil {
			return err
		}
		lane, err = resolveDirtyOwnership(command, lane, options, collection.Items)
		if err != nil {
			return err
		}
	} else {
		lane, err = prfix.ApplyDirtyOwnership(lane, lane.DirtyOwnership, collection.Items)
		if err != nil {
			return err
		}
	}
	if err := runtime.state.Save(target, pullRequest.HeadRefOID, await, collection.Items); err != nil {
		return err
	}
	tasks := prfix.RenderFeedback(collection.Items)
	if shouldEditPRFix(options) {
		tasks, err = prFixEditor(command, options, tasks)
		if err != nil {
			return err
		}
	}
	prompt, err := prfix.BuildPrompt(target, lane, collection.Items, tasks, options.MaxSubagents)
	if err != nil {
		return err
	}
	return outputPRFixPrompt(command, prompt, options)
}

func outputPRFixPrompt(command *cobra.Command, prompt string, options prFixOptions) error {
	if options.OutputOnly {
		if options.Copy {
			if err := clipboardCopyFunc(prompt); err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(command.OutOrStdout(), prompt)
		return err
	}
	if err := clipboardCopyFunc(prompt); err != nil {
		return fmt.Errorf("copy PR repair prompt: %w", err)
	}
	_, err := fmt.Fprintln(command.OutOrStdout(), "PR repair prompt copied to clipboard; paste it into a coding agent.")
	return err
}

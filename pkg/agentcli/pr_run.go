package agentcli

import (
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/internal/prfix"
	"github.com/jamesonstone/kit/internal/registry"
	"github.com/spf13/cobra"
)

func runPRFix(command *cobra.Command, options prFixOptions) error {
	if err := validatePRFixOptions(options); err != nil {
		return err
	}
	runtime, err := newPRFixRuntime(prfix.Output{Stdout: command.ErrOrStderr(), Stderr: command.ErrOrStderr()})
	if err != nil {
		return err
	}
	cwd := mustGetwd()
	root, err := registry.FindProjectRoot(cwd)
	if err != nil {
		return err
	}
	contract, err := prfix.LoadContract(root)
	if err != nil {
		return err
	}
	current, err := runtime.github.CurrentRepository(command.Context(), cwd)
	if err != nil {
		return err
	}
	reference := strings.TrimSpace(options.PRRef)
	if reference == "" {
		pullRequests, listErr := runtime.github.ListOpenPullRequests(command.Context(), cwd, current)
		if listErr != nil {
			return listErr
		}
		reference, err = selectOpenPullRequest(command.InOrStdin(), command.OutOrStdout(), pullRequests)
		if err != nil {
			return err
		}
	}
	target, err := prfix.ParseTarget(reference, current)
	if err != nil {
		return err
	}
	if !prfix.SameRepository(target.Repository, current) {
		return fmt.Errorf("target repository %s does not match current repository %s", target.Slug(), current.Slug())
	}
	if options.Resolve {
		return runPRFixResolution(command, runtime, cwd, target, contract, options)
	}
	return runPRFixPrompt(command, runtime, cwd, target, contract, options)
}

func validatePRFixOptions(options prFixOptions) error {
	if err := prfix.ValidateMaxSubagents(options.MaxSubagents); err != nil {
		return err
	}
	if options.IncludeDirty && options.ExcludeDirty {
		return fmt.Errorf("--include-dirty and --exclude-dirty are mutually exclusive")
	}
	editorModes := 0
	if options.Edit {
		editorModes++
	}
	if options.UseVim {
		editorModes++
	}
	if strings.TrimSpace(options.Editor) != "" {
		editorModes++
	}
	if editorModes > 1 {
		return fmt.Errorf("use only one of --edit, --vim, or --editor")
	}
	if options.Resolve {
		if options.Wait || options.OutputOnly || options.Copy || editorModes > 0 ||
			options.IncludeDirty || options.ExcludeDirty {
			return fmt.Errorf("--resolve cannot be combined with wait, prompt-output, editor, or dirty-ownership flags")
		}
		if !options.Yes {
			return fmt.Errorf("--resolve mutates GitHub; rerun with --yes after exact-head reflection")
		}
	} else if len(options.Threads) > 0 || options.Head != "" || options.Yes {
		return fmt.Errorf("--thread, --head, and --yes require --resolve")
	}
	return nil
}

func resolveDirtyOwnership(command *cobra.Command, lane prfix.Lane, options prFixOptions, feedback []prfix.Feedback) (prfix.Lane, error) {
	if len(lane.DirtyPaths) == 0 {
		return prfix.ApplyDirtyOwnership(lane, "none", feedback)
	}
	ownership := ""
	if options.IncludeDirty {
		ownership = "include"
	} else if options.ExcludeDirty {
		ownership = "exclude"
	} else {
		if _, err := fmt.Fprintln(command.ErrOrStderr(), "Repair worktree has existing changes:"); err != nil {
			return prfix.Lane{}, err
		}
		if _, err := fmt.Fprintln(command.ErrOrStderr(), lane.DirtyStatus); err != nil {
			return prfix.Lane{}, err
		}
		if _, err := fmt.Fprint(command.ErrOrStderr(), "Ownership [include/exclude/abort]: "); err != nil {
			return prfix.Lane{}, err
		}
		if _, err := fmt.Fscanln(command.InOrStdin(), &ownership); err != nil {
			return prfix.Lane{}, fmt.Errorf("read dirty-change ownership: %w", err)
		}
		ownership = strings.ToLower(strings.TrimSpace(ownership))
		if ownership == "abort" {
			return prfix.Lane{}, fmt.Errorf("dirty repair lane was not assigned explicit ownership")
		}
	}
	return prfix.ApplyDirtyOwnership(lane, ownership, feedback)
}

func verifyHeadUnchanged(expected string, refreshed prfix.PullRequest) error {
	if refreshed.HeadRefOID != expected {
		return fmt.Errorf("pull request head changed from %s to %s", expected, refreshed.HeadRefOID)
	}
	return nil
}

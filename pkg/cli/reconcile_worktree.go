package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/worktreeprep"
)

type reconcileRefreshDecision struct {
	Apply    bool
	Deferred bool
}

func buildDeferredReconcileCommand(args []string) string {
	command := []string{"kit", "reconcile"}
	if len(args) == 1 {
		command = append(command, shellQuoteArgument(args[0]))
	}
	if reconcileAll {
		command = append(command, "--all")
	}
	command = append(command, "--include-files")
	if reconcileForce {
		command = append(command, "--force")
	}
	for _, file := range reconcileRefreshFiles {
		command = append(command, "--file", shellQuoteArgument(file))
	}
	if reconcileMigrateReferences {
		command = append(command, "--migrate-references")
	}
	if reconcileMigrateVerification {
		command = append(command, "--migrate-verification")
	}
	if profile := currentPromptProfile(); profile != promptProfileNone {
		command = append(command, "--profile="+shellQuoteArgument(string(profile)))
	}
	if singleAgent {
		command = append(command, "--single-agent")
	}
	command = append(command, "--output-only")
	return strings.Join(command, " ")
}

var inspectReconcileWorktree = func(projectRoot string) (worktreeprep.Location, error) {
	return worktreeprep.New().Inspect(context.Background(), projectRoot)
}

func resolveReconcileRefreshDecision(
	projectRoot string,
	requested bool,
	dryRun bool,
) (reconcileRefreshDecision, error) {
	decision := reconcileRefreshDecision{Apply: requested}
	if !requested || dryRun {
		return decision, nil
	}

	location, err := inspectReconcileWorktree(projectRoot)
	if err != nil {
		return reconcileRefreshDecision{}, fmt.Errorf("inspect reconcile worktree: %w", err)
	}
	if location.InsideGit && location.IsPrimary {
		decision.Apply = false
		decision.Deferred = true
	}
	return decision, nil
}

package cli

import (
	"context"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func planInitProject(root string, cfg *config.Config, release mintActionRelease) (*initRefreshPlan, error) {
	opts := initRefreshOptions{
		mintRelease:   &release,
		mint:          initMint,
		mintSelection: initMintSelection,
		issueTemplate: cfg.Issues.IssueTemplateEnabled(),
		seed:          cfg,
	}
	plan, err := buildInitRefreshPlan(context.Background(), root, opts)
	if err != nil {
		return nil, err
	}
	if initMint {
		return plan, nil
	}
	mintChanges, mintNotes, err := planInitMintIntegration(root, false, initMintSelection)
	if err != nil {
		return nil, err
	}
	mintChanges, updateNotes, err := planMintActionUpdates(root, mintChanges, release)
	if err != nil {
		return nil, err
	}
	plan.notes = append(plan.notes, mintNotes...)
	plan.notes = append(plan.notes, updateNotes...)
	seen := map[string]bool{}
	for _, change := range plan.changes {
		seen[change.relativePath] = true
	}
	for _, change := range mintChanges {
		if seen[change.relativePath] {
			continue
		}
		plan.changes = append(plan.changes, change)
		plan.stats.recordFileChange(change)
	}
	return plan, nil
}

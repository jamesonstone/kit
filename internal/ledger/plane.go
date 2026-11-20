package ledger

import (
	"fmt"

	"github.com/jamesonstone/kit/v3/internal/config"
)

type plane struct {
	cfg config.PlaneLedgerConfig
}

func (p plane) Provider() string { return "plane" }
func (p plane) Enabled() bool    { return true }

func (p plane) Description() string {
	states := p.cfg.States.Resolved()
	project := p.cfg.Project
	if p.cfg.Name != "" {
		project = fmt.Sprintf("%s (%s)", p.cfg.Name, p.cfg.Project)
	}
	return fmt.Sprintf("Plane work ledger in workspace %s, project %s. Kit stores no credentials and does not call Plane. The native agent uses its already-authorized connector. At start, find or create the matching item, link the issue, set status %s, and apply owner label agent:<x> (Grok Bots use grokbot:<bot_name>). When the pull request opens, link it and set status %s. On merge, set status %s.",
		p.cfg.Workspace, project, states.Active, states.InReview, states.Done)
}

func (p plane) Plan(phase Phase, item Item) ([]Action, error) {
	if err := validateOwnerLabel(item.OwnerLabel); err != nil {
		return nil, err
	}
	states := p.cfg.States.Resolved()
	base := Action{Provider: "plane", Workspace: p.cfg.Workspace, Project: p.cfg.Project}
	switch phase {
	case PhaseStart:
		return []Action{
			step(base, OpFindOrCreate, LinkIssue, item.IssueURL, "", "", findOrCreateSummary(item.Title)),
			step(base, OpLink, LinkIssue, item.IssueURL, "", "", "Link the issue."),
			step(base, OpSetState, "", "", states.Active, "", "Set status "+states.Active+"."),
			step(base, OpLabel, "", "", "", item.OwnerLabel, "Apply owner label agent:<x>. Grok Bots use grokbot:<bot_name>."),
		}, nil
	case PhaseReview:
		return []Action{
			step(base, OpFind, LinkIssue, item.IssueURL, "", "", "Find the work item that links this issue. Do not create another."),
			step(base, OpLink, LinkPullRequest, item.PullURL, "", "", "Link the pull request."),
			step(base, OpSetState, "", "", states.InReview, "", "Set status "+states.InReview+"."),
		}, nil
	case PhaseDone:
		return []Action{
			step(base, OpFind, LinkIssue, item.IssueURL, "", "", "Find the work item that links this issue. Do not create another."),
			step(base, OpSetState, "", "", states.Done, "", "Set status "+states.Done+"."),
		}, nil
	default:
		return nil, fmt.Errorf("unknown work ledger phase %q", phase)
	}
}

func findOrCreateSummary(title string) string {
	if title == "" {
		return "Find the work item that links this issue. If none, create one with the issue title. A failed search is not an empty project."
	}
	return fmt.Sprintf("Find the work item that links this issue. If none, create one titled %q. A failed search is not an empty project.", title)
}

func step(base Action, operation, linkKind, linkURL, state, label, summary string) Action {
	base.Operation = operation
	base.LinkKind = linkKind
	base.LinkURL = linkURL
	base.State = state
	base.Label = label
	base.Summary = summary
	return base
}

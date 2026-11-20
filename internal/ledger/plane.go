package ledger

import (
	"fmt"

	"github.com/jamesonstone/kit/v3/internal/config"
)

type plane struct {
	cfg        config.PlaneLedgerConfig
	ownerLabel string
}

func (p plane) Provider() string { return "plane" }
func (p plane) Enabled() bool    { return true }

func (p plane) Description() string {
	states := p.cfg.States.Resolved()
	project := p.cfg.Project
	if p.cfg.Name != "" {
		project = fmt.Sprintf("%s (%s)", p.cfg.Name, p.cfg.Project)
	}
	label := "the repository's work_ledger.owner_label when set"
	if p.ownerLabel != "" {
		label = p.ownerLabel
	}
	return fmt.Sprintf("Plane work ledger in workspace %s, project %s. Kit stores no credentials and does not call Plane. The native agent uses its already-authorized connector. The repository owner supplies the target. At start, find or create the matching item, link the issue, set status %s, and apply owner label %s. If a pull request is already open, find the item, link the pull request, and set status %s instead of %s. When the pull request opens, find the item, link it, and set status %s. Do not create a second item. On merge, find the item and set status %s.",
		p.cfg.Workspace, project, states.Active, label, states.InReview, states.Active, states.InReview, states.Done)
}

func (p plane) Plan(phase Phase, item Item) ([]Action, error) {
	label, err := p.resolveLabel(item.OwnerLabel)
	if err != nil {
		return nil, err
	}
	states := p.cfg.States.Resolved()
	base := Action{Provider: "plane", Workspace: p.cfg.Workspace, Project: p.cfg.Project}
	switch phase {
	case PhaseStart:
		actions := []Action{
			step(base, OpFindOrCreate, LinkIssue, item.IssueURL, "", "", findOrCreateSummary(item.Title)),
			step(base, OpLink, LinkIssue, item.IssueURL, "", "", "Link the issue on the item returned by the find or create."),
		}
		if item.PullURL != "" {
			actions = append(actions,
				step(base, OpLink, LinkPullRequest, item.PullURL, "", "", "Link the pull request on that same item."),
				step(base, OpSetState, "", "", states.InReview, "", "Set status "+states.InReview+"."),
			)
		} else {
			actions = append(actions, step(base, OpSetState, "", "", states.Active, "", "Set status "+states.Active+"."))
		}
		actions = append(actions, step(base, OpLabel, "", "", "", label, ownerLabelSummary(label)))
		return actions, nil
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

func (p plane) resolveLabel(supplied string) (string, error) {
	if supplied != "" && p.ownerLabel != "" && supplied != p.ownerLabel {
		return "", fmt.Errorf("owner label must match work_ledger.owner_label")
	}
	label := supplied
	if label == "" {
		label = p.ownerLabel
	}
	if err := validateOwnerLabel(label); err != nil {
		return "", err
	}
	return label, nil
}

func ownerLabelSummary(label string) string {
	if label == "" {
		return "Apply work_ledger.owner_label when the repository sets one."
	}
	return "Apply owner label " + label + "."
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

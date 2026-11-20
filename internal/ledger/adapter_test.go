package ledger

import (
	"strings"
	"testing"

	"github.com/jamesonstone/kit/v3/internal/config"
)

func exampleConfig() config.WorkLedgerConfig {
	return config.WorkLedgerConfig{
		Provider:   "plane",
		OwnerLabel: "agent:example",
		Plane: config.PlaneLedgerConfig{
			Workspace: "example-workspace",
			Project:   "PROJ",
			Name:      "Ledger",
		},
	}
}

func TestNoopAdapter(t *testing.T) {
	for _, cfg := range []config.WorkLedgerConfig{{}, {Provider: "none"}} {
		adapter, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if adapter.Enabled() || adapter.Provider() != "" || !strings.Contains(adapter.Description(), "off") {
			t.Fatalf("noop = %q %v", adapter.Provider(), adapter.Enabled())
		}
		for _, phase := range []Phase{PhaseStart, PhaseReview, PhaseDone} {
			actions, err := adapter.Plan(phase, Item{IssueURL: "https://github.com/acme/app/issues/1"})
			if err != nil || len(actions) != 1 || actions[0].Operation != OpNoop || actions[0].Workspace != "" {
				t.Fatalf("%s %#v %v", phase, actions, err)
			}
		}
		if _, err := adapter.Plan("abandoned", Item{}); err == nil {
			t.Fatal("unknown phase was accepted")
		}
	}
}

func TestPlaneAdapterLifecycle(t *testing.T) {
	adapter, err := New(exampleConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !adapter.Enabled() || adapter.Provider() != "plane" {
		t.Fatalf("provider = %q enabled %v", adapter.Provider(), adapter.Enabled())
	}
	description := adapter.Description()
	for _, want := range []string{"example-workspace", "Ledger (PROJ)", "does not call Plane", "agent:example", "repository owner", "Active", "In review", "Done", "find the item", "already open"} {
		if !strings.Contains(description, want) {
			t.Errorf("description missing %q", want)
		}
	}
	item := Item{IssueURL: "https://github.com/acme/app/issues/300", Title: "ledger hook", OwnerLabel: "agent:example"}
	start, err := adapter.Plan(PhaseStart, item)
	if err != nil {
		t.Fatal(err)
	}
	assertOps(t, start, []string{OpFindOrCreate, OpLink, OpSetState, OpLabel})
	if start[0].LinkURL != item.IssueURL || start[1].LinkKind != LinkIssue || start[2].State != "Active" || start[3].Label != "agent:example" {
		t.Fatalf("start = %#v", start)
	}
	if start[0].Workspace != "example-workspace" || start[0].Project != "PROJ" {
		t.Fatalf("target = %#v", start[0])
	}
	open := item
	open.PullURL = "https://github.com/acme/app/pull/301"
	resumed, err := adapter.Plan(PhaseStart, open)
	if err != nil {
		t.Fatal(err)
	}
	assertOps(t, resumed, []string{OpFindOrCreate, OpLink, OpLink, OpSetState, OpLabel})
	if resumed[2].LinkKind != LinkPullRequest || resumed[2].LinkURL != open.PullURL || resumed[3].State != "In review" {
		t.Fatalf("open pull request start = %#v", resumed)
	}
	review, err := adapter.Plan(PhaseReview, open)
	if err != nil {
		t.Fatal(err)
	}
	assertOps(t, review, []string{OpFind, OpLink, OpSetState})
	if review[0].LinkURL != open.IssueURL || review[1].LinkKind != LinkPullRequest || review[1].LinkURL != open.PullURL || review[2].State != "In review" {
		t.Fatalf("review = %#v", review)
	}
	done, err := adapter.Plan(PhaseDone, item)
	if err != nil {
		t.Fatal(err)
	}
	assertOps(t, done, []string{OpFind, OpSetState})
	if done[1].State != "Done" {
		t.Fatalf("done = %#v", done)
	}
	if !strings.Contains(start[0].Summary, `"ledger hook"`) {
		t.Fatalf("create summary = %q", start[0].Summary)
	}
	for _, actions := range [][]Action{start, review, done} {
		for _, action := range actions {
			blob := action.Summary + action.Label + action.LinkURL + action.Workspace
			for _, secret := range []string{"token", "secret", "api_key", "password"} {
				if strings.Contains(strings.ToLower(blob), secret) {
					t.Fatalf("action carries %q: %#v", secret, action)
				}
			}
		}
	}
}

func TestPlaneAdapterRejectsBadLabelAndUnknownPhase(t *testing.T) {
	adapter, err := New(exampleConfig())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Plan(PhaseStart, Item{OwnerLabel: "not a label"}); err == nil {
		t.Fatal("accepted an owner label with spaces")
	}
	if _, err := adapter.Plan(PhaseStart, Item{OwnerLabel: "other:agent"}); err == nil {
		t.Fatal("accepted an owner label other than work_ledger.owner_label")
	}
	if _, err := adapter.Plan("later", Item{}); err == nil {
		t.Fatal("unknown phase was accepted")
	}
	custom := exampleConfig()
	custom.Plane.States = config.PlaneLedgerStates{Active: "Doing", InReview: "Review", Done: "Finished"}
	adapter, err = New(custom)
	if err != nil {
		t.Fatal(err)
	}
	start, err := adapter.Plan(PhaseStart, Item{})
	if err != nil || start[2].State != "Doing" {
		t.Fatalf("%v %#v", err, start)
	}
}

func TestNewTrimsPlaneProvider(t *testing.T) {
	cfg := exampleConfig()
	cfg.Provider = " plane "
	adapter, err := New(cfg)
	if err != nil || adapter.Provider() != "plane" || !adapter.Enabled() {
		t.Fatalf("%v %#v", err, adapter)
	}
}

func TestNewRejectsUnavailableProviders(t *testing.T) {
	for _, provider := range []string{"linear", "jira", "notion", "trello", "plane"} {
		cfg := config.WorkLedgerConfig{Provider: provider}
		if provider == "plane" {
			cfg.Plane.Project = "WL"
		}
		if _, err := New(cfg); err == nil {
			t.Fatalf("provider %q was accepted", provider)
		}
	}
}

func assertOps(t *testing.T, actions []Action, want []string) {
	t.Helper()
	if len(actions) != len(want) {
		t.Fatalf("ops = %#v, want %v", actions, want)
	}
	for i, op := range want {
		if actions[i].Operation != op {
			t.Fatalf("op %d = %q, want %q", i, actions[i].Operation, op)
		}
	}
}

package templates

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

type capabilityRuntimeFixture struct {
	Scenario                 string   `yaml:"scenario"`
	EvidenceBasis            []string `yaml:"evidence_basis"`
	SeparateExecution        string   `yaml:"separate_execution"`
	ParallelExecution        string   `yaml:"parallel_execution"`
	StableAgentReferences    string   `yaml:"stable_agent_references"`
	SameAgentFollowUp        string   `yaml:"same_agent_follow_up"`
	ModelSelection           string   `yaml:"model_selection"`
	EffortSelection          string   `yaml:"effort_selection"`
	FreshVerification        string   `yaml:"fresh_verification"`
	WaitStatusControls       string   `yaml:"wait_status_controls"`
	EffectiveCapacity        string   `yaml:"effective_capacity"`
	SelectedTopology         string   `yaml:"selected_topology"`
	DelegationDepth          string   `yaml:"delegation_depth"`
	ExecutionKind            string   `yaml:"execution_kind"`
	RoutingAction            string   `yaml:"routing_action"`
	TaskOutcome              string   `yaml:"task_outcome"`
	OrchestrationConformance string   `yaml:"orchestration_conformance"`
	VerificationIndependent  string   `yaml:"verification_independent"`
	Degradations             []string `yaml:"degradations"`
	ContinuityLoss           bool     `yaml:"continuity_loss"`
	ReplacementRebrief       string   `yaml:"replacement_rebrief"`
}

func TestCapabilityAwareRuntimeFixturesCoverDegradationMatrix(t *testing.T) {
	content, err := os.ReadFile("testdata/capability-aware-runtime-fixtures.yaml")
	if err != nil {
		t.Fatalf("read capability-aware runtime fixtures: %v", err)
	}
	var document struct {
		Fixtures []capabilityRuntimeFixture `yaml:"fixtures"`
	}
	if err := yaml.Unmarshal(content, &document); err != nil {
		t.Fatalf("parse capability-aware runtime fixtures: %v", err)
	}

	fixtures := make(map[string]capabilityRuntimeFixture, len(document.Fixtures))
	for _, fixture := range document.Fixtures {
		if _, duplicate := fixtures[fixture.Scenario]; duplicate {
			t.Fatalf("duplicate runtime fixture scenario %q", fixture.Scenario)
		}
		assertCompleteCapabilityFixture(t, fixture)
		fixtures[fixture.Scenario] = fixture
	}

	requireFixture(t, fixtures, "full_codex_controls", func(f capabilityRuntimeFixture) bool {
		return f.SeparateExecution == "confirmed" && f.ParallelExecution == "confirmed" &&
			f.StableAgentReferences == "confirmed" && f.SameAgentFollowUp == "confirmed" &&
			f.ModelSelection == "confirmed" && f.EffortSelection == "confirmed" &&
			f.FreshVerification == "confirmed" && f.WaitStatusControls == "confirmed" &&
			f.EffectiveCapacity == "4" &&
			f.OrchestrationConformance == "full" && f.VerificationIndependent == "true"
	})
	requireFixture(t, fixtures, "unknown_capacity", func(f capabilityRuntimeFixture) bool {
		return f.EffectiveCapacity == "unknown" && f.RoutingAction == "serialize"
	})
	requireFixture(t, fixtures, "no_child_primitive", func(f capabilityRuntimeFixture) bool {
		return f.SeparateExecution == "unavailable" && f.ExecutionKind == "logical_lane" &&
			f.SelectedTopology == "single-supervisor" && f.VerificationIndependent == "false"
	})
	requireFixture(t, fixtures, "unavailable_exact_model_pin", func(f capabilityRuntimeFixture) bool {
		return f.ModelSelection == "unavailable" && f.TaskOutcome == "BLOCKED" &&
			f.OrchestrationConformance == "unsatisfied"
	})
	requireFixture(t, fixtures, "replacement_after_continuity_loss", func(f capabilityRuntimeFixture) bool {
		return f.ContinuityLoss && f.ReplacementRebrief == "required"
	})
	for _, scenario := range []string{
		"limited_routing_or_continuation",
		"host_managed_copilot_or_warp",
	} {
		requireFixture(t, fixtures, scenario, func(capabilityRuntimeFixture) bool { return true })
	}
}

func assertCompleteCapabilityFixture(t *testing.T, fixture capabilityRuntimeFixture) {
	t.Helper()
	if fixture.Scenario == "" || len(fixture.EvidenceBasis) == 0 || fixture.Degradations == nil {
		t.Errorf("runtime fixture %+v lacks scenario, evidence, or degradation reporting", fixture)
	}
	for name, value := range map[string]string{
		"separate_execution":        fixture.SeparateExecution,
		"parallel_execution":        fixture.ParallelExecution,
		"stable_agent_references":   fixture.StableAgentReferences,
		"same_agent_follow_up":      fixture.SameAgentFollowUp,
		"model_selection":           fixture.ModelSelection,
		"effort_selection":          fixture.EffortSelection,
		"fresh_verification":        fixture.FreshVerification,
		"wait_status_controls":      fixture.WaitStatusControls,
		"effective_capacity":        fixture.EffectiveCapacity,
		"selected_topology":         fixture.SelectedTopology,
		"delegation_depth":          fixture.DelegationDepth,
		"execution_kind":            fixture.ExecutionKind,
		"routing_action":            fixture.RoutingAction,
		"task_outcome":              fixture.TaskOutcome,
		"orchestration_conformance": fixture.OrchestrationConformance,
		"verification_independent":  fixture.VerificationIndependent,
	} {
		if value == "" {
			t.Errorf("runtime fixture %q missing %s", fixture.Scenario, name)
		}
	}
}

func requireFixture(
	t *testing.T,
	fixtures map[string]capabilityRuntimeFixture,
	scenario string,
	valid func(capabilityRuntimeFixture) bool,
) {
	t.Helper()
	fixture, ok := fixtures[scenario]
	if !ok {
		t.Fatalf("runtime fixture scenario %q is missing", scenario)
	}
	if !valid(fixture) {
		t.Errorf("runtime fixture scenario %q has invalid semantics: %+v", scenario, fixture)
	}
}

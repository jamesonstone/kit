package cli

import (
	"encoding/json"
	"testing"
)

func TestCapabilitiesTargetedJSON(t *testing.T) {
	assertCapabilitiesInitTarget(t)
	assertCapabilitiesSpecTarget(t)
	assertCapabilitiesLegacyVerifyTarget(t)
	assertCapabilitiesCITarget(t)
	assertCapabilitiesDispatchTarget(t)
	assertCapabilitiesImproveTarget(t)
	assertCapabilitiesLoopPromptTarget(t)
	assertCapabilitiesLoopWorkflowTarget(t)
	assertCapabilitiesLoopReviewTarget(t)
	assertCapabilitiesPRFixTarget(t)
	assertCapabilitiesProjectRefreshTarget(t)
	assertCapabilitiesRemovedReviewLoopTarget(t)
}

func mustCapabilityDetail(t *testing.T, label string, args ...string) capabilityDetailPayload {
	t.Helper()

	output, err := executeCapabilitiesCommand(append([]string{"--json"}, args...)...)
	if err != nil {
		t.Fatalf("kit capabilities %s --json error = %v", label, err)
	}

	var payload capabilityDetailPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v\noutput: %s", label, err, output)
	}
	return payload
}

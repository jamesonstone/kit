package cli

import (
	"testing"
)

// The real mixed-history contract test verifies working input bindings, artifact
// retention and tagged retries. Keep the unit boundary focused on authority/pins.
func TestMintPreparationStepsContracts(t *testing.T) {
	steps := mintReleasePreparationSteps()
	if steps[0].(map[string]any)["id"] != "kit_mint_release" {
		t.Fatal("resolve output references require the kit_mint_release step ID")
	}
	for i, raw := range steps {
		step := raw.(map[string]any)
		wantAction := "jamesonstone/mint@" + mintPipelineActionRef
		if i == 2 {
			wantAction = "actions/upload-artifact@" + mintPipelineArtifactRef
		}
		if step["uses"] != wantAction || step["run"] != nil {
			t.Fatalf("unexpected preparation action: %+v", step)
		}
		with := step["with"].(map[string]any)
		if with["github-token"] != nil || with["version-file"] != nil || with["command"] == "release-publish" || with["command"] == "version" {
			t.Fatalf("preparation gained publication authority: %+v", step)
		}
	}
}

func TestMintPublisherKeepsCanonicalNotes(t *testing.T) {
	p := mintReleasePublisherStep()
	with := p["with"].(map[string]any)
	if with["command"] != "release-publish" || with["commitish"] != "${{ github.sha }}" || with["github-token"] != "${{ github.token }}" {
		t.Fatalf("publisher: %+v", p)
	}
	if with["release-notes-file"] != nil || with["output"] != nil || p["run"] != nil {
		t.Fatal("enriched artifact must not replace canonical publisher notes")
	}
}

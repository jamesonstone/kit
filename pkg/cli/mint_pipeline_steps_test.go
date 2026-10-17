package cli

import (
	stdreflect "reflect"
	"testing"
)

func TestMintPreparationStepsContracts(t *testing.T) {
	steps := mintReleasePreparationSteps()
	if len(steps) != 3 {
		t.Fatal("expected resolve, changelog, artifact retention")
	}
	resolve := steps[0].(map[string]any)
	if resolve["id"] != "kit_mint_release" || resolve["uses"] != "jamesonstone/mint@"+mintPipelineActionRef {
		t.Fatalf("resolve: %+v", resolve)
	}
	if !stdreflect.DeepEqual(resolve["with"], map[string]any{"command": "release-resolve", "commitish": "${{ github.sha }}"}) {
		t.Fatalf("resolve inputs: %+v", resolve)
	}
	changelog := steps[1].(map[string]any)
	if changelog["id"] != "kit_mint_changelog" || changelog["uses"] != "jamesonstone/mint@"+mintPipelineActionRef || changelog["if"] != mintPipelineNewRelease {
		t.Fatalf("changelog: %+v", changelog)
	}
	wantInputs := map[string]any{
		"command": "changelog", "prev-tag": "${{ steps.kit_mint_release.outputs.base_tag }}",
		"current-tag": "${{ steps.kit_mint_release.outputs.version_tag }}",
		"current-ref": "${{ steps.kit_mint_release.outputs.target_sha }}",
		"owner":       "${{ github.repository_owner }}", "repo": "${{ github.event.repository.name }}",
		"output": "${{ runner.temp }}/kit-mint-CHANGELOG.md",
	}
	if !stdreflect.DeepEqual(changelog["with"], wantInputs) {
		t.Fatalf("changelog bindings: %+v", changelog["with"])
	}
	artifact := steps[2].(map[string]any)
	if artifact["uses"] != "actions/upload-artifact@"+mintPipelineArtifactRef || artifact["if"] != mintPipelineNewRelease {
		t.Fatalf("artifact: %+v", artifact)
	}
	if !stdreflect.DeepEqual(artifact["with"], map[string]any{
		"name": "mint-changelog-${{ steps.kit_mint_release.outputs.version_tag }}",
		"path": "${{ runner.temp }}/kit-mint-CHANGELOG.md", "if-no-files-found": "error",
	}) {
		t.Fatalf("artifact inputs: %+v", artifact["with"])
	}
	for _, raw := range steps {
		step := raw.(map[string]any)
		if step["run"] != nil {
			t.Fatal("preparation must not introduce shell mutations")
		}
		with := step["with"].(map[string]any)
		if with["github-token"] != nil || with["version-file"] != nil || with["command"] == "release-publish" || with["command"] == "version" {
			t.Fatalf("preparation publishes or only checks version: %+v", step)
		}
	}
}

func TestMintPreparationGuardsAndFreshValues(t *testing.T) {
	want := "${{ steps.kit_mint_release.outputs.needs_git_tag == 'true' && steps.kit_mint_release.outputs.commit_count != '0' }}"
	if mintPipelineNewRelease != want {
		t.Fatal("tagged/no-release reruns must skip changelog and artifact")
	}
	a, b := mintReleasePreparationSteps(), mintReleasePreparationSteps()
	if !stdreflect.DeepEqual(a, b) {
		t.Fatal("preparation is nondeterministic")
	}
	a[0].(map[string]any)["id"] = "changed"
	if b[0].(map[string]any)["id"] != "kit_mint_release" {
		t.Fatal("calls share mutable step state")
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

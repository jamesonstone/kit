package cli

import "github.com/jamesonstone/kit/v3/internal/mintassets"

const mintPipelineActionPath = ".git/kit-mint-action"
const mintPipelineAction = "./" + mintPipelineActionPath

const mintPipelineArtifactRef = "ea165f8d65b6e75b540449e92b4886f43607fa02" // upload-artifact v4.6.2
const mintPipelineChangelogPath = "${{ runner.temp }}/kit-mint-CHANGELOG.md"
const mintPipelineNewRelease = "${{ steps.kit_mint_release.outputs.needs_git_tag == 'true' && steps.kit_mint_release.outputs.commit_count != '0' }}"

// mintLatestSteps loads the published action outside the project's working tree.
func mintLatestSteps() []any {
	return []any{
		map[string]any{
			"name": "Resolve latest stable Mint release", "id": "kit_mint_latest",
			"shell": "bash", "run": mintassets.ResolveRelease,
			"env": map[string]any{"GH_TOKEN": "${{ github.token }}"},
		},
		map[string]any{
			"name": "Check out published Mint Action",
			"uses": "actions/checkout@" + mintPipelineCheckoutRef,
			"with": map[string]any{
				"repository":          "jamesonstone/mint",
				"ref":                 "${{ steps.kit_mint_latest.outputs.ref }}",
				"path":                mintPipelineActionPath,
				"persist-credentials": false,
			},
		},
	}
}

// mintReleasePreparationSteps resolves source metadata and retains an enriched
// changelog without publishing tags, changing runtime version files or deploying.
func mintReleasePreparationSteps() []any {
	return []any{
		map[string]any{
			"name": "Resolve Mint release metadata", "id": "kit_mint_release",
			"uses": mintPipelineAction,
			"with": map[string]any{"command": "release-resolve", "commitish": "${{ github.sha }}"},
		},
		map[string]any{
			"name": "Generate enriched Mint changelog", "id": "kit_mint_changelog",
			"if":   mintPipelineNewRelease,
			"uses": mintPipelineAction,
			"with": map[string]any{
				"command":     "changelog",
				"prev-tag":    "${{ steps.kit_mint_release.outputs.base_tag }}",
				"current-tag": "${{ steps.kit_mint_release.outputs.version_tag }}",
				"current-ref": "${{ steps.kit_mint_release.outputs.target_sha }}",
				"owner":       "${{ github.repository_owner }}",
				"repo":        "${{ github.event.repository.name }}",
				"output":      mintPipelineChangelogPath,
			},
		},
		map[string]any{
			"name": "Retain enriched Mint changelog", "if": mintPipelineNewRelease,
			"uses": "actions/upload-artifact@" + mintPipelineArtifactRef,
			"with": map[string]any{
				"name":              "mint-changelog-${{ steps.kit_mint_release.outputs.version_tag }}",
				"path":              mintPipelineChangelogPath,
				"if-no-files-found": "error",
			},
		},
	}
}

// mintReleasePublisherStep retains Mint's canonical subject-based release notes;
// the enriched Linear/GitHub changelog is a separate retained workflow artifact.
func mintReleasePublisherStep() map[string]any {
	return map[string]any{
		"name": "Publish immutable release state", "uses": mintPipelineAction,
		"with": map[string]any{
			"command": "release-publish", "commitish": "${{ github.sha }}",
			"owner": "${{ github.repository_owner }}", "repo": "${{ github.event.repository.name }}",
			"github-token": "${{ github.token }}",
		},
	}
}

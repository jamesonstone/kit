package cli

import (
	"gopkg.in/yaml.v3"
)

const mintPipelineActionRef = "eac91490f772adbd2ab030cb6eba3910c1bca93c" // Mint v0.9.0
const mintPipelineCheckoutRef = "11d5960a326750d5838078e36cf38b85af677262"

// The starter is project owned after generation. Existing workflows keep their settings.
func renderMintReleasePipeline(defaultBranch string) []byte {
	steps := []any{map[string]any{"name": "Check out the triggering commit", "uses": "actions/checkout@" + mintPipelineCheckoutRef, "with": map[string]any{"fetch-depth": 0, "ref": "${{ github.sha }}"}}}
	steps = append(steps, mintReleasePreparationSteps()...)
	steps = append(steps, mintReleasePublisherStep())
	workflow := map[string]any{
		"name":        "Mint release",
		"on":          map[string]any{"push": map[string]any{"branches": []string{defaultBranch}}},
		"permissions": map[string]string{"contents": "read"},
		"concurrency": map[string]any{"group": "mint-release-${{ github.repository }}", "cancel-in-progress": false},
		"jobs": map[string]any{
			"publish": map[string]any{
				"if":          "${{ github.event_name == 'push' && github.ref == format('refs/heads/{0}', github.event.repository.default_branch) && github.event.deleted == false }}",
				"runs-on":     "ubuntu-latest",
				"permissions": map[string]string{"contents": "write"},
				"steps":       steps,
			},
		},
	}
	data, err := yaml.Marshal(workflow)
	if err != nil {
		return nil
	} // The fixed scalar/map schema cannot fail encoding.
	header := "# Scaffolded by Kit with Mint v0.9.0. This workflow is project owned.\n" +
		"# Publishes tags and GitHub Releases only; it does not build or deploy the application.\n" +
		"# Enriched Linear/GitHub changelog is retained as an artifact; published release notes stay subject based.\n"
	return append([]byte(header), data...)
}

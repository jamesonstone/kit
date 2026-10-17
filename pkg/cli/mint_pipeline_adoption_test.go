package cli

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMintPublicationExplicitOwnershipAndOrder(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/deploy.yaml"
	source := strings.Replace(pipelineDeploy, "contents: read", "contents: write", 1)
	source = strings.Replace(source, "push:\n    branches: [trunk]", "push:", 1)
	source = strings.Replace(source, "jobs:", "concurrency:\n  group: production-release\n  cancel-in-progress: false\njobs:", 1)
	writePipelineFixture(t, root, path, source)
	changes, notes, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy", releaseOwner: "mint"})
	if err != nil || len(changes) != 1 || strings.Contains(strings.Join(notes, " "), "INCOMPLETE") {
		t.Fatalf("%+v %v %v", changes, notes, err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(changes[0].after), &doc); err != nil {
		t.Fatal(err)
	}
	steps := mintField(mintField(mintField(&doc, "jobs"), "deploy"), "steps").Content
	if len(steps) != 6 {
		t.Fatal(changes[0].after)
	}
	for i, command := range map[int]string{1: "release-resolve", 2: "changelog", 5: "release-publish"} {
		if mintField(mintField(steps[i], "with"), "command").Value != command {
			t.Fatal(changes[0].after)
		}
	}
	if mintField(steps[4], "id").Value != "deploy" || mintField(steps[5], "if").Value != mintPublishGuard {
		t.Fatal(changes[0].after)
	}
	if !strings.Contains(changes[0].after, "group: production-release\n  cancel-in-progress: false") {
		t.Fatal("changed serialization")
	}
}
func TestMintPublicationRequiresExistingAuthority(t *testing.T) {
	base := strings.Replace(pipelineDeploy, "contents: read", "contents: write", 1)
	base = strings.Replace(base, "push:\n    branches: [trunk]", "push:", 1)
	base = strings.Replace(base, "jobs:", "concurrency:\n  group: release\n  cancel-in-progress: false\njobs:", 1)
	cases := map[string]string{
		"unknown token":                strings.Replace(base, "fetch-depth: 0", "fetch-depth: 0\n          token: ${{ secrets.CUSTOM_TOKEN }}", 1),
		"unknown SSH key":              strings.Replace(base, "fetch-depth: 0", "fetch-depth: 0\n          ssh-key: ${{ secrets.CUSTOM_SSH }}", 1),
		"dynamic credential retention": strings.Replace(base, "fetch-depth: 0", "fetch-depth: 0\n          persist-credentials: ${{ inputs.persist }}", 1),
		"read permissions":             strings.Replace(base, "contents: write", "contents: read", 1),
		"cancellation":                 strings.Replace(base, "cancel-in-progress: false", "cancel-in-progress: true", 1),
		"dynamic serialization":        strings.Replace(base, "group: release", "group: release-${{ github.ref }}", 1),
		"tag trigger":                  strings.Replace(base, "push:", "push:\n    tags: ['v*']", 1),
		"no credentials":               strings.Replace(base, "fetch-depth: 0", "fetch-depth: 0\n          persist-credentials: false", 1),
		"matrix":                       strings.Replace(base, "    needs: []", "    strategy:\n      matrix:\n        os: [ubuntu-latest]\n    needs: []", 1),
		"shallow":                      strings.Replace(base, "fetch-depth: 0", "fetch-depth: 1", 1),
		"custom checkout":              strings.Replace(base, "fetch-depth: 0", "fetch-depth: 0\n          ref: release", 1),
		"ncipollo publisher":           strings.Replace(base, "run: echo original", "uses: ncipollo/release-action@v1", 1),
		"existing publisher":           strings.Replace(base, "echo original", "gh release create v1.0.0", 1),
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			path := ".github/workflows/deploy.yaml"
			writePipelineFixture(t, root, path, source)
			if changes, _, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy", releaseOwner: "mint"}); err == nil || len(changes) != 0 {
				t.Fatalf("accepted: %+v %v", changes, err)
			}
		})
	}
	t.Run("publisher in another workflow", func(t *testing.T) {
		root := t.TempDir()
		path := ".github/workflows/deploy.yaml"
		writePipelineFixture(t, root, path, base)
		writePipelineFixture(t, root, ".github/workflows/publish.yaml", "on: push\njobs:\n  publish:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: ncipollo/release-action@v1\n")
		if _, _, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy", releaseOwner: "mint"}); err == nil {
			t.Fatal("duplicate publisher accepted")
		}
	})
}
func TestMintExistingConsumerHandoff(t *testing.T) {
	root := t.TempDir()
	path := ".github/workflows/deploy.yaml"
	source := strings.Replace(pipelineDeploy, "echo original", "echo '${{ steps.kit_mint_release.outputs.version_tag }}' '${{ runner.temp }}/kit-mint-CHANGELOG.md'", 1)
	writePipelineFixture(t, root, path, source)
	changes, notes, err := planMintPipeline(root, mintPipelineSelection{workflow: path, job: "deploy", beforeStep: "deploy", releaseOwner: "existing"})
	if err != nil || len(changes) != 1 || strings.Contains(strings.Join(notes, " "), "INCOMPLETE") || !strings.Contains(strings.Join(notes, " "), "references the resolved version") {
		t.Fatalf("%+v %v %v", changes, notes, err)
	}
}
func TestMintVersionDiagnosticNeverClaimsIntegration(t *testing.T) {
	root := t.TempDir()
	writePipelineFixture(t, root, ".github/workflows/deploy.yaml", strings.Replace(pipelineDeploy, "run: echo original", "uses: jamesonstone/mint@old\n        with:\n          command: version", 1))
	changes, notes, err := planMintPipeline(root, mintPipelineSelection{})
	if err != nil || len(changes) != 1 || !strings.Contains(strings.Join(notes, " "), "INCOMPLETE") {
		t.Fatalf("%+v %v %v", changes, notes, err)
	}
}

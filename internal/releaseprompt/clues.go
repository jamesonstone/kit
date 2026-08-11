package releaseprompt

import (
	"os"
	"path/filepath"
)

var verificationCandidates = []string{
	"scripts/verify-production.sh",
	"scripts/verify-prod.sh",
	"scripts/production-smoke.sh",
	"tests/end-to-end/production/run.sh",
}

var integrationCandidates = []string{
	"tests/end-to-end/production/run.sh",
	"scripts/integration-test.sh",
	"scripts/test-integration.sh",
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func kitManaged(root string) bool {
	return fileExists(filepath.Join(root, ".kit.yaml")) ||
		fileExists(filepath.Join(root, "docs", "CONSTITUTION.md")) ||
		fileExists(filepath.Join(root, "docs", "agents", "README.md"))
}

func firstScriptHint(root string, candidates []string) string {
	for _, candidate := range candidates {
		if fileExists(filepath.Join(root, filepath.FromSlash(candidate))) {
			return "script:" + candidate
		}
	}
	return ""
}

type infrastructureHint struct {
	Mode, Provider, CLI string
}

func discoverInfrastructure(repositories []Repository) infrastructureHint {
	for _, repository := range repositories {
		root := repository.Path
		switch {
		case fileExists(filepath.Join(root, "cdk.json")):
			return infrastructureHint{Mode: "iac", Provider: "aws", CLI: "cdk"}
		case fileExists(filepath.Join(root, "azure.yaml")):
			return infrastructureHint{Mode: "iac", Provider: "azure", CLI: "azd"}
		case fileExists(filepath.Join(root, "Pulumi.yaml")):
			return infrastructureHint{Mode: "iac", Provider: "infer from Pulumi configuration", CLI: "pulumi"}
		case len(mustGlob(filepath.Join(root, "*.tf"))) > 0 || directoryExists(filepath.Join(root, "terraform")):
			return infrastructureHint{Mode: "iac", Provider: "infer from Terraform configuration", CLI: "terraform"}
		case directoryExists(filepath.Join(root, "infrastructure")) || directoryExists(filepath.Join(root, "infra")):
			return infrastructureHint{Mode: "auto", Provider: "infer from repository infrastructure", CLI: "infer from repository tooling"}
		}
	}
	return infrastructureHint{}
}

func mustGlob(pattern string) []string {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}
	return matches
}

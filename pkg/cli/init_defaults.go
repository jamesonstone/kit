package cli

import "github.com/jamesonstone/kit/v3/internal/config"

func defaultInitConfig() *config.Config {
	cfg := config.Default()
	cfg.InstructionScaffoldVersion = config.CurrentInstructionScaffoldVersion
	return cfg
}

func projectInitDeliveryPaths(cfg *config.Config) []string {
	paths := []string{
		config.ConfigFileName,
		gitignorePath,
		makefilePath,
		codeRabbitConfigPath,
		pullRequestTemplatePath,
		autoAssignWorkflowPath,
		readmePath,
		cfg.ConstitutionPath,
	}
	paths = append(paths, instructionArtifactPaths(cfg)...)
	return paths
}
